package oauth

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/owen-d/beacon-api/config"
	"github.com/owen-d/beacon-api/lib/cass"
	"github.com/owen-d/beacon-api/lib/crypt"
	"github.com/owen-d/beacon-api/lib/route"
	"github.com/owen-d/beacon-api/lib/validator"
	"github.com/urfave/negroni"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	userinfoEndpoint = "https://www.googleapis.com/oauth2/v3/userinfo"
)

func NewOAuthConf(vars *config.OAuth) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     vars.ClientID,
		ClientSecret: vars.ClientSecret,
		RedirectURL:  vars.RedirectUri,
		Scopes:       vars.Scopes,
		Endpoint:     google.Endpoint,
	}
}

type GoogleAuth interface {
	HandleAuth(http.ResponseWriter, *http.Request, http.HandlerFunc)
	Redirect(http.ResponseWriter, *http.Request, http.HandlerFunc)
}

type GoogleAuthMethods struct {
	OAuth      *oauth2.Config
	Coder      *crypt.OmniCrypter
	CassClient cass.Client
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

// loginUser upserts a user
func (self *GoogleAuthMethods) loginUser(user *GoogleUser) *cass.UpsertResult {
	return self.CassClient.CreateUser(user.ToCass(), cass.Google, []byte(user.Sub), nil)
}

// HandleAuth handles redirect w/ state & code params. validate state & exchange code for user
func (self *GoogleAuthMethods) HandleAuth(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	params := r.URL.Query()

	codes, okcode := params["code"]
	states, okstate := params["state"]

	if !okstate || !okcode {
		err := &validator.RequestErr{Status: http.StatusBadRequest, Message: "requires state & code params"}
		err.Flush(rw)
		return
	}

	token, exchangeErr := self.exchange(states[0], codes[0])

	if exchangeErr != nil {
		err := &validator.RequestErr{Status: http.StatusInternalServerError, Message: exchangeErr.Error()}
		err.Flush(rw)
		return
	}

	googleUser, userExchangeErr := self.getUser(token)

	if userExchangeErr != nil {
		err := &validator.RequestErr{Status: http.StatusInternalServerError, Message: userExchangeErr.Error()}
		err.Flush(rw)
		return
	}

	res := self.loginUser(googleUser)

	if res.Err != nil {
		err := &validator.RequestErr{Status: http.StatusInternalServerError, Message: res.Err.Error()}
		err.Flush(rw)
		return
	}

	rw.Write([]byte("created user"))

}

// RedirectToGoogle generates a state & redirects to the correct google login endpoint
func (self *GoogleAuthMethods) Redirect(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	state, stateErr := GenState(self.Coder)
	if stateErr != nil {
		err := &validator.RequestErr{Status: 500, Message: stateErr.Error()}
		err.Flush(rw)
		return
	}
	redirectUrl := self.getLoginURL(state)
	http.Redirect(rw, r, redirectUrl, http.StatusFound)
}

func (self *GoogleAuthMethods) getLoginURL(state string) string {
	return self.OAuth.AuthCodeURL(state)
}

func GenState(encoder *crypt.OmniCrypter) (string, error) {
	// default to 5 minutes
	expiry := time.Now().Add(time.Minute * 10).Unix()

	// 64 bit int = 8 bytes (assuming there isn't some leading/trailing identifiers)
	// we then add 1 byte, to facilitate 'variable-length encoding'
	// see https://medium.com/go-walkthrough/go-walkthrough-encoding-binary-96dc5d4abb5d
	msgBuf := make([]byte, 9)
	binary.PutVarint(msgBuf, expiry)
	encrypted, err := encoder.Encrypt(msgBuf)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encrypted), nil
}

func ValidateState(decoder *crypt.OmniCrypter, sig string) error {
	buf, decodeErr := hex.DecodeString(sig)
	// str -> []byte
	if decodeErr != nil {
		return decodeErr
	}

	// decrypt
	decrypted, decryptErr := decoder.Decrypt(buf)
	if decryptErr != nil {
		return decryptErr
	}

	// time validation
	expiry, expiryErr := binary.ReadVarint(bytes.NewReader(decrypted))

	if expiryErr != nil {
		return expiryErr
	}

	// expiry check
	if time.Now().After(time.Unix(expiry, 0)) {
		return errors.New("state expired")
	} else {
		return nil
	}

}

type GoogleUser struct {
	Sub           string `json:"sub"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
}

func (self *GoogleUser) ToCass() *cass.User {
	return &cass.User{
		Email:            self.Email,
		GivenName:        self.GivenName,
		FamilyName:       self.FamilyName,
		PublicPictureUrl: self.Picture,
	}
}

func (self *GoogleAuthMethods) Router() *route.Router {
	endpoints := []*route.Endpoint{
		&route.Endpoint{
			Method:   http.MethodGet,
			Handlers: []negroni.Handler{negroni.HandlerFunc(self.Redirect)},
			SubPath:  "/init",
		},
		&route.Endpoint{
			Method:   http.MethodGet,
			Handlers: []negroni.Handler{negroni.HandlerFunc(self.HandleAuth)},
			SubPath:  "/authorize",
		},
	}

	r := route.Router{
		Path:      "/auth/google",
		Endpoints: endpoints,
		Name:      "beaconRouter",
	}

	return &r
}
