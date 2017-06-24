package bolt

import (
	"encoding/binary"

	"github.com/boltdb/bolt"
)

var googleBucket = []byte("google")

type GoogleRepository struct {
	driver *Driver
}

func NewGoogleRepository(driver *Driver) *GoogleRepository {
	return &GoogleRepository{
		driver: driver,
	}
}

func (r *GoogleRepository) Get(googleID string) (int, error) {
	var id int
	err := r.driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(googleBucket)

		data := bucket.Get([]byte(googleID))
		if data == nil {
			id = 0
			return nil
		}

		id = btoi(data)
		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *GoogleRepository) Insert(googleID string, id int) error {
	return r.driver.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(googleBucket)
		return bucket.Put([]byte(googleID), itob(id))
	})
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func btoi(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}
