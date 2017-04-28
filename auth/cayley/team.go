package cayley

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/cayley/quad"

	"github.com/bobinette/papernet/auth"
)

var (
	maxTeamIDNode = quad.Raw("maxTeamID")
	maxTeamIDEdge = quad.Raw("value")

	allTeamsNode = quad.Raw("allTeams")
	allTeamsEdge = quad.Raw("team")
)

type TeamRepository struct {
	store *Store
}

// NewUserRepository creates a new user repository based on a store.
func NewTeamRepository(store *Store) *TeamRepository {
	return &TeamRepository{
		store: store,
	}
}

func (r *TeamRepository) Get(id int) (auth.Team, error) {
	startingPoint := cayley.StartPath(r.store, teamQuad(id))
	startingPoint = startingPoint.Except(startingPoint.HasReverse(deletedEdge, deletedNode))
	p := startingPoint.SaveOptional(nameEdge, "name")

	it := r.store.buildIterator(p)
	defer it.Close()

	team := auth.Team{
		Members: make([]auth.TeamMember, 0),
		CanSee:  make([]int, 0),
		CanEdit: make([]int, 0),
	}
	for it.Next() {
		teamID, err := r.store.entity(it.Result(), "team")
		if err != nil {
			return auth.Team{}, err
		} else if teamID == 0 {
			// TODO: log or return error, to be decided
			continue
		}

		team.ID = teamID

		m := make(map[string]graph.Value)
		it.TagResults(m)
		for tag, token := range m {
			switch tag {
			case "name":
				name, err := r.store.string(token)
				if err != nil {
					return auth.Team{}, err
				}
				team.Name = name
			default:
				// Do nothing
				fmt.Println("unsupported tag", tag)
			}
		}

	}

	admins := startingPoint.InWithTags(
		[]string{"isAdminOf"},
		isAdminOfEdge,
	) //.SaveOptional(nameEdge, "name").SaveOptional(emailEdge, "email")
	members := startingPoint.In(
		isMemberOfEdge,
	) //.SaveOptional(emailEdge, "name").SaveOptional(emailEdge, "email")

	p = admins.Or(members)
	it = r.store.buildIterator(p)
	defer it.Close()

	for it.Next() {
		memberID, err := r.store.entity(it.Result(), "user")
		if err != nil {
			return auth.Team{}, err
		} else if memberID == 0 {
			// TODO: log or return error, to be decided
			continue
		}

		member := auth.TeamMember{
			ID: memberID,
		}

		m := make(map[string]graph.Value)
		it.TagResults(m)

		for tag, token := range m {
			switch tag {
			case "isAdminOf":
				member.IsTeamAdmin = true
			case "name":
				name, err := r.store.string(token)
				if err != nil {
					return auth.Team{}, err
				}
				member.Name = name
			case "email":
				email, err := r.store.string(token)
				if err != nil {
					return auth.Team{}, err
				}
				member.Email = email
			default:
				fmt.Println("unsupported tag", tag)
			}
		}

		team.Members = append(team.Members, member)
	}

	canSee := startingPoint.OutWithTags([]string{"canSee"}, canSeeEdge)
	canEdit := startingPoint.OutWithTags([]string{"canEdit"}, canEditEdge)

	p = canSee.Or(canEdit)
	it = r.store.buildIterator(p)
	defer it.Close()

	for it.Next() {
		paperID, err := r.store.entity(it.Result(), "paper")
		if err != nil {
			return auth.Team{}, err
		} else if paperID == 0 {
			// TODO: log or return error, to be decided
			continue
		}

		m := make(map[string]graph.Value)
		it.TagResults(m)

		for tag, _ := range m {
			switch tag {
			case "canSee":
				team.CanSee = append(team.CanSee, paperID)
			case "canEdit":
				team.CanEdit = append(team.CanEdit, paperID)
			default:
				fmt.Println("unsupported tag", tag)
			}
		}
	}

	return team, nil
}

