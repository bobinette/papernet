package services

import (
	"github.com/bobinette/papernet/papernet"
)

type TagService struct {
	index papernet.TagIndex
}

func NewTagService(index papernet.TagIndex) *TagService {
	return &TagService{
		index: index,
	}
}

func (s *TagService) Search(q string) ([]string, error) {
	return s.index.Search(q)
}
