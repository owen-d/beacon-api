package main

import (
	"encoding/json"
	"fmt"
	bc "github.com/owen-d/beacon-api/api/controllers/beacons"
	"io/ioutil"
	"net/http"
)

var ExampleJWT string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTQzOTYzOTksImlhdCI6MTQ5ODg0NDM5OSwidXNlcl9pZCI6IjZiYTdiODEwLTlkYWQtMTFkMS04MGI0LTAwYzA0ZmQ0MzBjOCJ9._Mn0COXwcs9l4NqqAbbosXWCTMentdy4xj9ZqgKhEF0"

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
