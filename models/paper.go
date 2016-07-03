package models

type Paper struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Read    bool   `json:"read"`
	Summary string `json:"summary"`

	Authors    []string `json:"authors"`
	References []int    `json:"references"`
	Tags       []string `json:"tags"`
}

func (p *Paper) Node() Node {
	return Node{
		ID:    p.ID,
		Label: string(p.Title),
		Type:  NodePaper,
	}
}
