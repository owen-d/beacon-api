package beacons

import (
	"encoding/json"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/route"
	"google.golang.org/api/proximitybeacon/v1beta1"
	"net/http"
)

type BeaconRoutes interface {
	GetBeacons(http.ResponseWriter, *http.Request)
}

type BeaconMethods struct {
	Client beaconclient.Client
}

type beaconResponse struct {
	Beacons []*proximitybeacon.Beacon `json:"beacons"`
}

func (self *BeaconMethods) GetBeacons(rw http.ResponseWriter, r *http.Request) {
	res, _ := self.Client.GetOwnedBeaconNames()
	beacons := res.Beacons

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	data, _ := json.Marshal(beaconResponse{beacons})

	rw.Write(data)
}

func (self *BeaconMethods) Router() *route.Router {
	endpoints := []*route.Endpoint{
		&route.Endpoint{
			Method: "GET",
			Fns:    []http.HandlerFunc{self.GetBeacons},
		},
	}

	r := route.Router{
		Path:      "/beacons",
		Endpoints: endpoints,
	}

	return &r
}
