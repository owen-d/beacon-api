package api

import (
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/owen-d/beacon-api/api/controllers/beacons"
	"github.com/owen-d/beacon-api/api/controllers/deployments"
	"github.com/owen-d/beacon-api/api/controllers/messages"
	"github.com/owen-d/beacon-api/api/controllers/oauth"
	"github.com/owen-d/beacon-api/config"
	"github.com/owen-d/beacon-api/lib/auth/jwt"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/cass"
	"github.com/owen-d/beacon-api/lib/crypt"
	"github.com/owen-d/beacon-api/lib/route"
	"github.com/urfave/negroni"
	"log"
	"net"
	"net/http"
)

type Env struct {
	Conf *config.JsonConfig
}

func (self *Env) Init() http.Handler {

	JWTDecoder := jwt.Decoder{[]byte(self.Conf.JWTSecret)}
	JWTEncoder := jwt.Encoder{[]byte(self.Conf.JWTSecret)}

	httpClient := beaconclient.JWTConfigFromJSON(self.Conf.GCloudConfigPath, self.Conf.Scope)
	svc, bknClientErr := beaconclient.NewBeaconClient(httpClient)
	safeExit(bknClientErr)

	cassClient := createCassClient(self.Conf.CassKeyspace, self.Conf.CassEndpoint)

	beacons := beacons.BeaconMethods{JWTDecoder, svc, cassClient}
	deployments := deployments.DeploymentMethods{JWTDecoder, svc, cassClient}
	messages := messages.MessageMethods{JWTDecoder, cassClient}

	googleCrypter, googleCrypterErr := crypt.NewOmniCrypter(self.Conf.GoogleOAuth.StateKey)
	safeExit(googleCrypterErr)
	googleOAuth := oauth.GoogleAuthMethods{
		OAuth:      oauth.NewOAuthConf(&self.Conf.GoogleOAuth),
		Coder:      googleCrypter,
		CassClient: cassClient,
		JWTEncoder: &JWTEncoder,
	}

	v1Router := &route.Router{
		Path:      "/v1",
		SubRoutes: []*route.Router{beacons.Router(), deployments.Router(), messages.Router(), googleOAuth.Router()},
	}

	// default root handler (for healthchecks/welcome msg)
	root := mux.NewRouter()
	root.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte("welcome to the sharecrows api"))
	})

	root = route.Inject(v1Router, root)
	return negroni.New(negroni.NewLogger(), route.CorsHandler, negroni.Wrap(root))
}

func createCassClient(keyspace string, address string) *cass.CassClient {
	if address == "" {
		address = "localhost"
	}

	addrs, lookupErr := net.LookupHost(address)
	if lookupErr != nil {
		log.Fatal("couldn't match cassandra host:\n", lookupErr)
	}

	cluster := gocql.NewCluster(addrs...)
	client, err := cass.Connect(cluster, keyspace)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func safeExit(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
