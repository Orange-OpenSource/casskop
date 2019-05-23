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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func helperInitCluster(t *testing.T, name string) (*ReconcileCassandraCluster, *api.CassandraCluster) {
	var cc api.CassandraCluster
	err := yaml.Unmarshal(helperLoadBytes(t, name), &cc)
	if err != nil {
		log.Error(err, "error: helpInitCluster")
		os.Exit(-1)
	}

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
		//&ccList,
	}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(api.SchemeGroupVersion, &cc)
	s.AddKnownTypes(api.SchemeGroupVersion, &ccList)
	cl := fake.NewFakeClient(objs...)
	// Create a ReconcileCassandraCluster object with the scheme and fake client.
	rcc := ReconcileCassandraCluster{client: cl, scheme: s}

	cc.InitCassandraRackList()
	return &rcc, &cc
}

func helperGetStatefulset(t *testing.T, dcRackName string) *appsv1.StatefulSet {
	var sts appsv1.StatefulSet
	name := fmt.Sprintf("cassandracluster-2DC-%s-sts.yaml", dcRackName)
	err := yaml.Unmarshal(helperLoadBytes(t, name), &sts)
	if err != nil {
		log.Error(err, "error: helperGetStatefulset")
		os.Exit(-1)
	}
	return &sts
}

func TestFlipCassandraClusterUpdateSeedListStatus_ScaleDC2(t *testing.T) {
	assert := assert.New(t)

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	//Allow Update SeedList
	cc.Spec.AutoUpdateSeedList = true

	assert.Equal(api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)
	assert.Equal(3, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	var a = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.ns",
	}

	assert.Equal(3, len(cc.Status.SeedList))
	assert.Equal(true, reflect.DeepEqual(a, cc.Status.SeedList))

	//Ask for Scaling
	var nodesPerRack int32 = 2
	cc.Spec.Topology.DC[1].NodesPerRacks = &nodesPerRack
	status := cc.Status.DeepCopy()

	var b = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.ns",
		"cassandra-demo-dc2-rack1-1.cassandra-demo-dc2-rack1.ns",
	}

	dc1rack1sts := helperGetStatefulset(t, "dc1-rack1")

	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)

	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Status)
	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	//UpdateClusterStatus
	UpdateCassandraClusterStatusPhase(cc, status)

	//Flip with AutoUpdateSeedList= true -> update status
	FlipCassandraClusterUpdateSeedListStatus(cc, status)

	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	//Simulate the Update of SeedList
	dc1rack1sts.Spec.Template.Spec.Containers[0].Env[1].Value = cc.GetSeedList(&b)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	//No status must have been changed
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	assert.Equal(4, len(status.SeedList))
	assert.Equal(true, reflect.DeepEqual(b, status.SeedList))

}

func TestFlipCassandraClusterUpdateSeedListStatus_scaleDC1(t *testing.T) {
	assert := assert.New(t)

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	//Allow Update SeedList
	cc.Spec.AutoUpdateSeedList = true

	//1. Init
	cc.Status.SeedList = cc.InitSeedList()
	var a = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.ns",
	}
	assert.Equal(true, reflect.DeepEqual(a, cc.Status.SeedList))

	//2. Ask for Scaling
	var nodesPerRack int32 = 2
	cc.Spec.NodesPerRacks = nodesPerRack

	status := cc.Status.DeepCopy()

	//Add pod of dc1-rack1 at the end of existing seedlist
	var b = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.ns",
		"cassandra-demo-dc1-rack1-1.cassandra-demo-dc1-rack1.ns",
	}

	dc1rack1sts := helperGetStatefulset(t, "dc1-rack1")

	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)

	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Status)
	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	//UpdateClusterStatus
	UpdateCassandraClusterStatusPhase(cc, status)

	//Flip with AutoUpdateSeedList= true -> update status
	FlipCassandraClusterUpdateSeedListStatus(cc, status)

	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	assert.Equal(true, reflect.DeepEqual(b, status.SeedList))

}

