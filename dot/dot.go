package dot

import (
	"fmt"
	"io"
	"strings"

	"github.com/bobinette/papernet/models"
)

var nodeColors = map[models.NodeType]string{
	models.NodePaper: "#000000",
	models.NodeTag:   "#FF0000",
}

var edgeColors = map[models.EdgeType]string{
	models.EdgeReference: "#000000",
	models.EdgeTag:       "#FF0000",
}

type node struct {
	models.Node
	ref string
}

type edge struct {
	e models.Edge
	s int
	t int
}

type Graph struct {
	vertices map[int]node
	edges    []edge
}

func NewGraph() *Graph {
	return &Graph{
		vertices: make(map[int]node),
		edges:    make([]edge, 0),
	}
}

func (g *Graph) AddVertex(n models.Node) {
	ref := ""
	switch n.Type {
	case models.NodeTag:
		ref = "T"
	case models.NodePaper:
		ref = "P"
	}

	ref = fmt.Sprintf("%s%d", ref, n.ID)

	g.vertices[n.ID] = node{n, ref}
}

func (g *Graph) AddEdge(s, t models.Node, e models.Edge) {
	g.AddVertex(s)
	g.AddVertex(t)
	g.edges = append(g.edges, edge{e: e, s: s.ID, t: t.ID})
}

func (g *Graph) WriteTo(w io.Writer) (int64, error) {
	c := "digraph {\n"
	c += "  bgcolor=\"#00000000\"\n"
	c += "  size=\"8.0,8.0!\"\n"
	c += "  overlap=false\n"

	for _, n := range g.vertices {
		c += fmt.Sprintf("  %s\n", g.formatNode(n))
	}

	for _, e := range g.edges {
		c += fmt.Sprintf("  %s -> %s", g.vertices[e.s].ref, g.vertices[e.t].ref)
		c += " " + fmt.Sprintf(`[color="%s"]`, edgeColors[e.e.Style]) + "\n"
	}
	c += "}"

	n, err := w.Write([]byte(c))
	return int64(n), err
}

func (g *Graph) formatNode(n node) string {
	attr := map[string]string{
		"label":     n.Label,
		"color":     nodeColors[n.Type],
		"fontcolor": nodeColors[n.Type],
		"fontsize":  "16.0",
	}

	if n.Type == models.NodePaper {
		attr["URL"] = fmt.Sprintf("/papers/%d/show", n.ID)
	}

	attrl := make([]string, len(attr))
	i := 0
	for k, v := range attr {
		attrl[i] = fmt.Sprintf("%s=\"%s\"", k, v)
		i++
	}

	return fmt.Sprintf("%s [%s]", n.ref, strings.Join(attrl, ","))
}
