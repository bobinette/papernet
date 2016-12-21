package scrapers

import (
	"github.com/bobinette/papernet/etl"
)

var registry = struct {
	r map[string]etl.Scraper
}{
	r: make(map[string]etl.Scraper),
}

func register(name string, scraper etl.Scraper) {
	registry.r[name] = scraper
}

func New(name string) (etl.Scraper, bool) {
	s, ok := registry.r[name]
	return s, ok
}
