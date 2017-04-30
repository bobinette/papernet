package errors

import (
	"net/http"
)

func BadRequest() ErrorEnricher { return WithCode(http.StatusBadRequest) }
func Forbidden() ErrorEnricher  { return WithCode(http.StatusForbidden) }
func NotFound() ErrorEnricher   { return WithCode(http.StatusNotFound) }
