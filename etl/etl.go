package etl

import (
	"github.com/bobinette/papernet"
)

type Node interface {
	Find(string) Node
	Text() string
}

type Crawler interface {
	Crawl(string) ([]Node, error)
}

type Scraper interface {
	Scrap(Node) (papernet.Paper, error)
}

type Importer struct{}

func (i Importer) Import(resource string, crawler Crawler, scraper Scraper) ([]papernet.Paper, error) {
	nodes, err := crawler.Crawl(resource)
	if err != nil {
		return nil, err
	}

	// Only option for now
	papers := make([]papernet.Paper, len(nodes))
	for i, node := range nodes {
		p, err := scraper.Scrap(node)
		if err != nil {
			return nil, err
		}

		papers[i] = p
	}
	return papers, nil
}
