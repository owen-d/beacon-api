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

func Attach(n *negroni.Negroni, handlers []negroni.Handler) *negroni.Negroni {
	for _, handler := range handlers {
		n.Use(handler)
	}
	return n
}

type Router struct {
	Path              string
	Router            *mux.Router
	DefaultMiddleware []negroni.Handler
	SubRoutes         []*Router
	Endpoints         []*Endpoint
}

func (r *Router) build(rootRouter *mux.Router) {
	// build a new router from the path prefix, upon which all subsequent method routs will be mounted
	if len(r.DefaultMiddleware) == 0 {
		r.Router = rootRouter.PathPrefix(r.Path).Subrouter().StrictSlash(true)
	} else {
		// if DefaultMiddleware is specified, we apply it to all endpoints derived from this router
		handler := Attach(negroni.New(), r.DefaultMiddleware)
		r.Router = rootRouter.PathPrefix(r.Path).Handler(handler).Subrouter().StrictSlash(true)
	}

	// instantiate a new negroni middleware manageer,
	// attach all the middleware functions to it, & bind those functions to a method on a subrouter
	for _, endpoint := range r.Endpoints {
		// allow a specified subpath to be handled on the same router for convenience, i.e. GET /item/:id can use a router on /item with a subpath /:id
		sPath := endpoint.SubPath

		if sPath == "" {
			sPath = "/"
		}

		handler := Attach(negroni.New(), endpoint.Handlers)
		r.Router.Handle(sPath, handler).Methods(endpoint.Method)
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
