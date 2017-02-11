package papernet

import (
	"context"
	"net/http"
)

type Request struct {
	*http.Request

	Params map[string]string
}

func (r *Request) Query(k string) string {
	return r.Request.URL.Query().Get(k)
}

func (r *Request) Param(k string) string {
	return r.Params[k]
}

func (r *Request) WithContext(ctx context.Context) *Request {
	return &Request{
		Request: r.Request.WithContext(ctx),
		Params:  r.Params,
	}
}

// HandlerFunc defines the signature of a web endpoint
type HandlerFunc func(*Request) (interface{}, error)

type Route struct {
	// Route defines the url of the end point. It can contain parameters,
	// the form depending on the framework used for implementation
	Route string

	// Method of the endpoint. Typically GET, PUT, POST, DELETE, etc.
	Method string

	// Renderer defines the renderer used to marshal the first returned
	// value of the handler. Typically JSON, Text, etc.
	Renderer string

	// Authenticated should be set to true so the server activates
	// authentication for that route, i.e. loads the user that will
	// be available in the context.
	Authenticated bool

	// HandlerFunc is the function used to handle incoming requests on
	// that route.
	HandlerFunc HandlerFunc
}

type Server interface {
	http.Handler
	Register(Route) error
}
