package gin

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/mock"
)

func createRouter(t *testing.T) (*gin.Engine, *PaperHandler) {
	handler := &PaperHandler{
		Repository: &mock.PaperRepository{},
	}

	gin.SetMode(gin.ReleaseMode) // avoid unnecessary log
	router := gin.New()
	handler.RegisterRoutes(router)

	return router, handler
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
	router, handler := createRouter(t)

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
			Query: "/papernet/papers/1",
			Code:  200,
		},
		{
			// test cannot be decoded as an int
			Query: "/papernet/papers/test",
			Code:  400,
		},
		{
			// 2 is not in the database
			Query: "/papernet/papers/2",
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
	router, _ := createRouter(t)

	url := "/papernet/papers"
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
	router, handler := createRouter(t)

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
			Query: "/papernet/papers/1",
			Paper: papernet.Paper{
				ID:      1,
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 200,
		},
		{
			Name:  "test cannot be decoded as an int",
			Query: "/papernet/papers/test",
			Paper: papernet.Paper{
				ID:      1,
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 400,
		},
		{
			Name:  "2 is not in the database",
			Query: "/papernet/papers/2",
			Paper: papernet.Paper{
				ID:      2,
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 404,
		},
		{
			Name:  "IDs do not correspond",
			Query: "/papernet/papers/1",
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
			t.Error("%s - could not decode response as JSON:", tt.Name, err)
		}
	}
}

func TestDelete(t *testing.T) {
	router, handler := createRouter(t)

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
			Query: "/papernet/papers/1",
			Code:  200,
		},
		{
			// test cannot be decoded as an int
			Query: "/papernet/papers/test",
			Code:  400,
		},
		{
			// 2 is not in the database
			Query: "/papernet/papers/2",
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
	router, handler := createRouter(t)

	// Load fixtures
	papers := []*papernet.Paper{
		&papernet.Paper{ID: 1, Title: "Test", Summary: "Test"},
		&papernet.Paper{ID: 2, Title: "Test 2", Summary: "Summary"},
	}
	for _, paper := range papers {
		err := handler.Repository.Upsert(paper)
		if err != nil {
			t.Fatal("could not insert paper:", err)
		}
	}

	var tts = []struct {
		Query string
		Code  int
		Len   int
	}{
		{
			// List all papers, simply
			Query: "/papernet/papers",
			Code:  200,
			Len:   2,
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
			t.Error("wrong number of papers extracted: expected %d got %d", tt.Len, len(r.Data))
		}
	}
}
