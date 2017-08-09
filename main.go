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
	defaultConfigsDir = filepath.Join(os.Getenv("GOPATH"), "src/github.com/owen-d/beacon-api/conf/settings")
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

	http.ListenAndServe(":"+strconv.Itoa(conf.Port), handler)

}

func listNamespaces(svc *beaconclient.BeaconClient) {
	res, _ := svc.Svc.Namespaces.List().Do()
	for _, ns := range res.Namespaces {
		fmt.Printf("ns: %+v\n", ns)
	}
}

func beaconCycle(svc *beaconclient.BeaconClient) {
	// get list of beacon names
	res, err := svc.GetOwnedBeaconNames()
	safeExit(err)
	bNames := make([]string, 0, len(res.Beacons))

	for _, b := range res.Beacons {
		bNames = append(bNames, b.BeaconName)
	}

	// delete old attachments on beacon
	numDeleted, _ := svc.BatchDeleteAttachments(bNames[0])
	fmt.Printf("deleted %v\n", numDeleted)

	// add new attachment to beacon
	newAttachment := beaconclient.AttachmentData{
		Title: "Welcome home, qtpi",
		Url:   "https://www.eff.org",
	}

	fmt.Println("attempting to attach to beacon:", bNames[0], "\n")
	attachment, err := svc.CreateAttachment(bNames[0], &newAttachment)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created attachment:\n%+v\n", attachment)

	fmt.Printf("done")

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
