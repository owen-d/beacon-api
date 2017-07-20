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
	"strings"
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
		if strings.HasPrefix(b.Name, "beacons/3!") {
			dep.BeaconNames = append(dep.BeaconNames, b.Name)
		}
	}

	// fmt.Printf("built:\n%+v\n", dep)

	PostDeployment(dep)
}

func GetBeacons() (*bc.BeaconResponse, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/beacons", nil)
	req.Header.Set("x-jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTQzOTYzOTksImlhdCI6MTQ5ODg0NDM5OSwidXNlcl9pZCI6IjZiYTdiODEwLTlkYWQtMTFkMS04MGI0LTAwYzA0ZmQ0MzBjOCJ9._Mn0COXwcs9l4NqqAbbosXWCTMentdy4xj9ZqgKhEF0")
	resp, _ := client.Do(req)

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

	fmt.Println(string(jsonData))
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://localhost:8080/deployments", bytes.NewReader(jsonData))
	req.Header.Set("x-jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTQzOTYzOTksImlhdCI6MTQ5ODg0NDM5OSwidXNlcl9pZCI6IjZiYTdiODEwLTlkYWQtMTFkMS04MGI0LTAwYzA0ZmQ0MzBjOCJ9._Mn0COXwcs9l4NqqAbbosXWCTMentdy4xj9ZqgKhEF0")
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var bodyS interface{}
	json.Unmarshal(body, &bodyS)

	fmt.Printf("res:\nstatusCode: %v\nstatus: %v\nbody: %+v\n", resp.StatusCode, resp.Status, bodyS)

}
