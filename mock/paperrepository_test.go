package mock

import (
	"reflect"
	"testing"

	"github.com/bobinette/papernet"
)

func createRepository(t *testing.T) (*PaperRepository, func()) {
	return &PaperRepository{}, func() {}
}

func TestRepository_GetNone(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	paper, err := repo.Get(1)
	if err != nil {
		t.Fatal("non existing id should not return an error, got", err)
	} else if paper != nil {
		t.Fatal("expected nil, got a non-nil pointer")
	}
}

func TestRepository_Insert(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	paper := papernet.Paper{Title: "Test"}
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}

	retrieved, err := repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if !reflect.DeepEqual(*retrieved, paper) {
		t.Fatalf("incorrect paper retrieved: expected %+v got %+v", paper, *retrieved)
	}
}

func TestReposistory_Update(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	paper := papernet.Paper{Title: "Test"}
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}

	paper.Title = "Updated"
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}

	retrieved, err := repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if !reflect.DeepEqual(*retrieved, paper) {
		t.Fatalf("incorrect paper retrieved: expected %+v got %+v", paper, *retrieved)
	}
}
