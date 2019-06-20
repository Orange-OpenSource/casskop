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
	"fmt"
	"k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/kubernetes/scheme"
	//"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"testing"

	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

var name string = "cassandra-demo"
var namespace string = "ns"

var cc2Dcs = `
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-demo
  labels:
    cluster: k8s.pic
  namespace: ns
spec:
  nodesPerRacks: 6
  baseImage: orangeopensource/cassandra-image
  version: latest
  rollingPartition: 0
  dataCapacity: "3Gi"
  dataStorageClass: "local-storage"
  hardAntiAffinity: false
  deletePVC: true
  autoPilot: true
  resources:         
    requests:
      cpu: '1'
      memory: 2Gi
    limits:
      cpu: '1'
      memory: 2Gi
  topology:
    dc:
      - name: dc1
        labels:
          location.dfy.orange.com/site : mts
        rack:
          - name: rack1
            labels: 
              location.dfy.orange.com/street : street1
          - name: rack2
            labels: 
              location.dfy.orange.com/street : street2
      - name: dc2
        nodesPerRacks: 2
        labels: 
          location.dfy.orange.com/site : mts
        rack:
          - name: rack1
            labels: 
              location.dfy.orange.com/street : street3
`

func TestUpdateStatusIfSeedListHasChanged(t *testing.T) {
	assert := assert.New(t)

	var cc api.CassandraCluster
	err := yaml.Unmarshal([]byte(cc2Dcs), &cc)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	cc.InitCassandraRackList()

	assert.Equal(api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)
	assert.Equal(3, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	var a = []string{"cassandra-demo-dc1-rack1-0.cassandra-demo-dc1.ns",
		"cassandra-demo-dc1-rack1-1.cassandra-demo-dc1.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo-dc1.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo-dc2.ns",
		"cassandra-demo-dc2-rack1-1.cassandra-demo-dc2.ns"}

	assert.Equal(5, len(cc.Status.SeedList))

	assert.Equal(true, reflect.DeepEqual(a, cc.Status.SeedList))

}

//helperCreateCassandraCluster fake create a cluster from the yaml specified
func helperCreateCassandraCluster(t *testing.T, cassandraClusterFileName string) (*ReconcileCassandraCluster,
	*reconcile.Request) {
	assert := assert.New(t)
	rcc, cc := helperInitCluster(t, cassandraClusterFileName)


	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cc.Name,
			Namespace: cc.Namespace,
		},
	}

	//The first Reconcile Just make Init
	res, err := rcc.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	err = rcc.client.Get(context.TODO(), req.NamespacedName, cc)
	if err != nil {
		t.Fatalf("can't get cassandracluster: (%v)", err)
	}
	//Check that we have set the Finalizers
	f := cc.GetFinalizers()
	assert.Equal(f[0], "kubernetes.io/pvc-to-delete", "set finalizer for PVC")
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	//Second Reconcile create objects
	res, err = rcc.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	dcRackNames := cc.GetDCRackNames()
	for _, dcRackName := range dcRackNames{
		//Update Statefulset fake status
		sts := &appsv1.StatefulSet{}
		err = rcc.client.Get(context.TODO(), types.NamespacedName{Name: cc.Name+"-"+dcRackName,
			Namespace: cc.Namespace},
		sts)
		if err != nil {
			t.Fatalf("get statefulset: (%v)", err)
		}
		// Check if the quantity of Replicas for this deployment is equals the specification
		dsize := *sts.Spec.Replicas
		if dsize != 1 {
			t.Errorf("dep size (%d) is not the expected size (%d)", dsize, cc.Spec.NodesPerRacks)
		}
		//Now simulate sts to be ready for CassKop
		sts.Status.Replicas = *sts.Spec.Replicas
		sts.Status.ReadyReplicas = *sts.Spec.Replicas
		rcc.UpdateStatefulSet(sts)

		//Create Statefulsets associated fake Pods
		pod := &v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "template",
				Namespace: namespace,
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					v1.ContainerStatus{
						Name: "cassandra",
						Ready: true,
					},
				},
			},
		}

		for i := 0; i < int(sts.Status.Replicas); i++{
			pod.Name = sts.Name+strconv.Itoa(i)
			err = rcc.CreatePod(pod)
			if err != nil {
				t.Fatalf("can't create pod: (%v)", err)
			}
		}

		//We recall Reconcile to update Next rack
		res, err = rcc.Reconcile(req)
		if err != nil {
			t.Fatalf("reconcile: (%v)", err)
		}
	}

	//Check creation Statuses
	err = rcc.client.Get(context.TODO(), req.NamespacedName, cc)
	if err != nil {
		t.Fatalf("can't get cassandracluster: (%v)", err)
	}
	assert.Equal(cc.Status.Phase, api.ClusterPhaseRunning)

	for _, dcRackName := range dcRackNames{
		assert.Equal(cc.Status.CassandraRackStatus[dcRackName].Phase, api.ClusterPhaseRunning)
		assert.Equal(cc.Status.CassandraRackStatus[dcRackName].CassandraLastAction.Name, api.ClusterPhaseInitial)
		assert.Equal(cc.Status.CassandraRackStatus[dcRackName].CassandraLastAction.Status, api.StatusDone)
	}
			assert.Equal(cc.Status.LastClusterAction, api.ClusterPhaseInitial)
		assert.Equal(cc.Status.LastClusterActionStatus, api.StatusDone)

	return rcc, &req
}

func TestReconcileCassandraCluster(t *testing.T) {
	//assert := assert.New(t)

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .

	rcc, req := helperCreateCassandraCluster(t, "cassandracluster-2DC.yaml")

	//WARNING: ListPod with fieldselector is not working on client-side
	//So CassKop will try to execute podActions in pods without succeed (they are fake pod)
	//https://github.com/kubernetes/client-go/issues/326
	res, err := rcc.Reconcile(*req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if !res.Requeue && res.RequeueAfter == 0 {
		t.Error("reconcile did not requeue request as expected")
	}

}