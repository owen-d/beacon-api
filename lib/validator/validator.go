package validator

import (
	"encoding/json"
	"net/http"
)

// Validator must provide a Validate function, which ensures a http request is valid
type JSONValidator interface {
	Validate(*http.Request) (json.Unmarshaler, error)
}

type RequestErr struct {
	Message string `json:"message"`
	status  int    `json:-`
}

func (self *RequestErr) Flush(rw http.ResponseWriter) {
}
