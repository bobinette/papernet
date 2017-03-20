package bolt

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"

	"github.com/bobinette/papernet"
)

var teamBucket = []byte("teams")

type TeamStore struct {
	Driver *Driver
}

func (s *TeamStore) Get(id int) (papernet.Team, error) {
	var team papernet.Team
	err := s.Driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(teamBucket)

		data := bucket.Get(itob(id))
		if data == nil {
			return nil
		}

		team = papernet.Team{}
		return json.Unmarshal(data, &team)
	})
	if err != nil {
		return papernet.Team{}, err
	}

	return team, nil
}

func (s *TeamStore) Upsert(team *papernet.Team) error {
	return s.Driver.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(teamBucket)

		if team.ID <= 0 {
			id, err := bucket.NextSequence()
			if err != nil {
				return fmt.Errorf("error incrementing id: %v", err)
			}
			team.ID = int(id)
		}

		data, err := json.Marshal(team)
		if err != nil {
			return err
		}

		return bucket.Put(itob(team.ID), data)
	})
}

func (s *TeamStore) All() ([]*papernet.Team, error) {
	teams := make([]*papernet.Team, 0)

	err := s.Driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(teamBucket)

		c := bucket.Cursor()
		for id, data := c.First(); id != nil; id, data = c.Next() {
			var team papernet.Team
			if err := json.Unmarshal(data, &team); err != nil {
				return err
			}
			teams = append(teams, &team)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *TeamStore) List(userID string) ([]papernet.Team, error) {
	teams := make([]papernet.Team, 0)

	err := s.Driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(teamBucket)

		c := bucket.Cursor()
		for id, data := c.First(); id != nil; id, data = c.Next() {
			var team papernet.Team
			if err := json.Unmarshal(data, &team); err != nil {
				return err
			}

			for _, uID := range team.Members {
				if uID == userID {
					teams = append(teams, team)
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *TeamStore) Delete(id int) error {
	return s.Driver.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(teamBucket)
		return bucket.Delete(itob(id))
	})
}
