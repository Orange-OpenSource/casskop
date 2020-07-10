package cassandrabackup

import (
	"errors"
)

var ErrCassandraSidecarNotReturned200 	= errors.New("non 200 response from sidecar cluster")
var ErrCassandraSidecarNotReturned201 	= errors.New("non 201 response from sidecar cluster")
var ErrNoCassandraBackupClientAvailable = errors.New("cannot create a cassandra backup client to perform actions")