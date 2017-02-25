package papernet

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bobinette/papernet/errors"
)

var (
	arxivURL         = "http://export.arxiv.org/api/query"
	arxivSummaryPipe = CleaningPipe(
		strings.TrimSpace,
		OneLine,
		strings.TrimSpace,
	)
	arxivIDRegExp     *regexp.Regexp
	arxivImportRegExp *regexp.Regexp
)

func init() {
	arxivIDRegExp = regexp.MustCompile("http://arxiv.org/abs/([0-9.]*)(v[0-9]+)?")
	arxivImportRegExp = regexp.MustCompile("https?://arxiv.org/(abs|pdf)/([0-9.]*)(v[0-9]+)?")

	// Check if arxiv URL is valid
	_, err := url.Parse(arxivURL)
	if err != nil {
		log.Fatal(err)
	}
}

type ArxivSearch struct {
	Q          string
	IDs        []string
	Start      int
	MaxResults int
}

type ArxivResult struct {
	Papers     []*Paper
	Pagination Pagination
}

type ArxivSpider struct {
	Client *http.Client
}

func (s *ArxivSpider) Search(search ArxivSearch) (ArxivResult, error) {
	// No need to check for error
	u, _ := url.Parse(arxivURL)
	query := u.Query()

	if search.Q != "" {
		re, _ := regexp.Compile("[A-Za-z0-9]+")
		matches := re.FindAllStringSubmatch(search.Q, -1)
		q := make([]string, len(matches))
		for i, match := range matches {
			q[i] = match[0]
		}
		query.Add("search_query", fmt.Sprintf("all:%s", strings.Join(q, " AND ")))
	}
	if len(search.IDs) > 0 {
		query.Add("id_list", strings.Join(search.IDs, ","))
	}
	if search.Start > 0 {
		query.Add("start", strconv.Itoa(search.Start))
	}
	if search.MaxResults > 0 {
		query.Add("max_results", strconv.Itoa(search.MaxResults))
	}

	query.Add("sortBy", "submittedDate")
	query.Add("sortOrder", "descending")

	u.RawQuery = query.Encode()

	if s.Client == nil {
		s.Client = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := s.Client.Get(u.String())
	if err != nil {
		return ArxivResult{}, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ArxivResult{}, err
	}

	r := struct {
		Title string `xml:"title"`
		ID    string `xml:"id"`
		Total struct {
			Value uint64 `xml:",chardata"`
		} `xml:"totalResults"`
		Offset struct {
			Value uint64 `xml:",chardata"`
		} `xml:"startIndex"`
		Limit struct {
			Value uint64 `xml:",chardata"`
		} `xml:"itemsPerPage"`
		Entries []struct {
			Title   string `xml:"title"`
			ID      string `xml:"id"`
			Summary string `xml:"summary"`
			Links   []struct {
				HRef string `xml:"href,attr"`
				Type string `xml:"type,attr"`
			} `xml:"link"`
			Categories []struct {
				Term string `xml:"term,attr"`
			} `xml:"category"`
			Published time.Time `xml:"published"`
			Updated   time.Time `xml:"updated"`
		} `xml:"entry"`
	}{}
	err = xml.Unmarshal(data, &r)
	if err != nil {
		return ArxivResult{}, err
	}

	papers := make([]*Paper, len(r.Entries))
	for i, entry := range r.Entries {
		tags := make([]string, 0, len(entry.Categories))
		for _, cat := range entry.Categories {
			tag, ok := arxivCategories[cat.Term]
			if ok {
				tags = append(tags, tag)
			}
		}

		var arxivID string
		matches := arxivIDRegExp.FindAllStringSubmatch(entry.ID, -1)
		if len(matches) > 0 && len(matches[0]) > 1 {
			arxivID = matches[0][1]
		}

		papers[i] = &Paper{
			Title:   entry.Title,
			Summary: arxivSummaryPipe(entry.Summary),
			References: []string{
				entry.Links[0].HRef, // link to arXiv
				entry.Links[1].HRef, // PDF
			},
			Tags:      tags,
			CreatedAt: entry.Published,
			UpdatedAt: entry.Updated,
			ArxivID:   arxivID,
		}
	}

	return ArxivResult{
		Papers: papers,
		Pagination: Pagination{
			Total:  r.Total.Value,
			Limit:  r.Limit.Value,
			Offset: r.Offset.Value,
		},
	}, nil
}

func (s *ArxivSpider) Import(url string) (*Paper, error) {
	matches := arxivImportRegExp.FindAllStringSubmatch(url, -1)
	if len(matches) == 0 || len(matches[0]) < 4 || matches[0][2] == "" {
		return nil, errors.New(fmt.Sprintf("could not extract arxiv ID from %s", url), errors.WithCode(http.StatusNotFound))
	}

	arxivID := matches[0][2]
	res, err := s.Search(ArxivSearch{
		IDs: []string{matches[0][2]},
	})
	if err != nil {
		return nil, err
	}

	if len(res.Papers) == 0 {
		return nil, errors.New(fmt.Sprintf("no paper found for id %s", arxivID), errors.WithCode(http.StatusNotFound))
	}

	return res.Papers[0], nil
}
