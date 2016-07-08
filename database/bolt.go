package database

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/cayley"
	"github.com/google/cayley/graph"
	_ "github.com/google/cayley/graph/bolt"

	"github.com/bobinette/papernet/models"
)

type boltDB struct {
	store *bolt.DB
	graph *cayley.Handle
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

	// Open graph db next to the main db
	graphDBPath := fmt.Sprintf("%s.cayley", dbpath)
	*graph.IgnoreDup = true
	err = graph.InitQuadStore("bolt", graphDBPath, graph.Options{"ignore_duplicate": true})
	if err != nil && err != graph.ErrDatabaseExists {
		return nil, err
	}
	graph, err := cayley.NewGraph("bolt", graphDBPath, graph.Options{"ignore_duplicate": true})
	if err != nil {
		return nil, err
	}

	return &boltDB{
		store: store,
		graph: graph,
	}, nil
}

func (db *boltDB) Close() error {
	db.graph.Close() // As surprising as it is, Close does not return an error...
	return db.store.Close()
}

// Get retrieves papers from the DB based on their ids
//
// It supposes everything is correctly saved in bolt such that reading is quick
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
	err := db.updateReferences(p)
	if err != nil {
		return err
	}

	err = db.store.Update(func(tx *bolt.Tx) error {
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

	id := strconv.Itoa(p.ID)
	trx := graph.NewTransaction()
	for _, ref := range p.References {
		trx.AddQuad(cayley.Quad(id, "references", strconv.Itoa(ref.ID), ""))
	}
	err = db.graph.ApplyTransaction(trx)
	if err != nil {
		return err
	}

	return nil
}

func (db *boltDB) Update(p *models.Paper) error {
	err := db.updateReferences(p)
	if err != nil {
		return err
	}

	err = db.store.Update(func(tx *bolt.Tx) error {
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

	id := strconv.Itoa(p.ID)
	trx := graph.NewTransaction()
	for _, ref := range p.References {
		trx.AddQuad(cayley.Quad(id, "references", strconv.Itoa(ref.ID), ""))
	}
	// TODO: remove old links
	err = db.graph.ApplyTransaction(trx)
	if err != nil {
		return err
	}

	return nil
}

func (db *boltDB) Delete(id int) error {
	return db.store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))
		return b.Delete(itob(id))
	})
}

func (db *boltDB) updateReferences(p *models.Paper) error {
	err := db.store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("papers"))
		for i, ref := range p.References {
			data := b.Get(itob(ref.ID))

			if len(data) == 0 {
				return fmt.Errorf("no paper for id %d", ref.ID)
			}

			var rp models.Paper
			if err := json.Unmarshal(data, &rp); err != nil {
				return err
			}

			p.References[i].Title = rp.Title
			fmt.Printf("%+v", p)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// TODO: update papers referencing this paper

	return nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
