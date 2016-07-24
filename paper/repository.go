package paper

import (
	"encoding/binary"
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
		_, err := tx.CreateBucketIfNotExists([]byte("papers"))
		if err != nil {
			return fmt.Errorf("error creating papers bucket: %s", err)
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

func (r *Repository) Get(ids ...int) ([]*Paper, error) {
	ps := make([]*Paper, 0, len(ids))
	err := r.store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))
		for _, id := range ids {
			data := b.Get(itob(id))

			if data == nil {
				continue
			}

			var p Paper
			if err := json.Unmarshal(data, &p); err != nil {
				return err
			}

			ps = append(ps, &p)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (r *Repository) List() ([]*Paper, error) {
	var ps []*Paper
	err := r.store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))
		c := b.Cursor()

		for id, data := c.First(); id != nil; id, data = c.Next() {
			var p Paper
			if err := json.Unmarshal(data, &p); err != nil {
				return err
			}
			ps = append(ps, &p)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (r *Repository) Insert(p *Paper) error {
	err := r.store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))

		id, err := b.NextSequence()
		if err != nil {
			return fmt.Errorf("error incrementing id: %v", err)
		}

		p.ID = int(id)
		data, err := json.Marshal(p)
		if err != nil {
			return err
		}

		return b.Put(itob(p.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Update(p *Paper) error {
	err := r.store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))

		data, err := json.Marshal(p)
		if err != nil {
			return err
		}

		return b.Put(itob(p.ID), data)
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Delete(id int) error {
	return r.store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))
		return b.Delete(itob(id))
	})
}

// ------------------------------------------------------------------------------------------------
// Helpers
// ------------------------------------------------------------------------------------------------

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
