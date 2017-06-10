package route

import (
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
)

type Endpoint struct {
	// Must be a HTTP Method
	Method string
	Fns    []http.HandlerFunc
}

func (e *Endpoint) Attach(n *negroni.Negroni) {
	for _, fn := range e.Fns {
		n.UseHandlerFunc(fn)
	}
}

type Router struct {
	Path      string
	Router    *mux.Router
	SubRoutes []*Router
	Endpoints []*Endpoint
}

func (r *Router) build(rootRouter *mux.Router) {
	// build a new router from the path prefix, upon which all subsequent method routs will be mounted
	r.Router = rootRouter.PathPrefix(r.Path).Subrouter().StrictSlash(true)

	// instantiate a new negroni middleware manageer,
	// attach all the middleware functions to it, & bind those functions to a method on a subrouter
	for _, endpoint := range r.Endpoints {
		n := negroni.New()
		endpoint.Attach(n)
		r.Router.Handle("/", n).Methods(endpoint.Method)
	}

	//recursively build subroutes
	for _, route := range r.SubRoutes {
		route.build(r.Router)
	}
}

// BuildRouter recursively builds all routes & related middleware via annotations, returning the resulting mux router.
func BuildRouter(root *Router) *Router {
	// base case, must instantiate a new router
	// recursive call to build all deps
	root.build(mux.NewRouter())
	return root

}
