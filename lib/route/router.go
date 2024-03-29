package route

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

var (
	tenMinutesInSeconds = 60 * 10
	CorsHandler         = cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{"*"},
		MaxAge:         tenMinutesInSeconds,
	})
)

type Endpoint struct {
	// Must be a HTTP Method
	Method   string
	Handlers []negroni.Handler
	SubPath  string
}

type Router struct {
	Path              string
	Router            *mux.Router
	DefaultMiddleware []negroni.Handler
	SubRoutes         []*Router
	Endpoints         []*Endpoint
	Name              string
}

func (r *Router) build(rootRouter *mux.Router, prependMiddleware []negroni.Handler) {
	// build a new router from the path prefix, upon which all subsequent method routs will be mounted
	r.Router = rootRouter.PathPrefix(r.Path).Subrouter()

	// inherit default middleware from parent
	concatedDefaultMiddleware := make([]negroni.Handler, 0, len(prependMiddleware)+len(r.DefaultMiddleware))
	concatedDefaultMiddleware = append(concatedDefaultMiddleware, prependMiddleware...)
	r.DefaultMiddleware = append(concatedDefaultMiddleware, r.DefaultMiddleware...)

	// instantiate a new negroni middleware manageer,
	// attach all the middleware functions to it, & bind those functions to a method on a subrouter
	for _, endpoint := range r.Endpoints {
		// allow a specified subpath to be handled on the same router for convenience, i.e. GET /item/:id can use a router on /item with a subpath /:id
		sPath := endpoint.SubPath

		handler := negroni.New(r.DefaultMiddleware...).With(endpoint.Handlers...)
		fmtStr := fmt.Sprintf("\n\trPath: %+v\n\trName: %+v\n\tsPath: %+v\n\tmethod: %+v\n\n", r.Path, r.Name, sPath, endpoint.Method)
		r.Router.Handle(sPath, handler).Methods(endpoint.Method).Name(fmtStr)
	}

	//recursively build subroutes
	for _, route := range r.SubRoutes {
		route.build(r.Router, r.DefaultMiddleware)
	}
}

// Inject recursively builds all routes & related middleware via endpoints, adding their routes onto the root mux router & returning it.
func Inject(router *Router, root *mux.Router) *mux.Router {
	// base case, must instantiate a new router
	if root == nil {
		root = mux.NewRouter()
	}

	// recursive call to build all deps
	router.build(root, nil)

	return root

}
