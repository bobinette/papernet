package user

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

type Repository struct {
	store *bolt.DB
}

func NewRepository(store *bolt.DB) (*Repository, error) {
	// Check buckets
	err := store.Update(func(tx *bolt.Tx) error {
		// Create papers bucket
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("error creating users bucket: %s", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Repository{
		store: store,
	}, nil
}

func (r *Repository) Get(name string) (*User, error) {
	var u *User
	err := r.store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		data := b.Get([]byte(name))

		if data == nil {
			return nil
		}

		u = &User{}
		if err := json.Unmarshal(data, u); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *Repository) Upsert(u *User) error {
	err := r.store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))

		data, err := json.Marshal(u)
		if err != nil {
			return err
		}

		return b.Put([]byte(u.Name), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Delete(name string) error {
	return r.store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		return b.Delete([]byte(name))
	})
}

// ------------------------------------------------------------------------------------------------
// Helpers
// ------------------------------------------------------------------------------------------------
