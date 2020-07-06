package cassandrabackup

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/sidecar"
	csd "github.com/instaclustr/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
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
	sidecarClient *sidecar.Client,
	instance *backupClient,
	logging *logrus.Entry,
	recorder record.EventRecorder) {

	bandwidth := strings.Replace(instance.backup.Spec.Bandwidth, " ", "", -1)
	bandwidthDataRate, err := parseBandwidth(bandwidth)

	if err != nil {
		recorder.Event(instance.backup,
			corev1.EventTypeNormal,
			"BackupNotInitiated",
			fmt.Sprintf("Backup of datacenter %s of cluster %s to %s under snapshot %s can't be initiated with bandwidth %s",
				instance.backup.Spec.Datacenter, instance.backup.Spec.CassandraCluster,
				instance.backup.Spec.StorageLocation, instance.backup.Spec.SnapshotTag, bandwidth))
		return
	}

	backupRequest := csd.BackupOperationRequest{
		Type_:                 "backup",
		StorageLocation:       instance.backup.Spec.StorageLocation,
		SnapshotTag:           instance.backup.Spec.SnapshotTag,
		Duration:              instance.backup.Spec.Duration,
		Bandwidth:             bandwidthDataRate,
		ConcurrentConnections: instance.backup.Spec.ConcurrentConnections,
		Entities:              instance.backup.Spec.Entities,
		K8sSecretName:         instance.backup.Spec.Secret,
		Dc:                    instance.backup.Spec.Datacenter,
		GlobalRequest:         true,
		K8sNamespace:          instance.backup.Namespace,
	}

	if operationID, err := sidecarClient.StartOperation(backupRequest); err != nil {
		logging.Error(err, fmt.Sprintf("Error while starting backup operation %v", backupRequest))
	} else {
		recorder.Event(instance.backup,
			corev1.EventTypeNormal,
			"BackupInitiated",
			fmt.Sprintf("Task initiated to backup datacenter %s of cluster %s to %s under snapshot %s",
				instance.backup.Spec.Datacenter, instance.backup.Spec.CassandraCluster,
				instance.backup.Spec.StorageLocation, instance.backup.Spec.SnapshotTag))
		podHostname := sidecarClient.Host
		for range time.NewTicker(2 * time.Second).C {
			if r, err := sidecarClient.GetOperation(operationID); err != nil {
				logging.Error(err, fmt.Sprintf("Error while finding submitted backup operation %v", operationID))
				break
			} else {
				instance.updateStatus(podHostname, r, logging)

				if api.BackupState(r.State) == api.BackupFailed {
					recorder.Event(instance.backup,
						corev1.EventTypeWarning,
						"BackupFailed",
						fmt.Sprintf("Backup operation %v on node %s has failed", operationID, podHostname))
					break
				}

				if api.BackupState(r.State) == api.BackupCompleted {
					recorder.Event(instance.backup,
						corev1.EventTypeNormal,
						"BackupCompleted",
						fmt.Sprintf("Backup operation %v on node %s was completed.", operationID, podHostname))
					break
				}
			}
		}
	}
}
func (si *backupClient) updateStatus(podHostname string, r *csd.BackupOperationResponse,
	logging *logrus.Entry) {

	status := &api.CassandraBackupStatus{
		Node:     podHostname,
		Progress: fmt.Sprintf("%v%%", strconv.Itoa(int(r.Progress*100))),
		State:    api.BackupState(r.State),
	}

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
