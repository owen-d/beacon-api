package jwtLib

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gocql/gocql"
)

// JWT is a wrapper struct for jwt & related fields
type JWT struct {
	Str    string
	Secret []byte
	Token  *jwt.Token
}

// Bindings is an application struct extracted & casted from JWT claims
type Bindings struct {
	UserId *gocql.UUID
}

// ConvertFromJwt casts a *jwt.MapClaims into a Bindings struct
func (self *Bindings) ConvertFromJwt(claims *jwt.MapClaims) error {
	idStr := claims["user_id"]
	if !idStr {
		return errors.New("no user_id field")
	}

	userId, parseErr := gocql.ParseUUID(idStr)
	if parseErr != nil {
		return parseErr
	}

	self.UserId = &userId
}

// Validtae ensures a JWT uses a valid signature & contains the correct data, returning *Bindings with the result & optionally an error
func (self *JWT) Validate() (*Bindings, error) {
	token, parseErr := jwt.Parse(self.Str, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return self.Secret, nil
	})

	if parseErr != nil {
		return nil, parseErr
	}

	self.Token = token

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		bindings := &Bindings{}
		if castErr := bindings.ConvertFromJwt(claims); castErr != nil {
			return nil, castErr
		}
		return bindings, nil
	} else {
		return nil, errors.New("invalid jwt")
	}
}
