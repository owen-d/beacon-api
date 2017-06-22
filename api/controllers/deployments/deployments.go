package deployments

import (
	"encoding/json"
	"github.com/gocql/gocql"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/cass"
	"github.com/owen-d/beacon-api/lib/route"
	"github.com/owen-d/beacon-api/lib/validator"
	"google.golang.org/api/proximitybeacon/v1beta1"
	"io/ioutil"
	"net/http"
)

type DeploymentRoutes interface {
	PostDeployment(http.ResponseWriter, *http.Request)
}

type DeploymentMethods struct {
	BeaconClient beaconclient.Client
	CassClient   cass.Client
}

type Deployment struct {
	UserId      string   `json:"user_id"`
	Name        string   `json:"name"`
	BeaconNames []string `json:beacon_names"`
	MessageName string   `json:"message_name`
	Message     *Message `json:"message"`
}

type Message struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Url   string `json:"url"`
}

// Validate fulfills the validator.JSONValidator interface
func (self *Deployment) Validate(r *http.Request) *validator.RequestErr {
	// validate deployment
	jsonBody, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		return &validator.RequestErr{400, "invalid json"}
	}

	unmarshalErr := json.Unmarshal(jsonBody, self)
	if unmarshalErr != nil {
		return &validator.RequestErr{Status: 400}
	}

	return nil
}

// ToCass coerces a Deployment into the cassandra lib version (mainly handling uuid conversions)
func (self *Deployment) ToCass() (*cass.Deployment, error) {
	userId, parseErr := gocql.ParseUUID(self.UserId)

	if parseErr != nil {
		return nil, parseErr
	}

	cassDep := &cass.Deployment{
		UserId:      &userId,
		DeployName:  self.Name,
		BeaconNames: self.BeaconNames,
	}

	if self.MessageName != "" {
		cassDep.MessageName = self.MessageName
	}

	if self.Message != nil {
		cassDep.Message = &cass.Message{
			Name:  self.Message.Name,
			Title: self.Message.Title,
			Url:   self.Message.Url,
		}
	}

	return cassDep, nil
}

func (self *DeploymentMethods) PostDeployment(rw http.ResponseWriter, r *http.Request) {

	deployment := &Deployment{}

	if invalid := deployment.Validate(r); invalid != nil {
		invalid.Flush(rw)
		return
	}

	cassDep, castErr := deployment.ToCass()

	if castErr != nil {
		err := &validator.RequestErr{500, castErr.Error()}
		err.Flush(rw)
		return
	}

	// insert deployment to cassandra (acts as upsert)
	res := self.CassClient.PostDeployment(cassDep)
	if res.Err != nil {
		err := &validator.RequestErr{500, res.Err.Error()}
		err.Flush(rw)
		return
	}

	// iterate over affected beacons, removing old attachment & creating new attachment
	attachment := &beaconclient.AttachmentData{
		Title: cassDep.Message.Title,
		Url:   cassDep.Message.Url,
	}
	attachmentResults := self.postAttachments(cassDep.BeaconNames, attachment)

	rw.WriteHeader(http.StatusCreated)

	data, _ := json.Marshal(attachmentResults)
	rw.Write(data)
}

type AttachmentResult struct {
	Name       string
	Err        error
	Attachment *proximitybeacon.BeaconAttachment
}

func (self *DeploymentMethods) postAttachments(bNames []string, attachment *beaconclient.AttachmentData) []*AttachmentResult {
	res := make([]*AttachmentResult, 0, len(bNames))

	ch := make(chan *AttachmentResult)

	// delete old attachments & apply new one
	for _, bName := range bNames {
		go func(bName string, ch chan<- *AttachmentResult) {
			resp := &AttachmentResult{Name: bName}

			// remove old attachments on beacon
			_, deleteErr := self.BeaconClient.BatchDeleteAttachments(bName)
			if deleteErr != nil {
				resp.Err = deleteErr
				ch <- resp
				return
			}

			postedAttachment, postErr := self.BeaconClient.CreateAttachment(bName, attachment)

			if postErr != nil {
				resp.Err = postErr
				ch <- resp
				return
			} else {
				resp.Attachment = postedAttachment
				ch <- resp
				return
			}
		}(bName, ch)
	}

	for range bNames {
		res = append(res, <-ch)
	}

	return res
}

func (self *DeploymentMethods) Router() *route.Router {
	endpoints := []*route.Endpoint{
		// &route.Endpoint{
		// 	Method: "GET",
		// 	Fns:    []http.HandlerFunc{self.GetBeacons},
		// },
		&route.Endpoint{
			Method: "POST",
			Fns:    []http.HandlerFunc{self.PostDeployment},
		},
	}

	r := route.Router{
		Path:      "/deployments",
		Endpoints: endpoints,
	}

	return &r
}
