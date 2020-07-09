package errorfactory

import "emperror.dev/errors"

// ResourceNotReady states that resource is not ready
type ResourceNotReady struct{ error }

// APIFailure states that something went wrong with the api
type APIFailure struct{ error }

// StatusUpdateError states that the operator failed to update the Status
type StatusUpdateError struct{ error }

// SidecarNotReady states that Sidecar is not ready to receive connection
type SidecarNotReady struct{ error }

// SidecarOperationRunning states that Sidecar Operation is still running
type SidecarOperationRunning struct{ error }

//SidecarOperationFailure states that Sidecar Operation was not found (Sidecar restart?) or failed
type SidecarOperationFailure struct{ error }

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
	case SidecarNotReady:
		return SidecarNotReady{wrapped}
	case SidecarOperationRunning:
		return SidecarOperationRunning{wrapped}
	case SidecarOperationFailure:
		return SidecarOperationFailure{wrapped}
	}
	return wrapped
}
