package etl

import (
	"github.com/PuerkitoBio/goquery"
)

type HtmlCrawler struct{}

func (HtmlCrawler) Crawl(url string) (Node, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}
	return &GoQueryNode{sel: doc.Selection}, nil
}

type GoQueryNode struct {
	sel *goquery.Selection
}

func (n *GoQueryNode) Find(selector string) Node {
	return &GoQueryNode{sel: n.sel.Find(selector)}
}

func (n *GoQueryNode) Text() string { return n.sel.Text() }
