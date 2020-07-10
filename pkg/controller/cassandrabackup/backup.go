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
	client *backrest.Client,
	instance *backupClient,
	logging *logrus.Entry,
	recorder record.EventRecorder) {

	operationID, err := client.PerformBackup(instance.backup)

	if err == nil {
		logging.Error(err, fmt.Sprintf("Error while starting backup operation"))
		recorder.Event(instance.backup,
			corev1.EventTypeNormal,
			"BackupNotInitiated",
			fmt.Sprintf("Backup of datacenter %s of cluster %s to %s under snapshot %s failed.",
				instance.backup.Spec.Datacenter, instance.backup.Spec.CassandraCluster,
				instance.backup.Spec.StorageLocation, instance.backup.Spec.SnapshotTag))
		return
	}

	recorder.Event(instance.backup,
		corev1.EventTypeNormal,
		"BackupInitiated",
		fmt.Sprintf("Task initiated to backup datacenter %s of cluster %s to %s under snapshot %s",
			instance.backup.Spec.Datacenter, instance.backup.Spec.CassandraCluster,
			instance.backup.Spec.StorageLocation, instance.backup.Spec.SnapshotTag))

	for range time.NewTicker(2 * time.Second).C {
		if status, err := client.GetBackupById(operationID); err != nil {
			logging.Error(err, fmt.Sprintf("Error while finding submitted backup operation %v", operationID))
			break
		} else {
			instance.updateStatus(status, logging)

			if status.State == api.BackupFailed {
				recorder.Event(instance.backup,
					corev1.EventTypeWarning,
					"BackupFailed",
					fmt.Sprintf("Backup operation %v on node %s has failed", operationID, status.Node))
				break
			}

			if status.State == api.BackupCompleted {
				recorder.Event(instance.backup,
					corev1.EventTypeNormal,
					"BackupCompleted",
					fmt.Sprintf("Backup operation %v on node %s was completed.", operationID, status.Node))
				break
			}
		}
	}
}

func (si *backupClient) updateStatus(status *api.CassandraBackupStatus,
	logging *logrus.Entry) {

	si.backup.Status = status

	jsonPatch := fmt.Sprintf(`{"status":{"node": "%s", "state": "%s", "progress": "%s"}}`,
		status.Node, status.State, status.Progress)
	patchToApply := client.RawPatch(types.MergePatchType, []byte(jsonPatch))
	objToPatch := &api.CassandraBackup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: si.backup.Namespace,
			Name:      si.backup.Name,
		}}

	if err := si.client.Patch(context.Background(), objToPatch, patchToApply); err != nil {

		logging.Error(err, "Error updating CassandraBackup backup")
	}
}
