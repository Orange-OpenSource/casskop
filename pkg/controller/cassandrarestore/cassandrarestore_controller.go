package cassandrarestore

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/backrest"
	"github.com/Orange-OpenSource/casskop/pkg/controller/common"
	"github.com/Orange-OpenSource/casskop/pkg/errorfactory"
	k8sutil "github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/Orange-OpenSource/casskop/pkg/util"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	AnnotationLastApplied string = "cassandrarestores.db.orange.com/last-applied-configuration"
)

var log = logf.Log.WithName("controller_cassandracluster")


// Add creates a new CassandraRestore Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCassandraRestore{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {

	// Create a new controller
	c, err := controller.New("cassandrarestore-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			object, err := meta.Accessor(e.Object)
			if err != nil {
				return false
			}
			restore, _ := object.(*api.CassandraRestore)
			reqLogger := log.WithValues("Restore.Namespace", restore.Namespace, "Restore.Name", restore.Name)
			cond := api.GetRestoreCondition(&restore.Status, api.RestoreRequired)
			if cond != nil {
				reqLogger.Info("Restore is already scheduled on Cluster member %s", restore.Spec.CoordinatorMember)
				return false
			}

			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			object, err := meta.Accessor(e.ObjectNew)
			if err != nil {
				return false
			}
			restore, _ := object.(*api.CassandraRestore)

			reqLogger := log.WithValues("Restore.Namespace", restore.Namespace, "Restore.Name", restore.Name)
			//old := e.ObjectOld.(*api.CassandraRestore)
			new := e.ObjectNew.(*api.CassandraRestore)
			if new.Status.Condition == nil {
				return false
			}
			if new.Status.Condition.Type.IsCompleted() {
				reqLogger.Info("Restore is completed, skipping.")
				return false
			}

			if new.Status.Condition.Type.IsInError() {
				reqLogger.Info("Restore is in error state, skipping.")
				return false
			}
			return true
		},
	}

	// Watch for changes to primary resource CassandraRestore
	err = c.Watch(&source.Kind{Type: &api.CassandraRestore{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCassandraRestore{}

// ReconcileCassandraCluster reconciles a CassandraCluster object
type ReconcileCassandraRestore struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	cc       *api.CassandraCluster
	recorder record.EventRecorder
	client   client.Client
	scheme   *runtime.Scheme
}

func (r ReconcileCassandraRestore) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CassandraBackup")

	ctx := context.TODO()

	// Fetch the CassandraRestore instance
	instance := &api.CassandraRestore{}
	err := r.client.Get(ctx, request.NamespacedName, instance)

	if err != nil {
		if apiErrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return common.Reconciled()
		}
		// Error reading the object - requeue the request.
		return common.RequeueWithError(reqLogger, err.Error(), err)
	}

	// Check the referenced Cluster exists.
	cc := &api.CassandraCluster{}
	if cc, err = k8sutil.LookupCassandraCluster(r.client, instance.Spec.CassandraClusterRef, instance.Namespace); err != nil {
		// This shouldn't trigger anymore, but leaving it here as a safetybelt
		if k8sutil.IsMarkedForDeletion(instance.ObjectMeta) {
			reqLogger.Info("Cluster is gone already, there is nothing we can do")
			return common.Reconciled()
		}
		r.recorder.Event(
			instance,
			corev1.EventTypeWarning,
			"CassandraClusterNotFound",
			fmt.Sprintf("Cassandra Cluster %s to restore not found",instance.Spec.CassandraClusterRef))
		return common.RequeueWithError(reqLogger, "failed to lookup referenced cluster", err)
	}

	// Check the referenced Backup exists.
	backup := &api.CassandraBackup{}
	if backup, err = k8sutil.LookupCassandraBackup(r.client, instance.Spec.BackupRef, instance.Namespace); err != nil {
		r.recorder.Event(
			instance,
			corev1.EventTypeWarning,
			"BackupNotFound",
			fmt.Sprintf("Backup %s to restore not found",instance.Spec.BackupRef))
		return common.RequeueWithError(reqLogger, "failed to lookup referenced backup", err)
	}

	// Require restore on first pod of the cluster.
	if instance.Status.Condition == nil   {
		err = r.requiredRestore(instance, cc, backup, reqLogger)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.ResourceNotReady:
				return ctrl.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			default:
				return common.RequeueWithError(log, err.Error(), err)
			}
		}
		r.recorder.Event(instance,
			corev1.EventTypeNormal,
			"RestoreRequired",
			fmt.Sprintf("Restore task required from backup of datacenter %s of cluster %s to %s under snapshot %s. Restore operation on pod %s",
				backup.Spec.Datacenter, backup.Spec.CassandraCluster,
				backup.Spec.StorageLocation, backup.Spec.SnapshotTag,
				instance.Spec.CoordinatorMember))
	}

	if len(instance.Spec.CoordinatorMember) == 0 {
		return common.RequeueWithError(reqLogger, "No coordinator member to perform the restore", err)
	}

	if instance.Status.Condition.Type.IsRequired(){
		err = r.handleRequiredRestore(instance, cc, backup, reqLogger)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.SidecarNotReady, errorfactory.ResourceNotReady:
				r.recorder.Event(
					instance,
					corev1.EventTypeWarning,
					"PerformRestoreOperationFailed",
					fmt.Sprintf("Restore task from backup of datacenter %s of cluster %s to %s under snapshot %s failed to run, will retry. Restore operation on pod %s",backup.Spec.Datacenter, backup.Spec.CassandraCluster,
						backup.Spec.StorageLocation, backup.Spec.SnapshotTag,
						instance.Spec.CoordinatorMember))
				return ctrl.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			default:
				return common.RequeueWithError(log, err.Error(), err)
			}
		}
		r.recorder.Event(instance,
			corev1.EventTypeNormal,
			"RestoreInitiated",
			fmt.Sprintf("Restore task initiated from backup of datacenter %s of cluster %s to %s under snapshot %s. Restore operation %v on pod %s.",
				backup.Spec.Datacenter, backup.Spec.CassandraCluster,
				backup.Spec.StorageLocation, backup.Spec.SnapshotTag,
				instance.Status.Id, instance.Spec.CoordinatorMember))
	}

	if instance.Status.Condition.Type.IsInProgress() {

		err = r.checkRestoreOperationState(instance, cc, backup, reqLogger)
		if err != nil {
			switch errors.Cause(err).(type) {
			case errorfactory.SidecarNotReady, errorfactory.ResourceNotReady:
				return ctrl.Result{
					RequeueAfter: time.Duration(15) * time.Second,
				}, nil
			case errorfactory.SidecarOperationRunning:
				return ctrl.Result{
					RequeueAfter: time.Duration(20) * time.Second,
				}, nil
			case errorfactory.SidecarOperationFailure:
				return ctrl.Result{
					RequeueAfter: time.Duration(20) * time.Second,
				}, nil
			default:
				return common.RequeueWithError(log, err.Error(), err)
			}
		}
		r.recorder.Event(instance,
			corev1.EventTypeNormal,
			"RestoreCompleted",
			fmt.Sprintf("Restore task from backup of datacenter %s of cluster %s to %s under snapshot %s is completed. Restore operation %v on pod %s.",
				backup.Spec.Datacenter, backup.Spec.CassandraCluster,
				backup.Spec.StorageLocation, backup.Spec.SnapshotTag,
				instance.Status.Id, instance.Spec.CoordinatorMember))
	}
	return common.Reconciled()
}

