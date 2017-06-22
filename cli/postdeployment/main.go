package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	bc "github.com/owen-d/beacon-api/api/controllers/beacons"
	"github.com/owen-d/beacon-api/api/controllers/deployments"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	userId  = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	depName = "live_dep"
	title   = "welcome home fam!"
	url     = "https://www.google.com"
)

func main() {
	beacons, _ := GetBeacons()
	dep := &deployments.Deployment{
		UserId:      userId,
		Name:        "live1",
		BeaconNames: make([]string, 0, len(beacons.Beacons)),
		Message: &deployments.Message{
			Name:  "testmsg",
			Title: title,
			Url:   url,
		},
	}

	for _, b := range beacons.Beacons {
		dep.BeaconNames = append(dep.BeaconNames, b.BeaconName)
	}

	// fmt.Printf("built:\n%+v\n", dep)

	PostDeployment(dep)
}

func GetBeacons() (*bc.BeaconResponse, error) {
	resp, _ := http.Get("http://localhost:8080/beacons/")

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	beacons := bc.BeaconResponse{}
	json.Unmarshal(body, &beacons)
	return &beacons, nil

}

func PostDeployment(dep *deployments.Deployment) {
	jsonData, err := json.Marshal(dep)

	if err != nil {
		log.Fatal(err)
	}

	resp, _ := http.Post("http://localhost:8080/deployments/", "application/json", bytes.NewReader(jsonData))

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var bodyS interface{}
	json.Unmarshal(body, &bodyS)

	fmt.Printf("res:\nstatusCode: %v\nstatus: %v\nbody: %+v\n", resp.StatusCode, resp.Status, bodyS)

}
