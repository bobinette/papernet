package bolt

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"

	"github.com/bobinette/papernet"
)

var bucketName = "papers"

// PaperRepository is used to store and retrieve papers from a bolt database.
type PaperRepository struct {
	store *bolt.DB
}

// Get retrieves the paper defined by id in the database. If no paper can be found with the
// given id, Get returns nil.
func (r *PaperRepository) Get(id int) (*papernet.Paper, error) {
	var paper *papernet.Paper

	err := r.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))

		data := bucket.Get(itob(id))
		if data == nil {
			return nil
		}

		paper = new(papernet.Paper)
		return json.Unmarshal(data, paper)
	})
	if err != nil {
		return nil, err
	}

	return paper, nil
}

// Upsert inserts or update a paper in the database, depending on paper.ID.
func (r *PaperRepository) Upsert(paper *papernet.Paper) error {
	return r.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))

		if paper.ID <= 0 {
			id, err := bucket.NextSequence()
			if err != nil {
				return fmt.Errorf("error incrementing id: %v", err)
			}
			paper.ID = int(id)
		}

		data, err := json.Marshal(paper)
		if err != nil {
			return err
		}

		return bucket.Put(itob(paper.ID), data)
	})
}

func (r *PaperRepository) Delete(id int) error {
	return r.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		return bucket.Delete(itob(id))
	})
}

func (r *PaperRepository) List() ([]*papernet.Paper, error) {
	var papers []*papernet.Paper

	err := r.store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()

		for id, data := c.First(); id != nil; id, data = c.Next() {
			var paper papernet.Paper
			if err := json.Unmarshal(data, &paper); err != nil {
				return err
			}
			papers = append(papers, &paper)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return papers, nil
}

// ------------------------------------------------------------------------------------------------
// Connection
// ------------------------------------------------------------------------------------------------

// Open opens the connection to the bolt database defined by path.
func (r *PaperRepository) Open(path string) error {
	if r.store != nil {
		return errors.New("repository alread open")
	}

	store, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	err = store.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})

	r.store = store
	return nil
}

// Close closes the underlying database.
func (r *PaperRepository) Close() error {
	if r.store != nil {
		err := r.store.Close()
		r.store = nil
		return err
	}
	return nil
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
