package common

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// requeueWithError is a convenience wrapper around logging an error message
// separate from the stacktrace and then passing the error through to the controller
// manager
func RequeueWithError(logger logr.Logger, msg string, err error) (reconcile.Result, error) {
	// Info log the error message and then let the reconciler dump the stacktrace
	logger.Info(msg)
	return reconcile.Result{}, err
}

// reconciled returns an empty result with nil error to signal a successful reconcile
// to the controller manager
func Reconciled() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}