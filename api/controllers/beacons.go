package controllers

import (
	"fmt"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/route"
	"net/http"
)

type BeaconRoutes interface {
	GetBeacons(http.ResponseWriter, *http.Request)
}

type BeaconMethods struct {
	Client beaconclient.Client
}

func (self *BeaconMethods) GetBeacons(rw http.ResponseWriter, r *http.Request) {
	res, _ := self.Client.GetOwnedBeaconNames()
	beacons := res.Beacons
	fmt.Printf("found beacons:\n%+v", beacons)
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