// requiredRestore select restore coordinator on a specific member of a Cluster
func (r *ReconcileCassandraRestore) requiredRestore(restore *api.CassandraRestore, cc *api.CassandraCluster, backup *api.CassandraBackup, reqLogger logr.Logger) error {
	ns := restore.Namespace

	pods, err := r.listPods(ns, k8sutil.LabelsForCassandraDC(cc, backup.Spec.Datacenter))
	if err != nil {
		return errorfactory.New(errorfactory.ResourceNotReady{}, err, "no pods founds for this dc")
	}

	if len(pods.Items) > 0 {
		restore.Spec.CoordinatorMember = pods.Items[0].Name
		if err := UpdateRestoreStatus(r.client, restore,
			api.CassandraRestoreStatus{
				Condition: &api.RestoreCondition{
					Type: api.RestoreRequired,
					LastTransitionTime: metav1.Now().Format(util.TimeStampLayout),
				},
			}, log); err != nil{
			return errors.WrapIfWithDetails(err, "could not update status for restore", "restore", restore)
		}

		return nil
	}

	return errors.New("No pods found.")
}

func (r *ReconcileCassandraRestore) handleRequiredRestore(restore *api.CassandraRestore, cc *api.CassandraCluster, backup *api.CassandraBackup, reqLogger logr.Logger) error {
	pods, err := r.listPods(restore.Namespace, k8sutil.LabelsForCassandraDC(cc, backup.Spec.Datacenter))
	if err != nil {
		return errorfactory.New(errorfactory.ResourceNotReady{}, err, "no pods founds for this dc")
	}

	sr, err := backrest.NewSidecarRestore(r.client, cc, restore, pods)
	if err != nil {
		reqLogger.Info("Sidecar communication error checking running restore operation")
		return errorfactory.New(errorfactory.SidecarNotReady{}, err, "sidecar communication error")
	}

	restoreStatus, err := sr.PerformRestore(restore, backup)
	if err != nil {
		reqLogger.Info("Sidecar communication error checking running restore operation")
		return errorfactory.New(errorfactory.SidecarNotReady{}, err, "sidecar communication error")
	}

	if err := UpdateRestoreStatus(r.client, restore, *restoreStatus, log); err != nil {
		return errors.WrapIfWithDetails(err, "could not update status for restore", "restore", restore)
	}

	return nil
}

