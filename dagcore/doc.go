// Package dagcore is a dependency-driven scheduler for static DAGs.
//
// A DAG is built by declaring its nodes and their dependencies, verified and
// locked with Freeze, then instantiated and executed. Nodes run in parallel as
// soon as their dependencies are satisfied; dependency tracking is lock-free,
// using per-node atomic counters rather than locks.
//
// # Stability
//
// dagcore is a long-term stable API, covered by the same stability as the rest
// of this module.
//
// dagfunc, the type-driven builder layered on top of it, is experimental. See
// that package for its status.
package dagcore
