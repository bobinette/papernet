package gin

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/mock"
)

func TestGet(t *testing.T) {
	handler := &PaperHandler{
		Repository: &mock.PaperRepository{},
	}
	err := handler.Repository.Upsert(&papernet.Paper{
		ID:      1,
		Title:   "Test",
		Summary: "Test",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	gin.SetMode(gin.ReleaseMode) // avoid unnecessary log
	router := gin.New()
	router.GET("/get/:id", handler.Get)

	var tts = []struct {
		Query string
		Code  int
	}{
		{
			// Paper is inserted above
			Query: "/get/1",
			Code:  200,
		},
		{
			// test cannot be decoded as an int
			Query: "/get/test",
			Code:  400,
		},
		{
			// 2 is not in the database
			Query: "/get/2",
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
