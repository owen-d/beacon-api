package route

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
	"net/http"
)

var (
	CorsHandler = cors.AllowAll()
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

// BuildRouter recursively builds all routes & related middleware via endpoints, returning the resulting mux router.
func BuildRouter(root *Router, initRouter *mux.Router) *Router {
	// base case, must instantiate a new router
	if initRouter == nil {
		initRouter = mux.NewRouter()
	}

	// recursive call to build all deps
	root.build(initRouter, nil)
	return root

}

func Encase(initHandler http.Handler, prependMiddlewares ...http.Handler) http.Handler {
	resHandler := negroni.New()
	for _, h := range prependMiddlewares {
		resHandler.UseHandler(h)
	}

	resHandler.UseHandler(initHandler)
	return resHandler
}
