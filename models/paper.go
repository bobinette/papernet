package models

type Paper struct {
	ID      int
	Title   []byte
	Read    bool
	Summary []byte

	Authors    []string
	References []int
	Tags       []string
}

func (p *Paper) Node() Node {
	return Node{
		ID:    p.ID,
		Label: string(p.Title),
		Type:  NodePaper,
	}
}
