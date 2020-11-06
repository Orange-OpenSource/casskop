package common

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/cassandrabackup"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RequeueWithError is a convenience wrapper around logging an error message
// separate from the stacktrace and then passing the error through to the controller
// manager
func RequeueWithError(logger *logrus.Entry, msg string, err error) (reconcile.Result, error) {
	// Info log the error message and then let the reconciler dump the stacktrace
	logger.Info(msg)
	return reconcile.Result{}, err
}

// reconciled returns an empty result with nil error to signal a successful reconcile
// to the controller manager
func Reconciled() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// NewCassandraBackupConnection is a convenience wrapper for creating a sidecars connection
// and creating a safer close function
func NewCassandraBackupConnection(client client.Client, cluster *api.CassandraCluster,
	pod *corev1.Pod) (csClient cassandrabackup.Client, err error) {
	// Get a cassandra backup connection
	logrus.Info(fmt.Sprintf("Retrieving Cassandra Sidecar client for %s/%s", cluster.Namespace, cluster.Name))
	csClient, err = cassandrabackup.ClientFromCluster(client, cluster, pod)
	if err != nil {
		return
	}
	return
}

func JsonPatch(restore map[string]interface{}) (client.Patch, error) {
	statusBytes, err := json.Marshal(restore)
	if err != nil {
		return nil, err
	}

	jsonPatch := string(statusBytes)
	patchToApply := client.RawPatch(types.MergePatchType, []byte(jsonPatch))
	return patchToApply, nil
}
