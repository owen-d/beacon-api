package handlers

import (
	"github.com/owen-d/beacon-api/lib/beaconclient"
	"net/http"
)

type BeaconHandlers interface {
	func GetBeacons(http.ResponseWriter, *http.Request, http.HandlerFunc) {}
}
