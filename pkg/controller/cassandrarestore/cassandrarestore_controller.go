package cassandrarestore

import (
	"context"
	"time"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
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
	AnnotationLastApplied string = "cassandrabackups.db.orange.com/last-applied-configuration"
)

var log = logf.Log.WithName("controller_cassandracluster")

// Add creates a new CassandraCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
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

	// Watch for changes to primary resource CassandraCluster
	err = c.Watch(&source.Kind{Type: &api.CassandraRestore{}}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				object, err := meta.Accessor(e.Object)
				if err != nil {
					return false
				}
				if restore, ok := object.(*api.CassandraRestore); ok {
					reqLogger := log.WithValues("Restore.Namespace", restore.Namespace, "Restore.Name", restore.Name)
					_, cond := api.GetRestoreCondition(&restore.Status, api.RestoreScheduled)
					if cond != nil && cond.Status == corev1.ConditionTrue {
						reqLogger.Info("Restoreis already scheduled on Cluster member %s", restore.Spec.ScheduledMember)
						return false
					}

					return true
				}
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				object, err := meta.Accessor(e.ObjectNew)
				if err != nil {
					return false
				}
				if _, ok := object.(*api.CassandraRestore); ok {
					old := e.ObjectOld.(*api.CassandraRestore)
					new := e.ObjectNew.(*api.CassandraRestore)
					reqLogger := log.WithValues("Restore.Namespace", new.Namespace, "Restore.Name", new.Name)

					_, cond :=  api.GetRestoreCondition(&new.Status, api.RestoreComplete)
					if cond != nil && cond.Status == corev1.ConditionTrue {
						reqLogger.Info("Restore is Complete, skipping.")
						return false
					}

					_, cond = api.GetRestoreCondition(&new.Status, api.RestoreRunning)
					if cond != nil && cond.Status == corev1.ConditionTrue {
						reqLogger.Info("Restore is Running, skipping.")
						return false
					}

					_, cond = api.GetRestoreCondition(&new.Status, api.RestoreScheduled)
					if cond != nil && cond.Status == corev1.ConditionTrue && new.Spec.ScheduledMember ==  c.podName {
						return true
					}

					reqLogger.Info("Restore is not Scheduled on this agent")
				}
			},
		})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner CassandraRestart
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &api.CassandraRestore{},
	})

	// Watch for changes to secondary resource CassandraBackup and requeue the owner CassandraRestart
	/*err = c.Watch(&source.Kind{Type: &api.CassandraBackup{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &api.CassandraRestore{},
	})

	if err != nil {
		return err
	}*/
	return nil
}

var _ reconcile.Reconciler = &ReconcileCassandraRestore{}

// ReconcileCassandraCluster reconciles a CassandraCluster object
type ReconcileCassandraRestore struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	cc     *api.CassandraCluster
	client client.Client
	scheme *runtime.Scheme
}

func (r ReconcileCassandraRestore) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	requeue30 := reconcile.Result{RequeueAfter: 30 * time.Second}
	requeue5 := reconcile.Result{RequeueAfter: 5 * time.Second}
	requeue := reconcile.Result{Requeue: true}
	forget := reconcile.Result{}

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CassandraBackup")

	ctx := context.TODO()

	// Fetch the CassandraRestore instance
	instance := &api.CassandraRestore{}
	err := r.client.Get(ctx, request.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return forget, nil
		}
		// Error reading the object - requeue the request.
		return forget, err
	}


}
