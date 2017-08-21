// Cassandra lib
package cass

import (
	"errors"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type providerId []uint8

func (self providerId) ToUUID() [16]byte {
	return uuid.NewSHA1(uuid.Nil, self)
}

func (self providerId) UUIDFromBytes(data []byte) [16]byte {
	return uuid.NewSHA1(self.ToUUID(), data)
}

var (
	Google providerId = []uint8{1}
)

type User struct {
	Id    *gocql.UUID `cql:"id"`
	Email string      `cql:"email"`
}

func (self *CassClient) CreateUser(u *User, provider providerId, providerKey []byte, batch *gocql.Batch) *UpsertResult {

	uuidBytes := provider.UUIDFromBytes(providerKey)
	uuid, uuidErr := gocql.UUIDFromBytes((&uuidBytes)[:])

	// earl return for uuid generation errors
	if uuidErr != nil {
		if batch != nil {
			return &UpsertResult{
				Batch: batch, Err: uuidErr,
			}
		} else {
			return &UpsertResult{
				Batch: nil, Err: uuidErr,
			}
		}
	}

	template := `INSERT INTO users (id, email) VALUES (?, ?)`
	args := []interface{}{
		&uuid,
		u.Email,
	}

	if batch != nil {
		batch.Query(template, args...)
		return &UpsertResult{Batch: batch, Err: nil}
	} else {
		return &UpsertResult{
			Batch: nil,
			Err:   self.Sess.Query(template, args...).Exec(),
		}
	}

}

func (self *CassClient) FetchUser(u *User) (*User, error) {
	// instantiate user struct for unmarshalling
	matchedUser := &User{}
	var err error
	if u.Id != nil {
		err = self.Sess.Query(`SELECT id, email FROM users WHERE id = ?`, u.Id).Scan(&matchedUser.Id, &matchedUser.Email)
	} else {
		err = self.Sess.Query(`SELECT id, email FROM users_by_email WHERE email = ?`, u.Email).Scan(&matchedUser.Id, &matchedUser.Email)
	}

	if err != nil {
		return nil, err
	}

	if matchedUser.Id != nil {
		return matchedUser, nil
	} else {
		return nil, errors.New("no matched user")
	}
}
