package cayley

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/bolt"
	"github.com/cayleygraph/cayley/graph/path"
	"github.com/cayleygraph/cayley/quad"

	"github.com/bobinette/papernet/auth"
)

func init() {
	graph.IgnoreDuplicates = true
	graph.IgnoreMissing = true
}

type UserRepository struct {
	store *Store
}

var (
	maxUserIDNode = quad.Raw("maxUserID")
	maxUserIDEdge = quad.Raw("value")

	allUsersNode = quad.Raw("allUsers")
	allUsersEdge = quad.Raw("user")
)

// NewUserRepository creates a new user repository based on a store.
func NewUserRepository(store *Store) *UserRepository {
	return &UserRepository{
		store: store,
	}
}

// Get retrieves a user from its id.
func (r *UserRepository) Get(id int) (auth.User, error) {
	startingPoint := cayley.StartPath(r.store, userQuad(id))
	startingPoint = startingPoint.Except(startingPoint.HasReverse(deletedEdge, deletedNode))

	return r.userFromStartingPoint(startingPoint)
}

// GetByGoogleID retrieves a user by its google id instead of its internal id.
func (r *UserRepository) GetByGoogleID(googleID string) (auth.User, error) {
	startingPoint := cayley.StartPath(r.store, quad.Raw(googleID)).In(googleIDEdge)
	startingPoint = startingPoint.Except(startingPoint.HasReverse(deletedEdge, deletedNode))

	return r.userFromStartingPoint(startingPoint)
}

func (r *UserRepository) GetByEmail(email string) (auth.User, error) {
	startingPoint := cayley.StartPath(r.store, quad.Raw(email)).In(emailEdge)
	startingPoint = startingPoint.Except(startingPoint.HasReverse(deletedEdge, deletedNode))

	return r.userFromStartingPoint(startingPoint)
}

// Upsert updates the user passed as argument in the database. If the user has no ID (i.e. user.ID == 0),
// this method sets the user ID before inserting.
func (r *UserRepository) Upsert(user *auth.User) error {
	if user.ID == 0 {
		id, err := r.store.incrementMaxID(maxUserIDNode, maxUserIDEdge)
		if err != nil {
			return err
		}

		user.ID = id
	}

	// Upsert
	oldUser, err := r.Get(user.ID)
	if err != nil {
		return err
	}
	tx := graph.NewTransaction()

	// Update user profile
	replaceTarget(tx, userQuad(user.ID), nameEdge, quad.Raw(oldUser.Name), quad.Raw(user.Name))
	replaceTarget(tx, userQuad(user.ID), emailEdge, quad.Raw(oldUser.Email), quad.Raw(user.Email))
	replaceTarget(tx, userQuad(user.ID), googleIDEdge, quad.Raw(oldUser.GoogleID), quad.Raw(user.GoogleID))
	replaceTarget(tx, userQuad(user.ID), isAdminEdge, strconv.FormatBool(oldUser.IsAdmin), strconv.FormatBool(user.IsAdmin))

	// Update user owned papers
	for _, paperID := range oldUser.Owns {
		removeQuad(tx, userQuad(user.ID), ownsEdge, paperQuad(paperID))
	}

	for _, paperID := range user.Owns {
		addQuad(tx, userQuad(user.ID), ownsEdge, paperQuad(paperID))
	}

	// Update user bookmarks
	for _, paperID := range oldUser.Bookmarks {
		removeQuad(tx, userQuad(user.ID), bookmarksEdge, paperQuad(paperID))
	}

	for _, paperID := range user.Bookmarks {
		addQuad(tx, userQuad(user.ID), bookmarksEdge, paperQuad(paperID))
	}

	// Add user to all users
	addQuad(tx, allUsersNode, allUsersEdge, userQuad(user.ID))

	return r.store.ApplyTransaction(tx)
}

// Delete removes a user from the database based on its id.
func (r *UserRepository) Delete(id int) error {
	tx := graph.NewTransaction()
	addQuad(tx, deletedNode, deletedEdge, userQuad(id))
	return r.store.ApplyTransaction(tx)
}

