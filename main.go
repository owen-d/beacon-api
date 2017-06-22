package main

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/owen-d/beacon-api/api"
	"github.com/owen-d/beacon-api/config"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"github.com/owen-d/beacon-api/lib/cass"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	defaultConfigPath = filepath.Join(os.Getenv("GOPATH"), "src/github.com/owen-d/beacon-api/config.json")
)

func safeExit(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func loadConf() *config.JsonConfig {
	var confPath string
	if envPath := os.Getenv("CONFIG_FILE"); envPath != "" {
		confPath = envPath
	} else {
		confPath = defaultConfigPath
	}

	conf, err := config.LoadConfFromFile(confPath)

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
	cassClient := createLocalhostCassClient("bkn")

	// inject necessary backend (google api svc) into env
	env := api.Env{svc, cassClient}

	// build router from bound env
	router := env.Init()

	http.ListenAndServe(":8080", router.Router)

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

func createLocalhostCassClient(keyspace string) *cass.CassClient {
	cluster := gocql.NewCluster("localhost")
	client, err := cass.Connect(cluster, keyspace)
	if err != nil {
		log.Fatal(err)
	}

	return client
}
