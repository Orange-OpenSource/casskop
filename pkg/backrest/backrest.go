package backrest

import (
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/controller/common"
	"github.com/Orange-OpenSource/casskop/pkg/sidecarclient"
	csapi "github.com/instaclustr/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("backrest-methods")

type SidecarRestore struct {
	csClient 		sidecarclient.CassandraSidecarClient
	CoordinatorMember string
}

func NewSidecarRestore(client client.Client, cc *api.CassandraCluster, restore *api.CassandraRestore, pods *corev1.PodList) (*SidecarRestore, error) {
	// Create new sidecars clients
	csClient, err := common.NewSidecarsConnection(log, client, cc, pods)
	if err != nil {
		return nil, err
	}

	return &SidecarRestore{csClient:csClient, CoordinatorMember: restore.Spec.CoordinatorMember}, nil

}

// PerformRestore, perform a restore
func (sr *SidecarRestore) PerformRestore(restore *api.CassandraRestore, backup *api.CassandraBackup) (*api.CassandraRestoreStatus, error) {
	// Prepare restore request
	restoreOperationRequest := &csapi.RestoreOperationRequest {
		Type_: "restore",
		StorageLocation: backup.Spec.StorageLocation,
		SnapshotTag: backup.Spec.SnapshotTag,
		NoDeleteTruncates: restore.Spec.NoDeleteTruncates,
		ExactSchemaVersion: restore.Spec.ExactSchemaVersion,
		RestorationPhase: string(api.RestorationPhaseDownload),
		GlobalRequest: true,
		Import_: &csapi.AllOfRestoreOperationRequestImport_{
			Type_: "import",
			SourceDir: "/var/lib/cassandra/data/downloadedsstables",
		},
		Entities: restore.Spec.Entities,
		K8sSecretName: restore.Spec.SecretName,
		CassandraDirectory: restore.Spec.CassandraDirectory,
		SchemaVersion: restore.Spec.SchemaVersion,
		RestorationStrategyType: restore.Spec.RestorationStrategyType,
		ConcurrentConnections: *restore.Spec.ConcurrentConnection,
	}

	if len(restore.Spec.Entities) == 0 {
		restoreOperationRequest.Entities = backup.Spec.Entities
	}

	if len(restore.Spec.SecretName) == 0 {
		restoreOperationRequest.K8sSecretName = backup.Spec.Secret
	}

	// Perform restore operation
	restoreOperation, err := sr.csClient.PerformRestoreOperation(sr.CoordinatorMember, *restoreOperationRequest)
	if err != nil && err != sidecarclient.ErrCassandraSidecarNotReturned201 {
		log.Error(err, "could not communicate with sidecar")
		return nil, err
	}

	if err == sidecarclient.ErrCassandraSidecarNotReturned201 {
		log.Error(err, "Restore gracefully failed since sidecar returned non 201")
		return nil, err
	}

	log.Info("Restore using sidecar")
	restoreStatus := api.ComputeStatusFromRestoreOperation(restoreOperation)
	return &restoreStatus, nil
}

// GetRestorebyId performs a restore
func (sr *SidecarRestore) GetRestorebyId(restoreId string) (*api.CassandraRestoreStatus, error) {

	// Perform restore operation
	restoreOperation, err := sr.csClient.GetRestoreOperation(sr.CoordinatorMember, restoreId)
	if err != nil && err != sidecarclient.ErrCassandraSidecarNotReturned200 {
		log.Error(err, "could not communicate with sidecar")
		return nil, err
	}

	if err == sidecarclient.ErrCassandraSidecarNotReturned200 {
		log.Error(err, "Restore gracefully failed since sidecar returned non 200")
		return nil, err
	}

	log.Info("Restore status using sidecar")
	restoreStatus := api.ComputeStatusFromRestoreOperation(restoreOperation)
	return &restoreStatus, nil
}
