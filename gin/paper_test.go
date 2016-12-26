package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/auth"
	"github.com/bobinette/papernet/bleve"
	"github.com/bobinette/papernet/bolt"
)

type handlers struct {
	PaperHandler *PaperHandler
	UserHandler  *UserHandler
}

func createRouter(t *testing.T) (*gin.Engine, handlers, func()) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("could not create tmp file:", err)
	}

	index := &bleve.PaperIndex{}
	err = index.Open(path.Join(dir, "test", "index"))
	if err != nil {
		t.Fatal("error creating index", err)
	}

	driver := &bolt.Driver{}
	err = driver.Open(path.Join(dir, "test", "db"))
	if err != nil {
		t.Fatal("error creating index", err)
	}

	encoder := auth.Encoder{Key: "test"}

	paperStore := &bolt.PaperStore{Driver: driver}
	userRepo := &bolt.UserRepository{Driver: driver}
	tagIndex := &bolt.TagIndex{Driver: driver}

	paperHandler := &PaperHandler{
		Store:          paperStore,
		Searcher:       index,
		TagIndex:       tagIndex,
		UserRepository: userRepo,
		Authenticator: Authenticator{
			UserRepository: userRepo,
			Encoder:        encoder,
		},
	}

	userHandler := &UserHandler{
		Authenticator: Authenticator{
			UserRepository: userRepo,
			Encoder:        encoder,
		},
		Repository: userRepo,
	}

	gin.SetMode(gin.ReleaseMode) // avoid unnecessary log
	router := gin.New()
	paperHandler.RegisterRoutes(router)
	userHandler.RegisterRoutes(router)

	return router, handlers{
			PaperHandler: paperHandler,
			UserHandler:  userHandler,
		}, func() {
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

func TestPaperHandler_Get(t *testing.T) {
	router, handlers, f := createRouter(t)
	handler := handlers.PaperHandler
	defer f()

	// Load fixtures
	err := handler.Store.Upsert(&papernet.Paper{
		ID:      1,
		Title:   "Test",
		Summary: "Test",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	err = handler.Store.Upsert(&papernet.Paper{
		ID:      2,
		Title:   "Test 2",
		Summary: "Test 2",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	user := &papernet.User{
		ID:        "1",
		Name:      "Test user",
		Bookmarks: []int{1},
		CanSee:    []int{1},
	}
	if err := handler.Authenticator.UserRepository.Upsert(user); err != nil {
		t.Fatal("could not insert user:", err)
	}
	token, err := handler.Authenticator.Encoder.Encode(user.ID)
	if err != nil {
		t.Fatal("could not fake token:", err)
	}
	bearer := fmt.Sprint("Bearer ", token)

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
			// User is not allowed to see 2 -> 404 à la Github
			Query: "/api/papers/2",
			Code:  404,
		},
		{
			// 3 is not in the database
			Query: "/api/papers/3",
			Code:  404,
		},
	}

	for _, tt := range tts {
		req := httptest.NewRequest("GET", tt.Query, nil)
		req.Header.Add("Authorization", bearer)

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

func TestPaperHandler_Insert(t *testing.T) {
	router, handlers, f := createRouter(t)
	handler := handlers.PaperHandler
	defer f()

	user := &papernet.User{
		ID:        "1",
		Name:      "Test user",
		Bookmarks: []int{1},
	}
	if err := handler.Authenticator.UserRepository.Upsert(user); err != nil {
		t.Fatal("could not insert user:", err)
	}
	token, err := handler.Authenticator.Encoder.Encode(user.ID)
	if err != nil {
		t.Fatal("could not fake token:", err)
	}
	bearer := fmt.Sprint("Bearer ", token)

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
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
				Tags:    []string{"Supa tag"},
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
		req.Header.Add("Authorization", bearer)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != tt.Code {
			t.Errorf("incorrect code: expected %d got %d", tt.Code, resp.Code)
		}
		if resp.Code >= 400 {
			continue
		}

		var r struct {
			Data papernet.Paper `json:"data"`
		}
		err := json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Error("could not decode response as JSON:", err)
		}

		if r.Data.ID <= 0 {
			t.Errorf("response should have a positive ID, got %d", r.Data.ID)
		} else if r.Data.Title != tt.Paper.Title {
			t.Errorf("incorrect title: expected %s got %s", tt.Paper.Title, r.Data.Title)
		}

		// Check that the paper is available for the user
		user, err = handler.Authenticator.UserRepository.Get(user.ID)
		if err != nil {
			t.Error("error getting user:", err)
		} else if !isIn(r.Data.ID, user.CanSee) {
			t.Error("Paper should be added to user can see")
		} else if !isIn(r.Data.ID, user.CanEdit) {
			t.Error("Paper should be added to user can edit")
		}
	}
}

func TestPaperHandler_Update(t *testing.T) {
	router, handlers, f := createRouter(t)
	handler := handlers.PaperHandler
	defer f()

	// Load fixtures
	err := handler.Store.Upsert(&papernet.Paper{
		ID:      1,
		Title:   "Test",
		Summary: "Test",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	err = handler.Store.Upsert(&papernet.Paper{
		ID:      2,
		Title:   "Test",
		Summary: "Test",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	user := &papernet.User{
		ID:        "1",
		Name:      "Test user",
		Bookmarks: []int{1},
		CanSee:    []int{1},
		CanEdit:   []int{1},
	}
	if err := handler.Authenticator.UserRepository.Upsert(user); err != nil {
		t.Fatal("could not insert user:", err)
	}
	token, err := handler.Authenticator.Encoder.Encode(user.ID)
	if err != nil {
		t.Fatal("could not fake token:", err)
	}
	bearer := fmt.Sprint("Bearer ", token)

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
			Name:  "User not allowed to edit 2",
			Query: "/api/papers/2",
			Paper: papernet.Paper{
				ID:      2,
				Title:   "Pizza Yolo",
				Summary: "Paper for test",
			},
			Code: 403,
		},
		{
			Name:  "3 is not in the database",
			Query: "/api/papers/3",
			Paper: papernet.Paper{
				ID:      3,
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
		req.Header.Add("Authorization", bearer)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != tt.Code {
			t.Errorf("%s - incorrect code: expected %d got %d (%s)", tt.Name, tt.Code, resp.Code, resp.Body.String())
		}

		r := make(map[string]interface{})
		err := json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Errorf("%s - could not decode response as JSON: %v", tt.Name, err)
		}
	}
}

func TestPaperHandler_Delete(t *testing.T) {
	router, handlers, f := createRouter(t)
	handler := handlers.PaperHandler
	defer f()

	// Load fixtures
	err := handler.Store.Upsert(&papernet.Paper{
		ID:      1,
		Title:   "Test",
		Summary: "Test",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	err = handler.Store.Upsert(&papernet.Paper{
		ID:      2,
		Title:   "Test",
		Summary: "Test",
	})
	if err != nil {
		t.Fatal("could not insert paper:", err)
	}

	user := &papernet.User{
		ID:        "1",
		Name:      "Test user",
		Bookmarks: []int{1},
		CanSee:    []int{1},
		CanEdit:   []int{1},
	}
	if err := handler.Authenticator.UserRepository.Upsert(user); err != nil {
		t.Fatal("could not insert user:", err)
	}
	token, err := handler.Authenticator.Encoder.Encode(user.ID)
	if err != nil {
		t.Fatal("could not fake token:", err)
	}
	bearer := fmt.Sprint("Bearer ", token)

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
			// User is not allwed to delete 2
			Query: "/api/papers/2",
			Code:  403,
		},
		{
			// 3 is not in the database
			Query: "/api/papers/3",
			Code:  404,
		},
	}

	for i, tt := range tts {
		req := httptest.NewRequest("DELETE", tt.Query, nil)
		req.Header.Add("Authorization", bearer)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != tt.Code {
			t.Errorf("%d - incorrect code: expected %d got %d", i, tt.Code, resp.Code)
		}

		if tt.Code >= 400 {
			continue
		}

		r := make(map[string]interface{})
		err := json.Unmarshal(resp.Body.Bytes(), &r)
		if err != nil {
			t.Error("%d - could not decode response as JSON:", i, err)
		}

		req = httptest.NewRequest("GET", tt.Query, nil)
		req.Header.Add("Authorization", bearer)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != 404 {
			t.Errorf("%d - seems like I can still GET %s: %v", i, tt.Query, resp.Body.String())
		}
	}
}

func TestPaperHandler_List(t *testing.T) {
	router, handlers, f := createRouter(t)
	handler := handlers.PaperHandler
	defer f()

	// Load fixtures
	papers := []*papernet.Paper{
		&papernet.Paper{ID: 1, Title: "Test", Summary: "Test"},
		&papernet.Paper{ID: 2, Title: "Pizza yolo", Summary: "Summary"},
		&papernet.Paper{ID: 3, Title: "I am the invisible", Summary: "Shhhht!"},
	}
	for _, paper := range papers {
		err := handler.Store.Upsert(paper)
		if err != nil {
			t.Fatal("could not insert paper:", err)
		}

		err = handler.Searcher.Index(paper)
		if err != nil {
			t.Fatal("could not index paper:", err)
		}
	}

	user := &papernet.User{
		ID:        "1",
		Name:      "Test user",
		Bookmarks: []int{1},
		CanSee:    []int{1, 2},
	}
	if err := handler.Authenticator.UserRepository.Upsert(user); err != nil {
		t.Fatal("could not insert user:", err)
	}
	token, err := handler.Authenticator.Encoder.Encode(user.ID)
	if err != nil {
		t.Fatal("could not fake token:", err)
	}
	bearer := fmt.Sprint("Bearer ", token)

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
		{
			// List all papers with title starting by piz
			Query: "/api/papers?q=piz&bookmarked=true",
			Code:  200,
			Len:   0,
		},
		{
			// List all papers with title starting by piz
			Query: "/api/papers?bookmarked=true",
			Code:  200,
			Len:   1,
		},
	}

	for _, tt := range tts {
		req := httptest.NewRequest("GET", tt.Query, nil)
		req.Header.Add("Authorization", bearer)

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
