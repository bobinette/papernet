package scrapers

import (
	"strings"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/etl"
)

func init() {
	register("arxiv", ArxivScraper{})
}

type ArxivScraper struct{}

func (ArxivScraper) Scrap(node etl.Node) (papernet.Paper, error) {
	title := CleanString(
		node.Find(".title").Text(),
		strings.TrimSpace,
		OneLine,
		RemovePrefix("Title:"),
		strings.TrimSpace,
	)

	summary := CleanString(
		node.Find(".abstract").Text(),
		strings.TrimSpace,
		OneLine,
		RemovePrefix("Abstract:"),
		strings.TrimSpace,
	)

	paper := papernet.Paper{
		Title:   title,
		Summary: summary,
	}
	return paper, nil
}
