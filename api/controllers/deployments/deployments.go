package deployments

import (
	"encoding/json"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/route"
	"google.golang.org/api/proximitybeacon/v1beta1"
	"net/http"
)

type DeploymentRoutes interface {
	PostDeployment(http.ResponseWriter, *http.Request)
}

type DeploymentMethods struct {
	BeaconClient beaconclient.Client
}

func (self *DeploymentMethods) PostDeployment(rw http.ResponseWriter, r *http.Request) {
	// validate deployment
	// ensure message, schedule exist

	// insert deployment to cassandra (acts as upsert)
	// for each affected beacon, update it, aetting cur_deployment to new deployment, & delete any old deployment references for that beacon.

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
