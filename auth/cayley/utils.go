package cayley

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/cayley/quad"
)

var (
	deletedNode = quad.Raw("deleted")
	deletedEdge = quad.Raw("isDeleted")

	nameEdge      = quad.Raw("name")
	emailEdge     = quad.Raw("email")
	isAdminEdge   = quad.Raw("isAdmin")
	ownsEdge      = quad.Raw("owns")
	bookmarksEdge = quad.Raw("bookmarks")

	isAdminOfEdge  = quad.Raw("isAdminOf")
	isMemberOfEdge = quad.Raw("isMemberOf")
	canSeeEdge     = quad.Raw("canSee")
	canEditEdge    = quad.Raw("canEdit")
)

// userQuad crafts a user quad.IRI from an id: <user:id>
func userQuad(id int) quad.IRI {
	return quad.IRI(fmt.Sprintf("user:%d", id))
}

// teamQuad crafts a team quad.IRI from an id: <team:id>
func teamQuad(id int) quad.IRI {
	return quad.IRI(fmt.Sprintf("team:%d", id))
}

// paperQuad crafts a paper quad.IRI from an id: <paper:id>
func paperQuad(id int) quad.IRI {
	return quad.IRI(fmt.Sprintf("paper:%d", id))
}

// splitIRI splits the iri into prefix and data, both as strings.
func splitIRI(iri quad.IRI) (string, string) {
	iriString := iri.String()
	// ":" is allowed in the prefix but not in the id
	index := strings.LastIndex(iriString, ":")
	if index < 0 {
		return "", ""
	}
	return iriString[1:index], iriString[index+1 : len(iriString)-1]
}

// splitIRIInt splits the iri into a string prefix and an int id. It
// returns an error if it fails to convert the id to an int.
func splitIRIInt(iri quad.IRI) (string, int, error) {
	prefix, data := splitIRI(iri)
	id, err := strconv.Atoi(data)
	if err != nil {
		return "", 0, err
	}
	return prefix, id, nil
}

func addQuad(tx *graph.Transaction, source, predicate, target interface{}) {
	tx.AddQuad(quad.Make(
		source,
		predicate,
		target,
		"",
	))
}

func removeQuad(tx *graph.Transaction, source, predicate, target interface{}) {
	tx.RemoveQuad(quad.Make(
		source,
		predicate,
		target,
		"",
	))
}

func replaceTarget(tx *graph.Transaction, source, predicate, oldValue, newValue interface{}) {
	// Remove old value
	removeQuad := quad.Make(
		source,
		predicate,
		oldValue,
		"",
	)
	tx.RemoveQuad(removeQuad)

	// Set new value
	addQuad := quad.Make(
		source,
		predicate,
		newValue,
		"",
	)
	tx.AddQuad(addQuad)
}

func replacePredicate(tx *graph.Transaction, source, oldPredicate, newPredicate, target interface{}) {
	// Remove old value
	removeQuad := quad.Make(
		source,
		oldPredicate,
		target,
		"",
	)
	tx.RemoveQuad(removeQuad)

	// Set new value
	addQuad := quad.Make(
		source,
		newPredicate,
		target,
		"",
	)
	tx.AddQuad(addQuad)
}
