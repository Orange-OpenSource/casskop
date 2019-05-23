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
	"time"

	"github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/k8s"

	"github.com/sirupsen/logrus"

	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_cassandracluster")

// Add creates a new CassandraCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCassandraCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("cassandracluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CassandraCluster
	err = c.Watch(&source.Kind{Type: &api.CassandraCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	/* We currently don't have secondary resource to watch
	// Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner CassandraCluster
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &dbv1alpha1.CassandraCluster{},
	})
	if err != nil {
		return err
	}
	*/

	return nil
}

var _ reconcile.Reconciler = &ReconcileCassandraCluster{}

// ReconcileCassandraCluster reconciles a CassandraCluster object
type ReconcileCassandraCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	cc     *api.CassandraCluster
	client client.Client
	scheme *runtime.Scheme

	storedPdb         *policyv1beta1.PodDisruptionBudget
	storedStatefulSet *appsv1.StatefulSet
}

// Reconcile reads that state of the cluster for a CassandraCluster object and makes changes based on the state read
// and what is in the CassandraCluster.Spec
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (rcc *ReconcileCassandraCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CassandraCluster")

	requeue30 := reconcile.Result{RequeueAfter: 30 * time.Second}
	requeue5 := reconcile.Result{RequeueAfter: 5 * time.Second}
	requeue := reconcile.Result{Requeue: true}
	forget := reconcile.Result{}

	//If not yet created, create a kubernetes client
	k8s.InitClient()

	// Fetch the CassandraCluster instance
	rcc.cc = &api.CassandraCluster{}
	cc := rcc.cc
	err := rcc.client.Get(context.TODO(), request.NamespacedName, cc)
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
			return requeue, rcc.client.Update(context.TODO(), cc)
		}
	}

	err = rcc.CheckDeletePVC(cc)
	if err != nil {
		return forget, err
	}

	status := cc.Status.DeepCopy()

	//We Update Status at the end
	defer rcc.updateCassandraStatus(cc, status)

	//If non allowed changes on CRD, we exit until an operator fix the CRD
	if rcc.CheckNonAllowedChanged(cc, status) {
		return requeue30, nil
	}

	if err = rcc.ensureCassandraPodDisruptionBudget(cc); err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Errorf("ensureCassandraPodDisruptionBudget Error: %v", err)
	}

	//ReconcileRack will also add and initiate new racks, we must not go through racks before this method
	err = rcc.ReconcileRack(cc, status)
	if err != nil {
		return requeue5, err
	}

	//Do we need to UpdateSeedList
	FlipCassandraClusterUpdateSeedListStatus(cc, status)

	UpdateCassandraClusterStatusPhase(cc, status)

	//We could set different requeue based on current Operation
	return requeue5, nil

}
