package models

type Paper struct {
	// Core attributes
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`

	// Fancy attributes
	Read bool      `json:"read"`
	Type PaperType `json:"type"`
	Year int       `json:"year"`
	URLs []string  `json:"urls"`

	// Relations
	Authors    []string    `json:"authors"`
	References []Reference `json:"references"`
	Tags       []string    `json:"tags"`
}

func (p *Paper) Node() Node {
	return Node{
		ID:    p.ID,
		Label: string(p.Title),
		Type:  NodePaper,
	}
}

type Reference struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type PaperType int

const (
	PaperTypePaper PaperType = iota
	PaperTypeBook
	PaperTypeSlides
)
