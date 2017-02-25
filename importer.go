package papernet

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/bobinette/papernet/errors"
)

type Importer interface {
	Import(string) (*Paper, error)
}

type ImporterRegistry map[string]Importer

func (reg ImporterRegistry) Register(host string, imp Importer) {
	reg[host] = imp
}

func (reg ImporterRegistry) Import(addr string) (*Paper, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	imp, ok := reg[u.Host]
	if !ok {
		return nil, errors.New(fmt.Sprintf("no importer registered for host %s", addr), errors.WithCode(http.StatusBadRequest))
	}

	return imp.Import(addr)
}
