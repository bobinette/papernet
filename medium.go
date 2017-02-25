package papernet

import (
	"encoding/json"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

var (
	mediumContentRegex *regexp.Regexp
)

func init() {
	mediumContentRegex = regexp.MustCompile(`\/\/ <!\[CDATA\[ window\["obvInit"\]\((.*)\) \/\/ \]\]>`)
}

type MediumImporter struct {
}

type mediumPost struct {
	Value struct {
		ID      string `json:"id"`
		Creator struct {
			Name string `json:"name"`
		} `json:"creator"`
		Title   string `json:"title"`
		Content struct {
			BodyModel struct {
				Paragraphs []struct {
					Type int    `json:"type"`
					Text string `json:"text"`
				} `json:"paragraphs"`
			} `json:"bodyModel"`
		} `json:"content"`
		Virtuals struct {
			Tags []struct {
				Name string `json:"name"`
			} `json:"tags"`
		} `json:"virtuals"`
	} `json:"value"`
}

func (MediumImporter) Import(addr string) (*Paper, error) {
	doc, err := goquery.NewDocument(addr)
	if err != nil {
		return nil, err
	}

	var post mediumPost
	// There is only html script tag matching the regexp
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		content := OneLine(s.Text())
		matches := mediumContentRegex.FindAllStringSubmatch(content, -1)
		if len(matches) == 0 {
			return
		}

		contentJSON := matches[0][1]
		err = json.Unmarshal([]byte(contentJSON), &post)
		return
	})

	if err != nil {
		return nil, err
	}

	// We consider the abstract to be the first paragraph of type 1. No need to import
	// the whole content of the blog post, that is not the goal of papernet
	var abstract string
	for _, paragraph := range post.Value.Content.BodyModel.Paragraphs {
		if paragraph.Type == 1 {
			abstract = paragraph.Text
			break
		}
	}

	tags := make([]string, len(post.Value.Virtuals.Tags))
	for i, tag := range post.Value.Virtuals.Tags {
		tags[i] = tag.Name
	}

	paper := Paper{
		Title:   post.Value.Title,
		Summary: abstract,
		Tags:    tags,
		Authors: []string{post.Value.Creator.Name},
	}

	return &paper, nil
}
