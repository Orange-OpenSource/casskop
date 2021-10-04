package cassandrarestore

import (
	"github.com/sirupsen/logrus"
	"math/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	api "github.com/Orange-OpenSource/casskop/api/v2"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	AnnotationLastApplied string = "cassandrarestores.db.orange.com/last-applied-configuration"
)

// initialize local pseudorandom generator
var random = rand.New(rand.NewSource(time.Now().Unix()))


func (r *CassandraRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
			if len(new.Status.CoordinatorMember)<1 {
				return false
			}
			if new.Status.Condition == nil {
				return true
			}
			restoreConditionType := api.RestoreConditionType(new.Status.Condition.Type)
			if restoreConditionType.IsCompleted() {
				reqLogger.Info("Restore is completed, skipping.")
				return false
			}

			if restoreConditionType.IsInError() {
				reqLogger.Info("Restore is in error state, skipping.")
				return false
			}
			return true
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.CassandraRestore{}).
		WithEventFilter(pred).
		Complete(r)
}

var _ reconcile.Reconciler = &CassandraRestoreReconciler{}
