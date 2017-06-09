package main

import (
	"github.com/owen-d/beacon-api/lib/route"
)

// We want a couple primitives for routes: Beacons, Messages(backed by attachments), Schedules.
// These will exist independently so we can detach & reattach them to each other:
// A beacon is just a beacon, but Messages can be attached to 1 or more beacons and
// attachments are the the product of conjoining a message and a schedule (nil schedule = always on)

// Path Annotations
var (
	RootRouter = route.PathAnnotation{
		// note: we don't add a Path property to the root router;
		// it will be mounted via http.Handle("/", rootRouter)
		SubRoutes: []*route.PathAnnotation{
			&BeaconRouter,
		},
	}
	BeaconRouter = route.PathAnnotation{
		Path: "beacons",
		SubRoutes: []*route.PathAnnotation{
			&BeaconIdRouter,
		},
		Handlers: BeaconHandlers,
	}
	BeaconIdRouter = route.PathAnnotation{
		Path:     "{beaconId}",
		Handlers: BeaconByIdHandlers,
	}
)

// Handler Annotations
var (
	BeaconHandlers = []*route.HandlerAnnotation{
		&route.HandlerAnnotation{
			Method: "POST",
			// TBD: add Fn
		},
	}
	BeaconByIdHandlers = []*route.HandlerAnnotation{
		&route.HandlerAnnotation{
			Method: "GET",
			// TBD: add Fn
		},
		&route.HandlerAnnotation{
			Method: "DELETE",
			// TBD: add Fn
		},
	}
)

// Beacon Router
