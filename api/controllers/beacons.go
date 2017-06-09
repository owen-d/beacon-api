package controllers

import (
	"net/http"
)

type BeaconRoutes interface {
	GetBeacons(http.ResponseWriter, *http.Request)
}

type BeaconMethods struct {
	Client *beaconclient.Client
}

func (self *BeaconMethods) GetBeacons(rw http.ResponseWriter, r *http.Request) {
	res, err := self.Client.GetOwnedBeaconNames()
}
