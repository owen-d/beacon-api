package validator

import (
	"encoding/json"
	"net/http"
)

// Validator must provide a Validate function, which ensures a http request is valid
type JSONValidator interface {
	Validate(*http.Request) *RequestErr
}

type RequestErr struct {
	Status  int    `json:-`
	Message string `json:"message"`
}

func (self *RequestErr) Flush(rw http.ResponseWriter) {
	rw.WriteHeader(self.Status)
	if self.Message == "" {
		self.Message = http.StatusText(self.Status)
	}
	jsonData, _ := json.Marshal(self)
	rw.Write(jsonData)
}
