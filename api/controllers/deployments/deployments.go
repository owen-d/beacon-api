package deployments

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/owen-d/beacon-api/api/controllers/beacons"
	"github.com/owen-d/beacon-api/lib/auth/jwt"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/cass"
	"github.com/owen-d/beacon-api/lib/route"
	"github.com/owen-d/beacon-api/lib/validator"
	"github.com/urfave/negroni"
	"google.golang.org/api/proximitybeacon/v1beta1"
	"io/ioutil"
	"net/http"
)

type DeploymentRoutes interface {
	PostDeployment(http.ResponseWriter, *http.Request, http.HandlerFunc)
	FetchDeploymentsMetadata(http.ResponseWriter, *http.Request, http.HandlerFunc)
	FetchDeploymentBeacons(http.ResponseWriter, *http.Request, http.HandlerFunc)
}

type DeploymentMethods struct {
	JWTDecoder   jwt.Decoder
	BeaconClient beaconclient.Client
	CassClient   cass.Client
}

type DeploymentsResponse struct {
	Deployments []*cass.Deployment `json:"deployments"`
}

type IncomingDeployment struct{ *cass.Deployment }

func NewIncomingDep() *IncomingDeployment {
	return &IncomingDeployment{&cass.Deployment{}}
}

// Validate fulfills the validator.JSONValidator interface
func (self *IncomingDeployment) Validate(r *http.Request) *validator.RequestErr {
	// validate deployment
	jsonBody, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		return &validator.RequestErr{400, "invalid json"}
	}

	unmarshalErr := json.Unmarshal(jsonBody, self)
	if unmarshalErr != nil {
		return &validator.RequestErr{Status: 400}
	}

	//assign userId into deployment (forcefully overwrite a potentially malicious userId)
	bindings := r.Context().Value(jwt.JWTNamespace).(*jwt.Bindings)

	self.UserId = bindings.UserId
	return nil
}

// PostDeployment is middleware which creates a deployment (composed of its parts) in cassandra
func (self *DeploymentMethods) PostDeployment(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	deployment := NewIncomingDep()

	if invalid := deployment.Validate(r); invalid != nil {
		invalid.Flush(rw)
		next(rw, r)
		return
	}

	// deployment wraps type cass.Deployment
	cassDep := deployment.Deployment

	// insert deployment to cassandra (acts as upsert)
	res := self.CassClient.PostDeployment(cassDep)
	if res.Err != nil {
		err := &validator.RequestErr{500, res.Err.Error()}
		err.Flush(rw)
		next(rw, r)
		return
	}

	// iterate over affected beacons, removing old attachment & creating new attachment
	attachment := &beaconclient.AttachmentData{
		Title: cassDep.Message.Title,
		Url:   cassDep.Message.Url,
	}
	attachmentResults := self.postAttachments(cass.MapBytesToHex(cassDep.BeaconNames), attachment)

	rw.WriteHeader(http.StatusCreated)

	data, _ := json.Marshal(attachmentResults)
	rw.Write(data)
}

// AttachmentResult is a wrapper type hol,ding response data from google beacon platform about attachment deletions and creations
type AttachmentResult struct {
	Name       string
	Err        error
	Attachment *proximitybeacon.BeaconAttachment `json:-`
}

// postAttachments is a private method which deletes & re-adds attachments to a beacon registered in google's beacon platform
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

func (self *DeploymentMethods) FetchDeploymentsMetadata(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	bindings := r.Context().Value(jwt.JWTNamespace).(*jwt.Bindings)

	mds, fetchErr := self.CassClient.FetchDeploymentsMetadata(bindings.UserId)

	if fetchErr != nil {
		err := &validator.RequestErr{Status: 500, Message: fetchErr.Error()}
		err.Flush(rw)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	data, _ := json.Marshal(DeploymentsResponse{mds})

	rw.Write(data)

}
func (self *DeploymentMethods) FetchDeploymentBeacons(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	bindings := r.Context().Value(jwt.JWTNamespace).(*jwt.Bindings)

	reqVars := mux.Vars(r)
	name, nameExists := reqVars["name"]

	if !nameExists {
		err := &validator.RequestErr{Status: 500}
		err.Flush(rw)
		return
	}

	dep := &cass.Deployment{
		UserId:     bindings.UserId,
		DeployName: name,
	}

	bkns, fetchErr := self.CassClient.FetchDeploymentBeacons(dep)

	if fetchErr != nil {
		err := &validator.RequestErr{Status: 500, Message: fetchErr.Error()}
		err.Flush(rw)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	data, _ := json.Marshal(beacons.BeaconResponse{bkns})

	rw.Write(data)

}

// Router instantiates a Router object from the related lib
func (self *DeploymentMethods) Router() *route.Router {
	endpoints := []*route.Endpoint{
		&route.Endpoint{
			Method:   "GET",
			Handlers: []negroni.Handler{negroni.HandlerFunc(self.FetchDeploymentsMetadata)},
		},
		&route.Endpoint{
			Method:   "POST",
			Handlers: []negroni.Handler{negroni.HandlerFunc(self.PostDeployment)},
		},
		// /:id routes
		&route.Endpoint{
			Method:   "GET",
			Handlers: []negroni.Handler{negroni.HandlerFunc(self.FetchDeploymentBeacons)},
			SubPath:  "/{name}/beacons",
		},
	}

	r := route.Router{
		Path:              "/deployments",
		Endpoints:         endpoints,
		DefaultMiddleware: []negroni.Handler{negroni.HandlerFunc(self.JWTDecoder.Validate)},
		Name:              "deploymentsRouter",
	}

	return &r
}
