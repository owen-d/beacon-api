package oauth

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/owen-d/beacon-api/config"
	"github.com/owen-d/beacon-api/lib/crypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"time"
)

func NewOAuthConf(vars *config.OAuth, env string) *oauth2.Config {
	var redirectUri string
	if env == "production" {
		redirectUri = vars.RedirectUris.Prod
	} else {
		redirectUri = vars.RedirectUris.Dev
	}
	return &oauth2.Config{
		ClientID:     vars.ClientID,
		ClientSecret: vars.ClientSecret,
		RedirectURL:  redirectUri,
		Scopes:       vars.Scopes,
		Endpoint:     google.Endpoint,
	}
}

type GoogleAuth interface{}

type GoogleAuthMethods struct {
	OAuth *oauth2.Config
}

func (self *GoogleAuthMethods) getLoginURL(state string) string {
	return self.OAuth.AuthCodeURL(state)
}

func GenState(encoder *crypt.OmniCrypter) (string, error) {
	// default to 5 minutes
	expiry := time.Now().Add(time.Minute * 10).Unix()

	var msgBuf []byte
	binary.PutUvarint(msgBuf, expiry)
	encrypted, err := encoder.encrypt(msgBuf)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(msgBuf), nil
}

func ValidateState(decoder *crypt.OmniCrypter, sig string) error {
	buf, decodeErr := hex.DecodeString(sig)
	// str -> []byte
	if decodeErr != nil {
		return decodeErr
	}

	// decrypt
	decrypted, decryptErr := decoder.decrypt(buf)
	if decryptErr != nil {
		return decryptErr
	}

	// time validation
	expiry, expiryErr := binary.ReadVarint(decrypted)

	if expiryErr != nil {
		return expiryErr
	}

	// expiry check
	if time.Now().After(time.Unix(expiry)) {
		return errors.New("state expired")
	} else {
		return nil
	}

}
