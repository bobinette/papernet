package google

import (
	"testing"
)

func TestInmemRepository(t *testing.T) {
	repo := NewInmemRepository()
	TestRepository(t, repo)
}
