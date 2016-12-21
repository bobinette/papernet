package scrapers

import (
	"testing"

	"github.com/bobinette/papernet/etl"
)

type MapNode struct {
	Children map[string]*MapNode
	TextVal  string
}

func (m MapNode) Find(key string) etl.Node {
	return m.Children[key]
}

func (m MapNode) Text() string { return m.TextVal }

func TestArxiv_Scrap(t *testing.T) {
	tts := []struct {
		Name             string
		Title            string
		Abstract         string
		ExpectedTitle    string
		ExpectedAbstract string
	}{
		{
			Name:             "Basic",
			Title:            "Title",
			Abstract:         "Abstract",
			ExpectedTitle:    "Title",
			ExpectedAbstract: "Abstract",
		},
		{
			Name:             "With prefix",
			Title:            "Title: Title",
			Abstract:         "Abstract: Abstract",
			ExpectedTitle:    "Title",
			ExpectedAbstract: "Abstract",
		},
		{
			Name:             "With new lines",
			Title:            "Title: Title on\n2 lines",
			Abstract:         "Abstract: Abstract\non\n3 lines",
			ExpectedTitle:    "Title on 2 lines",
			ExpectedAbstract: "Abstract on 3 lines",
		},
	}

	for _, tt := range tts {
		node := MapNode{
			Children: map[string]*MapNode{
				".title":    &MapNode{TextVal: tt.Title},
				".abstract": &MapNode{TextVal: tt.Abstract},
			},
			TextVal: "",
		}

		scraper := ArxivScraper{}
		paper, err := scraper.Scrap(node)
		if err != nil {
			t.Fatal(err.Error())
		}

		if paper.Title != tt.ExpectedTitle {
			t.Errorf("Incorrect title: expected %s got %s", tt.ExpectedTitle, paper.Title)
		}

		if paper.Summary != tt.ExpectedAbstract {
			t.Errorf("Incorrect abstract: expected %s got %s", tt.ExpectedAbstract, paper.Summary)
		}
	}
}
