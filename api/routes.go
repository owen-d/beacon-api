package api

import (
	"github.com/owen-d/beacon-api/api/controllers"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/route"
)

type Env struct {
	Beaconclient beaconclient.Client
}

func (self *Env) Init() *route.Router {
	beacons := controllers.BeaconMethods{self.Beaconclient}
	r := beacons.Router()
	return route.BuildRouter(r)
}
