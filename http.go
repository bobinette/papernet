package papernet

import (
	"context"
	"io"
)

// Request defines the basic interface that the handlers
// need to serve an HTTP request
type Request interface {
	// Param returns the value of the URI parameter for a given name
	Param(string) string

	// Query returns the value of the first query parameter for a
	// given name. If you need the list of parameters use QueryAll
	Query(string) string

	// QueryAll returns all the values found for the given name in
	// the query parameters
	QueryAll(string) []string

	// Body returns an io.ReadCloser that can be used to read the
	// body. Do not forget to close it once you have finished
	// reading
	Body() io.ReadCloser

	// Context returns the context of the request. Typically, if the
	// route requires authentication, the user should be found under
	// "user"
	Context() context.Context
}

// HandlerFunc defines the signature of a web endpoint
type HandlerFunc func(Request) (interface{}, error)

type Route struct {
	// Method of the endpoint. Typically GET, PUT, POST, DELETE, etc.
	Method string

	// URL define the url of the end point. It can contain parameters,
	// the form depending on the framework used for implementation
	URL string

	// HandlerFunc is the function used to handle incoming requests on
	// that route.
	HandlerFunc HandlerFunc

	// Renderer defines the renderer used to marshal the first returned
	// value of the handler. Typically JSON, Text, etc.
	Renderer string

	// Authenticated should be set to true so the server activates
	// authentication for that route, i.e. loads the user that will
	// be available in the context.
	Authenticated bool
}

type Server interface {
	Register(...Route)
}
