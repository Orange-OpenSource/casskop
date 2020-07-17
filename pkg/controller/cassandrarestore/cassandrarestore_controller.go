package cassandrarestore

import (
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	AnnotationLastApplied string = "cassandrarestores.db.orange.com/last-applied-configuration"
)

// initialize local pseudorandom generator
var random = rand.New(rand.NewSource(time.Now().Unix()))

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
			reqLogger := logrus.WithFields(logrus.Fields{"Request.Namespace": restore.Namespace,
				"Request.Name": restore.Name})
			cond := api.GetRestoreCondition(&restore.Status, api.RestoreRequired)
			if cond != nil {
				reqLogger.Infof("Restore is already scheduled on Cluster member %s",
					restore.Status.CoordinatorMember)
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
			reqLogger := logrus.WithFields(logrus.Fields{"Request.Namespace": restore.Namespace,
				"Request.Name": restore.Name})
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
