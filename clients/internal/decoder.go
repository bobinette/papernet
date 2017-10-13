package internal

import (
	"encoding/json"
	"io"
)

type JSONDecoder struct {
	readCloser io.ReadCloser
}

func NewJSONDecoder(readCloser io.ReadCloser) *JSONDecoder {
	return &JSONDecoder{
		readCloser: readCloser,
	}
}

func (d *JSONDecoder) Decode(v interface{}) error {
	defer d.readCloser.Close()
	return json.NewDecoder(d.readCloser).Decode(v)
}

type ErrorDecoder struct {
	err error
}

func NewErrorDecoder(err error) *ErrorDecoder {
	return &ErrorDecoder{
		err: err,
	}
}
func (d *ErrorDecoder) Decode(_ interface{}) error {
	return d.err
}
