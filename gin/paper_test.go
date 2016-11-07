package gin

import (
	"bytes"
	"encoding/json"
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

		r := make(map[string]interface{})
		err := json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Error("could not decode response as JSON:", err)
		}
		if resp.Code != tt.Code {
			t.Errorf("incorrect code: expected %d got %d", tt.Code, resp.Code)
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
		data, err := json.Marshal(tt.Paper)
		if err != nil {
			t.Fatal("cannot unmarshal:", err)
		}

		buf := bytes.Buffer{}
		_, err = buf.Write(data)
		if err != nil {
			t.Fatal("cannot write:", err)
		}

		req := httptest.NewRequest("POST", url, &buf)
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
		err = json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Error("could not decode response as JSON:", err)
		}

		if r.Data.ID <= 0 {
			t.Errorf("response should have a positive ID, got %d", r.Data.ID)
		}
	}
}
