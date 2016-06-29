package models

type NodeType int

const (
	NodePaper NodeType = iota
	NodeTag
)

type Node struct {
	ID    int
	Type  NodeType
	Label string
}

type EdgeType int

const (
	EdgeReference EdgeType = iota
	EdgeTag
)

type Edge struct {
	Style EdgeType
}
