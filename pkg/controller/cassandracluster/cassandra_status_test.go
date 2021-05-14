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
	"github.com/Orange-OpenSource/casskop/pkg/controller/common"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"reflect"

	v1 "k8s.io/api/core/v1"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"strconv"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

var clusterName = "cassandra-demo"
var namespace   = "ns"
var clusterUID = "cassandra-demo-uid"

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
  baseImage: cassandra
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

func HelperInitCluster(t *testing.T, name string) (*ReconcileCassandraCluster,
	*api.CassandraCluster) {
	var cc api.CassandraCluster
	yaml.Unmarshal(common.HelperLoadBytes(t, name), &cc)

	cc.ObjectMeta.UID = types.UID(clusterUID) // Set default UID for cc

	ccList := api.CassandraClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CassandraClusterList",
			APIVersion: api.SchemeGroupVersion.String(),
		},
	}
	//Create Fake client
	//Objects to track in the Fake client
	objs := []runtime.Object{
		&cc,
	}
	// Register operator types with the runtime scheme.
	fakeClientScheme := scheme.Scheme
	fakeClientScheme.AddKnownTypes(api.SchemeGroupVersion, &cc)
	fakeClientScheme.AddKnownTypes(api.SchemeGroupVersion, &ccList)
	cl := fake.NewFakeClientWithScheme(fakeClientScheme, objs...)
	// Create a ReconcileCassandraCluster object with the scheme and fake client.
	rcc := ReconcileCassandraCluster{Client: cl, Scheme: fakeClientScheme}

	cc.InitCassandraRackList()
	return &rcc, &cc
}

func TestUpdateStatusIfSeedListHasChanged(t *testing.T) {
	assert := assert.New(t)

	var cc api.CassandraCluster
	err := yaml.Unmarshal([]byte(cc2Dcs), &cc)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	cc.InitCassandraRackList()

	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)
	assert.Equal(3, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	var expectedSeedList = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc1-rack1-1.cassandra-demo.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-1.cassandra-demo.ns",
	}

	assert.Equal(len(expectedSeedList), len(cc.Status.SeedList))

	assert.Equal(true, reflect.DeepEqual(expectedSeedList, cc.Status.SeedList))

}

//helperCreateCassandraCluster fake create a cluster from the yaml specified
func helperCreateCassandraCluster(t *testing.T, cassandraClusterFileName string) (*ReconcileCassandraCluster,
	*reconcile.Request) {
	assert := assert.New(t)
	rcc, cc := HelperInitCluster(t, cassandraClusterFileName)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cc.Name,
			Namespace: cc.Namespace,
		},
	}

	//The first Reconcile just makes Init
	res, err := rcc.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	err = rcc.Client.Get(context.TODO(), req.NamespacedName, cc)
	if err != nil {
		t.Fatalf("can't get cassandracluster: (%v)", err)
	}
	if cc.Spec.DeletePVC {
		//Check that we have set the Finalizers
		f := cc.GetFinalizers()
		assert.Equal(f[0], "kubernetes.io/pvc-to-delete", "set finalizer for PVC")
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	//Second Reconcile creates objects
	res, err = rcc.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	for _, dc := range cc.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			dcRackName := cc.GetDCRackName(dc.Name, rack.Name)
			//Update Statefulset fake status
			sts := &appsv1.StatefulSet{}
			err = rcc.Client.Get(context.TODO(), types.NamespacedName{Name: cc.Name + "-" + dcRackName,
				Namespace: cc.Namespace},
				sts)
			if err != nil {
				t.Fatalf("get statefulset: (%v)", err)
			}

			//Now simulate sts to be ready for CassKop
			sts.Status.Replicas = *sts.Spec.Replicas
			sts.Status.ReadyReplicas = *sts.Spec.Replicas
			rcc.UpdateStatefulSet(sts)

			//Create Statefulsets associated fake Pods
			podTemplate := v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template",
					Namespace: namespace,
					Labels: map[string]string{
						"cluster":                              cc.Labels["cluster"],
						"dc-rack":                              dcRackName,
						"cassandraclusters.db.orange.com.dc":   dc.Name,
						"cassandraclusters.db.orange.com.rack": rack.Name,
						"app":                                  "cassandracluster",
						"cassandracluster":                     cc.Name,
						"cassandracluster-uid":                 string(cc.GetUID()),
					},
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "cassandra",
							Ready: true,
						},
					},
				},
			}

			for i := 0; i < int(sts.Status.Replicas); i++ {
				pod := podTemplate.DeepCopy()
				pod.Name = sts.Name + strconv.Itoa(i)
				pod.Spec.Hostname = pod.Name
				pod.Spec.Subdomain = cc.Name
				if err = rcc.CreatePod(pod); err != nil {
					t.Fatalf("can't create pod: (%v)", err)
				}
			}

			//We recall Reconcile to update Next rack
			if res, err = rcc.Reconcile(req); err != nil {
				t.Fatalf("reconcile: (%v)", err)
			}
		}
	}

	//Check creation Statuses
	if err = rcc.Client.Get(context.TODO(), req.NamespacedName, cc); err != nil {
		t.Fatalf("can't get cassandracluster: (%v)", err)
	}

	assert.Equal(api.ClusterPhaseRunning.Name, cc.Status.Phase)

	for _, dcRackName := range cc.GetDCRackNames() {
		assert.Equal(cc.Status.CassandraRackStatus[dcRackName].Phase, api.ClusterPhaseRunning.Name,
			"dc-rack: %s", dcRackName)
		assert.Equal(cc.Status.CassandraRackStatus[dcRackName].CassandraLastAction.Name, api.ClusterPhaseInitial.Name,
			"dc-rack: %s", dcRackName)
		assert.Equal(cc.Status.CassandraRackStatus[dcRackName].CassandraLastAction.Status, api.StatusDone,
			"dc-rack %s", dcRackName)
	}
	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.LastClusterAction)
	assert.Equal(api.StatusDone, cc.Status.LastClusterActionStatus)

	return rcc, &req
}

