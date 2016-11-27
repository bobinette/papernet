package etl

import (
	"strings"

	"github.com/bobinette/papernet"
)

type ArxivScraper struct{}

func (ArxivScraper) Scrap(node Node) (papernet.Paper, error) {
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
