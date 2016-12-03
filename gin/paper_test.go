package gin

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/mock"
)

func createRouter(t *testing.T) (*gin.Engine, *PaperHandler, func()) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("could not create tmp file:", err)
	}

	// Too lazy to mock the search now...
	index := &bleve.PaperIndex{}
	err = index.Open(path.Join(dir, "test"))
	if err != nil {
		t.Fatal("error creating index", err)
	}

	handler := &PaperHandler{
		Repository: &mock.PaperRepository{},
		Searcher:   index,
	}

	gin.SetMode(gin.ReleaseMode) // avoid unnecessary log
	router := gin.New()
	handler.RegisterRoutes(router)

	return router, handler, func() {
		if err := index.Close(); err != nil {
			t.Log(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Log(err)
		}
	}
}

func createReader(i interface{}, t *testing.T) io.Reader {
	data, err := json.Marshal(i)
	if err != nil {
		t.Fatal("cannot unmarshal:", err)
	}

	buf := bytes.Buffer{}
	_, err = buf.Write(data)
	if err != nil {
		t.Fatal("cannot write:", err)
	}

	return &buf
}

func TestGet(t *testing.T) {
	router, handler, f := createRouter(t)
	defer f()

	// Load fixtures
	err := handler.Repository.Upsert(&papernet.Paper{
		ID:      1,
		Title:   "Test",
		Summary: "Test",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	var tts = []struct {
		Query string
		Code  int
	}{
		{
			// Paper is inserted above
			Query: "/api/papers/1",
			Code:  200,
		},
		{
			// test cannot be decoded as an int
			Query: "/api/papers/test",
			Code:  400,
		},
		{
			// 2 is not in the database
			Query: "/api/papers/2",
			Code:  404,
		},
	}

	for _, tt := range tts {
		req := httptest.NewRequest("GET", tt.Query, nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != tt.Code {
			t.Errorf("incorrect code: expected %d got %d", tt.Code, resp.Code)
		}

		r := make(map[string]interface{})
		err := json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Error("could not decode response as JSON:", err)
		}
	}
}

func TestInsert(t *testing.T) {
	router, _, f := createRouter(t)
	defer f()

	url := "/api/papers"
	var tts = []struct {
		Paper papernet.Paper
		Code  int
	}{
		{
			Paper: papernet.Paper{
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 200,
		},
		{
			Paper: papernet.Paper{
				ID:      3,
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 400,
		},
	}

	for _, tt := range tts {
		reader := createReader(tt.Paper, t)
		req := httptest.NewRequest("POST", url, reader)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		if resp.Code != tt.Code {
			t.Errorf("incorrect code: expected %d got %d", tt.Code, resp.Code)
		}
		if resp.Code >= 400 {
			continue
		}

		var r struct {
			Data papernet.Paper
		}
		err := json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Error("could not decode response as JSON:", err)
		}

		if r.Data.ID <= 0 {
			t.Errorf("response should have a positive ID, got %d", r.Data.ID)
		}
	}
}

func TestUpdate(t *testing.T) {
	router, handler, f := createRouter(t)
	defer f()

	// Load fixtures
	err := handler.Repository.Upsert(&papernet.Paper{
		ID:      1,
		Title:   "Test",
		Summary: "Test",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	var tts = []struct {
		Name  string
		Query string
		Paper papernet.Paper
		Code  int
	}{
		{
			Name:  "Paper is inserted above",
			Query: "/api/papers/1",
			Paper: papernet.Paper{
				ID:      1,
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 200,
		},
		{
			Name:  "test cannot be decoded as an int",
			Query: "/api/papers/test",
			Paper: papernet.Paper{
				ID:      1,
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 400,
		},
		{
			Name:  "2 is not in the database",
			Query: "/api/papers/2",
			Paper: papernet.Paper{
				ID:      2,
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 404,
		},
		{
			Name:  "IDs do not correspond",
			Query: "/api/papers/1",
			Paper: papernet.Paper{
				ID:      2,
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 400,
		},
	}

	for _, tt := range tts {
		reader := createReader(tt.Paper, t)
		req := httptest.NewRequest("PUT", tt.Query, reader)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		if resp.Code != tt.Code {
			t.Errorf("%s - incorrect code: expected %d got %d (%s)", tt.Name, tt.Code, resp.Code, resp.Body.String())
		}

		r := make(map[string]interface{})
		err := json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Errorf("%s - could not decode response as JSON:", tt.Name, err)
		}
	}
}

func TestDelete(t *testing.T) {
	router, handler, f := createRouter(t)
	defer f()

	// Load fixtures
	err := handler.Repository.Upsert(&papernet.Paper{
		ID:      1,
		Title:   "Test",
		Summary: "Test",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	var tts = []struct {
		Query string
		Code  int
	}{
		{
			// Paper is inserted above
			Query: "/api/papers/1",
			Code:  200,
		},
		{
			// test cannot be decoded as an int
			Query: "/api/papers/test",
			Code:  400,
		},
		{
			// 2 is not in the database
			Query: "/api/papers/2",
			Code:  404,
		},
	}

	for _, tt := range tts {
		req := httptest.NewRequest("DELETE", tt.Query, nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != tt.Code {
			t.Errorf("incorrect code: expected %d got %d", tt.Code, resp.Code)
		}

		if tt.Code >= 400 {
			continue
		}

		r := make(map[string]interface{})
		err := json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Error("could not decode response as JSON:", err)
		}

		req = httptest.NewRequest("GET", tt.Query, nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != 404 {
			t.Errorf("seems like I can still GET %s", tt.Query)
		}
	}
}

func TestList(t *testing.T) {
	router, handler, f := createRouter(t)
	defer f()

	// Load fixtures
	papers := []*papernet.Paper{
		&papernet.Paper{ID: 1, Title: "Test", Summary: "Test"},
		&papernet.Paper{ID: 2, Title: "Pizza yolo", Summary: "Summary"},
	}
	for _, paper := range papers {
		err := handler.Repository.Upsert(paper)
		if err != nil {
			t.Fatal("could not insert paper:", err)
		}

		err = handler.Searcher.Index(paper)
		if err != nil {
			t.Fatal("could not index paper:", err)
		}
	}

	var tts = []struct {
		Query string
		Code  int
		Len   int
	}{
		{
			// List all papers, simply
			Query: "/api/papers",
			Code:  200,
			Len:   2,
		},
		{
			// List all papers with title starting by piz
			Query: "/api/papers?q=piz",
			Code:  200,
			Len:   1,
		},
	}

	for _, tt := range tts {
		req := httptest.NewRequest("GET", tt.Query, nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != tt.Code {
			t.Errorf("incorrect code: expected %d got %d", tt.Code, resp.Code)
		}

		var r struct {
			Data []*papernet.Paper `json:"data"`
		}
		err := json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Error("could not decode response as JSON:", err)
		}

		if len(r.Data) != tt.Len {
			t.Errorf("wrong number of papers extracted: expected %d got %d", tt.Len, len(r.Data))
		}
	}
}