func TestReconcileCassandraCluster(t *testing.T) {

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	rcc, req := helperCreateCassandraCluster(t, "cassandracluster-2DC.yaml")

	//WARNING: ListPod with fieldselector is not working on Client-side
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

// test that we detect an addition of a configmap
func TestUpdateStatusIfconfigMapHasChangedWithNoConfigMap(t *testing.T) {
	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	rcc, req := helperCreateCassandraCluster(t, "cassandracluster-2DC.yaml")

	//WARNING: ListPod with fieldselector is not working on Client-side
	//So CassKop will try to execute podActions in pods without succeed (they are fake pod)
	//https://github.com/kubernetes/client-go/issues/326
	res, err := rcc.Reconcile(*req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if !res.Requeue && res.RequeueAfter == 0 {
		t.Error("reconcile did not requeue request as expected")
	}

	//Test on each statefulset
	for _, dc := range rcc.cc.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			dcRackName := rcc.cc.GetDCRackName(dc.Name, rack.Name)
			//Update Statefulset fake status
			sts := &appsv1.StatefulSet{}
			err = rcc.Client.Get(context.TODO(), types.NamespacedName{Name: rcc.cc.Name + "-" + dcRackName,
				Namespace: rcc.cc.Namespace},
				sts)
			if err != nil {
				t.Fatalf("get statefulset: (%v)", err)
			}
			assert.Equal(t, false, UpdateStatusIfconfigMapHasChanged(rcc.cc, dcRackName, sts, &rcc.cc.Status))
		}
	}

	//Ask for a new ConfigMap
	rcc.cc.Spec.ConfigMapName = "my-super-configmap"
	//Test on each statefulset
	for _, dc := range rcc.cc.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			dcRackName := rcc.cc.GetDCRackName(dc.Name, rack.Name)
			//Update Statefulset fake status
			sts := &appsv1.StatefulSet{}
			err = rcc.Client.Get(context.TODO(), types.NamespacedName{Name: rcc.cc.Name + "-" + dcRackName,
				Namespace: rcc.cc.Namespace},
				sts)
			if err != nil {
				t.Fatalf("get statefulset: (%v)", err)
			}
			assert.Equal(t, true, UpdateStatusIfconfigMapHasChanged(rcc.cc, dcRackName, sts, &rcc.cc.Status))
		}
	}

}