func TestFlipCassandraClusterUpdateSeedListStatus_scaleDown(t *testing.T) {
	assert := assert.New(t)

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	//Allow Update SeedList
	cc.Spec.AutoUpdateSeedList = true

	dc1rack1sts := helperGetStatefulset(t, "dc1-rack1")

	//1. Init
	cc.Status.SeedList = cc.InitSeedList()
	var a = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.ns",
	}
	assert.Equal(true, reflect.DeepEqual(a, cc.Status.SeedList))

	//2. Ask for Scaling
	var nodesPerRack int32 = 2
	cc.Spec.NodesPerRacks = cc.Spec.NodesPerRacks + 1
	cc.Spec.Topology.DC[1].NodesPerRacks = &nodesPerRack

	status := cc.Status.DeepCopy()

	//Add pod of dc1-rack1 at the end of existing seedlist
	var b = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.ns",
		"cassandra-demo-dc1-rack1-1.cassandra-demo-dc1-rack1.ns",
		"cassandra-demo-dc2-rack1-1.cassandra-demo-dc2-rack1.ns",
	}

	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)

	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Status)
	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	//UpdateClusterStatus
	UpdateCassandraClusterStatusPhase(cc, status)

	//Flip with AutoUpdateSeedList= true -> update status
	FlipCassandraClusterUpdateSeedListStatus(cc, status)

	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	assert.Equal(true, reflect.DeepEqual(b, status.SeedList))

	//3. Simulate the Update of SeedList
	dc1rack1sts.Spec.Template.Spec.Containers[0].Env[1].Value = cc.GetSeedList(&b)

	//4. Ask for ScaleDown on dc1
	cc.Spec.NodesPerRacks = cc.Spec.NodesPerRacks - 1

	//expecetd : remove element (dc1-rack1-1) in middle of seedlist
	var c = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.ns",
		"cassandra-demo-dc2-rack1-1.cassandra-demo-dc2-rack1.ns",
	}

	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)

	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Status)
	assert.Equal(api.StatusConfiguring, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	//Flip with AutoUpdateSeedList= true -> update status
	FlipCassandraClusterUpdateSeedListStatus(cc, status)

	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Status)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	assert.Equal(true, reflect.DeepEqual(c, status.SeedList))

}

//mock example https://github.com/operator-framework/operator-sdk/blob/e74dd322b291b111f78702cf71e5ac843a0c8912/doc/user/unit-testing.md
func TestCheckNonAllowedChanged_NodesTo0(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	res := rcc.CheckNonAllowedChanged(cc, status)
	assert.Equal(false, res)

	//Global ScaleDown to 0 must be ignored
	cc.Spec.NodesPerRacks = 0
	res = rcc.CheckNonAllowedChanged(cc, status)
	assert.Equal(true, res)
	assert.Equal(int32(1), cc.Spec.NodesPerRacks)

}

func TestCheckNonAllowedChanged_Mix1(t *testing.T) {
	assert := assert.New(t)
	rcc, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	//Forbidden Changes
	//Global ScaleDown to 0 must be ignored
	cc.Spec.NodesPerRacks = 0         //instead of 1
	cc.Spec.DataCapacity = "4Gi"      //instead of "3Gi"
	cc.Spec.DataStorageClass = "fast" //instead of "local-storage"
	//Allow Changed
	cc.Spec.AutoPilot = false //instead of true

	res := rcc.CheckNonAllowedChanged(cc, status)
	assert.Equal(true, res)

	//Forbidden Changes
	assert.Equal(int32(1), cc.Spec.NodesPerRacks)
	assert.Equal("3Gi", cc.Spec.DataCapacity)
	assert.Equal("local-storage", cc.Spec.DataStorageClass)

	//Allow Change
	assert.Equal(false, cc.Spec.AutoPilot)

}

func TestCheckNonAllowedChanged_ResourcesIsAllowedButNeedAttention(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	//Allow Changes but need sequential rolling pdate
	//Global ScaleDown to 0 must be ignored
	cc.Spec.Resources.Requests.CPU = "2"      //instead of '1'
	cc.Spec.Resources.Requests.Memory = "2Gi" //instead of 2Gi

	res := rcc.CheckNonAllowedChanged(cc, status)
	assert.Equal(false, res)

	assert.Equal("2", cc.Spec.Resources.Requests.CPU)
	assert.Equal("2Gi", cc.Spec.Resources.Requests.Memory)

	dcRackName := "dc1-rack1"
	assert.Equal(api.ActionUpdateResources, status.CassandraRackStatus[dcRackName].CassandraLastAction.Name)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus[dcRackName].CassandraLastAction.Status)
	dcRackName = "dc1-rack2"
	assert.Equal(api.ActionUpdateResources, status.CassandraRackStatus[dcRackName].CassandraLastAction.Name)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus[dcRackName].CassandraLastAction.Status)
	dcRackName = "dc2-rack1"
	assert.Equal(api.ActionUpdateResources, status.CassandraRackStatus[dcRackName].CassandraLastAction.Name)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus[dcRackName].CassandraLastAction.Status)

}

