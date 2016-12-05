package bolt

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/boltdb/bolt"

	"github.com/bobinette/papernet"
)

func createRepository(t *testing.T) (*PaperRepository, func()) {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal("could not create tmp file:", err)
	}

	filename := tmpFile.Name()
	store, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		os.Remove(filename)
		t.Fatalf("could not open bolt on file %s: %v", filename, err)
	}

	err = store.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(paperBucket)
		return err
	})
	if err != nil {
		os.Remove(filename)
		t.Fatal("could not create bucket: ", err)
	}
	driver := Driver{store: store}
	repo := PaperRepository{Driver: &driver}

	return &repo, func() {
		store.Close()
		os.Remove(filename)
	}
}

func TestRepository_Insert_Get(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	nilTime := time.Time{}

	paper := papernet.Paper{Title: "Test"}
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}
	if paper.ID <= 0 {
		t.Fatal("inserting should have set the id")
	}
	if paper.CreatedAt == nilTime {
		t.Fatal("inserting should have set the created at")
	}
	if paper.UpdatedAt == nilTime {
		t.Fatal("inserting should have set the updated at")
	}

	papers, err := repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	}
	if len(papers) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected 1 got %d", len(papers))
	}
	retrieved := papers[0]
	assertPaper(&paper, retrieved, t)

	papers, err = repo.Get(paper.ID, paper.ID+1)
	if err != nil {
		t.Fatal("error getting:", err)
	}
	if len(papers) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected 1 got %d", len(papers))
	}
	retrieved = papers[0]
	assertPaper(&paper, retrieved, t)

	papers, err = repo.Get(paper.ID + 1)
	if err != nil {
		t.Fatal("error getting:", err)
	}
	if len(papers) != 0 {
		t.Fatalf("incorrect number of papers retrieved: expected 0 got %d", len(papers))
	}
}

func TestReposistory_Update(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	date := time.Now()
	paper := papernet.Paper{ID: 1, Title: "Test", CreatedAt: date, UpdatedAt: date}
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}

	paper.Title = "Updated"
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	} else if paper.CreatedAt != date {
		t.Fatal("inserting should NOT have changed the created at")
	} else if paper.UpdatedAt == date {
		t.Fatal("inserting should have changed the updated at")
	}

	papers, err := repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if len(papers) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected 1 got %d", len(papers))
	}
	retrieved := papers[0]
	assertPaper(&paper, retrieved, t)
}

func TestRepository_Delete(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	paper := papernet.Paper{Title: "Test"}
	if err := repo.Upsert(&paper); err != nil {
		t.Fatal("error inserting:", err)
	}

	papers, err := repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if len(papers) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected 1 got %d", len(papers))
	}
	retrieved := papers[0]
	assertPaper(&paper, retrieved, t)

	err = repo.Delete(paper.ID)
	if err != nil {
		t.Fatal("error deleting", err)
	}

	papers, err = repo.Get(paper.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if len(papers) != 0 {
		t.Fatalf("incorrect number of papers retrieved: expected 0 got %d", len(papers))
	}
}

func TestRepository_List(t *testing.T) {
	repo, f := createRepository(t)
	defer f()

	papers := []*papernet.Paper{
		&papernet.Paper{ID: 1, Title: "Test"},
		&papernet.Paper{ID: 2, Title: "Test 2", Summary: "Summary"},
		&papernet.Paper{ID: 3, Title: "Test 3", Summary: "Pizza yolo"},
	}
	for _, paper := range papers {
		if err := repo.Upsert(paper); err != nil {
			t.Fatal("error inserting:", err)
		}
	}

	retrieved, err := repo.List()
	if err != nil {
		t.Fatal("error getting:", err)
	}

	if len(papers) != len(retrieved) {
		t.Fatalf("incorrect number of papers retrieved: expected %d got %d", len(papers), len(retrieved))
	}

	for i, paper := range retrieved {
		assertPaper(papers[i], paper, t)
	}
}

func assertPaper(exp, got *papernet.Paper, t *testing.T) {
	if exp.Title != got.Title {
		t.Errorf("invalid title: expected %s got %s", exp.Title, got.Title)
	}

	if exp.Summary != got.Summary {
		t.Errorf("invalid summary: expected %s got %s", exp.Summary, got.Summary)
	}

	if !reflect.DeepEqual(exp.Tags, got.Tags) {
		t.Errorf("invalid tags: expected %v got %v", exp.Tags, got.Tags)
	}
}
