package cayley

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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
	return r.userFromStartingPoint(startingPoint)
}

// GetByGoogleID retrieves a user by its google id instead of its internal id.
func (r *UserRepository) GetByGoogleID(googleID string) (auth.User, error) {
	startingPoint := cayley.StartPath(r.store, quad.Raw(googleID)).In(quad.Raw("googleID"))
	return r.userFromStartingPoint(startingPoint)
}

func (r *UserRepository) GetByEmail(email string) (auth.User, error) {
	startingPoint := cayley.StartPath(r.store, quad.Raw(email)).In(quad.Raw("email"))
	return r.userFromStartingPoint(startingPoint)
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
// this method sets the user ID before inserting. The following fields are updated in the database:
// name, email, googleID and isAdmin. For links to other entities, such as owned papers, bookmarks or teams,
// use the appropriate functions.
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

	return r.store.ApplyTransaction(tx)
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

	err = r.store.ApplyTransaction(tx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// PaperOwner retrieves for a paper defined by its id the user id of the owner of that paper.
// This supposes that there is only one owner for a given paper. It is the responsibility of
// the caller to ensure that by checking if a paper already has an owner before adding a
// new link (for now).
func (r *UserRepository) PaperOwner(paperID int) (int, error) {
	p := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("paper:%d", paperID))).In(quad.Raw("owns"))

	it := r.buildIterator(p)
	defer it.Close()

	if !it.Next() {
		return 0, nil
	}

	token := it.Result()                // get a ref to a node (backend-specific)
	value := r.store.NameOf(token)      // get the value in the node (RDF)
	nativeValue := quad.NativeOf(value) // convert value to normal Go type

	if nativeValue == nil {
		return 0, nil
	}

	_, idStr := splitIRI(nativeValue.(quad.IRI))
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *UserRepository) UpdatePaperOwner(userID, paperID int, owns bool) error {
	ownsQuad := quad.Make(
		quad.IRI(fmt.Sprintf("user:%d", userID)),
		quad.Raw("owns"),
		quad.IRI(fmt.Sprintf("paper:%d", paperID)),
		"",
	)

	if !owns {
		return r.store.RemoveQuad(ownsQuad)

	}

	return r.store.AddQuad(ownsQuad)
}

// -----------------------------------------------------------------------------
// Teams

func (r *UserRepository) GetTeam(id int) (auth.Team, error) {
	startingPoint := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("team:%d", id)))
	p := startingPoint.Clone().SaveOptional(quad.Raw("name"), "name")

	it := r.buildIterator(p)
	defer it.Close()

	team := auth.Team{
		Members: make([]auth.TeamMember, 0),
	}
	for it.Next() {
		token := it.Result()                // get a ref to a node (backend-specific)
		value := r.store.NameOf(token)      // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type

		if nativeValue == nil {
			continue
		}

		_, idStr := splitIRI(nativeValue.(quad.IRI))
		teamID, err := strconv.Atoi(idStr)
		if err != nil {
			return auth.Team{}, err
		}
		team.ID = teamID

		m := make(map[string]graph.Value)
		it.TagResults(m)
		for tag, token := range m {
			value := r.store.NameOf(token)      // get the value in the node (RDF)
			nativeValue := quad.NativeOf(value) // convert value to normal Go type
			switch tag {
			case "name":
				team.Name = string(nativeValue.(quad.Raw))
			default:
				// Do nothing
				fmt.Println("unsupported tag", tag, "with value", nativeValue)
			}
		}

	}

	admins := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("team:%d", team.ID))).InWithTags(
		[]string{"isAdminOf"},
		quad.Raw("isAdminOf"),
	).SaveOptional(
		quad.Raw("name"), "name",
	).SaveOptional(
		quad.Raw("email"), "email",
	)

	members := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("team:%d", team.ID))).InWithTags(
		[]string{"isMemberOf"},
		quad.Raw("isMemberOf"),
	).SaveOptional(
		quad.Raw("name"), "name",
	).SaveOptional(
		quad.Raw("email"), "email",
	)

	p = admins.Or(members)
	it = r.buildIterator(p)
	defer it.Close()

	for it.Next() {
		token := it.Result()                // get a ref to a node (backend-specific)
		value := r.store.NameOf(token)      // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type

		if nativeValue == nil {
			continue
		}

		_, idStr := splitIRI(nativeValue.(quad.IRI))
		memberID, err := strconv.Atoi(idStr)
		if err != nil {
			return auth.Team{}, err
		}

		member := auth.TeamMember{
			ID: memberID,
		}

		m := make(map[string]graph.Value)
		it.TagResults(m)

		fmt.Println(team, nativeValue, m)
		for tag, token := range m {
			value := r.store.NameOf(token)      // get the value in the node (RDF)
			nativeValue := quad.NativeOf(value) // convert value to normal Go type

			switch tag {
			case "isAdminOf":
				member.IsTeamAdmin = true
			case "name":
				member.Name = string(nativeValue.(quad.Raw))
			case "email":
				member.Email = string(nativeValue.(quad.Raw))
			default:
				fmt.Println("unsupported tag", tag, "with value", nativeValue)
			}
		}

		team.Members = append(team.Members, member)
	}

	return team, nil
}

