package cassandrabackup

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	"github.com/sirupsen/logrus"

	api "github.com/Orange-OpenSource/casskop/api/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// initialize local pseudorandom generator
var random = rand.New(rand.NewSource(time.Now().Unix()))

func (r *CassandraBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Filter event types for BackupCRD
	pred := predicate.Funcs{
		// Always handle create events
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if _, err := meta.Accessor(e.ObjectNew); err != nil {
				return false
			}
			if reflect.TypeOf(e.ObjectNew) != reflect.TypeOf(api.CassandraBackup{}) {
				return false
			}
			backup := e.ObjectNew.(*api.CassandraBackup)
			reqLogger := logrus.WithFields(logrus.Fields{"Request.Namespace": backup.Namespace,
				"Request.Name": backup.Name})
			if backup.Status.Condition == nil {
				return false
			}
			backupConditionType := api.BackupConditionType(backup.Status.Condition.Type)

			if backup.IsScheduled() {
				return true
			}

			if !backupConditionType.IsRunning() {
				reqLogger.Debug("Backup is not running, skipping.")
				return false
			}

			return true
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.CassandraBackup{}).
		Owns(&corev1.Pod{}).
		WithEventFilter(pred).
		Complete(r)
}

// blank assignment to verify that CassandraBackupReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &CassandraBackupReconciler{}

func existingNotScheduledSnapshot(c client.Client, instance *api.CassandraBackup) (bool, error) {

	backupsList := &api.CassandraBackupList{}

	if err := c.List(context.TODO(), backupsList); err != nil {
		return false, err
	}

	for _, existingBackup := range backupsList.Items {
		if !existingBackup.IsScheduled() &&
			existingBackup.Ran() &&
			existingBackup.Spec.SnapshotTag == instance.Spec.SnapshotTag &&
			existingBackup.Spec.StorageLocation == instance.Spec.StorageLocation &&
			existingBackup.Spec.CassandraCluster == instance.Spec.CassandraCluster &&
			existingBackup.Spec.Datacenter == instance.Spec.Datacenter {
			return true, nil
		}
	}

	return false, nil
}

func validateBackupSecret(secret *corev1.Secret, backup *api.CassandraBackup, logger *logrus.Entry) error {
	if backup.IsGcpBackup() {
		if len(secret.Data["gcp"]) == 0 {
			return fmt.Errorf("gcp key for secret %s is not set", secret.Name)
		}
	}

	if backup.IsAzureBackup() {
		if len(secret.Data["azurestorageaccount"]) == 0 {
			return fmt.Errorf("azurestorageaccount key for secret %s is not set", secret.Name)
		}

		if len(secret.Data["azurestoragekey"]) == 0 {
			return fmt.Errorf("azurestoragekey key for secret %s is not set", secret.Name)
		}
	}

	if backup.IsS3Backup() {
		// we are just logging here because node can have its credentials injected from AWS itself
		if len(secret.Data["awssecretaccesskey"]) == 0 {
			logger.Info(fmt.Sprintf("awssecretaccesskey key for secret %s is not set, backup "+
				"will failover to authentication mechanims of node itself against AWS.", secret.Name))
		}

		if len(secret.Data["awsaccesskeyid"]) == 0 {
			logger.Info(fmt.Sprintf("awsaccesskeyid key for secret %s is not set, backup "+
				"will failover to authentication mechanims of node itself against AWS.", secret.Name))
		}

		if len(secret.Data["awssecretaccesskey"]) != 0 && len(secret.Data["awsaccesskeyid"]) != 0 {
			if len(secret.Data["awsregion"]) == 0 {
				return fmt.Errorf("There is no awsregion property "+
					"while you have set both awssecretaccesskey and awsaccesskeyid in %s secret for backups", secret.Name)
			}
		}

		if len(secret.Data["awsendpoint"]) != 0 && len(secret.Data["awsregion"]) == 0 {
			return fmt.Errorf("awsendpoint is specified but awsregion is not set in %s secret for backups", secret.Name)
		}
	}

	return nil
}
