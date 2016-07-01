package database

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"

	"github.com/bobinette/papernet/models"
)

type boltDB struct {
	store *bolt.DB
}

func NewBoltDB(dbpath string) (DB, error) {
	store, err := bolt.Open(dbpath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	// Check buckets
	err = store.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("papers"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &boltDB{
		store: store,
	}, nil
}

func (db *boltDB) Close() error {
	return db.store.Close()
}

func (db *boltDB) Get(ids ...int) ([]*models.Paper, error) {
	ps := make([]*models.Paper, 0, len(ids))
	err := db.store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))
		for _, id := range ids {
			data := b.Get(itob(id))

			var p models.Paper
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

func (db *boltDB) List() ([]*models.Paper, error) {
	var ps []*models.Paper
	err := db.store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))
		c := b.Cursor()

		for id, data := c.First(); id != nil; id, data = c.Next() {
			var p models.Paper
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

func (db *boltDB) Insert(p *models.Paper) error {
	return db.store.Update(func(tx *bolt.Tx) error {
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
}

func (db *boltDB) Update(p *models.Paper) error {
	return db.store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))

		data, err := json.Marshal(p)
		if err != nil {
			return err
		}

		return b.Put(itob(p.ID), data)
	})
}

func (db *boltDB) Delete(id int) error {
	return db.store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))
		return b.Delete(itob(id))
	})
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
