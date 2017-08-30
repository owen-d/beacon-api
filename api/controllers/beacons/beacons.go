package beacons

import (
	"encoding/json"
	"github.com/gocql/gocql"
	"github.com/owen-d/beacon-api/lib/auth/jwt"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/cass"
	"github.com/owen-d/beacon-api/lib/route"
	"github.com/owen-d/beacon-api/lib/validator"
	"github.com/urfave/negroni"
	"io/ioutil"
	"net/http"
)

type BeaconRoutes interface {
	GetBeacons(http.ResponseWriter, *http.Request, http.HandlerFunc)
	ChangeDeployments(http.ResponseWriter, *http.Request, http.HandlerFunc)
	// RegisterBeacons(http.ResponseWriter, *http.Request, http.HandlerFunc)
	// DeregisterBeacons(http.ResponseWriter, *http.Request, http.HandlerFunc)
	// UpdateBeacons(http.ResponseWriter, *http.Request, http.HandlerFunc)
}

type BeaconMethods struct {
	JWTDecoder   jwt.Decoder
	BeaconClient beaconclient.Client
	CassClient   cass.Client
}

type BeaconResponse struct {
	Beacons []*cass.Beacon `json:"beacons"`
}

type IncBeacons struct {
	Beacons []*cass.Beacon `json:"beacons"`
}

func (self *IncBeacons) Validate(r *http.Request) ([]*cass.Beacon, *validator.RequestErr) {
	jsonBody, readErr := ioutil.ReadAll(r.Body)

	if readErr != nil {
		return nil, &validator.RequestErr{400, "invalid json"}
	}

	unmarshalErr := json.Unmarshal(jsonBody, self)

	if unmarshalErr != nil {
		return nil, &validator.RequestErr{Status: 400, Message: "invalid json"}
	}

	bindings := r.Context().Value(jwt.JWTNamespace).(*jwt.Bindings)

	// potentially overwrite malicious userId
	for _, bkn := range self.Beacons {
		bkn.UserId = bindings.UserId
	}

	return self.Beacons, nil
}

func (self *BeaconMethods) GetBeacons(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	bindings := r.Context().Value(jwt.JWTNamespace).(*jwt.Bindings)
	beacons, fetchErr := self.CassClient.FetchUserBeacons(bindings.UserId)

	if fetchErr != nil {
		err := &validator.RequestErr{Status: 500, Message: fetchErr.Error()}
		err.Flush(rw)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	data, _ := json.Marshal(BeaconResponse{beacons})

	rw.Write(data)
}

func (self *BeaconMethods) ChangeDeployments(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	bkns, validationErr := (&IncBeacons{}).Validate(r)

	if validationErr != nil {
		validationErr.Flush(rw)
		return
	}

	additions := make([]*cass.Beacon, 0)
	removals := make([]*cass.Beacon, 0)
	deploymentGrps := make(map[string][]*cass.Beacon)

	for _, bkn := range bkns {
		if bkn.DeployName == "" {
			removals = append(removals, bkn)
		} else {
			additions = append(additions, bkn)
			// add the beacon to the correct deployment grouping
			slice, ok := deploymentGrps[bkn.DeployName]
			if ok {
				slice = append(slice, bkn)
			} else {
				slice = []*cass.Beacon{bkn}
			}
		}
	}

	removalRes := self.CassClient.RemoveBeaconsDeployments(removals)
	additionRes := self.CassClient.UpdateBeacons(additions)

	if removalRes.Err != nil {
		(&validator.RequestErr{Status: 500, Message: removalRes.Err.Error()}).Flush(rw)
		return
	}

	if additionRes.Err != nil {
		(&validator.RequestErr{Status: 500, Message: additionRes.Err.Error()}).Flush(rw)
		return
	}

	// iterate over affected beacons & update proximity api.
	bindings := r.Context().Value(jwt.JWTNamespace).(*jwt.Bindings)
	attached, errs := self.handleDeploymentGroups(bindings.UserId, deploymentGrps)

	rw.Header().Set("Content-Type", "application/json")

	if len(errs) != 0 {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		rw.WriteHeader(http.StatusOK)
	}

	data, _ := json.Marshal(struct {
		Attached []string `json:"attached"`
		Errors   []error  `json:"errors"`
	}{attached, errs})

	rw.Write(data)
}

func (self *BeaconMethods) handleDeploymentGroups(userId *gocql.UUID, depGrps map[string][]*cass.Beacon) ([]string, []error) {
	resErrs := make([]error, 0)
	attachResults := make([]*beaconclient.AttachmentResult, 0)

	for depName, bkns := range depGrps {
		// fetch metatdata & then msg
		match, matchErr := self.CassClient.FetchDeploymentMetadata(userId, depName)
		if matchErr != nil {
			resErrs = append(resErrs, matchErr)
			break
		}

		if match.MessageName == "" {
			break
		}

		msg, msgErr := self.CassClient.FetchMessage(&cass.Message{
			UserId: userId,
			Name:   match.MessageName,
		})

		if msgErr != nil {
			resErrs = append(resErrs, msgErr)
			break
		}

		attachment := &beaconclient.AttachmentData{
			Title: msg.Title,
			Url:   msg.Url,
		}

		bknNames := make([][]byte, 0)
		for _, bkn := range bkns {
			bknNames = append(bknNames, bkn.Name)
		}

		results := self.BeaconClient.DeclarativeAttach(cass.MapBytesToHex(bknNames), attachment)
		attachResults = append(attachResults, results...)

	}

	successes := make([]string, 0)
	for _, attachResult := range attachResults {
		if attachResult.Err != nil {
			resErrs = append(resErrs, attachResult.Err)
		} else {
			successes = append(successes, attachResult.Name)
		}
	}

	return successes, resErrs
}

func (self *BeaconMethods) Router() *route.Router {
	endpoints := []*route.Endpoint{
		&route.Endpoint{
			Method:   "GET",
			Handlers: []negroni.Handler{negroni.HandlerFunc(self.GetBeacons)},
		},
		&route.Endpoint{
			Method:   http.MethodPut,
			Handlers: []negroni.Handler{negroni.HandlerFunc(self.ChangeDeployments)},
			SubPath:  "/deployments",
		},
	}

	r := route.Router{
		Path:              "/beacons",
		Endpoints:         endpoints,
		DefaultMiddleware: []negroni.Handler{negroni.HandlerFunc(self.JWTDecoder.Validate)},
		Name:              "beaconRouter",
	}

	return &r
}
