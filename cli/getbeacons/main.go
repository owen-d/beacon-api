package main

import (
	"encoding/json"
	"fmt"
	bc "github.com/owen-d/beacon-api/api/controllers/beacons"
	"io/ioutil"
	"net/http"
)

func main() {
	resp, _ := http.Get("http://localhost:8080/beacons/")

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	beacons := bc.BeaconResponse{}
	json.Unmarshal(body, &beacons)

	for i, b := range beacons.Beacons {
		fmt.Printf("beacon %v:\n%+v\n", i, *b)
	}

}
