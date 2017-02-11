package web

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"net/http/httptest"
// 	"os"
// 	"path"
// 	"reflect"
// 	"testing"

// 	"github.com/gin-gonic/gin"

// 	"github.com/bobinette/papernet"
// 	"github.com/bobinette/papernet/auth"
// 	"github.com/bobinette/papernet/bleve"
// 	"github.com/bobinette/papernet/bolt"
// )

// type handlers struct {
// 	PaperHandler *PaperHandler
// 	UserHandler  *UserHandler
// }

// func createRouter(t *testing.T) (*gin.Engine, handlers, func()) {
// 	dir, err := ioutil.TempDir("", "")
// 	if err != nil {
// 		t.Fatal("could not create tmp file:", err)
// 	}

// 	index := &bleve.PaperIndex{}
// 	err = index.Open(path.Join(dir, "test", "index"))
// 	if err != nil {
// 		t.Fatal("error creating index", err)
// 	}

// 	driver := &bolt.Driver{}
// 	err = driver.Open(path.Join(dir, "test", "db"))
// 	if err != nil {
// 		t.Fatal("error creating index", err)
// 	}

// 	encoder := auth.Encoder{Key: "test"}

// 	paperStore := &bolt.PaperStore{Driver: driver}
// 	userRepo := &bolt.UserRepository{Driver: driver}
// 	tagIndex := &bolt.TagIndex{Driver: driver}

// 	paperHandler := &PaperHandler{
// 		Store:          paperStore,
// 		Searcher:       index,
// 		TagIndex:       tagIndex,
// 		UserRepository: userRepo,
// 		Authenticator: Authenticator{
// 			UserRepository: userRepo,
// 			Encoder:        encoder,
// 		},
// 	}

// 	userHandler := &UserHandler{
// 		Authenticator: Authenticator{
// 			UserRepository: userRepo,
// 			Encoder:        encoder,
// 		},
// 		Repository: userRepo,
// 	}

// 	gin.SetMode(gin.ReleaseMode) // avoid unnecessary log
// 	router := gin.New()
// 	paperHandler.RegisterRoutes(router)
// 	userHandler.RegisterRoutes(router)

// 	return router, handlers{
// 			PaperHandler: paperHandler,
// 			UserHandler:  userHandler,
// 		}, func() {
// 			if err := index.Close(); err != nil {
// 				t.Log(err)
// 			}
// 			if err := os.RemoveAll(dir); err != nil {
// 				t.Log(err)
// 			}
// 		}
// }

// func createReader(i interface{}, t *testing.T) io.Reader {
// 	data, err := json.Marshal(i)
// 	if err != nil {
// 		t.Fatal("cannot unmarshal:", err)
// 	}

// 	buf := bytes.Buffer{}
// 	_, err = buf.Write(data)
// 	if err != nil {
// 		t.Fatal("cannot write:", err)
// 	}

// 	return &buf
// }

// func TestUserHandler_Me(t *testing.T) {
// 	t.Skip("need bleve driver")
// 	router, handlers, f := createRouter(t)
// 	handler := handlers.UserHandler
// 	defer f()

// 	// Load fixtures
// 	user := &papernet.User{
// 		ID:        "1",
// 		Name:      "Test user",
// 		Bookmarks: []int{1},
// 	}
// 	err := handler.Repository.Upsert(user)
// 	if err != nil {
// 		t.Fatal("could not insert user:", err)
// 	}

// 	userToken, err := handler.Authenticator.Encoder.Encode(user.ID)
// 	if err != nil {
// 		t.Fatal("could not encode token:", err)
// 	}

// 	var tts = []struct {
// 		Token string
// 		Code  int
// 	}{
// 		{
// 			Token: "",
// 			Code:  401,
// 		},
// 		{
// 			Token: "not a bearer",
// 			Code:  401,
// 		},
// 		{
// 			Token: "bearer not.a.token",
// 			Code:  401,
// 		},
// 		{
// 			Token: fmt.Sprintf("bearer %s", userToken),
// 			Code:  200,
// 		},
// 	}

// 	for i, tt := range tts {
// 		req := httptest.NewRequest("GET", "/api/me", nil)
// 		req.Header.Add("authorization", tt.Token)
// 		resp := httptest.NewRecorder()
// 		router.ServeHTTP(resp, req)
// 		if resp.Code != tt.Code {
// 			t.Errorf("%d - incorrect code: expected %d got %d (body: %v)", i, tt.Code, resp.Code, resp.Body.String())
// 		}
// 	}
// }

// func TestUserHandler_UpdateBookmarks(t *testing.T) {
// 	t.Skip("need bleve driver")
// 	router, handlers, f := createRouter(t)
// 	handler := handlers.UserHandler
// 	defer f()

// 	// Load fixtures
// 	user := &papernet.User{
// 		ID:        "1",
// 		Name:      "Test user",
// 		Bookmarks: []int{1},
// 	}
// 	err := handler.Repository.Upsert(user)
// 	if err != nil {
// 		t.Fatal("could not insert user:", err)
// 	}

// 	userToken, err := handler.Authenticator.Encoder.Encode(user.ID)
// 	if err != nil {
// 		t.Fatal("could not encode token:", err)
// 	}

// 	var tts = []struct {
// 		Token    string
// 		Add      []int
// 		Remove   []int
// 		Code     int
// 		Expected []int
// 	}{
// 		{
// 			Token:    "",
// 			Add:      []int{},
// 			Remove:   []int{},
// 			Code:     401,
// 			Expected: []int{},
// 		},
// 		{
// 			Token:    fmt.Sprintf("bearer %s", userToken),
// 			Add:      []int{2},
// 			Remove:   []int{},
// 			Code:     200,
// 			Expected: []int{1, 2},
// 		},
// 		{
// 			Token:    fmt.Sprintf("bearer %s", userToken),
// 			Add:      []int{3},
// 			Remove:   []int{2},
// 			Code:     200,
// 			Expected: []int{1, 3},
// 		},
// 		{
// 			Token:    fmt.Sprintf("bearer %s", userToken),
// 			Add:      []int{2},
// 			Remove:   []int{2, 3},
// 			Code:     200,
// 			Expected: []int{1},
// 		},
// 	}

// 	for i, tt := range tts {
// 		reader := createReader(map[string]interface{}{
// 			"add":    tt.Add,
// 			"remove": tt.Remove,
// 		}, t)
// 		req := httptest.NewRequest("POST", "/api/bookmarks", reader)
// 		req.Header.Add("authorization", tt.Token)
// 		resp := httptest.NewRecorder()
// 		router.ServeHTTP(resp, req)
// 		if resp.Code != tt.Code {
// 			t.Errorf("%d - incorrect code: expected %d got %d (body: %v)", i, tt.Code, resp.Code, resp.Body.String())
// 			continue
// 		}

// 		if tt.Code >= 400 {
// 			continue
// 		}

// 		var r struct {
// 			Data papernet.User `json:"data"`
// 		}
// 		if err := json.Unmarshal(resp.Body.Bytes(), &r); err != nil {
// 			t.Errorf("%d - could not read json body: %v", i, err)
// 		} else if !reflect.DeepEqual(tt.Expected, r.Data.Bookmarks) {
// 			t.Errorf("%d - incorrect bookmarks: expected %v got %v", i, tt.Expected, r.Data.Bookmarks)
// 		}
// 	}
// }
