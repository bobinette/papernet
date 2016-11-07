package papernet

type Paper struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type PaperRepository interface {
	Get(int) (*Paper, error)
	Upsert(*Paper) error
}
