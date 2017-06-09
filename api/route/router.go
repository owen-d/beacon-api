package route

import (
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
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

// HandlerAnnotation is a struct for programatically constructing a route handler.
type HandlerAnnotation struct {
	Method string
	Fns    []func(http.ResponseWriter, *http.Request, http.HandlerFunc)
}

// Attach takes a negroni instance and binds all the functions in a HandlerAnnotation to it
func (a *HandlerAnnotation) Attach(n *negroni.Negroni) {
	for _, fn := range a.Fns {
		n.UseFunc(fn)
	}
}

// PathAnnotation is a struct for managing a path, router, relevant subroutes (via recursion),
// and handlers (a method & slice of middleware functions)
type PathAnnotation struct {
	Path      string
	Router    *mux.Router
	SubRoutes []*PathAnnotation
	Handlers  []*HandlerAnnotation
}

// build is responsible for recursively constructing the router & binding middleware. It's an internal function
// and will be called via the exported 'BuildRouter' function
func (annotation *PathAnnotation) build(rootRouter *mux.Router) {
	// build a new router from the path prefix, upon which all subsequent method routs will be mounted
	sr := rootRouter.PathPrefix(annotation.Path).Subrouter()
	annotation.Router = sr

	// instantiate a new negroni middleware manageer,
	// attach all the middleware functions to it, & bind those functions to a method on a subrouter
	for _, handler := range annotation.Handlers {
		n := negroni.New()
		handler.Attach(n)
		sr.Handle("/", n).Methods(handler.Method)
	}

	//recursively build subroutes
	for _, route := range annotation.SubRoutes {
		route.build(sr)
	}

}

// BuildRouter recursively builds all routes & related middleware via annotations, returning the resulting mux router.
func BuildRouter(root *PathAnnotation) *mux.Router {
	// base case, must instantiate a new router
	root.Router = mux.NewRouter()
	// recursive call to build all deps
	root.build(root.Router)
	return root.Router

}
