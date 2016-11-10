package mock

import (
	"github.com/bobinette/papernet"
)

type PaperRepository struct {
	db    []*papernet.Paper
	maxId int
}

func (r PaperRepository) Get(id int) (*papernet.Paper, error) {
	for _, paper := range r.db {
		if paper.ID == id {
			return paper, nil
		}
	}
	return nil, nil
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
		r.db = make([]*papernet.Paper, 0)
	}

	index := -1
	for i, dbPaper := range r.db {
		if dbPaper.ID == paper.ID {
			index = i
			break
		}
	}
	if index >= 0 && index < len(r.db) {
		r.db[index] = paper
	} else {
		r.db = append(r.db, paper)
	}

	return nil
}

func (r *PaperRepository) Delete(id int) error {
	index := -1
	for i, paper := range r.db {
		if paper.ID == id {
			index = i
			break
		}
	}
	r.db = append(r.db[:index], r.db[index+1:]...)
	return nil
}

func (r *PaperRepository) List() ([]*papernet.Paper, error) {
	return r.db, nil
}
