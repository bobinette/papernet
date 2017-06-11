package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	// "io/ioutil"
	"net/http"
	"time"

	"github.com/bobinette/papernet/imports"
)

type UserService interface {
	Token(int) (string, error)
}

type PaperService struct {
	userService UserService
	paperURL    string
	client      *http.Client
}

func NewPaperService(us UserService, paperURL string) *PaperService {
	return &PaperService{
		userService: us,
		paperURL:    paperURL,
		client:      &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *PaperService) Insert(userID int, paper *imports.Paper, ctx context.Context) error {
	token, err := s.userService.Token(userID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(paper)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", s.paperURL, bytes.NewReader(data))
	if err != nil {
		return nil
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	var p struct {
		Data imports.Paper `json:"data"`
	}
	err = json.NewDecoder(res.Body).Decode(&p)
	if err != nil {
		return err
	}

	paper.ID = p.Data.ID
	return nil
}
