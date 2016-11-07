package mock

import (
	"github.com/bobinette/papernet"
)

type PaperRepository struct {
	db    map[int]*papernet.Paper
	maxId int
}

func (r PaperRepository) Get(id int) (*papernet.Paper, error) {
	if r.db == nil {
		r.db = make(map[int]*papernet.Paper)
	}
	return r.db[id], nil
}

func (r *PaperRepository) Upsert(paper *papernet.Paper) error {
	if paper.ID <= 0 {
		r.maxId++
		paper.ID = r.maxId
	}

	if paper.ID > r.maxId {
		r.maxId = paper.ID
	}

	if r.db == nil {
		r.db = make(map[int]*papernet.Paper)
	}
	r.db[paper.ID] = paper
	return nil
}
