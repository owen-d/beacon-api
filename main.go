package main

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/owen-d/beacon-api/api"
	"github.com/owen-d/beacon-api/config"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/cass"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var (
	// for loading dev config
	defaultConfigsDir = filepath.Join(os.Getenv("GOPATH"), "src/github.com/owen-d/beacon-api/conf/dev-settings")
)

func safeExit(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func loadConf() *config.JsonConfig {
	var confPath string
	if envPath := os.Getenv("CONFIGS_DIR"); envPath != "" {
		confPath = envPath
	} else {
		confPath = defaultConfigsDir
	}

	conf, err := config.LoadConfFromDir(confPath)

	if err != nil {
		log.Fatal(err)
	}

	return conf
}

func main() {
	// init w/ google configs
	conf := loadConf()
	httpClient := beaconclient.JWTConfigFromJSON(conf.GCloudConfigPath, conf.Scope)
	svc, err := beaconclient.NewBeaconClient(httpClient)
	safeExit(err)

	// cassClient
	cassClient := createCassClient(conf.CassKeyspace, conf.CassEndpoint)

	// inject necessary backend (google api svc) into env
	env := api.Env{svc, cassClient, []byte(conf.JWTSecret)}

	// build router from bound env
	handler := env.Init()

	fmt.Printf("sharecrows api live on port %v\n", conf.Port)

	http.ListenAndServe(":"+strconv.Itoa(conf.Port), handler)

}

func listNamespaces(svc *beaconclient.BeaconClient) {
	res, _ := svc.Svc.Namespaces.List().Do()
	for _, ns := range res.Namespaces {
		fmt.Printf("ns: %+v\n", ns)
	}
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
