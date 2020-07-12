package cassandrabackup

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/backrest"
	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const annotationLastApplied string = "cassandrabackups.db.orange.com/last-applied-configuration"

// ReconcileCassandraBackup reconciles a CassandraBackup object
type ReconcileCassandraBackup struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	recorder  record.EventRecorder
	scheduler Scheduler
}

func (r *ReconcileCassandraBackup) listPods(namespace string, selector map[string]string) (*corev1.PodList, error) {

	clientOpt := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(selector),
	}

	opt := []client.ListOption{
		clientOpt,
	}

	pl := &corev1.PodList{}
	return pl, r.client.List(context.TODO(), pl, opt...)
}

func (r *ReconcileCassandraBackup) setFinalizer(cb *api.CassandraBackup, value bool) {
	// Add or Remove Finalizer depending on value
	if (len(cb.Finalizers) > 0) != value {
		cb.PreventBackupDeletion(value)
	}
}

// Reconcile reads that state of the cluster for a CassandraBackup object and makes changes based on the state read
// and what is in the CassandraBackup.Spec
func (r *ReconcileCassandraBackup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := logrus.WithFields(logrus.Fields{"Request.Namespace": request.Namespace, "Request.Name": request.Name})
	reqLogger.Info("Reconciling CassandraBackup")

	// Fetch the CassandraBackup backup
	cb := &api.CassandraBackup{}

	if err := r.client.Get(context.TODO(), request.NamespacedName, cb); err != nil {
		if k8sErrors.IsNotFound(err) {
			// if the resource is not found, that means all of
			// the finalizers have been removed, and the resource has been deleted,
			// so there is nothing left to do.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// If deleted and a schedule was configured, we remove the cron task
	if cb.DeletionTimestamp != nil && cb.IsScheduled() {
		r.scheduler.Remove(cb.Name)
		r.recorder.Event(
			cb,
			corev1.EventTypeNormal,
			"BackupTaskUnscheduled",
			fmt.Sprintf("Controller unscheduled cron task %s to back up cluster %s under snapshot %s",
				cb.Name, cb.Spec.CassandraCluster, cb.Spec.SnapshotTag))
		// Remove Finalizer
		r.setFinalizer(cb, false)
		err := r.client.Update(context.TODO(), cb)
		if err != nil {
			logrus.WithFields(logrus.Fields{"backup": request.NamespacedName, "cluster": cb.Name,
				"err": err}).Error("Issue when updating CassandraBackup")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	instanceChanged := true

	if lac, _ := cb.ComputeLastAppliedConfiguration(); lac == cb.Annotations[annotationLastApplied] {
		instanceChanged = false
	} else {
		cb.Annotations[annotationLastApplied] = lac
		defer r.client.Update(context.TODO(), cb)
	}

	if !cb.IsScheduled() && cb.Status != nil {
		logrus.WithFields(logrus.Fields{"backup": request.NamespacedName}).Info("Reconcilliation stopped as backup already run")
		return reconcile.Result{}, nil
	} else if cb.IsScheduled() && r.scheduler.Contains(cb.Name) && !instanceChanged {
		logrus.WithFields(logrus.Fields{"backup": cb.Name}).Info("Reconcilliation stopped as backup already scheduled and there are no new changes")
		return reconcile.Result{}, nil
	}

	cb.Status = &api.CassandraBackupStatus{}

	if exists, err := existingNotScheduledSnapshot(r.client, cb); err != nil {
		return reconcile.Result{}, err
	} else if exists {
		// We can not backup with same snapshot, CassandraCluster and storageLocation
		r.recorder.Event(
			cb,
			corev1.EventTypeWarning,
			"BackupSkipped",
			fmt.Sprintf("Datacenter %s in cluster %s was not backed up to %s under snapshot %s because such backup already exists",
				cb.Spec.Datacenter, cb.Spec.CassandraCluster, cb.Spec.StorageLocation, cb.Spec.SnapshotTag))
		return reconcile.Result{}, nil
	}

	// fetch secret and make sure it exists
	secret := &corev1.Secret{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: cb.Spec.Secret, Namespace: cb.Namespace}, secret); err != nil {
		if k8sErrors.IsNotFound(err) {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"BackupFailedSecretNotFound",
				fmt.Sprintf("Secret %s used for backups was not found", cb.Spec.Secret))
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Based on storage location, be sure that respective secret entry is there so we error out asap
	if err := validateBackupSecret(secret, cb, reqLogger); err != nil {
		return reconcile.Result{}, err
	}

	// Validate the duration if it's set
	if cb.Spec.Duration != "" {
		if _, err := time.ParseDuration(cb.Spec.Duration); err != nil {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"BackupFailedDurationParseError",
				fmt.Sprintf("Duration %s can't be parsed", cb.Spec.Duration))
			return reconcile.Result{}, nil
		}
	}

	// Get CassandraCluster object
	cc := &api.CassandraCluster{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: cb.Spec.CassandraCluster, Namespace: cb.Namespace}, cc); err != nil {
		if k8sErrors.IsNotFound(err) {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"CassandraClusterNotFound",
				fmt.Sprintf("Datacenter %s of cluster %s to backup not found", cb.Spec.Datacenter, cb.Spec.CassandraCluster))

			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if cb.IsScheduled() { // Add or update a cron task to handle the backup

		if err := cb.Spec.ValidateSchedule(); err != nil {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"BackupInvalidSchedule",
				fmt.Sprintf("Schedule %s is not valid: %s", cb.Spec.Schedule, err.Error()))

			// If schedule is invalid and the object already existed, we reset it to what it was
			if instanceChanged {
				logrus.WithFields(logrus.Fields{
					"backup": request.NamespacedName,
				}).Info("Resetting the schedule to what it was previously. Everything else will be updated")
				//We retrieved our last-applied-configuration stored in the object
				var oldCassandraBackup api.CassandraBackup
				err := json.Unmarshal([]byte(cb.Annotations[annotationLastApplied]), &oldCassandraBackup)
				if err != nil {
					logrus.WithFields(logrus.Fields{"backup": request.NamespacedName,
						"err": err}).Error("Issue when resetting schedule, we'll give it another try")
					return reconcile.Result{}, err
				}

				cb.Spec.Schedule = oldCassandraBackup.Spec.Schedule
				cb.Annotations[annotationLastApplied], _ = cb.ComputeLastAppliedConfiguration()
				return reconcile.Result{}, err
			}

			// The schedule is invalid and the object is new, we stop here
			return reconcile.Result{}, nil
		}

		if skip, err := r.scheduler.AddOrUpdate(cb,
			func() { r.backupData(cb, cc, reqLogger) },
			&r.recorder); err != nil {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"BackupScheduleError",
				fmt.Sprintf("Wasn't able to schedule job %s: %s", cb.Name, err.Error()))
		} else if skip {
			return reconcile.Result{}, nil
		}

		// Add Finalizer
		r.setFinalizer(cb, true)
	}

	if cb.IsScheduled() {
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, r.backupData(cb, cc, reqLogger)
}

func (r *ReconcileCassandraBackup) backupData(cassandraBackup *api.CassandraBackup, cc *api.CassandraCluster,
	reqLogger *logrus.Entry) error {

	pods, err := r.listPods(cassandraBackup.Namespace, k8s.LabelsForCassandraDC(cc, cassandraBackup.Spec.Datacenter))
	if err != nil {
		return fmt.Errorf("unable to list pods")
	}

	pod := pods.Items[random.Intn(len(pods.Items))]
	cassandraBackup.Status = &api.CassandraBackupStatus{
		CoordinatorMember: pod.Name,
	}
	client := &backupClient{backup: cassandraBackup, client: r.client}
	client.updateStatus(&api.CassandraBackupStatus{},reqLogger )

	backupClient, _ := backrest.NewClientFromBackup(r.client, cc, cassandraBackup, &pod)

	go backup(backupClient, client, reqLogger, r.recorder)

	return nil
}
