package mock

import (
	"github.com/bobinette/papernet"
)

type PaperRepository struct {
	db    []*papernet.Paper
	maxId int
}

func (r PaperRepository) Get(ids ...int) ([]*papernet.Paper, error) {
	papers := make([]*papernet.Paper, 0, len(ids))
	for _, id := range ids {
		for _, paper := range r.db {
			if paper.ID == id {
				papers = append(papers, paper)
			}
		}
	}
	return papers, nil
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
	if index+1 < len(r.db) {
		r.db = append(r.db[:index], r.db[index+1:]...)
	} else {
		r.db = r.db[:index]
	}
	return nil
}

func (r *PaperRepository) List() ([]*papernet.Paper, error) {
	return r.db, nil
}
