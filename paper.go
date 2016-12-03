package papernet

type Paper struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`

	Tags []string `json:"tags"`
}

type PaperSearch struct {
	Q   string
	IDs []int
}

type PaperRepository interface {
	Get(...int) ([]*Paper, error)
	List() ([]*Paper, error)
	Upsert(*Paper) error
	Delete(int) error
}

type PaperIndex interface {
	Index(*Paper) error
	Search(PaperSearch) ([]int, error)
	Delete(int) error
}

type TagIndex interface {
	Index(string) error
	Search(string) ([]string, error)
}
