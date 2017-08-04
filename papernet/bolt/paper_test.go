package bolt

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/bobinette/papernet/papernet"
)

func createStore(t *testing.T) (*PaperRepository, func()) {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal("could not create tmp file:", err)
	}

	filename := tmpFile.Name()
	driver := Driver{}
	err = driver.Open(filename)
	if err != nil {
		os.Remove(filename)
		t.Fatal("could not create bucket: ", err)
	}
	repo := PaperRepository{Driver: &driver}

	return &repo, func() {
		driver.Close()
		os.Remove(filename)
	}
}

func TestStore_Insert_Get(t *testing.T) {
	store, f := createStore(t)
	defer f()

	nilTime := time.Time{}

	p := papernet.Paper{Title: "Test"}
	if err := store.Upsert(&p); err != nil {
		t.Fatal("error inserting:", err)
	}
	if p.ID <= 0 {
		t.Fatal("inserting should have set the id")
	}
	if p.CreatedAt == nilTime {
		t.Fatal("inserting should have set the created at")
	}
	if p.UpdatedAt == nilTime {
		t.Fatal("inserting should have set the updated at")
	}

	papers, err := store.Get(p.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	}
	if len(papers) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected 1 got %d", len(papers))
	}
	retrieved := papers[0]
	assertPaper(&p, &retrieved, t)

	papers, err = store.Get(p.ID, p.ID+1)
	if err != nil {
		t.Fatal("error getting:", err)
	}
	if len(papers) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected 1 got %d", len(papers))
	}
	retrieved = papers[0]
	assertPaper(&p, &retrieved, t)

	papers, err = store.Get(p.ID + 1)
	if err != nil {
		t.Fatal("error getting:", err)
	}
	if len(papers) != 0 {
		t.Fatalf("incorrect number of papers retrieved: expected 0 got %d", len(papers))
	}
}

func TestStore_Update(t *testing.T) {
	store, f := createStore(t)
	defer f()

	date := time.Now()
	p := papernet.Paper{ID: 1, Title: "Test", CreatedAt: date, UpdatedAt: date}
	if err := store.Upsert(&p); err != nil {
		t.Fatal("error inserting:", err)
	}

	p.Title = "Updated"
	if err := store.Upsert(&p); err != nil {
		t.Fatal("error inserting:", err)
	} else if p.CreatedAt != date {
		t.Fatal("inserting should NOT have changed the created at")
	} else if p.UpdatedAt == date {
		t.Fatal("inserting should have changed the updated at")
	}

	papers, err := store.Get(p.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if len(papers) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected 1 got %d", len(papers))
	}
	retrieved := papers[0]
	assertPaper(&p, &retrieved, t)
}

func TestStore_Delete(t *testing.T) {
	store, f := createStore(t)
	defer f()

	p := papernet.Paper{Title: "Test"}
	if err := store.Upsert(&p); err != nil {
		t.Fatal("error inserting:", err)
	}

	papers, err := store.Get(p.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if len(papers) != 1 {
		t.Fatalf("incorrect number of papers retrieved: expected 1 got %d", len(papers))
	}
	retrieved := papers[0]
	assertPaper(&p, &retrieved, t)

	err = store.Delete(p.ID)
	if err != nil {
		t.Fatal("error deleting", err)
	}

	papers, err = store.Get(p.ID)
	if err != nil {
		t.Fatal("error getting:", err)
	} else if len(papers) != 0 {
		t.Fatalf("incorrect number of papers retrieved: expected 0 got %d", len(papers))
	}
}

func TestStore_List(t *testing.T) {
	store, f := createStore(t)
	defer f()

	papers := []*papernet.Paper{
		&papernet.Paper{ID: 1, Title: "Test"},
		&papernet.Paper{ID: 2, Title: "Test 2", Summary: "Summary"},
		&papernet.Paper{ID: 3, Title: "Test 3", Summary: "Pizza yolo"},
	}
	for _, p := range papers {
		if err := store.Upsert(p); err != nil {
			t.Fatal("error inserting:", err)
		}
	}

	retrieved, err := store.List()
	if err != nil {
		t.Fatal("error getting:", err)
	}

	if len(papers) != len(retrieved) {
		t.Fatalf("incorrect number of papers retrieved: expected %d got %d", len(papers), len(retrieved))
	}

	for i, p := range retrieved {
		assertPaper(papers[i], &p, t)
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
