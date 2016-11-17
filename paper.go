package papernet

type Paper struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`

	Tags []string `json:"tags"`
}

type PaperRepository interface {
	Get(...int) ([]*Paper, error)
	List() ([]*Paper, error)
	Upsert(*Paper) error
	Delete(int) error
}

type PaperSearch interface {
	Index(*Paper) error
	Search(titlePrefix string) ([]int, error)
}

type TagSearcher interface {
	Index(string) error
	Search(string) ([]string, error)
}
