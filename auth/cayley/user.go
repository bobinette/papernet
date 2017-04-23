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
	maxIDEdge = quad.Raw("value")

	allUsersNode = quad.Raw("allUsers")
	allUsersEdge = quad.Raw("user")
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

func (r *UserRepository) Close() error {
	return r.store.Close()
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

// List retrieves all the users in the database
func (r *UserRepository) List() ([]auth.User, error) {
	p := cayley.StartPath(r.store, allUsersNode).Out(allUsersEdge)

	it := r.buildIterator(p)
	defer it.Close()

	users := make([]auth.User, 0)
	for it.Next() {

		token := it.Result()                // get a ref to a node (backend-specific)
		value := r.store.NameOf(token)      // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type

		if nativeValue == nil {
			continue
		}

		user := auth.User{}
		_, idStr := splitIRI(nativeValue.(quad.IRI))
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, err
		}
		user.ID = userID
		users = append(users, user)
	}

	return users, nil
}

// Upsert updates the user passed as argument in the database. If the user has no ID (i.e. user.ID == 0),
// this method sets the user ID before inserting.
func (r *UserRepository) Upsert(user *auth.User) error {
	if user.ID == 0 {
		id, err := r.incrementMaxID()
		if err != nil {
			return err
		}

		user.ID = id
	}

	replace := func(tx *graph.Transaction, userID int, predicate, oldValue, newValue string) {
		// Remove old value
		removeQuad := quad.Make(
			quad.IRI(fmt.Sprintf("user:%d", userID)),
			quad.Raw(predicate),
			quad.Raw(oldValue),
			"",
		)
		tx.RemoveQuad(removeQuad)

		// Set new value
		addQuad := quad.Make(
			quad.IRI(fmt.Sprintf("user:%d", userID)),
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

	replace(tx, user.ID, "name", oldUser.Name, user.Name)
	replace(tx, user.ID, "email", oldUser.Email, user.Email)
	replace(tx, user.ID, "googleID", oldUser.GoogleID, user.GoogleID)
	replace(tx, user.ID, "isAdmin", strconv.FormatBool(oldUser.IsAdmin), strconv.FormatBool(user.IsAdmin))

	// Add user to all users
	tx.AddQuad(quad.Make(allUsersNode, allUsersEdge, quad.IRI(fmt.Sprintf("user:%d", user.ID)), ""))

	return r.store.ApplyDeltas(tx.Deltas, graph.IgnoreOpts{IgnoreMissing: true, IgnoreDup: true})
}

// Delete removes a user from the database based on its id. The first return argument is whether or not
// the user defined by id was actually removed from the database. In the case no user could be found
// for id, Delete returns (false, nil).
func (r *UserRepository) Delete(id int) (bool, error) {
	user, err := r.Get(id)
	if err != nil {
		return false, err
	} else if user.ID == 0 {
		// User does not exist
		return false, nil
	}

	deleteQuad := func(tx *graph.Transaction, userID int, predicate, value string) {
		removeQuad := quad.Make(
			quad.IRI(fmt.Sprintf("user:%d", userID)),
			quad.Raw(predicate),
			quad.Raw(value),
			"",
		)
		tx.RemoveQuad(removeQuad)
	}

	tx := graph.NewTransaction()

	deleteQuad(tx, user.ID, "name", user.Name)
	deleteQuad(tx, user.ID, "email", user.Email)
	deleteQuad(tx, user.ID, "googleID", user.GoogleID)
	deleteQuad(tx, user.ID, "isAdmin", strconv.FormatBool(user.IsAdmin))

	tx.RemoveQuad(quad.Make(allUsersNode, allUsersEdge, quad.IRI(fmt.Sprintf("user:%d", user.ID)), ""))

	err = r.store.ApplyDeltas(tx.Deltas, graph.IgnoreOpts{IgnoreMissing: true})
	if err != nil {
		return false, err
	}
	return true, nil
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
	)).Or(startingPoint.Clone().OutWithTags(
		[]string{"isAdmin"},
		quad.Raw("isAdmin"),
	))

	it := r.buildIterator(p)
	defer it.Close()

	user := auth.User{}
	for it.Next() {
		token := it.Result()                // get a ref to a node (backend-specific)
		value := r.store.NameOf(token)      // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type

		if nativeValue == nil {
			continue
		}

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
		case "isAdmin":
			user.IsAdmin = string(nativeValue.(quad.Raw)) == "true"
		case "userID":
			_, idStr := splitIRI(nativeValue.(quad.IRI))
			userID, err := strconv.Atoi(idStr)
			if err != nil {
				return auth.User{}, err
			}
			user.ID = userID
		default:
			// Do nothing
		}
	}

	return user, nil
}

func (r *UserRepository) getMaxID() (int, error) {
	p := cayley.StartPath(r.store, maxIDNode).Out(maxIDEdge)

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
	if current != 0 {
		removeQuad := quad.Make(
			maxIDNode,
			maxIDEdge,
			quad.Raw(strconv.Itoa(current)),
			"",
		)
		tx.RemoveQuad(removeQuad)
	}

	// Set new value
	addQuad := quad.Make(
		maxIDNode,
		maxIDEdge,
		quad.Raw(strconv.Itoa(current+1)),
		"",
	)
	tx.AddQuad(addQuad)

	err = r.store.ApplyDeltas(tx.Deltas, graph.IgnoreOpts{})
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
