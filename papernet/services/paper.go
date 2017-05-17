package services

import (
	"fmt"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/papernet"
	"github.com/bobinette/papernet/users"
)

func errPaperNotFound(id int) error {
	return errors.New(fmt.Sprintf("paper %d not found", id), errors.NotFound())
}

type UserService interface {
	CreatePaper(userID, paperID int) error
}

type PaperService struct {
	repository papernet.PaperRepository
	index      papernet.PaperIndex

	userService UserService
	tagService  *TagService
}

func NewPaperService(
	repo papernet.PaperRepository,
	index papernet.PaperIndex,
	us UserService,
	ts *TagService,
) *PaperService {
	return &PaperService{
		repository: repo,
		index:      index,

		userService: us,
		tagService:  ts,
	}
}

func (s *PaperService) Get(user users.User, id int) (papernet.Paper, error) {
	if err := aclCanSee(user, id); err != nil {
		return papernet.Paper{}, err
	}

	papers, err := s.repository.Get(id)
	if err != nil {
		return papernet.Paper{}, err
	} else if len(papers) != 1 {
		return papernet.Paper{}, errPaperNotFound(id)
	}

	return papers[0], nil
}

type SearchResults struct {
	Papers     []papernet.Paper    `json:"papers"`
	Facets     papernet.Facets     `json:"facets"`
	Pagination papernet.Pagination `json:"pagination"`
}

func (s *PaperService) Search(user users.User, q string, tags []string, bookmarked bool, offset, limit int) (SearchResults, error) {
	sp := papernet.SearchParams{
		IDs:  user.CanSee,
		Q:    q,
		Tags: tags,

		Offset: uint64(offset),
		Limit:  uint64(limit),
	}
	if bookmarked {
		sp.IDs = user.Bookmarks
	}

	if sp.Limit <= 0 {
		sp.Limit = 20
	}

	res, err := s.index.Search(sp)
	if err != nil {
		return SearchResults{}, err
	}

	papers, err := s.repository.Get(res.IDs...)
	if err != nil {
		return SearchResults{}, err
	}

	return SearchResults{
		Papers:     papers,
		Facets:     res.Facets,
		Pagination: res.Pagination,
	}, nil
}

func (s *PaperService) Create(callerID int, paper papernet.Paper) (papernet.Paper, error) {
	if paper.ID != 0 {
		return papernet.Paper{}, errors.New("id already set", errors.BadRequest())
	}

	err := s.repository.Upsert(&paper)
	if err != nil {
		return papernet.Paper{}, err
	}

	err = s.userService.CreatePaper(callerID, paper.ID)
	if err != nil {
		return papernet.Paper{}, err
	}

	err = s.index.Index(&paper)
	if err != nil {
		return papernet.Paper{}, err
	}

	return paper, nil
}

func (s *PaperService) Update(user users.User, paper papernet.Paper) (papernet.Paper, error) {
	if paper.ID == 0 {
		return papernet.Paper{}, errors.New("id already set", errors.BadRequest())
	}

	err := aclCanEdit(user, paper.ID)
	if err != nil {
		return papernet.Paper{}, err
	}

	err = s.repository.Upsert(&paper)
	if err != nil {
		return papernet.Paper{}, err
	}

	err = s.index.Index(&paper)
	if err != nil {
		return papernet.Paper{}, err
	}

	return paper, nil
}

func (s *PaperService) Delete(user users.User, paperID int) error {
	err := aclCanDelete(user, paperID)
	if err != nil {
		return err
	}

	err = s.repository.Delete(paperID)
	if err != nil {
		return err
	}

	err = s.index.Delete(paperID)
	if err != nil {
		return err
	}

	return nil
}

func aclCanSee(user users.User, paperID int) error {
	if !contains(paperID, user.CanSee) {
		return errPaperNotFound(paperID)
	}
	return nil
}

func aclCanEdit(user users.User, paperID int) error {
	// CanEdit < CanSee
	err := aclCanSee(user, paperID)
	if err != nil {
		return err
	}

	if !contains(paperID, user.CanEdit) {
		return errors.New("you do not have edit permission", errors.Forbidden())
	}
	return nil
}

func aclCanDelete(user users.User, paperID int) error {
	// Owns < CanSee
	err := aclCanSee(user, paperID)
	if err != nil {
		return err
	}

	if !contains(paperID, user.Owns) {
		return errors.New("only the owner can delete a paper", errors.Forbidden())
	}
	return nil
}

func contains(v int, a []int) bool {
	for _, i := range a {
		if i == v {
			return true
		}
	}
	return false
}
