package crawlers

import (
	"github.com/PuerkitoBio/goquery"

	"github.com/bobinette/papernet/etl"
)

func init() {
	register("html", HtmlCrawler{})
}

type HtmlCrawler struct{}

func (HtmlCrawler) Crawl(url string) ([]etl.Node, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}
	return []etl.Node{
		&GoQueryNode{sel: doc.Selection},
	}, nil
}

type GoQueryNode struct {
	sel *goquery.Selection
}

func (n *GoQueryNode) Find(selector string) etl.Node {
	return &GoQueryNode{sel: n.sel.Find(selector)}
}

func (n *GoQueryNode) Text() string { return n.sel.Text() }
