package deployments

import (
	"encoding/json"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/route"
	"github.com/owen-d/beacon-api/lib/validator"
	"google.golang.org/api/proximitybeacon/v1beta1"
	"io/ioutil"
	"net/http"
)

type DeploymentRoutes interface {
	ValidateDeployment(http.ResponseWriter, *http.Request)
	PostDeployment(http.ResponseWriter, *http.Request)
}

type DeploymentMethods struct {
	BeaconClient beaconclient.Client
}

type Deployment struct {
	UserId      int
	DeployName  string
	BeaconNames []string
}

func (self *DeploymentMethods) PostDeployment(rw http.ResponseWriter, r *http.Request) {
	// validate deployment
	jsonBody, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	deployment := &Deployment{}
	unmarshalErr := json.Unmarshal(jsonBody, deployment)
	if unmarshalErr {
		validator.RequestErr{400}.Flush(rw)
	}
	// ensure message, schedule exist

	// insert deployment to cassandra (acts as upsert)
	// for each affected beacon, update it, setting cur_deployment to new deployment, & delete any old deployment references for that beacon.

	// iterate over affected beacons, removing old attachment & creating new attachment
	// self.postAttachments()
}

type AttachmentResult struct {
	Name       string
	Err        error
	Attachment *promixitybeacon.BeaconAttachment
}

func (self *DeploymentMethods) postAttachments(bNames []string, attachment *beaconclient.AttachmentData) []error {
	res := make([]*AttachmentResult, len(bNames))

	ch := make(chan AttachmentResult)

	// delete old attachments & apply new one
	for _, bName := range bNames {
		go func(bName string, ch chan<- AttachmentResult) {
			resp := AttachmentResult{bName}

			// remove old attachments on beacon
			_, deleteErr := self.BeaconClient.BatchDeleteAttachments(bName)
			if deleteErr {
				resp.Err = deleteErr
				ch <- resp
				return
			}

			postedAttachment, postErr := self.BeaconClient.CreateAttachment(bName, attachment)

			if postErr {
				resp.Err = postErr
				ch <- resp
				return
			} else {
				resp.Attachment = postedAttachment
			}
		}(bName, ch)
	}

	for i := range bNames {
		res = append(res, <-ch)
	}

	return res
}
