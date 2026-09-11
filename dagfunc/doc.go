// Package dagfunc builds and runs type-driven DAGs on top of dagcore.
//
// Nodes are wired by Go type instead of by string id: Provide declares an input
// of some type, Use registers a function whose parameters are resolved to the
// nodes producing those types and whose results are published under their own
// types.
//
//	b := dagfunc.New()
//	_ = b.Provide(Question(""))
//	_ = b.Use(retrieve)   // func(ctx context.Context, q Question) (Docs, error)
//	_ = b.Use(answer)     // func(ctx context.Context, q Question, d Docs) (Answer, error)
//	_ = b.Freeze()
//
//	prog, _ := b.Compile([]any{Question("why is the sky blue?")})
//	_, _ = prog.Run(ctx)
//	ans, _ := prog.Value[Answer]()
//
// # Node ids
//
// Every node has a string id, used by NodeByID, by Wrap and by dagviz. The
// default id is derived from what the node produces:
//
//	input:<type>   for Provide
//	func:<type>    for Use, or func:<t1>,<t2> for a node with several results
//	func:<symbol>  for a node without results, from the runtime name of the function
//
// Name overrides it. Group prefixes the id of every node registered inside it
// with "<namespace>.". Two nodes may produce the same type as long as at least
// one of them carries an explicit Name; a type produced by several nodes can no
// longer be resolved by type alone, and reading it fails with ErrAmbiguousType
// until From or NodeByID pins one down.
//
// # Function signatures
//
// Use accepts every shape that can be reduced to "take some dependencies,
// produce some results":
//
//	func(ctx context.Context, A, B) (R, error)       the canonical form
//	func(ctx context.Context, A, B) R                cannot fail
//	func(ctx context.Context, A, B) error            sink, produces no value
//	func(ctx context.Context, A, B) (R, S, error)    several results
//	func(ctx context.Context, A, B) (R, S)           several results, cannot fail
//	func(A, B) (R, error)                            no context
//	func(ctx context.Context) (R, error)             source, no dependencies
//	func(ctx context.Context, A) *future.Future[R]   asynchronous node
//
// Three rules decide what counts as an error:
//
//   - is-a, not as-a: only a result declared as exactly `error` is an error. A
//     named type that merely implements it is an ordinary value, so
//     `func(ctx) (Result, MyError)` produces two values and can never fail,
//     while `func(ctx) (Result, error)` produces one value and may fail. What
//     the signature says is what the graph does.
//   - at most one error, and it has to be the last result: `(R, error, error)`
//     and `(error, R)` are rejected with ErrFuncSignature instead of being
//     guessed at.
//   - a failed node publishes nothing, not even the values it computed before
//     failing, so a half finished result can never be mistaken for a good one.
//
// Every other result becomes an output of the node and can be depended on by
// its type. A node returning a *future.Future[R] waits for that Future instead
// of computing R inline, which is how a node integrates an already
// asynchronous API.
//
// A node with several results carries them as a []any internally; they are
// still read one by one, by type, through Value, Get or Node.
//
// # Options
//
// Name, After, From, Timeout, Retry, RetryWith, Wrap, Recover, OrElse and
// Default configure a single node at registration time:
//
//	_ = b.Use(search,
//	    dagfunc.Name("search"),
//	    dagfunc.Timeout(200*time.Millisecond),
//	    dagfunc.RetryWith(2, 10*time.Millisecond),
//	    dagfunc.OrElse[Result](Result{}),
//	)
//
// The built-in options are layered from the inside out as retry, timeout,
// fallback, and Wrap wrappers are applied outside of all of them, so a wrapper
// always observes the final result of the node. After declares an ordering-only
// dependency, which is how a sink node, or a node that merely has to run after
// another one, is wired in.
//
// # Composition
//
// Group registers nodes under a namespace. It is only a naming device: the type
// registry is shared with the parent, so the nodes stay part of the same graph
// and are scheduled together.
//
// Subgraph embeds a whole Builder as one node of the parent graph. The subgraph
// keeps its own scheduler and its nodes still run in parallel, while the parent
// sees a single node wired by type: its inputs are fed from the parent, and
// Outputs declares which of its results the parent can read. Subgraphs nest.
//
// # Execution
//
// Freeze verifies the graph and locks it, Compile binds the inputs and returns a
// Program, and Run or RunAsync executes it. A Program is single use, a frozen
// Builder is not: compile and run as many Programs as needed, in parallel.
//
// Compile also takes dagcore node wrappers, which are applied to every node of
// that run and are the hook for tracing, metrics and logging:
//
//	prog, _ := b.Compile(inputs, func(n *dagcore.NodeInstance, run dagcore.NodeFunc) dagcore.NodeFunc {
//	    return func(ctx context.Context, deps map[dagcore.NodeID]any) (any, error) {
//	        start := time.Now()
//	        out, err := run(ctx, deps)
//	        log.Printf("%s took %s", n.ID(), time.Since(start))
//	        return out, err
//	    }
//	})
//
// # Reading results
//
// Value[T] and ValueAsync[T] read one output by type, without a sample value
// and without an assertion. Get(sample) is the version independent form and
// returns an any. Node(sample) and NodeByID(id) expose the underlying
// dagcore.NodeInstance, which carries the per node Future, Duration and Cast,
// and Instance exposes the whole runtime, which is what dagviz renders.
//
// # Layering
//
// dagfunc is a layer over dagcore, not a subset of it: naming, per node
// wrappers, subgraphs, defaults and typed result access are all reachable
// without dropping down to dagcore, and Program.Instance hands over the
// dagcore runtime whenever something lower level is needed.
package dagfunc
