package beacons

import (
	"encoding/json"
	"github.com/owen-d/beacon-api/lib/auth/jwt"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/cass"
	"github.com/owen-d/beacon-api/lib/route"
	"github.com/owen-d/beacon-api/lib/validator"
	"github.com/urfave/negroni"
	"net/http"
)

type BeaconRoutes interface {
	GetBeacons(http.ResponseWriter, *http.Request, http.HandlerFunc)
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

func (self *BeaconMethods) Router() *route.Router {
	endpoints := []*route.Endpoint{
		&route.Endpoint{
			Method:   "GET",
			Handlers: []negroni.Handler{negroni.HandlerFunc(self.GetBeacons)},
		},
	}

	r := route.Router{
		Path:              "/beacons",
		Endpoints:         endpoints,
		DefaultMiddleware: []negroni.Handler{negroni.HandlerFunc(self.JWTDecoder.Validate)},
	}

	return &r
}
