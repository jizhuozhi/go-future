package dagfunc

import (
	"fmt"

	"github.com/jizhuozhi/go-future/dagcore"
)

type subgraphDef struct {
	id   dagcore.NodeID
	sub  *Builder
	opts *nodeOptions
}

type subgraphInput struct {
	subID dagcore.NodeID
	dep   *outputRef // nil when the value is a constant
	value any
}

// Subgraph embeds another Builder as a single node of this graph.
//
// The subgraph is a real dagcore subgraph: it keeps its own scheduler, so its
// nodes still run in parallel, but the parent only sees one node. Wiring is done
// by type:
//
//   - every input declared with Provide on sub is fed from the node of the
//     parent producing that type, or from the input's Default when it has one;
//   - Outputs declares which results of the subgraph the parent can read.
//
// A name is mandatory because a subgraph has no return type to derive an id
// from.
//
//	qa := dagfunc.New()
//	_ = qa.Provide(Question{})
//	_ = qa.Use(retrieve)
//	_ = qa.Use(rerank)
//
//	root := dagfunc.New()
//	_ = root.Provide(Question{})
//	_ = root.Subgraph(qa, dagfunc.Name("qa"), dagfunc.Outputs(Answer{}))
//	_ = root.Use(summarize) // func(ctx, Answer) (Summary, error)
func (b *Builder) Subgraph(sub *Builder, opts ...Option) error {
	if b.dag.Frozen() {
		return ErrFrozen
	}
	if sub == nil {
		return fmt.Errorf("%w: nil subgraph", ErrNodeNotFound)
	}
	o := newOptions(opts)
	id := dagcore.NodeID(b.join(o.name))
	if id == "" {
		return ErrSubgraphName
	}
	if _, ok := b.nodes[id]; ok {
		return fmt.Errorf("%w: %s", ErrNodeExisted, id)
	}
	// The declared outputs are published right away so that nodes registered
	// afterwards can depend on them, exactly like the results of Use. Freeze
	// resolves them inside the subgraph and creates the node itself.
	b.register(id, nil, o.outputs, false)
	b.subgraphs = append(b.subgraphs, &subgraphDef{id: id, sub: sub, opts: o})
	return nil
}

func (b *Builder) materializeSubgraphs() error {
	// Freeze the subgraphs first, they may embed subgraphs themselves.
	for _, sg := range b.subgraphs {
		if !sg.sub.dag.Frozen() {
			if err := sg.sub.Freeze(); err != nil {
				return err
			}
		}
	}
	for _, sg := range b.subgraphs {
		if err := b.materializeSubgraph(sg); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) materializeSubgraph(sg *subgraphDef) error {
	if len(sg.opts.outputs) == 0 {
		return fmt.Errorf("%w: %s", ErrSubgraphOutputs, sg.id)
	}

	outRefs := make([]outputRef, 0, len(sg.opts.outputs))
	for _, t := range sg.opts.outputs {
		ref, err := sg.sub.outputRef(t)
		if err != nil {
			return fmt.Errorf("subgraph %s: %w", sg.id, err)
		}
		outRefs = append(outRefs, ref)
	}

	pairs, deps, err := b.subgraphInputs(sg)
	if err != nil {
		return err
	}

	inputMapping := func(deps map[dagcore.NodeID]any) map[dagcore.NodeID]any {
		inputs := make(map[dagcore.NodeID]any, len(pairs))
		for _, p := range pairs {
			if p.dep == nil {
				inputs[p.subID] = p.value
				continue
			}
			inputs[p.subID] = unwrap(deps[p.dep.node], p.dep.index)
		}
		return inputs
	}
	outputMapping := func(results map[dagcore.NodeID]any) any {
		vals := make([]any, 0, len(outRefs))
		for _, r := range outRefs {
			vals = append(vals, unwrap(results[r.node], r.index))
		}
		if len(vals) == 1 {
			return vals[0]
		}
		return vals
	}

	if err := b.dag.AddSubgraph(sg.id, deps, sg.sub.dag, inputMapping, outputMapping); err != nil {
		return err
	}
	if len(sg.opts.wrappers) > 0 {
		b.wrappers[sg.id] = sg.opts.wrappers
	}
	return nil
}

func (b *Builder) subgraphInputs(sg *subgraphDef) ([]subgraphInput, []dagcore.NodeID, error) {
	var (
		pairs  []subgraphInput
		depIDs []dagcore.NodeID
	)
	// Registration order keeps the dependency list deterministic.
	for _, in := range sg.sub.inputList {
		t := in.typ
		subID := in.id
		if ref, err := b.outputRef(t); err == nil {
			ref := ref
			pairs = append(pairs, subgraphInput{subID: subID, dep: &ref})
			depIDs = append(depIDs, ref.node)
			continue
		}
		def, ok := sg.sub.defaults[t]
		if !ok {
			return nil, nil, fmt.Errorf("subgraph %s: %w: %v", sg.id, ErrMissingDependency, t)
		}
		pairs = append(pairs, subgraphInput{subID: subID, value: def})
	}

	extra, err := b.resolveAfter(sg.opts)
	if err != nil {
		return nil, nil, err
	}
	depIDs = append(depIDs, extra...)
	return pairs, dedupeNodeIDs(depIDs), nil
}
