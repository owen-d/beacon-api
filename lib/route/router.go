package route

import (
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
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
}

func (r *Router) build(rootRouter *mux.Router, prependMiddleware []negroni.Handler) {
	// build a new router from the path prefix, upon which all subsequent method routs will be mounted
	r.Router = rootRouter.PathPrefix(r.Path).Subrouter().StrictSlash(true)

	// inherit default middleware from parent
	concatedDefaultMiddleware := make([]negroni.Handler, 0, len(prependMiddleware)+len(r.DefaultMiddleware))
	copy(concatedDefaultMiddleware, prependMiddleware)
	r.DefaultMiddleware = append(concatedDefaultMiddleware, r.DefaultMiddleware...)

	// instantiate a new negroni middleware manageer,
	// attach all the middleware functions to it, & bind those functions to a method on a subrouter
	for _, endpoint := range r.Endpoints {
		// allow a specified subpath to be handled on the same router for convenience, i.e. GET /item/:id can use a router on /item with a subpath /:id
		sPath := endpoint.SubPath

		if sPath == "" {
			sPath = "/"
		}

		handler := negroni.New(r.DefaultMiddleware...).With(endpoint.Handlers...)
		r.Router.Handle(sPath, handler).Methods(endpoint.Method)
	}

	//recursively build subroutes
	for _, route := range r.SubRoutes {
		route.build(r.Router, r.DefaultMiddleware)
	}
}

// BuildRouter recursively builds all routes & related middleware via endpoints, returning the resulting mux router.
func BuildRouter(root *Router) *Router {
	// base case, must instantiate a new router
	// recursive call to build all deps
	root.build(mux.NewRouter(), nil)
	return root

}
