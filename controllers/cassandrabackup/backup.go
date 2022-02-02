package cassandrabackup

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	api "github.com/Orange-OpenSource/casskop/api/v2"
	"github.com/Orange-OpenSource/casskop/pkg/backrest"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type backupClient struct {
	backup *api.CassandraBackup
	client client.Client
}

func backup(
	backrestClient *backrest.Client,
	backupClient *backupClient,
	logging *logrus.Entry,
	recorder record.EventRecorder) {

	operationID, err := backrestClient.PerformBackup(backupClient.backup)

	if err != nil {
		logging.Error(err, fmt.Sprintf("Error while starting backup operation"))
		recorder.Event(backupClient.backup,
			corev1.EventTypeWarning,
			"BackupNotInitiated",
			fmt.Sprintf("Backup of datacenter %s of cluster %s to %s under snapshot %s failed.",
				backupClient.backup.Spec.Datacenter, backupClient.backup.Spec.CassandraCluster,
				backupClient.backup.Spec.StorageLocation, backupClient.backup.Spec.SnapshotTag))
		return
	}

	recorder.Event(backupClient.backup,
		corev1.EventTypeNormal,
		"BackupInitiated",
		fmt.Sprintf("Task initiated to backup datacenter %s of cluster %s to %s under snapshot %s",
			backupClient.backup.Spec.Datacenter, backupClient.backup.Spec.CassandraCluster,
			backupClient.backup.Spec.StorageLocation, backupClient.backup.Spec.SnapshotTag))

	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		if status, err := backrestClient.BackupStatusByID(operationID); err != nil {
			logging.Error(err, fmt.Sprintf("Error while finding submitted backup operation %v", operationID))
			ticker.Stop()
			break
		} else {
			if !backupClient.updateStatus(status, logging){
				continue
			}
			switch api.BackupConditionType(status.Condition.Type) {
			case api.BackupFailed:
				recorder.Event(backupClient.backup,
					corev1.EventTypeWarning,
					"BackupFailed",
					fmt.Sprintf("Backup operation %v on node %s has failed",
						operationID, status.CoordinatorMember))
				ticker.Stop()
				break
			case api.BackupCompleted:
				recorder.Event(backupClient.backup,
					corev1.EventTypeNormal,
					"BackupCompleted",
					fmt.Sprintf("Backup operation %v on node %s was completed.", operationID, status.CoordinatorMember))
				ticker.Stop()
				break
			}
		}
	}
}

func (backupClient *backupClient) updateStatus(status api.BackRestStatus, logging *logrus.Entry) bool {

	patch := client.MergeFrom(backupClient.backup.DeepCopy())
	backupClient.backup.Status = status

	if err := backupClient.client.Patch(context.Background(), backupClient.backup, patch); err != nil {
		logging.Error(err, errors.WrapIfWithDetails(err, "could not update status for restore",
			"restore", backupClient.backup))
		return false
	}

	return true
}