// test that we detect a change in a configmap
func TestUpdateStatusIfconfigMapHasChangedWithConfigMap(t *testing.T) {
	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	rcc, req := helperCreateCassandraCluster(t, "cassandracluster-2DC-configmap.yaml")

	//WARNING: ListPod with fieldselector is not working on Client-side
	//So CassKop will try to execute podActions in pods without succeed (they are fake pod)
	//https://github.com/kubernetes/client-go/issues/326
	res, err := rcc.Reconcile(*req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if !res.Requeue && res.RequeueAfter == 0 {
		t.Error("reconcile did not requeue request as expected")
	}

	//Test on each statefulset
	for _, dc := range rcc.cc.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			dcRackName := rcc.cc.GetDCRackName(dc.Name, rack.Name)
			//Update Statefulset fake status
			sts := &appsv1.StatefulSet{}
			err = rcc.Client.Get(context.TODO(), types.NamespacedName{Name: rcc.cc.Name + "-" + dcRackName,
				Namespace: rcc.cc.Namespace},
				sts)
			if err != nil {
				t.Fatalf("get statefulset: (%v)", err)
			}
			assert.Equal(t, false, UpdateStatusIfconfigMapHasChanged(rcc.cc, dcRackName, sts, &rcc.cc.Status))
		}
	}

	//Ask for a new ConfigMap
	rcc.cc.Spec.ConfigMapName = "my-super-configmap"
	//Test on each statefulset
	for _, dc := range rcc.cc.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			dcRackName := rcc.cc.GetDCRackName(dc.Name, rack.Name)
			//Update Statefulset fake status
			sts := &appsv1.StatefulSet{}
			err = rcc.Client.Get(context.TODO(), types.NamespacedName{Name: rcc.cc.Name + "-" + dcRackName,
				Namespace: rcc.cc.Namespace},
				sts)
			if err != nil {
				t.Fatalf("get statefulset: (%v)", err)
			}
			assert.Equal(t, true, UpdateStatusIfconfigMapHasChanged(rcc.cc, dcRackName, sts, &rcc.cc.Status))
		}
	}

	//Whant to remove the configmap
	rcc.cc.Spec.ConfigMapName = ""
	//Test on each statefulset
	for _, dc := range rcc.cc.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			dcRackName := rcc.cc.GetDCRackName(dc.Name, rack.Name)
			//Update Statefulset fake status
			sts := &appsv1.StatefulSet{}
			err = rcc.Client.Get(context.TODO(), types.NamespacedName{Name: rcc.cc.Name + "-" + dcRackName,
				Namespace: rcc.cc.Namespace},
				sts)
			if err != nil {
				t.Fatalf("get statefulset: (%v)", err)
			}
			assert.Equal(t, true, UpdateStatusIfconfigMapHasChanged(rcc.cc, dcRackName, sts, &rcc.cc.Status))
		}
	}

}

// test that we detect a change in a the docker image
func TestUpdateStatusIfDockerImageHasChanged(t *testing.T) {
	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	rcc, req := helperCreateCassandraCluster(t, "cassandracluster-2DC-configmap.yaml")

	//WARNING: ListPod with fieldselector is not working on Client-side
	//So CassKop will try to execute podActions in pods without succeed (they are fake pod)
	//https://github.com/kubernetes/client-go/issues/326
	res, err := rcc.Reconcile(*req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if !res.Requeue && res.RequeueAfter == 0 {
		t.Error("reconcile did not requeue request as expected")
	}

	//Test on each statefulset
	for _, dc := range rcc.cc.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			dcRackName := rcc.cc.GetDCRackName(dc.Name, rack.Name)
			//Update Statefulset fake status
			sts := &appsv1.StatefulSet{}
			err = rcc.Client.Get(context.TODO(), types.NamespacedName{Name: rcc.cc.Name + "-" + dcRackName,
				Namespace: rcc.cc.Namespace},
				sts)
			if err != nil {
				t.Fatalf("get statefulset: (%v)", err)
			}
			assert.Equal(t, false, UpdateStatusIfDockerImageHasChanged(rcc.cc, dcRackName, sts, &rcc.cc.Status))
		}
	}

	//Ask for a change in CassandraImage version
	rcc.cc.Spec.CassandraImage = "cassandra:new-versionftt"
	//Test on each statefulset
	for _, dc := range rcc.cc.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			dcRackName := rcc.cc.GetDCRackName(dc.Name, rack.Name)
			//Update Statefulset fake status
			sts := &appsv1.StatefulSet{}
			err = rcc.Client.Get(context.TODO(), types.NamespacedName{Name: rcc.cc.Name + "-" + dcRackName,
				Namespace: rcc.cc.Namespace},
				sts)
			if err != nil {
				t.Fatalf("get statefulset: (%v)", err)
			}
			assert.Equal(t, true, UpdateStatusIfDockerImageHasChanged(rcc.cc, dcRackName, sts, &rcc.cc.Status))
		}
	}

}
