package dot

import (
	"io"
	"strconv"
)

type Graph struct {
	vertices map[int]struct{}
	edges    map[int]int
}

func NewGraph() *Graph {
	return &Graph{
		vertices: make(map[int]struct{}),
		edges:    make(map[int]int),
	}
}

func (g *Graph) AddVertex(v int) {
	g.vertices[v] = struct{}{}
}

func (g *Graph) AddEdge(s, t int) {
	g.AddVertex(s)
	g.AddVertex(t)
	g.edges[s] = t
}

func (g *Graph) WriteTo(w io.Writer) (int64, error) {
	c := "digraph {\n"
	for s, t := range g.edges {
		c += "  " + strconv.Itoa(s) + " -> " + strconv.Itoa(t) + "\n"
	}
	c += "}"

	n, err := w.Write([]byte(c))
	return int64(n), err
}
