package services

import (
	"context"
	"fmt"

	"github.com/bobinette/papernet/errors"
	"github.com/bobinette/papernet/imports"
)

type ImportService struct {
	repository imports.PaperRepository
	importers  []imports.Importer
}

func NewImportService(repository imports.PaperRepository, importers ...imports.Importer) *ImportService {
	return &ImportService{
		repository: repository,
		importers:  importers,
	}
}

func (s *ImportService) Sources() []string {
	sources := make([]string, len(s.importers))
	for i, importer := range s.importers {
		sources[i] = importer.Source()
	}

	return sources
}

func (s *ImportService) Search(
	userID int,
	q string,
	limit int,
	offset int,
	sources []string,
	ctx context.Context,
) (map[string]imports.SearchResults, error) {
	// Select the importers
	var importers []imports.Importer
	if len(sources) != 0 {
		importers = make([]imports.Importer, len(sources))
		for i, source := range sources {
			found := false
			for _, importer := range s.importers {
				if importer.Source() == source {
					importers[i] = importer
					found = true
					break
				}
			}

			if !found {
				return nil, errors.New(fmt.Sprintf("unknown source %s", source), errors.BadRequest())
			}
		}
	} else {
		importers = s.importers
	}

	res := make(map[string]imports.SearchResults)
	for _, importer := range importers {
		// Use the importer to fetch the data
		r, err := importer.Search(q, limit, offset, ctx)
		if err != nil {
			return nil, err
		}

		// Get the ids of each entry from the repository
		for i, paper := range r.Papers {
			id, err := s.repository.Get(userID, importer.Source(), paper.Reference)
			if err != nil {
				return nil, err
			}

			paper.ID = id
			r.Papers[i] = paper
		}

		res[importer.Source()] = r
	}

	return res, nil
}

func isIn(str string, a []string) bool {
	for _, s := range a {
		if s == str {
			return true
		}
	}
	return false
}