func (r *UserRepository) UserTeams(userID int) ([]auth.Team, error) {
	startingPoint := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("user:%d", userID)))

	adminOf := startingPoint.Clone().Out(quad.Raw("isAdminOf")).Save(quad.Raw("name"), "name")

	memberOf := startingPoint.Clone().Out(quad.Raw("isMemberOf")).Save(quad.Raw("name"), "name")

	p := adminOf.Or(memberOf)
	it := r.buildIterator(p)
	defer it.Close()

	teams := make([]auth.Team, 0)
	for it.Next() {
		token := it.Result()                // get a ref to a node (backend-specific)
		value := r.store.NameOf(token)      // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type

		if nativeValue == nil {
			continue
		}

		_, idStr := splitIRI(nativeValue.(quad.IRI))
		teamID, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, err
		}

		team := auth.Team{
			ID:      teamID,
			Members: make([]auth.TeamMember, 0),
			CanSee:  make([]int, 0),
			CanEdit: make([]int, 0),
		}

		m := make(map[string]graph.Value)
		it.TagResults(m)

		for tag, token := range m {
			value := r.store.NameOf(token)      // get the value in the node (RDF)
			nativeValue := quad.NativeOf(value) // convert value to normal Go type

			switch tag {
			case "name":
				team.Name = string(nativeValue.(quad.Raw))
			default:
				fmt.Println("unsupported tag", tag, "with value", nativeValue)
			}
		}

		teams = append(teams, team)
	}

	// Retrieve team members and permissions
	for i, team := range teams {
		admins := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("team:%d", team.ID))).InWithTags(
			[]string{"isAdminOf"},
			quad.Raw("isAdminOf"),
		).SaveOptional(
			quad.Raw("name"), "name",
		).SaveOptional(
			quad.Raw("email"), "email",
		)

		members := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("team:%d", team.ID))).InWithTags(
			[]string{"isMemberOf"},
			quad.Raw("isMemberOf"),
		).SaveOptional(
			quad.Raw("name"), "name",
		).SaveOptional(
			quad.Raw("email"), "email",
		)

		p := admins.Or(members)
		it := r.buildIterator(p)
		defer it.Close()

		for it.Next() {
			token := it.Result()                // get a ref to a node (backend-specific)
			value := r.store.NameOf(token)      // get the value in the node (RDF)
			nativeValue := quad.NativeOf(value) // convert value to normal Go type

			if nativeValue == nil {
				continue
			}

			_, idStr := splitIRI(nativeValue.(quad.IRI))
			memberID, err := strconv.Atoi(idStr)
			if err != nil {
				return nil, err
			}

			member := auth.TeamMember{
				ID: memberID,
			}

			m := make(map[string]graph.Value)
			it.TagResults(m)

			fmt.Println(team, nativeValue, m)
			for tag, token := range m {
				value := r.store.NameOf(token)      // get the value in the node (RDF)
				nativeValue := quad.NativeOf(value) // convert value to normal Go type

				switch tag {
				case "isAdminOf":
					member.IsTeamAdmin = true
				case "name":
					member.Name = string(nativeValue.(quad.Raw))
				case "email":
					member.Email = string(nativeValue.(quad.Raw))
				default:
					fmt.Println("unsupported tag", tag, "with value", nativeValue)
				}
			}

			team.Members = append(team.Members, member)
		}

		canSee := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("team:%d", team.ID))).OutWithTags(
			[]string{"canSee"},
			quad.Raw("canSee"),
		)
		canEdit := cayley.StartPath(r.store, quad.IRI(fmt.Sprintf("team:%d", team.ID))).OutWithTags(
			[]string{"canEdit"},
			quad.Raw("canEdit"),
		)

		p = canSee.Or(canEdit)
		it = r.buildIterator(p)
		defer it.Close()

		for it.Next() {
			token := it.Result()                // get a ref to a node (backend-specific)
			value := r.store.NameOf(token)      // get the value in the node (RDF)
			nativeValue := quad.NativeOf(value) // convert value to normal Go type

			if nativeValue == nil {
				continue
			}

			_, idStr := splitIRI(nativeValue.(quad.IRI))
			paperID, err := strconv.Atoi(idStr)
			if err != nil {
				return nil, err
			}

			m := make(map[string]graph.Value)
			it.TagResults(m)

			fmt.Println(team, nativeValue, m)
			for tag, token := range m {
				value := r.store.NameOf(token)      // get the value in the node (RDF)
				nativeValue := quad.NativeOf(value) // convert value to normal Go type

				switch tag {
				case "canSee":
					team.CanSee = append(team.CanSee, paperID)
				case "canEdit":
					team.CanEdit = append(team.CanEdit, paperID)
				default:
					fmt.Println("unsupported tag", tag, "with value", nativeValue)
				}
			}
		}

		teams[i] = team
	}

	return teams, nil
}

