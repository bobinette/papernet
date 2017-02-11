package bolt

import (
	"encoding/json"

	"github.com/boltdb/bolt"

	"github.com/bobinette/papernet"
)

var userBucket = []byte("users")

type UserRepository struct {
	Driver *Driver
}

func (r *UserRepository) Get(id string) (*papernet.User, error) {
	var user *papernet.User
	err := r.Driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userBucket)

		data := bucket.Get([]byte(id))
		if data == nil {
			return nil
		}

		user = &papernet.User{}
		return json.Unmarshal(data, user)
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Upsert(user *papernet.User) error {
	return r.Driver.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userBucket)

		data, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(user.ID), data)
	})
}

func (s *UserRepository) List() ([]*papernet.User, error) {
	var users []*papernet.User

	err := s.Driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(userBucket)

		c := bucket.Cursor()
		for id, data := c.First(); id != nil; id, data = c.Next() {
			var user papernet.User
			if err := json.Unmarshal(data, &user); err != nil {
				return err
			}
			users = append(users, &user)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return users, nil
}
