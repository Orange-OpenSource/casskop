package common

import (
	"fmt"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/sidecarclient"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var newSidecarClientFromCluster = sidecarclient.NewFromCluster

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

// newSidecarsConnection is a convenience wrapper for creating a sidecars connection
// and creating a safer close function
func NewSidecarsConnection(log logr.Logger, client client.Client, cluster *api.CassandraCluster, podList *corev1.PodList) (csClient sidecarclient.CassandraSidecarClient, err error) {
	// Get a nifi connection
	log.Info(fmt.Sprintf("Retrieving Cassandra Sidecar client for %s/%s", cluster.Namespace, cluster.Name))
	csClient, err = newSidecarClientFromCluster(client, cluster, podList)
	if err != nil {
		return
	}
	return
}