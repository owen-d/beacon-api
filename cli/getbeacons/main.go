package main

import (
	"encoding/json"
	"fmt"
	bc "github.com/owen-d/beacon-api/api/controllers/beacons"
	"io/ioutil"
	"net/http"
)

func main() {
	beacons, _ := GetBeacons()
	for i, b := range beacons.Beacons {
		fmt.Printf("beacon %v:\n%+v\n", i, *b)
	}

}

func GetBeacons() (*bc.BeaconResponse, error) {
	resp, _ := http.Get("http://localhost:8080/beacons/")

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	beacons := bc.BeaconResponse{}
	json.Unmarshal(body, &beacons)
	return &beacons, nil

}
