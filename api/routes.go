package api

import (
	"github.com/gorilla/mux"
	"github.com/owen-d/beacon-api/api/controllers/beacons"
	"github.com/owen-d/beacon-api/api/controllers/deployments"
	"github.com/owen-d/beacon-api/lib/auth/jwt"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/cass"
	"github.com/owen-d/beacon-api/lib/route"
	"github.com/urfave/negroni"
	"net/http"
)

type Env struct {
	BeaconClient beaconclient.Client
	CassClient   cass.Client
	JWTSecret    []byte
}

func (self *Env) Init() http.Handler {

	JWTDecoder := jwt.Decoder{self.JWTSecret}
	beacons := beacons.BeaconMethods{JWTDecoder, self.BeaconClient, self.CassClient}
	deployments := deployments.DeploymentMethods{JWTDecoder, self.BeaconClient, self.CassClient}

	root := &route.Router{
		Path:      "/",
		SubRoutes: []*route.Router{beacons.Router(), deployments.Router()},
	}

	muxRouter := mux.NewRouter()
	route.BuildRouter(root, muxRouter)
	return route.Encase(root.Router, negroni.New(negroni.NewLogger(), route.CorsHandler))
}
