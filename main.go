package main

import (
	"fmt"
	"github.com/owen-d/beacon/api"
	"github.com/owen-d/beacon/config"
	"google.golang.org/api/proximitybeacon/v1beta1"
	"log"
	"os"
	"path/filepath"
)

var (
	defaultConfigPath = filepath.Join(os.Getenv("HOME"), "go/src/github.com/owen-d/beacon/config.json")
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
	conf := loadConf()
	client := api.JWTConfigFromJSON(conf.GCloudConfigPath, conf.Scope)
	svc, err := proximitybeacon.New(client)
	safeExit(err)
	res, e := svc.Beacons.List().Do()
	safeExit(e)
	for _, b := range res.Beacons {
		fmt.Printf("bk: %+v\n", b.BeaconName)
	}
}
