package cassandrabackup

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Orange-OpenSource/casskop/pkg/common/operations"
	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/sidecar"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	AnnotationLastApplied string = "cassandrabackups.db.orange.com/last-applied-configuration"
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
		// Always ignore changes. Meaning that the backup CRD is inactionable after creation.
		UpdateFunc: func(e event.UpdateEvent) bool {
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
	if cb.DeletionTimestamp != nil && cb.Spec.Schedule != "" {
		r.scheduler.Remove(cb.Name)
		r.recorder.Event(
			cb,
			corev1.EventTypeNormal,
			"BackupTaskUnscheduled",
			fmt.Sprintf("Controller unscheduled cron task %s to back up cluster %s under snapshot %s",
				cb.Name, cb.Spec.CassandraCluster, cb.Spec.SnapshotTag))
		return reconcile.Result{}, nil
	}

	instanceChanged := true
	scheduled := cb.Spec.Schedule != ""
	var lastAppliedConfiguration string

	if lac, _ := cb.ComputeLastAppliedConfiguration(); lac == cb.Annotations[AnnotationLastApplied] {
		instanceChanged = false
	} else {
		lastAppliedConfiguration = lac
	}

	if !scheduled && cb.Status != nil && len(cb.Status) != 0 {
		logrus.WithFields(logrus.Fields{"backup": request.NamespacedName}).Info("Reconcilliation stopped as backup already run")
		return reconcile.Result{}, nil
	} else if scheduled && r.scheduler.Contains(cb.Name) && !instanceChanged {
		logrus.WithFields(logrus.Fields{"backup": cb.Name}).Info("Reconcilliation stopped as backup already scheduled and there are no new changes")
		return reconcile.Result{}, nil
	}

	cb.Status = []*api.CassandraBackupStatus{}

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
				"FailureEvent",
				fmt.Sprintf("Secret %s used for backups was not found", cb.Spec.Secret))
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// based on storage location, be sure that respective secret entry is there so we error out asap
	if err := validateBackupSecret(secret, cb, reqLogger); err != nil {
		return reconcile.Result{}, err
	}

	// Get CassandraCluster
	cc := &api.CassandraCluster{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: cb.Spec.CassandraCluster, Namespace: cb.Namespace}, cc); err != nil {
		if k8sErrors.IsNotFound(err) {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"FailureEvent",
				fmt.Sprintf("Datacenter %s of cluster %s to backup not found", cb.Spec.Datacenter, cb.Spec.CassandraCluster))

			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if scheduled {
		// There is a schedule so we need to add or update a cron task to handle the backup
		schedule := cb.Spec.Schedule

		// Parse the schedule
		if err := cb.Spec.ValidateSchedule(); err != nil {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"FailureEvent",
				fmt.Sprintf("Schedule %s is not valid: %s", schedule, err.Error()))

			// If schedule is invalid and the object already existed, we reset it to what it was
			if instanceChanged {
				logrus.WithFields(logrus.Fields{"backup": request.NamespacedName,
					"cluster": cc.Name}).Info("Resetting the schedule to what it was previously. Everything else will be updated")
				//We retrieved our last-applied-configuration stored in the object
				var oldCassandraBackup api.CassandraBackup
				err := json.Unmarshal([]byte(cb.Annotations[AnnotationLastApplied]), &oldCassandraBackup)
				if err != nil {
					logrus.WithFields(logrus.Fields{"backup": request.NamespacedName, "cluster": cc.Name,
						"err": err}).Error("Issue when resetting schedule, we'll give it another try")
					return reconcile.Result{}, err
				}

				cb.Spec.Schedule = oldCassandraBackup.Spec.Schedule
				lastAppliedConfiguration, _ = cb.ComputeLastAppliedConfiguration()

			} else {
				// The schedule is invalid and the object is new, we stop here
				return reconcile.Result{}, nil
			}
		}

		if err := r.scheduler.AddOrUpdate(cb.Name, cb.Spec,
			func() { r.backupData(cb, cc, reqLogger) }); err != nil {
			r.recorder.Event(
				cb,
				corev1.EventTypeWarning,
				"FailureEvent",
				fmt.Sprintf("Wasn't able to schedule job %s: %s", cb.Name, err.Error()))
			return reconcile.Result{}, err
		}

		cb.Annotations[AnnotationLastApplied] = lastAppliedConfiguration

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

	syncedInstance := &syncedInstance{backup: instance, client: r.client}

	go backup(sidecarClient, syncedInstance, reqLogger, r.recorder)

	r.recorder.Event(
		instance,
		corev1.EventTypeNormal,
		"BackupFinished",
		fmt.Sprintf("Datacenter %s of cluster %s was backed up to %s under snapshot %s",
			instance.Spec.Datacenter, instance.Spec.CassandraCluster,
			instance.Spec.StorageLocation, instance.Spec.SnapshotTag))

	return nil
}

