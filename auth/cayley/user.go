package cayley

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/bolt"
	"github.com/cayleygraph/cayley/graph/path"
	"github.com/cayleygraph/cayley/quad"

	"github.com/bobinette/papernet/auth"
)

type UserRepository struct {
	store *cayley.Handle
}

var (
	maxIDNode = quad.Raw("maxID")
)

// New creates and new user repository based on a cayley graph db
// on a bolt file situated at dbPath. If the path does not exist,
// it will be created.
func New(dbpath string) (*UserRepository, error) {
	err := graph.InitQuadStore("bolt", dbpath, nil)
	if err != nil && err != graph.ErrDatabaseExists {
		return nil, err
	}

	store, err := cayley.NewGraph("bolt", dbpath, nil)
	if err != nil {
		return nil, err
	}

	return &UserRepository{
		store: store,
	}, nil
}

// Get retrieves a user from its id.
func (r *UserRepository) Get(id int) (auth.User, error) {
	startingPoint := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("user:%d", id)))
	return r.getFromStartingPoint(startingPoint)
}

// GetByGoogleID retrieves a user by its google id instead of its internal id.
func (r *UserRepository) GetByGoogleID(googleID string) (auth.User, error) {
	startingPoint := cayley.StartPath(r.store, quad.Raw(googleID)).In(quad.Raw("googleID"))
	return r.getFromStartingPoint(startingPoint)
}

// Upsert updates the user passed as argument in the database. If the user has no ID (i.e. user.ID == 0),
// this method sets the user ID before inserting.
func (r *UserRepository) Upsert(user *auth.User) error {
	isNew := false
	if user.ID == 0 {
		id, err := r.incrementMaxID()
		if err != nil {
			return err
		}

		user.ID = id
		isNew = true
	}

	replace := func(tx *graph.Transaction, predicate, oldValue, newValue string) {
		// Remove old value
		removeQuad := quad.Make(
			quad.IRI(fmt.Sprintf("user:%d", user.ID)),
			quad.Raw(predicate),
			quad.Raw(oldValue),
			"",
		)
		tx.RemoveQuad(removeQuad)

		// Set new value
		addQuad := quad.Make(
			quad.IRI(fmt.Sprintf("user:%d", user.ID)),
			quad.Raw(predicate),
			quad.Raw(newValue),
			"",
		)
		tx.AddQuad(addQuad)
	}

	// Upsert
	oldUser, err := r.Get(user.ID)
	if err != nil {
		return err
	}
	tx := graph.NewTransaction()

	replace(tx, "name", oldUser.Name, user.Name)
	replace(tx, "email", oldUser.Email, user.Email)
	replace(tx, "googleID", oldUser.GoogleID, user.GoogleID)

	return r.store.ApplyDeltas(tx.Deltas, graph.IgnoreOpts{IgnoreMissing: isNew})
}

// -----------------------------------------------------------------------------
// Helpers

func (r *UserRepository) getFromStartingPoint(startingPoint *path.Path) (auth.User, error) {
	p := startingPoint.Clone().Tag(
		"userID",
	).Or(startingPoint.Clone().OutWithTags(
		[]string{"name"},
		quad.Raw("name"),
	)).Or(startingPoint.Clone().OutWithTags(
		[]string{"email"},
		quad.Raw("email"),
	)).Or(startingPoint.Clone().OutWithTags(
		[]string{"googleID"},
		quad.Raw("googleID"),
	))

	it := r.buildIterator(p)
	defer it.Close()

	user := auth.User{}
	for it.Next() {

		token := it.Result()                // get a ref to a node (backend-specific)
		value := r.store.NameOf(token)      // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type

		m := make(map[string]graph.Value)
		it.TagResults(m)
		var tag string
		for t, _ := range m {
			// only one in map, so we just need the key
			tag = t
			break
		}

		switch tag {
		case "name":
			user.Name = string(nativeValue.(quad.Raw))
		case "email":
			user.Email = string(nativeValue.(quad.Raw))
		case "googleID":
			user.GoogleID = string(nativeValue.(quad.Raw))
		case "userID":
			_, idStr := splitIRI(nativeValue.(quad.IRI))
			userID, err := strconv.Atoi(idStr)
			if err != nil {
				return auth.User{}, err
			}
			user.ID = userID
		}
	}

	return user, nil
}

func (r *UserRepository) getMaxID() (int, error) {
	p := cayley.StartPath(r.store, maxIDNode).Out(quad.Raw("value"))

	it := r.buildIterator(p)
	defer it.Close()

	// We only care about the first node
	if !it.Next() {
		return 0, nil
	}

	token := it.Result()                // get a ref to a node (backend-specific)
	value := r.store.NameOf(token)      // get the value in the node (RDF)
	nativeValue := quad.NativeOf(value) // convert value to normal Go type

	if nativeValue == nil {
		return 0, nil
	}

	maxIDStr := nativeValue.(quad.Raw)
	maxID, err := strconv.Atoi(string(maxIDStr))
	if err != nil {
		return 0, err
	}

	return maxID, nil
}

func (r *UserRepository) incrementMaxID() (int, error) {
	current, err := r.getMaxID()
	if err != nil {
		return 0, nil
	}

	// Create transaction
	tx := graph.NewTransaction()

	// Remove old value
	removeQuad := quad.Make(
		maxIDNode,
		quad.Raw("value"),
		quad.Raw(strconv.Itoa(current)),
		"",
	)
	tx.RemoveQuad(removeQuad)

	// Set new value
	addQuad := quad.Make(
		maxIDNode,
		quad.Raw("value"),
		quad.Raw(strconv.Itoa(current+1)),
		"",
	)
	tx.AddQuad(addQuad)

	err = r.store.ApplyDeltas(tx.Deltas, graph.IgnoreOpts{IgnoreMissing: (current == 0)})
	if err != nil {
		return current, err
	}
	return current + 1, nil
}

func (r *UserRepository) buildIterator(p *path.Path) graph.Iterator {
	// Get an iterator for the path and optimize it.
	// The second return is if it was optimized, but we don't care.
	it, _ := p.BuildIterator().Optimize()

	// Optimize iterator on quad store level.
	// After this step iterators will be replaced with backend-specific ones.
	it, _ = r.store.OptimizeIterator(it)

	return it
}

// extradtID splits the iri into prefix and id
func splitIRI(iri quad.IRI) (string, string) {
	iriString := iri.String()
	// ":" is allowed in the prefix but not in the id
	index := strings.LastIndex(iriString, ":")
	return iriString[1:index], iriString[index+1 : len(iriString)-1]
}
