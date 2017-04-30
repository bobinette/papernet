package cayley

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cayleygraph/cayley/quad"
)

func TestSplitIRI(t *testing.T) {
	tts := map[string]struct {
		iri    quad.IRI
		prefix string
		data   string
	}{
		"valid iri": {
			iri:    quad.IRI("pizza:yolo"),
			prefix: "pizza",
			data:   "yolo",
		},
		"prefix with colon": {
			iri:    quad.IRI("piz:za:yolo"),
			prefix: "piz:za",
			data:   "yolo",
		},
		"no colon, but no panic": {
			iri:    quad.IRI("pizzayolo"),
			prefix: "",
			data:   "",
		},
	}

	for name, tt := range tts {
		prefix, data := splitIRI(tt.iri)
		assert.Equal(t, tt.prefix, prefix, "%s - invalid prefix", name)
		assert.Equal(t, tt.data, data, "%s - invalid data", name)
	}
}

func TestSplitIRIInt(t *testing.T) {
	tts := map[string]struct {
		iri    quad.IRI
		prefix string
		id     int
		fail   bool
	}{
		"valid iri": {
			iri:    quad.IRI("pizza:1"),
			prefix: "pizza",
			id:     1,
			fail:   false,
		},
		"prefix with colon": {
			iri:    quad.IRI("piz:za:1"),
			prefix: "piz:za",
			id:     1,
			fail:   false,
		},
		"not int, so error": {
			iri:    quad.IRI("pizza:yolo"),
			prefix: "",
			id:     0,
			fail:   true,
		},
	}

	for name, tt := range tts {
		prefix, id, err := splitIRIInt(tt.iri)
		assert.Equal(t, tt.prefix, prefix, "%s - invalid prefix", name)
		assert.Equal(t, tt.id, id, "%s - invalid id", name)
		assert.Equal(t, tt.fail, err != nil, "%s - invalid error status", name)
	}
}
