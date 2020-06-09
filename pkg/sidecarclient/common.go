package sidecarclient

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
)

var ErrCassandraSidecarNotConnected 	    = errors.New("The targeted sidecar is disconnected")
var ErrCassandraSidecarNotReturned200 	    = errors.New("non 200 response from sidecar cluster")
var ErrCassandraSidecarReturned404 		    = errors.New("404 response from sidecar cluster")
var ErrNoCassandraSidecarClientsAvailable	= errors.New("Cannot create a sidecar client to perform actions")

func decode(v interface{}, b []byte, contentType string) (err error) {
	if strings.Contains(contentType, "application/xml") {
		if err = xml.Unmarshal(b, v); err != nil {
			return err
		}
		return nil
	} else if strings.Contains(contentType, "application/json") {
		if err = json.Unmarshal(b, v); err != nil {
			return err
		}
		return nil
	}
	return errors.New("undefined response type")
}