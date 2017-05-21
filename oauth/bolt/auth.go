package bolt

import (
	"encoding/json"

	"github.com/boltdb/bolt"

	"github.com/bobinette/papernet/oauth"
)

var authBucket = []byte("auth")

type Authepository struct {
	driver *Driver
}

func NewAuthepository(driver *Driver) *Authepository {
	return &Authepository{
		driver: driver,
	}
}

func (r *Authepository) Get(email string) (oauth.User, error) {
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

func (r *Authepository) Insert(user oauth.User) error {
	return r.driver.store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(googleBucket)

		data, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(user.Email), data)
	})
}
