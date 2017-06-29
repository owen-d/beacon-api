package api

import (
	"github.com/owen-d/beacon-api/api/controllers/beacons"
	"github.com/owen-d/beacon-api/api/controllers/deployments"
	"github.com/owen-d/beacon-api/lib/auth/jwt"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/cass"
	"github.com/owen-d/beacon-api/lib/route"
)

type Env struct {
	BeaconClient beaconclient.Client
	CassClient   cass.Client
	JWTSecret    []byte
}

func (self *Env) Init() *route.Router {
	JWTDecoder := jwt.Decoder{self.JWTSecret}
	beacons := beacons.BeaconMethods{self.BeaconClient}
	deployments := deployments.DeploymentMethods{JWTDecoder, self.BeaconClient, self.CassClient}

	root := &route.Router{
		Path:      "/",
		SubRoutes: []*route.Router{beacons.Router(), deployments.Router()},
	}

	return route.BuildRouter(root)
}
