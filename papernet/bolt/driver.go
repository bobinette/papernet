package bolt

import (
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

type Driver struct {
	store *bolt.DB
}

// Open opens the connection to the bolt database defined by path.
func (d *Driver) Open(path string) error {
	if d.store != nil {
		return errors.New("store alread open")
	}

	store, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	err = store.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			paperBucket,
			tagBucket,
		}
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return err
			}
		}

		return nil
	})

	d.store = store
	return nil
}

// Close closes the underlying database.
func (d *Driver) Close() error {
	if d.store != nil {
		err := d.store.Close()
		d.store = nil
		return err
	}
	return nil
}
