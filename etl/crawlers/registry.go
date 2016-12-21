package crawlers

import (
	"github.com/bobinette/papernet/etl"
)

var registry = struct {
	r map[string]etl.Crawler
}{
	r: make(map[string]etl.Crawler),
}

func register(name string, crawler etl.Crawler) {
	registry.r[name] = crawler
}

func New(name string) (etl.Crawler, bool) {
	s, ok := registry.r[name]
	return s, ok
}
