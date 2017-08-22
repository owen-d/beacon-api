// Cassandra lib
package cass

import (
	"errors"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"time"
)

type providerId uint8

func (self providerId) ToUUID() [16]byte {
	var selfAsUint uint8 = uint8(self)
	return uuid.NewSHA1(uuid.Nil, []uint8{selfAsUint})
}

func (self providerId) Unwrap() uint8 {
	return uint8(self)
}

func (self providerId) UUIDFromBytes(data []byte) [16]byte {
	return uuid.NewSHA1(self.ToUUID(), data)
}

var (
	Google providerId = 1
)

type User struct {
	Id               *gocql.UUID `cql:"id" json:"id"`
	Email            string      `cql:"email" json:"email"`
	CreatedAt        time.Time   `cql:"created_at" json:"-"`
	UpdatedAt        time.Time   `cql:"updated_at" json:"-"`
	ProviderId       uint8       `cql:"provider_id" json:"-"`
	GivenName        string      `cql:"given_name" json:"given_name"`
	FamilyName       string      `cql:"family_name" json:"family_name"`
	PublicPictureUrl string      `cql:"public_picture_url json:"public_picture_url"`
}

func (self *CassClient) CreateUser(u *User, provider providerId, providerKey []byte, batch *gocql.Batch) *UpsertResult {

	uuidBytes := provider.UUIDFromBytes(providerKey)
	uuid, uuidErr := gocql.UUIDFromBytes((&uuidBytes)[:])
	u.Id = &uuid

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

	template := `INSERT INTO users (id, provider_id, email, given_name, family_name, public_picture_url, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	args := []interface{}{
		&uuid,
		// unwrap yields provider's id
		provider.Unwrap(),
		u.Email,
		u.GivenName,
		u.FamilyName,
		u.PublicPictureUrl,
		time.Now(),
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