func (r *TeamRepository) GetForUser(userID int) ([]auth.Team, error) {
	p := cayley.StartPath(r.store, userQuad(userID)).Out(isAdminOfEdge, isMemberOfEdge)
	p = p.Except(p.HasReverse(deletedEdge, deletedNode))
	it := r.store.buildIterator(p)
	defer it.Close()
	ids := make([]int, 0)
	for it.Next() {
		teamID, err := r.store.entity(it.Result(), "team")
		if err != nil {
			return nil, err
		} else if teamID == 0 {
			// TODO: log or return error, to be decided
			continue
		}

		ids = append(ids, teamID)
	}

	sort.Ints(ids)
	teams := make([]auth.Team, len(ids))
	for i, id := range ids {
		team, err := r.Get(id)
		if err != nil {
			return nil, err
		}
		teams[i] = team
	}

	return teams, nil
}

func (r *TeamRepository) Upsert(team *auth.Team) error {
	if team.ID == 0 {
		id, err := r.incrementMaxID()
		if err != nil {
			return err
		}

		team.ID = id
	}

	// Upsert
	oldTeam, err := r.Get(team.ID)
	if err != nil {
		return err
	}

	tx := graph.NewTransaction()
	replaceTarget(tx, teamQuad(team.ID), nameEdge, quad.Raw(oldTeam.Name), quad.Raw(team.Name))

	// Remove old members
	for _, m := range oldTeam.Members {
		if m.IsTeamAdmin {
			removeQuad(tx, userQuad(m.ID), isAdminOfEdge, teamQuad(team.ID))
		} else {
			removeQuad(tx, userQuad(m.ID), isMemberOfEdge, teamQuad(team.ID))
		}
	}

	// Add new members
	for _, m := range team.Members {
		if m.IsTeamAdmin {
			addQuad(tx, userQuad(m.ID), isAdminOfEdge, teamQuad(team.ID))
		} else {
			addQuad(tx, userQuad(m.ID), isMemberOfEdge, teamQuad(team.ID))
		}
	}

	// Remove old permissions
	for _, paperID := range oldTeam.CanSee {
		removeQuad(tx, teamQuad(team.ID), canSeeEdge, paperQuad(paperID))
	}

	for _, paperID := range oldTeam.CanEdit {
		removeQuad(tx, teamQuad(team.ID), canEditEdge, paperQuad(paperID))
	}

	// Add permissions
	for _, paperID := range team.CanSee {
		addQuad(tx, teamQuad(team.ID), canSeeEdge, paperQuad(paperID))
	}

	for _, paperID := range team.CanEdit {
		addQuad(tx, teamQuad(team.ID), canEditEdge, paperQuad(paperID))
	}

	// Add team to all teams
	addQuad(tx, allTeamsNode, allTeamsEdge, teamQuad(team.ID))

	return r.store.ApplyTransaction(tx)
}

func (r *TeamRepository) Delete(id int) error {
	tx := graph.NewTransaction()
	addQuad(tx, deletedNode, deletedEdge, teamQuad(id))
	return r.store.ApplyTransaction(tx)
}

// -----------------------------------------------------------------------------
// Helpers

func (r *TeamRepository) getMaxID() (int, error) {
	p := cayley.StartPath(r.store, maxTeamIDNode).Out(maxTeamIDEdge)

	it := r.store.buildIterator(p)
	defer it.Close()

	// We only care about the first node
	if !it.Next() {
		return 0, nil
	}

	maxID, err := r.store.int(it.Result())
	if err != nil {
		return 0, err
	}

	return maxID, nil
}

func (r *TeamRepository) incrementMaxID() (int, error) {
	current, err := r.getMaxID()
	if err != nil {
		return 0, nil
	}

	// Create transaction
	tx := graph.NewTransaction()

	// Remove old value
	if current != 0 {
		removeQuad := quad.Make(
			maxTeamIDNode,
			maxTeamIDEdge,
			quad.Raw(strconv.Itoa(current)),
			"",
		)
		tx.RemoveQuad(removeQuad)
	}

	// Set new value
	addQuad := quad.Make(
		maxTeamIDNode,
		maxTeamIDEdge,
		quad.Raw(strconv.Itoa(current+1)),
		"",
	)
	tx.AddQuad(addQuad)

	err = r.store.ApplyTransaction(tx)
	if err != nil {
		return 0, err
	}
	return current + 1, nil
}
