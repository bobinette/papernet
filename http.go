package papernet

import (
	"context"
	"io"
	"net/http"
)

type Request interface {
	Header(string) string
	Param(string) string
	Query(string) string
	Body() io.ReadCloser

	Context() context.Context
	WithContext(context.Context) Request
}

// HandlerFunc defines the signature of a web endpoint
type HandlerFunc func(Request) (interface{}, error)

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
