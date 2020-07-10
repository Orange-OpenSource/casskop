package cassandrabackup

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_cassandrabackup")

// initialize local pseudorandom generator
var random = rand.New(rand.NewSource(time.Now().Unix()))

// Add creates a new CassandraBackup Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCassandraBackup{
		client:    mgr.GetClient(),
		scheme:    mgr.GetScheme(),
		recorder:  mgr.GetEventRecorderFor("cassandrabackup-controller"),
		scheduler: NewScheduler(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("cassandrabackup-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Filter event types for BackupCRD
	pred := predicate.Funcs{
		// Always handle create events
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			object, err := meta.Accessor(e.ObjectNew)
			if err != nil {
				return false
			}
			if _, ok := object.(*api.CassandraBackup); ok {
				new := e.ObjectNew.(*api.CassandraBackup)
				if new.Status == nil || new.Status.State != api.BackupRunning {
					return true
				}

				return false
			}
			return false
		},
	}

	// Watch for changes to primary resource CassandraBackup
	err = c.Watch(&source.Kind{Type: &api.CassandraBackup{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner CassandraBackup
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &api.CassandraBackup{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCassandraBackup implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCassandraBackup{}

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
				return fmt.Errorf("there is no awsregion property "+
					"while you have set both awssecretaccesskey and awsaccesskeyid in %s secret for backups", secret.Name)
			}
		}

		if len(secret.Data["awsendpoint"]) != 0 && len(secret.Data["awsregion"]) == 0 {
			return fmt.Errorf("awsendpoint is specified but awsregion is not set in %s secret for backups", secret.Name)
		}
	}

	return nil
}

