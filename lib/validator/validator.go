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
	status  int    `json:-`
	Message string `json:"message"`
}

func (self *RequestErr) Flush(rw http.ResponseWriter) {
	rw.WriteHeader(self.status)
	if self.Message == nil {
		self.Message = http.StatusText(self.status)
	}
	jsonData, _ := json.Marshal(self)
	rw.Write(jsonData)
}
