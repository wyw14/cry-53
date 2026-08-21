package domain

import "sort"

type DependencyNode struct {
	ID          string
	Environment string
}

type DependencyEdge struct {
	From string
	To   string
}

type DependencyGraph struct {
	nodes map[string]DependencyNode
	edges map[string][]string
}

func NewDependencyGraph(nodes []DependencyNode) *DependencyGraph {
	graph := &DependencyGraph{nodes: make(map[string]DependencyNode), edges: make(map[string][]string)}
	for _, node := range nodes {
		graph.nodes[node.ID] = node
	}
	return graph
}

func (g *DependencyGraph) Connect(edge DependencyEdge) bool {
	from, fromKnown := g.nodes[edge.From]
	to, toKnown := g.nodes[edge.To]
	if !fromKnown || !toKnown {
		return false
	}
	if from.Environment != to.Environment {
		return false
	}
	g.edges[edge.From] = append(g.edges[edge.From], edge.To)
	return true
}

func (g *DependencyGraph) Cycles() [][]string {
	state := make(map[string]uint8)
	stack := make([]string, 0, len(g.nodes))
	cycles := make([][]string, 0)
	var visit func(string)
	visit = func(node string) {
		state[node] = 1
		stack = append(stack, node)
		neighbors := append([]string(nil), g.edges[node]...)
		sort.Strings(neighbors)
		for _, next := range neighbors {
			if state[next] == 0 {
				visit(next)
				continue
			}
			if state[next] != 1 {
				continue
			}
			start := 0
			for index, candidate := range stack {
				if candidate == next {
					start = index
					break
				}
			}
			cycle := append([]string(nil), stack[start:]...)
			cycles = append(cycles, append(cycle, next))
		}
		stack = stack[:len(stack)-1]
		state[node] = 2
	}
	keys := make([]string, 0, len(g.nodes))
	for key := range g.nodes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if state[key] == 0 {
			visit(key)
		}
	}
	return cycles
}
