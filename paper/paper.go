package paper

type Paper struct {
	// Core attributes
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`

	// Fancy attributes
	Read       Reading   `json:"read"`
	Type       PaperType `json:"type"`
	Year       int       `json:"year"`
	URLs       []string  `json:"urls"`
	Bookmarked bool      `json:"bookmarked"`

	// Relations
	Authors    []string    `json:"authors"`
	References []Reference `json:"references"`
	Tags       []string    `json:"tags"`
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
	PaperTypeWebPage
)

type Reading int

const (
	ReadingNotRead Reading = iota
	ReadingRead
	ReadingInProgress
)
