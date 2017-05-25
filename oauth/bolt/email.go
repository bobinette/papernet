package bolt

import (
	"encoding/json"

	"github.com/boltdb/bolt"

	"github.com/bobinette/papernet/oauth"
)

var authBucket = []byte("email")

type EmailRepository struct {
	driver *Driver
}

func NewEmailRepository(driver *Driver) *EmailRepository {
	return &EmailRepository{
		driver: driver,
	}
}

func (r *EmailRepository) Get(email string) (oauth.User, error) {
	var user oauth.User
	err := r.driver.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(googleBucket)

		data := bucket.Get([]byte(email))
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &user)
	})
	if err != nil {
		return oauth.User{}, err
	}

	return user, nil
}

func (r *EmailRepository) Insert(user oauth.User) error {
	return r.driver.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(googleBucket)

		data, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(user.Email), data)
	})
}
