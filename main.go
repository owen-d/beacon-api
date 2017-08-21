package main

import (
	"fmt"
	"github.com/owen-d/beacon-api/api"
	"github.com/owen-d/beacon-api/config"
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var (
	// for loading dev config
	defaultConfigsDir = filepath.Join(os.Getenv("GOPATH"), "src/github.com/owen-d/beacon-api/conf/dev-settings")
)

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
	env := api.Env{conf}

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
