package papernet

import (
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type chromeBookmarkLink struct {
	Title string
	URL   string
}

type chromeBookmarkPaper struct {
	Title string
	Links []chromeBookmarkLink
}

func ImportChromeBookmarks(r io.Reader) ([]Paper, error) {
	cbps, err := importChromeBookmarks(r)
	if err != nil {
		return nil, err
	}

	papers := make([]Paper, len(cbps))
	for i, cbp := range cbps {
		papers[i] = formatChromeBookmarks(cbp)
	}

	return papers, nil
}

func importChromeBookmarks(r io.Reader) ([]chromeBookmarkPaper, error) {
	z := html.NewTokenizer(r)

	title := ""
	var papers []chromeBookmarkPaper
	var err error
Loop:
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			err = z.Err()
			break Loop
		case html.StartTagToken:
			token := z.Token()
			if token.DataAtom == atom.H1 {
				next := z.Next()
				if next != html.TextToken {
					return nil, fmt.Errorf("expected text token, got %v", next)
				}

				textToken := z.Token()
				title = textToken.Data
			}

			if token.DataAtom == atom.Dl {
				papers, err = visitDl(title, z)
				if err != nil {
					break Loop
				}
			}
		}
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	return papers, nil
}

func visitDl(title string, z *html.Tokenizer) ([]chromeBookmarkPaper, error) {
	paper := chromeBookmarkPaper{
		Title: title,
	}
	papers := make([]chromeBookmarkPaper, 0)

Loop:
	for {
		tt := z.Next()
		tok := z.Token()
		switch tt {
		case html.ErrorToken:
			return nil, z.Err()
		case html.EndTagToken:
			if tok.DataAtom == atom.Dl {
				break Loop
			}
		case html.StartTagToken:
			if tok.DataAtom != atom.Dt && tok.DataAtom != atom.Dl {
				break
			}

			nextTT := z.Next()
			nextToken := z.Token()
			switch nextTT {
			case html.ErrorToken:
				return nil, z.Err()
			case html.StartTagToken:
				switch nextToken.DataAtom {
				case atom.A:
					link, err := createLink(nextToken, z)
					if err != nil {
						return nil, err
					}
					paper.Links = append(paper.Links, link)
				case atom.H3:
					h3TT := z.Next()
					if h3TT != html.TextToken {
						return nil, fmt.Errorf("expected text token, got %v", h3TT)
					}

					subTitle := z.Token().Data
					z.Next()

					ps, err := visitDl(subTitle, z)
					if err != nil {
						return nil, err
					}
					papers = append(papers, ps...)
				}
			}
		}
	}

	if len(paper.Links) > 0 {
		papers = append([]chromeBookmarkPaper{paper}, papers...)
	}

	return papers, nil
}

func createLink(a html.Token, z *html.Tokenizer) (chromeBookmarkLink, error) {
	link := chromeBookmarkLink{}

	for _, attr := range a.Attr {
		if attr.Key == "href" {
			link.URL = attr.Val
		}
	}

	next := z.Next()
	if next != html.TextToken {
		return chromeBookmarkLink{}, fmt.Errorf("expected text token, got %v", next)
	}

	textToken := z.Token()
	link.Title = textToken.Data

	return link, nil
}

func formatChromeBookmarks(cbp chromeBookmarkPaper) Paper {
	paper := Paper{
		Title:     cbp.Title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	linkFormat := "# %s\n%s"
	links := make([]string, len(cbp.Links))
	for i, link := range cbp.Links {
		links[i] = fmt.Sprintf(linkFormat, link.Title, link.URL)
	}

	paper.Summary = strings.Join(links, "\n\n")

	return paper
}