//Remove 2 dc is not allowed
func TestCheckNonAllowedChanged_Remove2DC(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := helperInitCluster(t, "cassandracluster-3DC.yaml")
	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	//Remove 1 rack/dc at specified index
	cc.Spec.Topology.DC.Remove(2)
	cc.Spec.Topology.DC.Remove(1)

	res := rcc.CheckNonAllowedChanged(cc, status)
	assert.Equal(true, res)

}

//remove only a rack is not allowed
func TestCheckNonAllowedChanged_RemoveRack(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := helperInitCluster(t, "cassandracluster-3DC.yaml")
	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)
	assert.Equal(4, cc.GetDCRackSize())

	//Remove 1 rack/dc at specified index
	cc.Spec.Topology.DC[0].Rack.Remove(1)

	res := rcc.CheckNonAllowedChanged(cc, status)
	assert.Equal(true, res)

	//Topology must have been restored
	assert.Equal(3, cc.GetDCSize())

	//Topology must have been restored
	assert.Equal(4, cc.GetDCRackSize())

}

//remove only a rack is not allowed
func TestCheckNonAllowedChanged_RemoveDCNot0(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := helperInitCluster(t, "cassandracluster-3DC.yaml")
	//Simulate old spec with nodes at 0

	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)
	assert.Equal(4, cc.GetDCRackSize())

	//Remove 1 rack/dc at specified index
	cc.Spec.Topology.DC.Remove(1)

	res := rcc.CheckNonAllowedChanged(cc, status)

	//Change not allowed because dc still has nodes
	assert.Equal(true, res)

	//Topology must have been restored
	assert.Equal(3, cc.GetDCSize())

	//Topology must have been restored
	assert.Equal(4, cc.GetDCRackSize())

}

func TestCheckNonAllowedChanged_RemoveDC(t *testing.T) {
	assert := assert.New(t)
	rcc, cc := helperInitCluster(t, "cassandracluster-3DC.yaml")
	//Simulate old spec with nodes at 0
	var nb int32
	cc.Spec.Topology.DC[1].NodesPerRacks = &nb

	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)
	assert.Equal(4, cc.GetDCRackSize())
	assert.Equal(4, len(status.CassandraRackStatus))

	//Remove 1 rack/dc at specified index
	cc.Spec.Topology.DC.Remove(1)

	res := rcc.CheckNonAllowedChanged(cc, status)

	//Change not allowed because dc still has nodes
	assert.Equal(true, res)

	//Topology must have been restored
	assert.Equal(2, cc.GetDCSize())

	//Topology must have been restored
	assert.Equal(3, cc.GetDCRackSize())

	//Check that status is updated
	assert.Equal(3, len(status.CassandraRackStatus))

}

