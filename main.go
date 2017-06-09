package main

import (
	"fmt"
	"github.com/owen-d/beacon-api/api/beaconclient"
	"github.com/owen-d/beacon-api/api/route"
	"github.com/owen-d/beacon-api/config"
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
	// conf := loadConf()
	// httpClient := beaconclient.JWTConfigFromJSON(conf.GCloudConfigPath, conf.Scope)
	// svc, err := beaconclient.NewBeaconClient(httpClient)
	// safeExit(err)

	RouteTester()

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

func RouteTester() {
	method1 := func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Printf("Hit method1")
		next(rw, r)
	}
	method2 := func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Printf("Hit method2")
		http.Error(rw, "note", 400)
	}

	//break
	ha := route.HandlerAnnotation{
		Method: "GET",
		Fns:    []func(http.ResponseWriter, *http.Request, http.HandlerFunc){method1, method2},
	}
	sr := route.PathAnnotation{
		Path:     "/test",
		Handlers: []*route.HandlerAnnotation{&ha},
	}
	a1 := route.PathAnnotation{
		Path:      "/",
		SubRoutes: []*route.PathAnnotation{&sr},
	}

	r := route.BuildRouter(&a1)

	log.Fatal(http.ListenAndServe(":8080", r))

}
