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
	if self.Message == "" {
		self.Message = http.StatusText(self.Status)
	}
	jsonData, _ := json.Marshal(self)
	// Add headers to header map before flushing them with WriteHeader
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(self.Status)
	rw.Write(jsonData)
}
