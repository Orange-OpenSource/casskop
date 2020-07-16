package cassandrabackup

import (
	"context"
	"fmt"
	"time"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/backrest"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	if err == nil {
		logging.Error(err, fmt.Sprintf("Error while starting backup operation"))
		recorder.Event(backupClient.backup,
			corev1.EventTypeNormal,
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

	for range time.NewTicker(2 * time.Second).C {
		if status, err := backrestClient.GetBackupStatusById(operationID); err != nil {
			logging.Error(err, fmt.Sprintf("Error while finding submitted backup operation %v", operationID))
			break
		} else {
			backupClient.updateStatus(status, logging)

			if status.State == api.BackupFailed {
				recorder.Event(backupClient.backup,
					corev1.EventTypeWarning,
					"BackupFailed",
					fmt.Sprintf("Backup operation %v on node %s has failed", operationID, status.CoordinatorMember))
				break
			}

			if status.State == api.BackupCompleted {
				recorder.Event(backupClient.backup,
					corev1.EventTypeNormal,
					"BackupCompleted",
					fmt.Sprintf("Backup operation %v on node %s was completed.", operationID, status.CoordinatorMember))
				break
			}
		}
	}
}

func (backupClient *backupClient) updateStatus(status *api.CassandraBackupStatus,
	logging *logrus.Entry) {

	backupClient.backup.Status = status

	jsonPatch := fmt.Sprintf(`{"status":{"node": "%s", "state": "%s", "progress": "%s"}}`,
		status.CoordinatorMember, status.State, status.Progress)
	patchToApply := client.RawPatch(types.MergePatchType, []byte(jsonPatch))
	cassandraBackup := &api.CassandraBackup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: backupClient.backup.Namespace,
			Name:      backupClient.backup.Name,
		}}

	if err := backupClient.client.Patch(context.Background(), cassandraBackup, patchToApply); err != nil {

		logging.Error(err, "Error updating CassandraBackup backup")
	}
}
