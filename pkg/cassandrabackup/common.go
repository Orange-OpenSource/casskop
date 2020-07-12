package cassandrabackup

import (
	"errors"
)

var ErrCassandraSidecarNotReturned200 	= errors.New("Non 200 response from cassandra backup sidecar")
var ErrCassandraSidecarNotReturned201 	= errors.New("Non 201 response from cassandra backup sidecar")
var ErrNoCassandraBackupClientAvailable = errors.New("Cannot create a cassandra backup client")