func (r *UserRepository) UpsertTeam(team *auth.Team) error {
	if team.ID == 0 {
		id, err := r.incrementMaxID()
		if err != nil {
			return err
		}

		team.ID = id
	}

	replace := func(tx *graph.Transaction, teamID int, predicate, oldValue, newValue string) {
		// Remove old value
		removeQuad := quad.Make(
			quad.IRI(fmt.Sprintf("team:%d", teamID)),
			quad.Raw(predicate),
			quad.Raw(oldValue),
			"",
		)
		tx.RemoveQuad(removeQuad)

		// Set new value
		addQuad := quad.Make(
			quad.IRI(fmt.Sprintf("team:%d", teamID)),
			quad.Raw(predicate),
			quad.Raw(newValue),
			"",
		)
		tx.AddQuad(addQuad)
	}

	// Upsert
	oldTeam, err := r.Get(team.ID)
	if err != nil {
		return err
	}
	tx := graph.NewTransaction()

	replace(tx, team.ID, "name", oldTeam.Name, team.Name)

	// Add new members
	for _, m := range team.Members {
		if m.IsTeamAdmin {
			tx.AddQuad(quad.Make(
				quad.IRI(fmt.Sprintf("user:%d", m.ID)),
				quad.Raw("isAdminOf"),
				quad.IRI(fmt.Sprintf("team:%d", team.ID)),
				"",
			))
		} else {
			tx.AddQuad(quad.Make(
				quad.IRI(fmt.Sprintf("user:%d", m.ID)),
				quad.Raw("isMemberOf"),
				quad.IRI(fmt.Sprintf("team:%d", team.ID)),
				"",
			))
		}
	}

	// Add team to all teams
	// tx.AddQuad(quad.Make(allUsersNode, allUsersEdge, quad.IRI(fmt.Sprintf("user:%d", user.ID)), ""))
	return r.store.ApplyTransaction(tx)
}

