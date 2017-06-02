package route

import (
	"github.com/gorilla/mux"
	"net/http"
)

// We want a couple primitives for routes: Beacons, Messages(backed by attachments), Schedules.
// These will exist independently so we can detach & reattach them to each other:
// A beacon is just a beacon, but Messages can be attached to 1 or more beacons and
// attachments are the the product of conjoining a message and a schedule (nil schedule = always on)

// Path Annotations
var (
	RootRouter = PathAnnotation{
		// note: we don't add a Path property to the root router;
		// it will be mounted via http.Handle("/", rootRouter)
		SubRoutes: []*PathAnnotation{
			&BeaconRouter,
		},
	}
	BeaconRouter = PathAnnotation{
		Path: "beacons",
		SubRoutes: []*PathAnnotation{
			&BeaconIdRouter,
		},
		Handlers: BeaconHandlers,
	}
	BeaconIdRouter = PathAnnotation{
		Path:     "{beaconId}",
		Handlers: BeaconByIdHandlers,
	}
)

// Handler Annotations
var (
	BeaconHandlers = []*HandlerAnnotation{
		&HandlerAnnotation{
			Method: "POST",
			// TBD: add Fn
		},
	}
	BeaconByIdHandlers = []*HandlerAnnotation{
		&HandlerAnnotation{
			Method: "GET",
			// TBD: add Fn
		},
		&HandlerAnnotation{
			Method: "DELETE",
			// TBD: add Fn
		},
	}
)

// struct for programatically constructing a route handler.
type HandlerAnnotation struct {
	Method string
	Fn     *func(http.ResponseWriter, *http.Request)
}

type PathAnnotation struct {
	Path      string
	Router    *mux.Router
	SubRoutes []*PathAnnotation
	Handlers  []*HandlerAnnotation
}

func (annotation *PathAnnotation) build(rootRouter *mux.Router) {
	sr := rootRouter.PathPrefix(annotation.Path).Subrouter()
	annotation.Router = sr
	for _, handler := range annotation.Handlers {
		sr.HandleFunc("/", *handler.Fn).
			Methods(handler.Method)
	}

	for _, route := range annotation.SubRoutes {
		route.build(sr)
	}

}

func BuildRouter(root *PathAnnotation) *mux.Router {
	// base case, must instantiate a new router
	root.Router = mux.NewRouter()
	// recursive call to build all deps
	root.build(root.Router)
	return root.Router

}
