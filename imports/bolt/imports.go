package bolt

import (
	"encoding/binary"
	"encoding/json"

	"github.com/boltdb/bolt"
)

var importsBucket = []byte("imports")

type mapping map[string]map[string]int

type PaperRepository struct {
	driver *Driver
}

func NewPaperRepository(driver *Driver) *PaperRepository {
	return &PaperRepository{
		driver: driver,
	}
}

func (r *PaperRepository) Get(userID int, source, ref string) (int, error) {
	var m mapping
	err := r.driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(importsBucket)

		data := bucket.Get(itob(userID))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &m)
	})

	if err != nil {
		return 0, err
	}

	return m[source][ref], nil
}

func (r *PaperRepository) Save(userID, paperID int, source, ref string) error {
	return r.driver.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(importsBucket)

		var m mapping
		data := bucket.Get(itob(userID))
		if data == nil {
			m = make(map[string]map[string]int)
		} else if err := json.Unmarshal(data, &m); err != nil {
			return err
		}

		references := m[source]
		if references == nil {
			references = make(map[string]int)
		}
		references[ref] = paperID
		m[source] = references

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		return bucket.Put(itob(userID), data)
	})
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