// PaperOwner retrieves for a paper defined by its id the user id of the owner of that paper.
// This supposes that there is only one owner for a given paper. It is the responsibility of
// the caller to ensure that by checking if a paper already has an owner before adding a
// new link (for now).
func (r *UserRepository) PaperOwner(paperID int) (int, error) {
	p := cayley.StartPath(r.store, paperQuad(paperID)).In(quad.Raw("owns"))
	p = p.Except(p.HasReverse(deletedEdge, deletedNode))

	it := r.store.buildIterator(p)
	defer it.Close()

	if !it.Next() {
		return 0, nil
	}

	userID, err := r.store.entity(it.Result(), "user")
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// -----------------------------------------------------------------------------
// Helpers

func (r *UserRepository) userFromStartingPoint(startingPoint *path.Path) (auth.User, error) {
	p := startingPoint.Clone().SaveOptional(
		nameEdge, "name",
	).SaveOptional(
		emailEdge, "email",
	).SaveOptional(
		googleIDEdge, "googleID",
	).SaveOptional(
		isAdminEdge, "isAdmin",
	)

	it := r.store.buildIterator(p)
	defer it.Close()

	user := auth.User{
		Owns:      make([]int, 0),
		CanSee:    make([]int, 0),
		CanEdit:   make([]int, 0),
		Bookmarks: make([]int, 0),
	}
	for it.Next() {
		userID, err := r.store.entity(it.Result(), "user")
		if err != nil {
			return auth.User{}, err
		}
		user.ID = userID

		m := make(map[string]graph.Value)
		it.TagResults(m)
		for tag, token := range m {
			switch tag {
			case "name":
				user.Name, err = r.store.string(token)
				if err != nil {
					return auth.User{}, err
				}
			case "email":
				user.Email, err = r.store.string(token)
				if err != nil {
					return auth.User{}, err
				}
			case "googleID":
				user.GoogleID, err = r.store.string(token)
				if err != nil {
					return auth.User{}, err
				}
			case "isAdmin":
				isAdmin, err := r.store.string(token)
				if err != nil {
					return auth.User{}, err
				}
				user.IsAdmin = isAdmin == "true"
			default:
				// Do nothing
				fmt.Println("unsupported tag", tag)
			}
		}

	}

	// Owned papers
	ownsPath := startingPoint.Clone().OutWithTags(
		[]string{"owns"},
		ownsEdge,
	)
	bookmarksPath := startingPoint.Clone().OutWithTags(
		[]string{"bookmarks"},
		bookmarksEdge,
	)
	canSeePath := startingPoint.Clone().Out(
		isAdminOfEdge,
		isMemberOfEdge,
	).OutWithTags(
		[]string{"canSee"},
		canSeeEdge,
	)
	canEditPath := startingPoint.Clone().Out(
		isAdminOfEdge,
		isMemberOfEdge,
	).OutWithTags(
		[]string{"canEdit"},
		canEditEdge,
	)

	p = ownsPath.Or(bookmarksPath).Or(canSeePath).Or(canEditPath)
	it = r.store.buildIterator(p)
	defer it.Close()

	// Int sets
	canSee := make(map[int]struct{})
	canEdit := make(map[int]struct{})
	for it.Next() {
		paperID, err := r.store.entity(it.Result(), "paper")
		if err != nil {
			return auth.User{}, err
		}

		m := make(map[string]graph.Value)
		it.TagResults(m)
		var tag string
		for k := range m {
			tag = k
			break
		}

		switch tag {
		case "canSee":
			canSee[paperID] = struct{}{}
		case "canEdit":
			canSee[paperID] = struct{}{}
			canEdit[paperID] = struct{}{}
		case "owns":
			user.Owns = append(user.Owns, paperID)
			canSee[paperID] = struct{}{}
			canEdit[paperID] = struct{}{}
		case "bookmarks":
			user.Bookmarks = append(user.Bookmarks, paperID)
		default:
			fmt.Println("unsupported tag", tag)
		}
	}

	for paperID := range canSee {
		user.CanSee = append(user.CanSee, paperID)
	}

	for paperID := range canEdit {
		user.CanEdit = append(user.CanEdit, paperID)
	}

	// Sort the paper ids to look normal
	sort.Ints(user.Owns)
	sort.Ints(user.CanSee)
	sort.Ints(user.CanEdit)
	sort.Ints(user.Bookmarks)

	return user, nil
}
