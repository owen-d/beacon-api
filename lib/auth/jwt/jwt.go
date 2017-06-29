package jwt

import (
	"errors"
	"fmt"
	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/gocql/gocql"
	"github.com/owen-d/beacon-api/lib/validator"
	"net/http"
)

var (
	jwtKeyword       = "x-jwt"
	JWTNamespace key = key{jwtKeyword}
)

// alias string as a namespaced type to avoid collisions when used w/ context map. Unexported
type key struct{ string }

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
		if castErr := bindings.ConvertFromJwt(claims); castErr != nil {
			return nil, castErr
		}
		return bindings, nil
	} else {
		return nil, errors.New("invalid jwtGo")
	}
}

// Validate ensures a request passes authentication
func (self *Decoder) Validate(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	jwtStr := r.Header.Get(jwtKeyword)

	if jwtStr == "" {
		err := &validator.RequestErr{Status: http.StatusUnauthorized, Message: "Authentication required"}
		err.Flush(rw)
		return
	}

	_, err := self.Decode(jwtStr)

	if err != nil {
		err := &validator.RequestErr{Status: 401, Message: err.Error()}
		err.Flush(rw)
		return
	}

	next(rw, r)

}

// Bindings is an application struct extracted & casted from JWTGO claims
type Bindings struct {
	Token  *jwtGo.Token
	UserId *gocql.UUID
}

// ConvertFromJwtGo casts a *jwtGo.MapClaims into a Bindings struct
func (self *Bindings) ConvertFromJwt(claims jwtGo.MapClaims) error {
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
