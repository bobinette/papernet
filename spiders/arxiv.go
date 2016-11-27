package spiders

import (
	"log"

	"github.com/PuerkitoBio/goquery"

	"github.com/bobinette/papernet"
)

type Arxiv struct{}

func (s Arxiv) Get(url string) (papernet.Paper, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	paper := papernet.Paper{
		Title:   doc.Find(".leftcolumn .title").Text(),
		Summary: doc.Find(".abstract").Text(),
	}

	return paper, nil
}
