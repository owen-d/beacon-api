package jwt

import (
	"errors"
	"fmt"
	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/gocql/gocql"
)

// Decoder is a wrapper struct which handles decoding
type Decoder struct {
	Secret []byte
}

// Decode parses a jwt and produces a relevant application bindings struct
func (self *Decoder) Decode(unparsed string) (*Bindings, error) {
	token, parseErr := jwtGo.Parse(unparsed, func(token *jwtGo.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtGo.SigningMethodHMAC); !ok {
			return nil, errors.New(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
		}

		return self.Secret, nil
	})

	if parseErr != nil {
		return nil, parseErr
	}

	bindings := &Bindings{Token: token}

	if claims, ok := token.Claims.(jwtGo.MapClaims); ok && token.Valid {
		if castErr := bindings.ConvertFromJwtGo(claims); castErr != nil {
			return nil, castErr
		}
		return bindings, nil
	} else {
		return nil, errors.New("invalid jwtGo")
	}
}

// Bindings is an application struct extracted & casted from JWTGO claims
type Bindings struct {
	Token  *jwtGo.Token
	UserId *gocql.UUID
}

// ConvertFromJwtGo casts a *jwtGo.MapClaims into a Bindings struct
func (self *Bindings) ConvertFromJwtGo(claims jwtGo.MapClaims) error {
	idStr := claims["user_id"].(string)
	if idStr != "" {
		return errors.New("no user_id field")
	}

	userId, parseErr := gocql.ParseUUID(idStr)
	if parseErr != nil {
		return parseErr
	}

	self.UserId = &userId
	return nil
}
