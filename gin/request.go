package gin

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
)

type Request struct {
	gc *gin.Context
}

func (r *Request) Header(key string) string {
	return r.gc.Request.Header.Get(key)
}

func (r *Request) Param(key string) string {
	return r.gc.Param(key)
}

func (r *Request) Query(key string) string {
	return r.gc.Query(key)
}

func (r *Request) Body() io.ReadCloser {
	return r.gc.Request.Body
}

func (r *Request) Context() context.Context {
	return r.gc.Request.Context()
}

func (r *Request) WithContext(ctx context.Context) papernet.Request {
	req := &Request{r.gc.Copy()}
	req.gc.Request = req.gc.Request.WithContext(ctx)
	return req
}
