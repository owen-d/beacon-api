package main

import (
	"fmt"
	"github.com/owen-d/beacon/api"
	"github.com/owen-d/beacon/config"
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
	httpClient := api.JWTConfigFromJSON(conf.GCloudConfigPath, conf.Scope)
	svc, err := api.NewBeaconClient(httpClient)
	safeExit(err)
	res, err := svc.GetOwnedBeaconNames()
	safeExit(err)
	bNames := make([]string, len(res.Beacons))
	for _, b := range res.Beacons {
		bNames = append(bNames, b.BeaconName)
	}

	beacons := svc.GetBeaconsByNames(bNames)
	for _, b := range beacons {
		fmt.Printf("beacon: %v", b)
	}
}
