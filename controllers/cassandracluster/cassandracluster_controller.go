// Copyright 2019 Orange
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// 	See the License for the specific language governing permissions and
// limitations under the License.

package cassandracluster

import (
	"context"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	"github.com/sirupsen/logrus"

	api "github.com/Orange-OpenSource/casskop/api/v2"
	appsv1 "k8s.io/api/apps/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("controller_cassandracluster")

func (r *CassandraClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.CassandraCluster{}).
		Complete(r)
}

var _ reconcile.Reconciler = &CassandraClusterReconciler{}

// CassandraClusterReconciler reconciles a CassandraCluster object
type CassandraClusterReconciler struct {
	// This Client, initialized using mgr.Client() above, is a split Client
	// that reads objects from the cache and writes to the apiserver
	cc     *api.CassandraCluster
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger

	storedPdb         *policyv1beta1.PodDisruptionBudget
	storedStatefulSet *appsv1.StatefulSet
}

// +kubebuilder:rbac:groups=db.orange.com,resources=cassandraclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=db.orange.com,resources=cassandraclusters/status,verbs=get;update;patch

// Reconcile reads that state of the cluster for a CassandraCluster object and makes changes based on the state read
// and what is in the CassandraCluster.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (rcc *CassandraClusterReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CassandraCluster")

	requeue30 := reconcile.Result{RequeueAfter: 30 * time.Second}
	requeue5 := reconcile.Result{RequeueAfter: 5 * time.Second}
	requeue := reconcile.Result{Requeue: true}
	forget := reconcile.Result{}

	// Fetch the CassandraCluster instance
	rcc.cc = &api.CassandraCluster{}
	cc := rcc.cc
	err := rcc.Client.Get(context.TODO(), request.NamespacedName, cc)
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

	// After first time reconcile, phase will switch to "Initializing".
	if cc.Status.Phase == "" {
		// Simulate initializer.
		changed := cc.SetDefaults()
		if changed {
			updateDeletePvcStrategy(cc)
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Info("Initialization: Update CassandraCluster")
			return requeue, rcc.Client.Update(context.TODO(), cc)
		}
	}
	cc.CheckDefaults()

	if err = rcc.CheckDeletePVC(cc); err != nil {
		return forget, err
	}

	status := cc.Status.DeepCopy()

	//We Update Status at the end
	defer rcc.updateCassandraStatus(cc, status)

	//If non allowed changes on CRD, we return here
	if rcc.CheckNonAllowedChanges(cc, status) {
		return requeue30, nil
	}

	if err = rcc.ensureCassandraPodDisruptionBudget(cc); err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Errorf("ensureCassandraPodDisruptionBudget Error: %v", err)
	}

	// check pods status
	if err = rcc.CheckPodsState(cc, status); err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Errorf("CheckPodsState Error: %v", err)
	}

	//ReconcileRack will also add and initiate new racks, we must not go through racks before this method
	if err = rcc.ReconcileRack(cc, status); err != nil {
		return requeue5, err
	}

	//Do we need to UpdateSeedList
	EnsureSeedListIsUpdatedWhenRequired(cc, status)

	UpdateCassandraClusterStatusPhase(cc, status)

	//We could set different requeue based on current Operation
	return requeue5, nil

}
