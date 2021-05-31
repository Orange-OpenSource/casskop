package cassandrabackup

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Orange-OpenSource/casskop/pkg/controller/common"
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

const (
	annotationLastApplied string = "cassandrabackups.db.orange.com/last-applied-configuration"
	backupAlreadyRun = "Reconcilliation stopped as backup already run"
	backupAlreadyScheduled = "Reconcilliation stopped as backup already scheduled and there are no new changes"
	undoScheduleChange = "Resetting the schedule to what it was previously. Everything else will be updated"
	retryFailedUndoSchedule = "Issue when resetting schedule, we'll give it another try"
	secretError = "Your secret is not valid"
    )

// ReconcileCassandraBackup reconciles a CassandraBackup object
type ReconcileCassandraBackup struct {
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
	cassandraBackup := &api.CassandraBackup{}

	if err := r.client.Get(context.TODO(), request.NamespacedName, cassandraBackup); err != nil {
		if k8sErrors.IsNotFound(err) {
			// if the resource is not found, that means all of
			// the finalizers have been removed, and the resource has been deleted,
			// so there is nothing left to do.
			return common.Reconciled()
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// If deleted and a schedule was configured, we remove the cron task
	if cassandraBackup.DeletionTimestamp != nil && cassandraBackup.IsScheduled() {
		r.scheduler.Remove(cassandraBackup.Name)
		r.recorder.Event(
			cassandraBackup,
			corev1.EventTypeNormal,
			"BackupTaskUnscheduled",
			fmt.Sprintf("Controller unscheduled cron task %s to back up cluster %s under snapshot %s",
				cassandraBackup.Name, cassandraBackup.Spec.CassandraCluster, cassandraBackup.Spec.SnapshotTag))
		// Remove Finalizer
		r.setFinalizer(cassandraBackup, false)
		err := r.client.Update(context.TODO(), cassandraBackup)
		if err != nil {
			logrus.WithFields(logrus.Fields{"backup": request.NamespacedName, "cluster": cassandraBackup.Name,
				"err": err}).Error("Issue when updating CassandraBackup")
			return reconcile.Result{}, err
		}
		return common.Reconciled()
	}

	instanceChanged := true

	if lac, _ := cassandraBackup.ComputeLastAppliedAnnotation(); lac == cassandraBackup.Annotations[annotationLastApplied] {
		instanceChanged = false
	} else {
		if cassandraBackup.Annotations == nil {
			cassandraBackup.Annotations = make(map[string] string)
		}
		cassandraBackup.Annotations[annotationLastApplied] = lac
		defer r.client.Update(context.TODO(), cassandraBackup)
	}

	if !cassandraBackup.IsScheduled() && cassandraBackup.Status.Condition != nil {
		logrus.WithFields(logrus.Fields{"backup": request.NamespacedName}).Info(backupAlreadyRun)
		return common.Reconciled()
	} else if cassandraBackup.IsScheduled() && r.scheduler.Contains(cassandraBackup.Name) && !instanceChanged {
		logrus.WithFields(logrus.Fields{"backup": cassandraBackup.Name}).Info(backupAlreadyScheduled)
		return common.Reconciled()
	}

	cassandraBackup.Status = api.BackRestStatus{}

	if exists, err := existingNotScheduledSnapshot(r.client, cassandraBackup); err != nil {
		return reconcile.Result{}, err
	} else if exists {
		// We can not backup with same snapshot, CassandraCluster and storageLocation
		r.recorder.Event(
			cassandraBackup,
			corev1.EventTypeWarning,
			"BackupSkipped",
			fmt.Sprintf("Datacenter %s in cluster %s was not backed up to %s under snapshot %s because such backup already exists",
				cassandraBackup.Spec.Datacenter, cassandraBackup.Spec.CassandraCluster, cassandraBackup.Spec.StorageLocation, cassandraBackup.Spec.SnapshotTag))
		return common.Reconciled()
	}

	// fetch secret and make sure it exists
	secret := &corev1.Secret{}
	if err := r.client.Get(context.TODO(),
		types.NamespacedName{Name: cassandraBackup.Spec.Secret, Namespace: cassandraBackup.Namespace}, secret); err != nil {
		if k8sErrors.IsNotFound(err) {
			r.recorder.Event(
				cassandraBackup,
				corev1.EventTypeWarning,
				"BackupFailedSecretNotFound",
				fmt.Sprintf("Secret %s used for backups was not found", cassandraBackup.Spec.Secret))
			return common.Reconciled()
		}
		return reconcile.Result{}, err
	}

	// Based on storage location, be sure that respective secret entry is there so we error out asap
	if err := validateBackupSecret(secret, cassandraBackup, reqLogger); err != nil {
		logrus.WithFields(logrus.Fields{"backup": request.NamespacedName,
			"err": err}).Error(secretError)
		return reconcile.Result{}, err
	}

	// Validate the duration if it's set
	if cassandraBackup.Spec.Duration != "" {
		if _, err := time.ParseDuration(cassandraBackup.Spec.Duration); err != nil {
			r.recorder.Event(
				cassandraBackup,
				corev1.EventTypeWarning,
				"BackupFailedDurationParseError",
				fmt.Sprintf("Duration %s can't be parsed", cassandraBackup.Spec.Duration))
			return common.Reconciled()
		}
	}

	// Get CassandraCluster object
	cc := &api.CassandraCluster{}
	if err := r.client.Get(context.TODO(),
		types.NamespacedName{Name: cassandraBackup.Spec.CassandraCluster,
			Namespace: cassandraBackup.Namespace}, cc); err != nil {
		if k8sErrors.IsNotFound(err) {
			r.recorder.Event(
				cassandraBackup,
				corev1.EventTypeWarning,
				"CassandraClusterNotFound",
				fmt.Sprintf("Datacenter %s of cluster %s to backup not found", cassandraBackup.Spec.Datacenter, cassandraBackup.Spec.CassandraCluster))

			return common.Reconciled()
		}
		return reconcile.Result{}, err
	}

	if cassandraBackup.IsScheduled() { // Add or update a cron task to handle the backup

		if err := cassandraBackup.Spec.ValidateScheduleFormat(); err != nil {
			r.recorder.Event(
				cassandraBackup,
				corev1.EventTypeWarning,
				"BackupInvalidSchedule",
				fmt.Sprintf("Schedule %s is not valid: %s", cassandraBackup.Spec.Schedule, err.Error()))

			// If schedule is invalid and the object already existed, we reset it to what it was
			if instanceChanged {
				logrus.WithFields(logrus.Fields{"backup": request.NamespacedName}).Info(undoScheduleChange)
				//We retrieved our last-applied-configuration stored in the object
				var oldCassandraBackup api.CassandraBackup
				err := json.Unmarshal([]byte(cassandraBackup.Annotations[annotationLastApplied]), &oldCassandraBackup)
				if err != nil {
					logrus.WithFields(logrus.Fields{"backup": request.NamespacedName,
						"err": err}).Error(retryFailedUndoSchedule)
					return reconcile.Result{}, err
				}

				cassandraBackup.Spec.Schedule = oldCassandraBackup.Spec.Schedule
				cassandraBackup.Annotations[annotationLastApplied], _ = cassandraBackup.ComputeLastAppliedAnnotation()
				return reconcile.Result{}, err
			}

			// The schedule is invalid and the object is new, we stop here
			return common.Reconciled()
		}

		if skipped, err := r.scheduler.AddOrUpdate(
			cassandraBackup, func() { r.backupData(cassandraBackup, cc, reqLogger) }, &r.recorder); err != nil {
			r.recorder.Event(cassandraBackup, corev1.EventTypeWarning, "BackupScheduleError",
				fmt.Sprintf("Wasn't able to schedule job %s: %s", cassandraBackup.Name, err.Error()))
		} else if skipped {
			return common.Reconciled()
		}

		// Add Finalizer
		r.setFinalizer(cassandraBackup, true)
	}

	if cassandraBackup.IsScheduled() {
		return common.Reconciled()
	}

	return reconcile.Result{}, r.backupData(cassandraBackup, cc, reqLogger)
}

func (r *ReconcileCassandraBackup) backupData(cassandraBackup *api.CassandraBackup, cc *api.CassandraCluster,
	reqLogger *logrus.Entry) error {

	pods, err := r.listPods(cassandraBackup.Namespace, k8s.LabelsForCassandraDC(cc, cassandraBackup.Spec.Datacenter))
	if err != nil {
		return fmt.Errorf("unable to list pods")
	}

	backupClient := &backupClient{backup: cassandraBackup, client: r.client}
	backupClient.updateStatus(api.BackRestStatus{},reqLogger)

	if len(pods.Items) == 0 {
		reqLogger.Error(err, fmt.Sprintf("Error while starting backup operation, no pods found"))
		r.recorder.Event(backupClient.backup,
			corev1.EventTypeWarning,
			"BackupNotInitiated",
			fmt.Sprintf("No pods found in datacenter %s of cluster %s, snapshot %s failed.",
				backupClient.backup.Spec.Datacenter, backupClient.backup.Spec.CassandraCluster,
				backupClient.backup.Spec.SnapshotTag))
		return nil
	}

	pod := pods.Items[random.Intn(len(pods.Items))]
	cassandraBackup.Status = api.BackRestStatus{CoordinatorMember: pod.Name}

	backrestClient, _ := backrest.NewClient(r.client, cc, &pod)

	go backup(backrestClient, backupClient, reqLogger, r.recorder)

	return nil
}
