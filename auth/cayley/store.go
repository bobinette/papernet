package cayley

import (
	"fmt"
	"strconv"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/cayley/graph/path"
	"github.com/cayleygraph/cayley/quad"

	"github.com/bobinette/papernet/errors"
)

type Store struct {
	*cayley.Handle
}

// NewStore creates a new store based on a cayley graph db
// on a bolt file situated at dbPath. If the path does not exist,
// it will be created.
func NewStore(dbpath string) (*Store, error) {
	err := graph.InitQuadStore("bolt", dbpath, nil)
	if err != nil && err != graph.ErrDatabaseExists {
		return nil, err
	}

	store, err := cayley.NewGraph("bolt", dbpath, nil)
	if err != nil {
		return nil, err
	}

	return &Store{
		store,
	}, nil
}

func (s *Store) Close() error {
	return s.Handle.Close()
}

func (s *Store) buildIterator(p *path.Path) graph.Iterator {
	// Get an iterator for the path and optimize it.
	// The second return is if it was optimized, but we don't care.
	it, _ := p.BuildIterator().Optimize()

	// Optimize iterator on quad store level.
	// After this step iterators will be replaced with backend-specific ones.
	it, _ = s.OptimizeIterator(it)

	return it
}

func (s *Store) int(token graph.Value) (int, error) {
	value := s.NameOf(token)            // get the value in the node
	nativeValue := quad.NativeOf(value) // convert value to normal Go type

	if nativeValue == nil {
		return 0, nil
	}

	// Try to get an int immediatly
	v, ok := nativeValue.(int)
	if ok {
		return v, nil
	}

	var str string
	switch nv := nativeValue.(type) {
	case string:
		str = nv
	case quad.Raw:
		str = string(nv)
	default:
		return 0, errors.New(fmt.Sprintf("invalid type %T for int node", nv))
	}

	v, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (s *Store) string(token graph.Value) (string, error) {
	value := s.NameOf(token)            // get the value in the node
	nativeValue := quad.NativeOf(value) // convert value to normal Go type

	if nativeValue == nil {
		return "", nil
	}

	switch nv := nativeValue.(type) {
	case string:
		return nv, nil
	case quad.Raw:
		return string(nv), nil
	}
	return "", errors.New(fmt.Sprintf("invalid type %T for string node", nativeValue))
}

func (s *Store) entity(token graph.Value, entityType string) (int, error) {
	value := s.NameOf(token)            // get the value in the node (RDF)
	nativeValue := quad.NativeOf(value) // convert value to normal Go type

	if nativeValue == nil {
		return 0, nil
	}

	prefix, id, err := splitIRIInt(nativeValue.(quad.IRI))
	if err != nil {
		return 0, err
	} else if prefix != entityType {
		return 0, errors.New(fmt.Sprintf("invalid entity type: %s, wanted %s", prefix, entityType))
	}

	return id, nil
}