func existingNotScheduledSnapshot(c client.Client, instance *api.CassandraBackup) (bool, error) {

	backupsList := &api.CassandraBackupList{}

	if err := c.List(context.TODO(), backupsList); err != nil {
		return false, err
	}

	for _, existingBackup := range backupsList.Items {
		if existingBackup.Spec.Schedule == "" && // Not scheduled
			existingBackup.Status != nil && // Not started
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

type syncedInstance struct {
	sync.RWMutex
	backup *v1alpha1.CassandraBackup
	client client.Client
}

func backup(
	sidecarClient *sidecar.Client,
	instance *syncedInstance,
	logging logr.Logger,
	recorder record.EventRecorder) {

	backupRequest := &sidecar.BackupRequest{
		// TODO Specify only the bucket name when the sidecar has been updated to remove that requirement
		StorageLocation:       fmt.Sprintf("%s/%s/dcx/nodex", instance.backup.Spec.StorageLocation, instance.backup.Spec.CassandraCluster),
		SnapshotTag:           instance.backup.Spec.SnapshotTag,
		Duration:              instance.backup.Spec.Duration,
		Bandwidth:             instance.backup.Spec.Bandwidth,
		ConcurrentConnections: instance.backup.Spec.ConcurrentConnections,
		Entities:              instance.backup.Spec.Entities,
		Secret:                instance.backup.Spec.Secret,
		Datacenter:            instance.backup.Spec.Datacenter,
		GlobalRequest:         true,
		KubernetesNamespace:   instance.backup.Namespace,
	}

	if operationID, err := sidecarClient.StartOperation(backupRequest); err != nil {
		logging.Error(err, fmt.Sprintf("Error while starting backup operation %v", backupRequest))
	} else {
		podHostname := sidecarClient.Host
		for range time.NewTicker(2 * time.Second).C {
			if r, err := sidecarClient.FindBackup(operationID); err != nil {
				logging.Error(err, fmt.Sprintf("Error while finding submitted backup operation %v", operationID))
				break
			} else {
				instance.updateStatus(podHostname, r)

				if r.State == operations.FAILED {

					recorder.Event(instance.backup,
						corev1.EventTypeWarning,
						"FailureEvent",
						fmt.Sprintf("Backup operation %v on node %s has failed", operationID, podHostname))

					break
				}

				if r.State == operations.COMPLETED {

					recorder.Event(instance.backup,
						corev1.EventTypeNormal,
						"SuccessEvent",
						fmt.Sprintf("Backup operation %v on node %s was completed.", operationID, podHostname))

					break
				}
			}
		}
	}
}

func (si *syncedInstance) updateStatus(podHostname string, r *sidecar.BackupResponse) {
	si.Lock()
	defer si.Unlock()

	status := &api.CassandraBackupStatus{Node: podHostname}

	var existingStatus = false

	for _, v := range si.backup.Status {
		if v.Node == podHostname {
			status = v
			existingStatus = true
			break
		}
	}

	status.Progress = fmt.Sprintf("%v%%", strconv.Itoa(int(r.Progress*100)))
	status.State = r.State

	if !existingStatus {
		si.backup.Status = append(si.backup.Status, status)
	}

	si.backup.GlobalProgress = func() string {
		var progresses = 0

		for _, s := range si.backup.Status {
			var i, _ = strconv.Atoi(strings.TrimSuffix(s.Progress, "%"))
			progresses = progresses + i
		}

		return strconv.FormatInt(int64(progresses/len(si.backup.Status)), 10) + "%"
	}()

	si.backup.GlobalStatus = func() operations.OperationState {
		var statuses backupStatuses = si.backup.Status

		if statuses.contains(operations.FAILED) {
			return operations.FAILED
		} else if statuses.contains(operations.PENDING) {
			return operations.PENDING
		} else if statuses.contains(operations.RUNNING) {
			return operations.RUNNING
		} else if statuses.allMatch(operations.COMPLETED) {
			return operations.COMPLETED
		}

		return operations.UNKNOWN
	}()

	if err := si.client.Update(context.TODO(), si.backup); err != nil {
		println("error updating CassandraBackup backup")
	}
}

type backupStatuses []*api.CassandraBackupStatus

func (statuses backupStatuses) contains(state operations.OperationState) bool {
	for _, s := range statuses {
		if s.State == state {
			return true
		}
	}
	return false
}

func (statuses backupStatuses) allMatch(state operations.OperationState) bool {
	for _, s := range statuses {
		if s.State != state {
			return false
		}
	}
	return true
}
