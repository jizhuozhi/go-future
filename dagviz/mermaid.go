package dagviz

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jizhuozhi/go-future/dagcore"
)

// ToMermaid converts the DAG static topology to a Mermaid.js-compatible graph string.
func ToMermaid(d *dagcore.DAGInstance) string {
	var b strings.Builder
	b.WriteString("graph LR\n")
	writeMermaidRecursive(&b, d, "", "\t")
	return b.String()
}

func writeMermaidRecursive(b *strings.Builder, d *dagcore.DAGInstance, prefix string, indent string) {
	ids := make([]string, 0, len(d.Nodes()))
	for id := range d.Nodes() {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	for _, id := range ids {
		node := d.Nodes()[dagcore.NodeID(id)]
		label := prefix + id

		if node.Subgraph() != nil {
			_, _ = fmt.Fprintf(b, "%ssubgraph %s [Subgraph %s]\n", indent, label, label)
			writeMermaidRecursive(b, node.Subgraph(), label+".", indent+"\t")
			_, _ = fmt.Fprintf(b, "%send\n", indent)
		} else {
			if node.Input() {
				_, _ = fmt.Fprintf(b, "%s%s[/%q/]\n", indent, label, label)
			} else {
				_, _ = fmt.Fprintf(b, "%s%s[%q]\n", indent, label, label)
			}
		}
	}

	for _, id := range ids {
		node := d.Nodes()[dagcore.NodeID(id)]
		srcLabel := prefix + id
		for _, dep := range node.Deps() {
			depLabel := prefix + string(dep)
			_, _ = fmt.Fprintf(b, "%s%s --> %s\n", indent, depLabel, srcLabel)
		}
	}
}
