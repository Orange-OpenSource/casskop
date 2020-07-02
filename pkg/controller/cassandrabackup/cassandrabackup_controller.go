package cassandrabackup

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	csd "github.com/cscetbon/cassandra-sidecar-go-client/pkg/cassandra_sidecar"

	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"

	"github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/sidecar"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
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

const annotationLastApplied string = "cassandrabackups.db.orange.com/last-applied-configuration"

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

// ReconcileCassandraBackup reconciles a CassandraBackup object
type ReconcileCassandraBackup struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	recorder  record.EventRecorder
	scheduler Scheduler
}

func (r *ReconcileCassandraBackup) listPods(namespace string, selector map[string]string) (*v1.PodList, error) {

	clientOpt := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(selector),
	}

	opt := []client.ListOption{
		clientOpt,
	}

	pl := &v1.PodList{}
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
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
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
	if cb.DeletionTimestamp != nil && !cb.IsScheduled() {
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
	var lastAppliedConfiguration string

	if lac, _ := cb.ComputeLastAppliedConfiguration(); lac == cb.Annotations[annotationLastApplied] {
		instanceChanged = false
	} else {
		lastAppliedConfiguration = lac
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

	// Get CassandraCluster
	cc := &api.CassandraCluster{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: cb.Spec.CassandraCluster, Namespace: cb.Namespace}, cc); err != nil {
		if k8sErrors.IsNotFound(err) {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"BackupNotFound",
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
				lastAppliedConfiguration, _ = cb.ComputeLastAppliedConfiguration()
				cb.Annotations[annotationLastApplied] = lastAppliedConfiguration
				r.client.Update(context.TODO(), cb)
				return reconcile.Result{}, err
			}

			// The schedule is invalid and the object is new, we stop here
			return reconcile.Result{}, nil
		}

		if skip, err := r.scheduler.AddOrUpdate(cb,
			func() {
				r.backupData(cb, cc, reqLogger)
			}, &r.recorder); err != nil {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"BackupScheduleError",
				fmt.Sprintf("Wasn't able to schedule job %s: %s", cb.Name, err.Error()))
		} else if skip {
			return reconcile.Result{}, nil
		}

		cb.Annotations[annotationLastApplied] = lastAppliedConfiguration

		// Add Finalizer
		r.setFinalizer(cb, true)

		err := r.client.Update(context.TODO(), cb)
		if err != nil {
			logrus.WithFields(logrus.Fields{"backup": request.NamespacedName, "cluster": cc.Name,
				"err": err}).Error("Issue when updating CassandraBackup")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil

	}

	return reconcile.Result{}, r.backupData(cb, cc, reqLogger)
}

func (r *ReconcileCassandraBackup) backupData(instance *api.CassandraBackup, cc *api.CassandraCluster,
	reqLogger logr.Logger) error {

	pods, err := r.listPods(instance.Namespace, k8s.LabelsForCassandraDC(cc, instance.Spec.Datacenter))
	if err != nil {
		return fmt.Errorf("Unable to list pods")
	}

	chosenPod := pods.Items[random.Intn(len(pods.Items))]
	sidecarClient := sidecar.NewSidecarClient(k8s.PodHostname(chosenPod), &sidecar.DefaultSidecarClientOptions)

	bc := &backupClient{backup: instance, client: r.client}

	go backup(sidecarClient, bc, reqLogger, r.recorder)

	return nil
}

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

func validateBackupSecret(secret *corev1.Secret, backup *v1alpha1.CassandraBackup, logger logr.Logger) error {
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

type backupClient struct {
	backup *v1alpha1.CassandraBackup
	client client.Client
}

func backup(
	sidecarClient *sidecar.Client,
	instance *backupClient,
	logging logr.Logger,
	recorder record.EventRecorder) {

	backupRequest := csd.BackupOperationRequest{
		// TODO Specify only the bucket name when the sidecar has been updated to remove that requirement
		Type_:                 "backup",
		StorageLocation:       instance.backup.Spec.StorageLocation,
		SnapshotTag:           instance.backup.Spec.SnapshotTag,
		Duration:              instance.backup.Spec.Duration,
		Bandwidth:             instance.backup.Spec.Bandwidth,
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

				if r.State == "FAILED" {
					recorder.Event(instance.backup,
						corev1.EventTypeWarning,
						"BackupFailed",
						fmt.Sprintf("Backup operation %v on node %s has failed", operationID, podHostname))
					break
				}

				if r.State == "COMPLETED" {
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
	logging logr.Logger) {

	status := &api.CassandraBackupStatus{Node: podHostname}
	status.Progress = fmt.Sprintf("%v%%", strconv.Itoa(int(r.Progress*100)))
	status.SetBackupStatusState(r.State)

	si.backup.Status = status

	if err := si.client.Update(context.TODO(), si.backup); err != nil {
		logging.Error(err, "Error updating CassandraBackup backup")
	}
}
