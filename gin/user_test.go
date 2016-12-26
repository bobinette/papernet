package gin

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bobinette/papernet"
)

func TestUserHandler_Me(t *testing.T) {
	t.Skip("need bleve driver")
	router, handlers, f := createRouter(t)
	handler := handlers.UserHandler
	defer f()

	// Load fixtures
	user := &papernet.User{
		ID:        "1",
		Name:      "Test user",
		Bookmarks: []int{1},
	}
	err := handler.Repository.Upsert(user)
	if err != nil {
		t.Fatal("could not insert user:", err)
	}

	userToken, err := handler.Authenticator.Encoder.Encode(user.ID)
	if err != nil {
		t.Fatal("could not encode token:", err)
	}

	var tts = []struct {
		Token string
		Code  int
	}{
		{
			Token: "",
			Code:  401,
		},
		{
			Token: "not a bearer",
			Code:  401,
		},
		{
			Token: "bearer not.a.token",
			Code:  401,
		},
		{
			Token: fmt.Sprintf("bearer %s", userToken),
			Code:  200,
		},
	}

	for i, tt := range tts {
		req := httptest.NewRequest("GET", "/api/me", nil)
		req.Header.Add("authorization", tt.Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != tt.Code {
			t.Errorf("%d - incorrect code: expected %d got %d (body: %v)", i, tt.Code, resp.Code, resp.Body.String())
		}
	}
}

func TestUserHandler_UpdateBookmarks(t *testing.T) {
	t.Skip("need bleve driver")
	router, handlers, f := createRouter(t)
	handler := handlers.UserHandler
	defer f()

	// Load fixtures
	user := &papernet.User{
		ID:        "1",
		Name:      "Test user",
		Bookmarks: []int{1},
	}
	err := handler.Repository.Upsert(user)
	if err != nil {
		t.Fatal("could not insert user:", err)
	}

	userToken, err := handler.Authenticator.Encoder.Encode(user.ID)
	if err != nil {
		t.Fatal("could not encode token:", err)
	}

	var tts = []struct {
		Token    string
		Add      []int
		Remove   []int
		Code     int
		Expected []int
	}{
		{
			Token:    "",
			Add:      []int{},
			Remove:   []int{},
			Code:     401,
			Expected: []int{},
		},
		{
			Token:    fmt.Sprintf("bearer %s", userToken),
			Add:      []int{2},
			Remove:   []int{},
			Code:     200,
			Expected: []int{1, 2},
		},
		{
			Token:    fmt.Sprintf("bearer %s", userToken),
			Add:      []int{3},
			Remove:   []int{2},
			Code:     200,
			Expected: []int{1, 3},
		},
		{
			Token:    fmt.Sprintf("bearer %s", userToken),
			Add:      []int{2},
			Remove:   []int{2, 3},
			Code:     200,
			Expected: []int{1},
		},
	}

	for i, tt := range tts {
		reader := createReader(map[string]interface{}{
			"add":    tt.Add,
			"remove": tt.Remove,
		}, t)
		req := httptest.NewRequest("POST", "/api/bookmarks", reader)
		req.Header.Add("authorization", tt.Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != tt.Code {
			t.Errorf("%d - incorrect code: expected %d got %d (body: %v)", i, tt.Code, resp.Code, resp.Body.String())
			continue
		}

		if tt.Code >= 400 {
			continue
		}

		var r struct {
			Data papernet.User `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &r); err != nil {
			t.Errorf("%d - could not read json body: %v", i, err)
		} else if !reflect.DeepEqual(tt.Expected, r.Data.Bookmarks) {
			t.Errorf("%d - incorrect bookmarks: expected %v got %v", i, tt.Expected, r.Data.Bookmarks)
		}
	}
}
