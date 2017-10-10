package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/bobinette/papernet/cron"
	"github.com/bobinette/papernet/errors"
)

type Cron struct {
	ID uint

	UserID  int
	Query   string
	Sources string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func newCron(c cron.Cron) Cron {
	return Cron{
		ID:      c.ID,
		UserID:  c.UserID,
		Query:   c.Q,
		Sources: strings.Join(c.Sources, ","),
	}
}

func (c Cron) format() cron.Cron {
	return cron.Cron{
		ID:      c.ID,
		UserID:  c.UserID,
		Q:       c.Query,
		Sources: strings.Split(c.Sources, ","),
	}
}

type dbPaper cron.Paper

func (d *dbPaper) Value() (driver.Value, error) {
	data, err := json.Marshal(d)
	return string(data), err
}

func (d *dbPaper) Scan(input interface{}) error {
	switch input := input.(type) {
	case string:
		if input == "" {
			return nil
		}
		return json.Unmarshal([]byte(input), d)
	case []byte:
		if len(input) == 0 {
			return nil
		}
		return json.Unmarshal(input, d)
	}
	return errors.New("not supported")
}

type SearchResult struct {
	ID uint

	CronID uint
	Source string

	Result *dbPaper

	CreatedAt time.Time
}

func newSearchResult(cronID uint, paper cron.Paper) SearchResult {
	dbp := dbPaper(paper)
	return SearchResult{
		CronID: cronID,
		Source: paper.Source,

		Result: &dbp,

		CreatedAt: paper.CreatedAt,
	}
}
