package sidecarclient

import (
	"errors"
)

var ErrCassandraSidecarNotConnected 	    = errors.New("The targeted sidecar is disconnected")
var ErrCassandraSidecarNotReturned200 	    = errors.New("non 200 response from sidecar cluster")
var ErrCassandraSidecarNotReturned201 	    = errors.New("non 201 response from sidecar cluster")
var ErrCassandraSidecarReturned404 		    = errors.New("404 response from sidecar cluster")
var ErrNoCassandraSidecarClientsAvailable	= errors.New("Cannot create a sidecar client to perform actions")