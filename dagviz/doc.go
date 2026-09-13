// Package dagviz renders a DAG as a Mermaid.js graph.
//
// ToMermaid walks the static topology of a dagcore.DAGInstance — its nodes,
// their dependencies and any subgraphs — and returns a string that Mermaid
// renders as a flowchart. Inputs are drawn as boxes, computed nodes as
// circles, and every dependency as an edge.
//
// It reads structure only: node results and timings are not included.
package dagviz
