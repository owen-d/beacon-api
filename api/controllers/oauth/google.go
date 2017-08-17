package oauth

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/owen-d/beacon-api/config"
	"github.com/owen-d/beacon-api/lib/crypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	userinfoEndpoint = "https://www.googleapis.com/oauth2/v3/userinfo"
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

type GoogleAuth interface {
	HandleAuth(http.ResponseWriter, *http.Request, http.HandlerFunc)
}

type GoogleAuthMethods struct {
	OAuth *oauth2.Config
	Coder *crypt.OmniCrypter
}

// Exchange validates a state string & exchanges the code for a token
func (self *GoogleAuthMethods) exchange(state string, code string) (*oauth2.Token, error) {
	if validationErr := ValidateState(self.Coder, state); validationErr != nil {
		return nil, validationErr
	}

	return self.OAuth.Exchange(oauth2.NoContext, code)
}

// getUser will use a token to retrieve a corresponding user
func (self *GoogleAuthMethods) getUser(tok *oauth2.Token) (*GoogleUser, error) {
	client := self.OAuth.Client(oauth2.NoContext, tok)

	resp, fetchErr := client.Get(userinfoEndpoint)
	if fetchErr != nil {
		return nil, fetchErr
	}
	defer resp.Body.Close()

	data, readErr := ioutil.ReadAll(resp.Body)
	fmt.Printf(string(data))

	if readErr != nil {
		return nil, readErr
	}

	user := &GoogleUser{}
	return user, json.Unmarshal(data, user)

}

// func to handle redirect w/ state & code params. validate state & exchange code for user
func (self *GoogleAuthMethods) HandleAuth(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
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

type GoogleUser struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
}