// TestCheckNonAllowedChanged_ScaleDown test that operator won't allowed a Scale Down to 0 if there are Pods in dc and
// still has datas replicated
//Uses K8s fake client, & Jolokia Mock
func TestCheckNonAllowedChanged_ScaleDown(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := helperInitCluster(t, "cassandracluster-3DC.yaml")
	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	//Create the Pods wanted by the statefulset dc2-rack1 (1 node)
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cassandra-demo-dc2-rack1-0",
			Namespace: "ns",
		},
	}
	pod.Status.Phase = v1.PodRunning
	pod.Spec.Hostname = "cassandra-demo2-dc2-rack1-0"
	pod.Spec.Subdomain = "cassandra-demo2-dc2-rack1"
	hostName := fmt.Sprintf("%s.%s", pod.Spec.Hostname, pod.Spec.Subdomain)
	rcc.CreatePod(pod)

	//Mock Jolokia Call to HasDataInDC
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	keyspacesDescribed := []string{}

	httpmock.RegisterResponder("POST", JolokiaURL(hostName, port),
		func(req *http.Request) (*http.Response, error) {
			var execrequestdata execRequestData
			if err := json.NewDecoder(req.Body).Decode(&execrequestdata); err != nil {
				t.Error("Can't decode request received")
			}
			if execrequestdata.Attribute == "Keyspaces" {
				return httpmock.NewStringResponse(200, keyspaceListString()), nil
			}
			keyspace, ok := execrequestdata.Arguments[0].(string)

			if !ok {
				t.Error("Keyspace can't be nil")
			}

			keyspacesDescribed = append(keyspacesDescribed, keyspace)

			response := `{"request": {"mbean": "org.apache.cassandra.db:type=StorageService",
						  "arguments": ["%s"],
						  "type": "exec",
						  "operation": "describeRingJMX"},
				      "timestamp": 1541908753,
				      "status": 200,`

			//  For keyspace demo1 and demo2 we return token ranges with some of them assigned to nodes on dc2
			if keyspace[:4] == "demo" {
				response += `"value":
					       ["TokenRange(start_token:4572538884437204647, end_token:4764428918503636065, endpoints:[10.244.3.8], rpc_endpoints:[10.244.3.8], endpoint_details:[EndpointDetails(host:10.244.3.8, datacenter:dc1, rack:rack1)])",
						"TokenRange(start_token:-8547872176065322335, end_token:-8182289314504856691, endpoints:[10.244.2.5], rpc_endpoints:[10.244.2.5], endpoint_details:[EndpointDetails(host:10.244.2.5, datacenter:dc1, rack:rack1)])",
						"TokenRange(start_token:-2246208089217404881, end_token:-2021878843377619999, endpoints:[10.244.2.5], rpc_endpoints:[10.244.2.5], endpoint_details:[EndpointDetails(host:10.244.2.5, datacenter:dc2, rack:rack1)])",
						"TokenRange(start_token:-1308323778199165410, end_token:-1269907200339273513, endpoints:[10.244.2.6], rpc_endpoints:[10.244.2.6], endpoint_details:[EndpointDetails(host:10.244.2.6, datacenter:dc1, rack:rack1)])",
						"TokenRange(start_token:8544184416734424972, end_token:8568577617447026631, endpoints:[10.244.2.6], rpc_endpoints:[10.244.2.6], endpoint_details:[EndpointDetails(host:10.244.2.6, datacenter:dc2, rack:rack1)])",
						"TokenRange(start_token:2799723085723957315, end_token:3289697029162626204, endpoints:[10.244.3.7], rpc_endpoints:[10.244.3.7], endpoint_details:[EndpointDetails(host:10.244.3.7, datacenter:dc1, rack:rack1)])"]}`
				return httpmock.NewStringResponse(200, fmt.Sprintf(response, keyspace)), nil
			}
			return httpmock.NewStringResponse(200, fmt.Sprintf(response+`"value": []}`, keyspace)), nil
		},
	)

	// ask scale down to 0
	var nb int32
	cc.Spec.Topology.DC[1].NodesPerRacks = &nb

	res := rcc.CheckNonAllowedChanged(cc, status)
	rcc.updateCassandraStatus(cc, status)
	//Change not allowed because dc still has nodes
	assert.Equal(true, res)

	//We have restore nodesperrack
	assert.Equal(int32(1), *cc.Spec.Topology.DC[1].NodesPerRacks)

	//--Step 3
	//Changes replicated keyspaces (remove demo1 and demo2 which still have replicated datas
	//all Keyspace is global test var
	allKeyspaces = []string{"system", "system_auth", "system_schema", "truc", "muche"}
	cc.Spec.Topology.DC[1].NodesPerRacks = &nb

	res = rcc.CheckNonAllowedChanged(cc, status)

	//Change  allowed because there is no more keyspace with replicated datas
	assert.Equal(false, res)

	//Nodes Per Rack is still 0
	assert.Equal(int32(0), *cc.Spec.Topology.DC[1].NodesPerRacks)

}

func TestInitClusterWithDeletePVC(t *testing.T) {
	assert := assert.New(t)
	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	updateDeletePvcStrategy(cc)
	assert.Equal([]string{"kubernetes.io/pvc-to-delete"}, cc.Finalizers)

	cc.Spec.DeletePVC = false
	updateDeletePvcStrategy(cc)
	assert.Equal([]string{}, cc.Finalizers)
}
