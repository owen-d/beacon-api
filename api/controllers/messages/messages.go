package messages

import (
	"encoding/json"
	"github.com/gocql/gocql"
	"github.com/owen-d/beacon-api/lib/auth/jwt"
	"github.com/owen-d/beacon-api/lib/cass"
	"github.com/owen-d/beacon-api/lib/route"
	"github.com/owen-d/beacon-api/lib/validator"
	"github.com/urfave/negroni"
	"io/ioutil"
	"net/http"
)

type MessagesRoutes interface {
	FetchMessages(http.ResponseWriter, *http.Request, http.HandlerFunc)
	PostMessage(http.ResponseWriter, *http.Request, http.HandlerFunc)
	UpdateMessage(http.ResponseWriter, *http.Request, http.HandlerFunc)
	DeleteMessage(http.ResponseWriter, *http.Request, http.HandlerFunc)
}

type MessageMethods struct {
	JWTDecoder jwt.Decoder
	CassClient cass.Client
}

type MessagesResponse struct {
	Messages []*cass.Message `json:"messages"`
}

type IncomingMessage struct {
	UserId *gocql.UUID `json:-`
	Name   string      `cql:"name" json:"name"`
	Title  string      `cql:"title" json:"title"`
	Url    string      `cql:"url" json:"url"`
}

// Validate fulfills the validator.JSONValidator interface
func (self *IncomingMessage) Validate(r *http.Request) *validator.RequestErr {
	// validate msg
	jsonBody, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		return &validator.RequestErr{400, "invalid json"}
	}

	unmarshalErr := json.Unmarshal(jsonBody, self)
	if unmarshalErr != nil {
		return &validator.RequestErr{Status: 400}
	}

	//assign userId into msg (forcefully overwrite a potentially malicious userId)
	bindings := r.Context().Value(jwt.JWTNamespace).(*jwt.Bindings)

	self.UserId = bindings.UserId
	return nil
}

// ToCass coerces a msg into the cassandra lib version
func (self *IncomingMessage) ToCass() (*cass.Message, error) {

	cassMsg := &cass.Message{
		UserId:      self.UserId,
		Name:        self.Name,
		Title:       self.Title,
		Url:         self.Url,
		Lang:        "en",
		Deployments: []string{},
	}

	return cassMsg, nil
}

func (self *MessageMethods) PostMessage(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	msg := &IncomingMessage{}

	if invalid := msg.Validate(r); invalid != nil {
		invalid.Flush(rw)
		next(rw, r)
		return
	}

	cassMsg, castErr := msg.ToCass()

	if castErr != nil {
		err := &validator.RequestErr{500, castErr.Error()}
		err.Flush(rw)
		next(rw, r)
		return
	}

	// insert msg to cassandra (acts as upsert)
	res := self.CassClient.CreateMessage(cassMsg, nil)
	if res.Err != nil {
		err := &validator.RequestErr{500, res.Err.Error()}
		err.Flush(rw)
		next(rw, r)
		return
	}

}

func (self *MessageMethods) FetchMessages(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	bindings := r.Context().Value(jwt.JWTNamespace).(*jwt.Bindings)

	msgs, fetchErr := self.CassClient.FetchMessages(bindings.UserId, cass.DefaultLimit)

	if fetchErr != nil {
		err := &validator.RequestErr{Status: 500, Message: fetchErr.Error()}
		err.Flush(rw)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	data, _ := json.Marshal(MessagesResponse{msgs})

	rw.Write(data)

}

// Router instantiates a Router object from the related lib
func (self *MessageMethods) Router() *route.Router {
	endpoints := []*route.Endpoint{
		&route.Endpoint{
			Method:   "GET",
			Handlers: []negroni.Handler{negroni.HandlerFunc(self.FetchMessages)},
		},
	}

	r := route.Router{
		Path:              "/messages",
		Endpoints:         endpoints,
		DefaultMiddleware: []negroni.Handler{negroni.HandlerFunc(self.JWTDecoder.Validate)},
		Name:              "messagesRouter",
	}

	return &r
}
