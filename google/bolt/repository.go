package bolt

import (
	"encoding/binary"
	"encoding/json"

	"github.com/boltdb/bolt"

	"github.com/bobinette/papernet/google"
)

var googleBucket = []byte("google")

type UserRepository struct {
	driver *Driver
}

func NewUserRepository(driver *Driver) *UserRepository {
	return &UserRepository{
		driver: driver,
	}
}

func (r *UserRepository) GetByID(id int) (google.User, error) {
	var user google.User
	err := r.driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(googleBucket)

		data := bucket.Get(itob(id))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &user)
	})
	if err != nil {
		return google.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByGoogleID(googleID string) (google.User, error) {
	var user google.User
	err := r.driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(googleBucket)

		data := bucket.Get([]byte(googleID))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &user)
	})
	if err != nil {
		return google.User{}, err
	}

	return user, nil
}

func (r *UserRepository) Upsert(user google.User) error {
	return r.driver.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(googleBucket)
		data, err := json.Marshal(user)
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte(user.GoogleID), data); err != nil {
			return err
		}

		if err := bucket.Put(itob(user.ID), data); err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}

		return nil
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
