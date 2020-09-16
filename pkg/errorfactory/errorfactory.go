package errorfactory

import "emperror.dev/errors"

// ResourceNotReady states that resource is not ready
type ResourceNotReady struct{ error }

// APIFailure states that something went wrong with the api
type APIFailure struct{ error }

// StatusUpdateError states that the operator failed to update the Status
type StatusUpdateError struct{ error }

// CassandraBackupSidecarNotReady states that Sidecar is not ready to receive connection
type CassandraBackupSidecarNotReady struct{ error }

// CassandraBackupOperationRunning states that Sidecar Operation is still running
type CassandraBackupOperationRunning struct{ error }

//CassandraBackupOperationFailure states that Sidecar Operation was not found (Sidecar restart?) or failed
type CassandraBackupOperationFailure struct{ error }

// New creates a new error factory error
func New(t interface{}, err error, msg string, wrapArgs ...interface{}) error {
	wrapped := errors.WrapIfWithDetails(err, msg, wrapArgs...)
	switch t.(type) {
	case ResourceNotReady:
		return ResourceNotReady{wrapped}
	case APIFailure:
		return APIFailure{wrapped}
	case StatusUpdateError:
		return StatusUpdateError{wrapped}
	case CassandraBackupSidecarNotReady:
		return CassandraBackupSidecarNotReady{wrapped}
	case CassandraBackupOperationRunning:
		return CassandraBackupOperationRunning{wrapped}
	case CassandraBackupOperationFailure:
		return CassandraBackupOperationFailure{wrapped}
	}
	return wrapped
}
