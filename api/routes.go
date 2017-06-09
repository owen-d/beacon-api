package api

import (
	"github.com/owen-d/beacon-api/api/handlers"
	"github.com/owen-d/beacon-api/lib/route"
	"net/http"
)

type Routes []route.Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		[]func(){},
	},
}
