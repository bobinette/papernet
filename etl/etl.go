package etl

import (
	"github.com/bobinette/papernet"
)

type Node interface {
	Find(string) Node
	Text() string
}

type Crawler interface {
	Get(string) (Node, error)
}

type Scraper interface {
	Scrap(Node) (papernet.Paper, error)
}

type Importer struct{}

func (i Importer) Import(resource string) (papernet.Paper, error) {
	// Only option for now
	crawler := HtmlCrawler{}
	node, err := crawler.Crawl(resource)
	if err != nil {
		return papernet.Paper{}, err
	}

	// Only option for now
	scraper := ArxivScraper{}
	return scraper.Scrap(node)
}
