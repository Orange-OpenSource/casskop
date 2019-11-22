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

package multicasskop

import (
	"context"
	"fmt"

	"admiralty.io/multicluster-controller/pkg/reconcile"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	"time"

	"admiralty.io/multicluster-controller/pkg/cluster"
	"admiralty.io/multicluster-controller/pkg/controller"
	apicmc "github.com/Orange-OpenSource/cassandra-k8s-operator/multi-casskop/pkg/apis"
	cmcv1 "github.com/Orange-OpenSource/cassandra-k8s-operator/multi-casskop/pkg/apis/db/v1alpha1"
	apicc "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis"
	ccv1 "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//Clusters defined each kubernetes we want to connect on it
type Clusters struct {
	Name    string
	Cluster *cluster.Cluster
}

//Client is the k8s client to use to connect to each kubernetes
type Client struct {
	name   string
	client client.Client
}

//reconciler is the base struc to be used for MultiCassKop
type reconciler struct {
	clients   []*Client
	cmc       *cmcv1.MultiCasskop
	namespace string
}

// NewController will create k8s clients for each k8s clusters,
// and watch for changes to MultiCasskop and CassandraCluster CRD objects
func NewController(clusters []Clusters, namespace string) (*controller.Controller, error) {
	var clients []*Client

	for i, cluster := range clusters {
		logrus.Infof("Create Client %d for Cluster %s", i+1, cluster.Name)
		client, err := cluster.Cluster.GetDelegatingClient()
		if err != nil {
			return nil, fmt.Errorf("getting delegating client %d for Cluster %s Cluster: %v", i, cluster.Name,
				err)
		}
		clients = append(clients, &Client{cluster.Name, client})
		logrus.Infof("Add CRDs to Cluster %s Scheme", cluster.Name)
		if err := apicc.AddToScheme(cluster.Cluster.GetScheme()); err != nil {
			return nil, fmt.Errorf("adding APIs CassandraCluster to Cluster %s Cluster's scheme: %v", cluster.Name, err)
		}
		if err := apicmc.AddToScheme(cluster.Cluster.GetScheme()); err != nil {
			return nil, fmt.Errorf("adding APIs MultiCasskop to Cluster %s Cluster's scheme: %v", cluster.Name,
				err)
		}
	}

	co := controller.New(&reconciler{clients: clients, namespace: namespace}, controller.Options{})
	for _, value := range clusters {
		logrus.Info("Configuring Watch for MultiCasskop")
		if err := co.WatchResourceReconcileObject(value.Cluster, &cmcv1.MultiCasskop{ObjectMeta: metav1.ObjectMeta{Namespace: namespace}},
			controller.WatchOptions{Namespace: namespace}); err != nil {
			return nil, fmt.Errorf("setting up MultiCasskop watch in Cluster %s Cluster: %v", value.Name, err)
		}
	}
	return co, nil
}

func (r *reconciler) preventClusterDeletion(value bool) {
	if value {
		r.cmc.SetFinalizers([]string{"kubernetes.io/multi-casskop"})
		return
	}
	r.cmc.SetFinalizers([]string{})
}
func (r *reconciler) updateDeletetrategy() bool {

	// Add Finalizer if DeleteCassandraCluster is enabled so that we can delete CassandraCluster
	if *r.cmc.Spec.DeleteCassandraCluster && len(r.cmc.Finalizers) == 0 {
		logrus.WithFields(logrus.Fields{"cluster": r.cmc.Name}).Info(
			"updateDeletetrategy: Will delete CassandraClusters when MultiCasskop is removed")
		r.preventClusterDeletion(true)
		return true
	}
	return false
}

func (r *reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	requeue30 := reconcile.Result{RequeueAfter: 30 * time.Second}
	requeue5 := reconcile.Result{RequeueAfter: 5 * time.Second}
	requeue := reconcile.Result{Requeue: true}
	forget := reconcile.Result{}

	if req.Namespace != r.namespace {
		logrus.Warningf("We don't watch the object in this namespace %s/%s", req.Name, req.Namespace)
		return forget, nil
	}

	logrus.Debugf("Reconcile %v.", req)

	// Fetch the MultiCasskop instance
	// It is stored in the Cluster with index 0 = the first kubernetes cluster given in parameter to multicasskop.
	masterClient := r.clients[0].client
	r.cmc = &cmcv1.MultiCasskop{}
	err := masterClient.Get(context.TODO(), req.NamespacedName, r.cmc)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return forget, nil
		}
		// Error reading the object - requeue the request.
		return requeue, err
	}

	if ok := r.updateDeletetrategy(); ok == true {
		err := masterClient.Update(context.TODO(), r.cmc)
		return requeue, err
	}

	//var storedCC *ccv1.CassandraCluster
	for _, client := range r.clients {
		var cc *ccv1.CassandraCluster
		var found bool
		if found, cc = r.computeCassandraClusterForContext(client); !found {
			logrus.WithFields(logrus.Fields{"kubernetes": client.name}).Warningf("No Cassandra Cluster defined for context: %v", err)
			break
		}

		//If deletion is asked
		if r.cmc.DeletionTimestamp != nil {
			r.deleteCassandraCluster(client, cc)
			continue
		}

		update, storedCC, err := r.CreateOrUpdateCassandraCluster(client, cc)
		if err != nil {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace,
				"kubernetes": client.name}).Errorf("error on CassandraCluster %v", err)
			return requeue5, err
		}
		if update {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace,
				"kubernetes": client.name}).Infof("Just Update CassandraCluster, returning for now..")
			return requeue30, err
		}

		if !r.ReadyCassandraCluster(storedCC) {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace,
				"kubernetes": client.name}).Infof("Cluster is not Ready, "+
				"we requeue [phase=%s / action=%s / status=%s]", storedCC.Status.Phase, storedCC.Status.LastClusterAction, storedCC.Status.LastClusterActionStatus)
			return requeue30, err
		}
	}

	if r.cmc.DeletionTimestamp != nil {
		//We remove the Finalizer
		r.preventClusterDeletion(false)
		err := masterClient.Update(context.TODO(), r.cmc)
		return forget, err
	}

	return requeue30, err
}

func (r *reconciler) namespacedName(name, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
}

//computeCassandraClusterForContext return the CassandraCluster object to create for the current context
//It merges the base definition, with the override part for the specified context in the MultiCasskop CRD
//If client.name don't match a kubernetes context name specified in the override section, it does nothing
func (r *reconciler) computeCassandraClusterForContext(client *Client) (bool, *ccv1.CassandraCluster) {
	base := r.cmc.Spec.Base.DeepCopy()
	for cmcclName, override := range r.cmc.Spec.Override {
		if client.name == cmcclName {
			mergo.Merge(base, override, mergo.WithOverride)
			//Force default values if missing
			base.CheckDefaults()
			return true, base
		}
	}
	return false, nil
}

func (r *reconciler) deleteCassandraCluster(client *Client, cc *ccv1.CassandraCluster) error {
	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace,
		"kubernetes": client.name}).Info("Delete CassandraCluster")
	if err := client.client.Delete(context.TODO(), cc); err != nil {
		return err
	}
	return nil
}
