package arxiv

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bobinette/papernet/imports"
)

var (
	apiURLStr   = "http://export.arxiv.org/api/query"
	summaryPipe = CleaningPipe(
		strings.TrimSpace,
		OneLine,
		strings.TrimSpace,
	)

	refRegexp *regexp.Regexp
)

func init() {
	refRegexp = regexp.MustCompile("http://arxiv.org/abs/([0-9.]*)(v[0-9]+)?")

	// Check if arxiv URL is valid
	_, err := url.Parse(apiURLStr)
	if err != nil {
		panic(err)
	}
}

type responseEntry struct {
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
}

type response struct {
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
	Entries []responseEntry `xml:"entry"`
}

type Importer struct {
	client *http.Client
	source string
}

func NewSearcher() *Importer {
	return &Importer{
		client: &http.Client{Timeout: 20 * time.Second},
		source: "arxiv",
	}
}

func (i *Importer) Source() string { return i.source }

func (i *Importer) Import(ctx context.Context, ref string) (imports.Paper, error) {
	u := craftRefURL(ref)
	req, err := http.NewRequest("GET", u.String(), nil)
	req = req.WithContext(ctx)
	resp, err := i.client.Do(req)
	if err != nil {
		return imports.Paper{}, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return imports.Paper{}, err
	}

	var r response
	err = xml.Unmarshal(data, &r)
	if err != nil {
		return imports.Paper{}, err
	}

	papers := i.parsePapers(r)
	if len(papers) == 0 {
		return imports.Paper{}, imports.ErrNotFound
	}
	return papers[0], nil
}

func (i *Importer) Search(ctx context.Context, q string, limit, offset int) (imports.SearchResults, error) {
	u := craftURL(q, limit, offset)
	req, err := http.NewRequest("GET", u.String(), nil)
	req = req.WithContext(ctx)
	resp, err := i.client.Do(req)
	if err != nil {
		return imports.SearchResults{}, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return imports.SearchResults{}, err
	}

	var r response
	err = xml.Unmarshal(data, &r)
	if err != nil {
		return imports.SearchResults{}, err
	}

	papers := i.parsePapers(r)
	return imports.SearchResults{
		Papers: papers,
		Pagination: imports.Pagination{
			Total:  uint(r.Total.Value),
			Limit:  uint(r.Limit.Value),
			Offset: uint(r.Offset.Value),
		},
	}, nil
}

func craftURL(q string, limit, offset int) *url.URL {
	// No need to check for error, done in the init
	u, _ := url.Parse(apiURLStr)
	query := u.Query()

	if q != "" {
		re, _ := regexp.Compile("[A-Za-z0-9]+")
		matches := re.FindAllStringSubmatch(q, -1)
		qs := make([]string, len(matches))
		for i, match := range matches {
			qs[i] = match[0]
		}
		query.Add("search_query", fmt.Sprintf("all:%s", strings.Join(qs, " AND ")))
	}

	query.Add("start", strconv.Itoa(offset))
	query.Add("max_results", strconv.Itoa(limit))
	query.Add("sortBy", "submittedDate")
	query.Add("sortOrder", "descending")

	u.RawQuery = query.Encode()
	return u
}

func craftRefURL(ref string) *url.URL {
	// No need to check for error, done in the init
	u, _ := url.Parse(apiURLStr)
	query := u.Query()

	query.Add("id_list", ref)

	u.RawQuery = query.Encode()
	return u
}

func (i *Importer) parsePapers(r response) []imports.Paper {
	papers := make([]imports.Paper, len(r.Entries))
	for n, entry := range r.Entries {
		tags := make([]string, 0, len(entry.Categories))
		for _, cat := range entry.Categories {
			tag, ok := categories[cat.Term]
			if ok {
				tags = append(tags, tag)
			}
		}

		papers[n] = imports.Paper{
			Source:    i.source,
			Reference: extractReference(entry.ID),

			Title:   entry.Title,
			Summary: summaryPipe(entry.Summary),
			Tags:    tags,
			// @TODO: authors
			References: []string{
				entry.Links[0].HRef, // link to arXiv
				entry.Links[1].HRef, // PDF
			},
		}
	}

	return papers
}

func extractReference(id string) string {
	ref := id

	urlParts := strings.Split(ref, "/")
	ref = strings.Split(urlParts[len(urlParts)-1], "v")[0] // Cut out version

	return ref
}