func (r *UserRepository) UpdateTeamPermission(teamID, paperID int, canSee, canEdit bool) error {
	canSeeQuad := quad.Make(
		quad.IRI(fmt.Sprintf("team:%d", teamID)),
		quad.Raw("canSee"),
		quad.IRI(fmt.Sprintf("paper:%d", paperID)),
		"",
	)

	if !canSee {
		err := r.store.RemoveQuad(canSeeQuad)
		if err != nil {
			return err
		}
	} else {
		err := r.store.AddQuad(canSeeQuad)
		if err != nil {
			return err
		}
	}

	canEditQuad := quad.Make(
		quad.IRI(fmt.Sprintf("team:%d", teamID)),
		quad.Raw("canEdit"),
		quad.IRI(fmt.Sprintf("paper:%d", paperID)),
		"",
	)

	if !canEdit {
		err := r.store.RemoveQuad(canEditQuad)
		if err != nil {
			return err
		}
	} else {
		err := r.store.AddQuad(canEditQuad)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *UserRepository) UpdateTeamMember(userID, teamID int, isMember, isAdmin bool) error {
	isMemberQuad := quad.Make(
		quad.IRI(fmt.Sprintf("user:%d", userID)),
		quad.Raw("isMemberOf"),
		quad.IRI(fmt.Sprintf("team:%d", teamID)),
		"",
	)

	if !isMember {
		err := r.store.RemoveQuad(isMemberQuad)
		if err != nil {
			return err
		}
	} else {
		err := r.store.AddQuad(isMemberQuad)
		if err != nil {
			return err
		}
	}

	isAdminQuad := quad.Make(
		quad.IRI(fmt.Sprintf("user:%d", userID)),
		quad.Raw("isAdminOf"),
		quad.IRI(fmt.Sprintf("team:%d", teamID)),
		"",
	)

	if !isAdmin {
		err := r.store.RemoveQuad(isAdminQuad)
		if err != nil {
			return err
		}
	} else {
		err := r.store.AddQuad(isAdminQuad)
		if err != nil {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Helpers

func (r *UserRepository) userFromStartingPoint(startingPoint *path.Path) (auth.User, error) {
	p := startingPoint.Clone().SaveOptional(
		quad.Raw("name"), "name",
	).SaveOptional(
		quad.Raw("email"), "email",
	).SaveOptional(
		quad.Raw("googleID"), "googleID",
	).SaveOptional(
		quad.Raw("isAdmin"), "isAdmin",
	)

	it := r.buildIterator(p)
	defer it.Close()

	user := auth.User{
		Owns:      make([]int, 0),
		CanSee:    make([]int, 0),
		CanEdit:   make([]int, 0),
		Bookmarks: make([]int, 0),
	}
	for it.Next() {
		token := it.Result()                // get a ref to a node (backend-specific)
		value := r.store.NameOf(token)      // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type

		if nativeValue == nil {
			continue
		}

		_, idStr := splitIRI(nativeValue.(quad.IRI))
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			return auth.User{}, err
		}
		user.ID = userID

		m := make(map[string]graph.Value)
		it.TagResults(m)
		for tag, token := range m {
			value := r.store.NameOf(token)      // get the value in the node (RDF)
			nativeValue := quad.NativeOf(value) // convert value to normal Go type

			switch tag {
			case "name":
				user.Name = string(nativeValue.(quad.Raw))
			case "email":
				user.Email = string(nativeValue.(quad.Raw))
			case "googleID":
				user.GoogleID = string(nativeValue.(quad.Raw))
			case "isAdmin":
				user.IsAdmin = string(nativeValue.(quad.Raw)) == "true"
			default:
				// Do nothing
				fmt.Println("unsupported tag", tag, "with value", nativeValue)
			}
		}

	}

	// Owned papers
	ownsPath := startingPoint.Clone().OutWithTags(
		[]string{"owns"},
		quad.Raw("owns"),
	)
	bookmarksPath := startingPoint.Clone().OutWithTags(
		[]string{"bookmarks"},
		quad.Raw("bookmarks"),
	)
	canSeePath := startingPoint.Clone().Out(
		quad.Raw("isAdminOf"),
		quad.Raw("isMemberOf"),
	).OutWithTags(
		[]string{"canSee"},
		quad.Raw("canSee"),
	)
	canEditPath := startingPoint.Clone().Out(
		quad.Raw("isAdminOf"),
		quad.Raw("isMemberOf"),
	).OutWithTags(
		[]string{"canEdit"},
		quad.Raw("canEdit"),
	)

	p = ownsPath.Or(bookmarksPath).Or(canSeePath).Or(canEditPath)
	it = r.buildIterator(p)
	defer it.Close()

	owns := make(map[int]struct{})
	canSee := make(map[int]struct{})
	canEdit := make(map[int]struct{})
	for it.Next() {
		token := it.Result()                // get a ref to a node (backend-specific)
		value := r.store.NameOf(token)      // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type

		if nativeValue == nil {
			continue
		}
		_, idStr := splitIRI(nativeValue.(quad.IRI))
		paperID, err := strconv.Atoi(idStr)
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
			owns[paperID] = struct{}{}
			canSee[paperID] = struct{}{}
			canEdit[paperID] = struct{}{}
		case "bookmarks":
			user.Bookmarks = append(user.Bookmarks, paperID)
		default:
			fmt.Printf("unsupported tag %s with value %v\n", tag, nativeValue)
		}
	}

	for paperID := range owns {
		user.Owns = append(user.Owns, paperID)
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

	err = r.store.ApplyTransaction(tx)
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