func (r *ReconcileCassandraRestore) checkRestoreOperationState(restore *api.CassandraRestore, cc *api.CassandraCluster, backup *api.CassandraBackup, reqLogger logr.Logger) error {

	pods, err := r.listPods(restore.Namespace, k8sutil.LabelsForCassandraDC(cc, backup.Spec.Datacenter))
	if err != nil {
		return errorfactory.New(errorfactory.ResourceNotReady{}, err, "no pods founds for this dc")
	}

	restoreId := restore.Status.Id
	if restoreId == "" {
		return errors.New("no Restore operation id provided to be checked")
	}

	// Check Restore operation status
	sr, err := backrest.NewSidecarRestore(r.client, cc, restore, pods)
	if err != nil {
		reqLogger.Info("Sidecar communication error checking running Operation", "OperationId", restoreId)
		return errorfactory.New(errorfactory.SidecarNotReady{}, err, "sidecar communication error")
	}
	restoreStatus, err := sr.GetRestorebyId(restoreId)
	if err != nil {
		reqLogger.Info("Sidecar communication error checking running Operation", "OperationId", restoreId)
		return errorfactory.New(errorfactory.SidecarNotReady{}, err, "sidecar communication error")
	}

	if err := UpdateRestoreStatus(r.client, restore, *restoreStatus, log); err != nil {
		return errors.WrapIfWithDetails(err, "could not update status for restore", "restore", restore)
	}

	// Restore operation failed or canceled,
	// TODO : reschedule it by marking restore Condition.State = RestoreRequired ?
	if restore.Status.Condition.Type.IsInError() {
		return errorfactory.New(errorfactory.SidecarOperationFailure{}, err, "Sidecar Operation failed", fmt.Sprintf("restore operation id : %s", restoreId))
	}

	// Restore operation completed successfully
	if restore.Status.Condition.Type.IsCompleted()  {
		return nil
	}

	// TODO : Implement timeout ?
	
	// restore operation still in progress
	log.Info("Sidecar operation is still running", "restoreId", restoreId)
	return errorfactory.New(errorfactory.SidecarOperationRunning{}, errors.New("sidecar restore operation still running"), fmt.Sprintf("restore operation id : %s", restoreId))
}

func (r *ReconcileCassandraRestore) updateAndFetchLatest(ctx context.Context, restore *api.CassandraRestore) (*api.CassandraRestore, error) {
	typeMeta := restore.TypeMeta
	err := r.client.Update(ctx, restore)
	if err != nil {
		return nil, err
	}
	restore.TypeMeta = typeMeta
	return restore, nil
}

func (r *ReconcileCassandraRestore) listPods(namespace string, selector map[string]string) (*corev1.PodList, error) {

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