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
	"github.com/Orange-OpenSource/casskop/pkg/controller/common"
	"k8s.io/apimachinery/pkg/api/resource"
	"net/http"
	"reflect"
	"testing"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/jarcoal/httpmock"
	"github.com/r3labs/diff"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFlipCassandraClusterUpdateSeedListStatusScaleDC2(t *testing.T) {
	assert := assert.New(t)

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	//Allow Update SeedList
	cc.Spec.AutoUpdateSeedList = true

	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)
	assert.Equal(3, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	var a = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo.ns",
	}

	assert.Equal(3, len(cc.Status.SeedList))
	assert.Equal(true, reflect.DeepEqual(a, cc.Status.SeedList))

	//Ask for Scaling
	var nodesPerRack int32 = 2
	cc.Spec.Topology.DC[1].NodesPerRacks = &nodesPerRack
	status := cc.Status.DeepCopy()

	var b = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-1.cassandra-demo.ns",
	}

	dc1rack1sts := common.HelperGetStatefulset(t, "dc1-rack1")

	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)

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

	//Simulate the Update of SeedList (field CASSANDRA_SEEDLIST of init-container bootstrap
	dc1rack1sts.Spec.Template.Spec.InitContainers[1].Env[1].Value = cc.SeedList(&b)
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

func TestFlipCassandraClusterUpdateSeedListStatusScaleDC2ManualSeedList(t *testing.T) {
	assert := assert.New(t)

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	//Allow Update SeedList
	cc.Spec.AutoUpdateSeedList = false

	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)
	assert.Equal(3, len(cc.Status.CassandraRackStatus))

	var firstSeedList = []string{
		"cassandra-demo-dc1-rack1-0.nodomain.com",
		"cassandra-demo-dc1-rack2-0.nodomain.com",
		"cassandra-demo-dc2-rack1-0.nodomain.com",
	}
	cc.Status.SeedList = firstSeedList

	assert.Equal(3, len(cc.Status.SeedList))

	status := cc.Status.DeepCopy()

	// We change the seedlist

	var newSeedList = []string{
		"cassandra-demo-dc1-rack1-0.nodomain.com",
		"cassandra-demo-dc1-rack2-0.nodomain.com",
		"cassandra-demo-dc2-rack1-0.nodomain.com",
		"cassandra-demo-dc2-rack2-0.nodomain.com",
	}
	cc.Status.SeedList = newSeedList

	dc1rack1sts := common.HelperGetStatefulset(t, "dc1-rack1")

	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)

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

}

func TestFlipCassandraClusterUpdateSeedListStatusscaleDC1(t *testing.T) {
	assert := assert.New(t)

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	//Allow Update SeedList
	cc.Spec.AutoUpdateSeedList = true

	//1. Init
	cc.Status.SeedList = cc.InitSeedList()
	var a = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo.ns",
	}
	assert.Equal(true, reflect.DeepEqual(a, cc.Status.SeedList))

	//2. Ask for Scaling
	var nodesPerRack int32 = 2
	cc.Spec.NodesPerRacks = nodesPerRack

	status := cc.Status.DeepCopy()

	//Add pod of dc1-rack1 at the end of existing seedlist
	var b = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc1-rack1-1.cassandra-demo.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo.ns",
	}

	dc1rack1sts := common.HelperGetStatefulset(t, "dc1-rack1")

	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)

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

func TestFlipCassandraClusterUpdateSeedListStatusScaleDown(t *testing.T) {
	assert := assert.New(t)

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	//Allow Update SeedList
	cc.Spec.AutoUpdateSeedList = true

	dc1rack1sts := common.HelperGetStatefulset(t, "dc1-rack1")

	//1. Init
	cc.Status.SeedList = cc.InitSeedList()
	var a = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo.ns",
	}
	assert.Equal(true, reflect.DeepEqual(a, cc.Status.SeedList))

	//2. Ask for Scaling
	var nodesPerRack int32 = 2
	cc.Spec.NodesPerRacks = cc.Spec.NodesPerRacks + 1
	cc.Spec.Topology.DC[1].NodesPerRacks = &nodesPerRack

	status := cc.Status.DeepCopy()

	//Add pod of dc1-rack1 at the end of existing seedlist
	var b = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc1-rack1-1.cassandra-demo.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-1.cassandra-demo.ns",
	}

	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)

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

	assert.Equal(true, reflect.DeepEqual(b, status.SeedList), "Status: %v", status.SeedList)

	//3. Simulate the Update of SeedList
	dc1rack1sts.Spec.Template.Spec.InitContainers[1].Env[1].Value = cc.SeedList(&b)

	//4. Ask for ScaleDown on dc1
	cc.Spec.NodesPerRacks = cc.Spec.NodesPerRacks - 1

	//expecetd : remove element (dc1-rack1-1) in middle of seedlist
	var c = []string{
		"cassandra-demo-dc1-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc1-rack2-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-0.cassandra-demo.ns",
		"cassandra-demo-dc2-rack1-1.cassandra-demo.ns",
	}

	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack1", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc1-rack2", dc1rack1sts, status)
	UpdateStatusIfSeedListHasChanged(cc, "dc2-rack1", dc1rack1sts, status)

	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc1-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ActionUpdateSeedList.Name, status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)

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
func TestCheckNonAllowedChangesNodesTo0(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")

	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	res := rcc.CheckNonAllowedChanges(cc, status)
	assert.Equal(false, res)

	//Global ScaleDown to 0 must be ignored
	cc.Spec.NodesPerRacks = 0
	res = rcc.CheckNonAllowedChanges(cc, status)
	assert.Equal(true, res)
	assert.Equal(int32(1), cc.Spec.NodesPerRacks)
}

func TestCheckNonAllowedChangesMix1(t *testing.T) {
	assert := assert.New(t)
	rcc, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	//Forbidden Changes
	//Global ScaleDown to 0 must be ignored
	cc.Spec.NodesPerRacks = 0         //instead of 1
	cc.Spec.DataCapacity = "4Gi"      //instead of "3Gi"
	cc.Spec.DataStorageClass = "fast" //instead of "local-storage"
	//Allow Changed
	cc.Spec.AutoPilot = false //instead of true

	res := rcc.CheckNonAllowedChanges(cc, status)
	assert.Equal(true, res)

	//Forbidden Changes
	assert.Equal(int32(1), cc.Spec.NodesPerRacks)
	assert.Equal("3Gi", cc.Spec.DataCapacity)
	assert.Equal("local-storage", cc.Spec.DataStorageClass)

	//Allow Change
	assert.Equal(false, cc.Spec.AutoPilot)
}

func TestCheckNonAllowedChangesResourcesIsAllowedButNeedAttention(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	//Allow Changes but need sequential rolling pdate
	//Global ScaleDown to 0 must be ignored
	cc.Spec.Resources.Requests = v1.ResourceList{
		"cpu":    resource.MustParse("2"), //instead of '1'
		"memory": resource.MustParse("2Gi"), //instead of 2Gi
	}

	res := rcc.CheckNonAllowedChanges(cc, status)
	assert.Equal(false, res)

	assert.Equal(resource.MustParse("2"), *cc.Spec.Resources.Requests.Cpu())
	assert.Equal(resource.MustParse("2Gi"), *cc.Spec.Resources.Requests.Memory())

	dcRackName := "dc1-rack1"
	assert.Equal(api.ActionUpdateResources.Name, status.CassandraRackStatus[dcRackName].CassandraLastAction.Name)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus[dcRackName].CassandraLastAction.Status)
	dcRackName = "dc1-rack2"
	assert.Equal(api.ActionUpdateResources.Name, status.CassandraRackStatus[dcRackName].CassandraLastAction.Name)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus[dcRackName].CassandraLastAction.Status)
	dcRackName = "dc2-rack1"
	assert.Equal(api.ActionUpdateResources.Name, status.CassandraRackStatus[dcRackName].CassandraLastAction.Name)
	assert.Equal(api.StatusToDo, status.CassandraRackStatus[dcRackName].CassandraLastAction.Status)
}

func TestCheckNonAllowedChangesRemove2DC(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := HelperInitCluster(t, "cassandracluster-3DC.yaml")
	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	cc.Spec.Topology.DC.Remove(2)
	cc.Spec.Topology.DC.Remove(1)

	// We can't remove more than one DC at once
	res := rcc.CheckNonAllowedChanges(cc, status)
	assert.Equal(true, res)
}

//Updating racks is not allowed
func TestCheckNonAllowedChangesUpdateRack(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := HelperInitCluster(t, "cassandracluster-3DC.yaml")
	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)
	assert.Equal(4, cc.GetDCRackSize())

	//Remove 1 rack/dc at specified index
	cc.Spec.Topology.DC[0].Rack.Remove(1)

	res := rcc.CheckNonAllowedChanges(cc, status)
	assert.Equal(true, res)

	//Topology must have been restored
	assert.Equal(3, cc.GetDCSize())

	//Topology must have been restored
	assert.Equal(4, cc.GetDCRackSize())

	needUpdate = false

	//Remove 1 rack/dc at specified index
	cc.Spec.Topology.DC[0].Rack = append(cc.Spec.Topology.DC[0].Rack, api.Rack{Name: "ForbiddenRack"})

	res = rcc.CheckNonAllowedChanges(cc, status)

	assert.Equal(true, res)

	//Topology must have been restored
	assert.Equal(3, cc.GetDCSize())

	//Topology must have been restored
	assert.Equal(4, cc.GetDCRackSize())
}

//remove only a rack is not allowed
func TestCheckNonAllowedChangesRemoveDCNot0(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := HelperInitCluster(t, "cassandracluster-3DC.yaml")

	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)
	assert.Equal(4, cc.GetDCRackSize())

	//Remove DC at specified index
	cc.Spec.Topology.DC.Remove(1)

	res := rcc.CheckNonAllowedChanges(cc, status)

	//Change not allowed because DC still has nodes
	assert.Equal(true, res)

	//Topology must have been restored
	assert.Equal(3, cc.GetDCSize())

	//Topology must have been restored
	assert.Equal(4, cc.GetDCRackSize())
}

func TestCheckNonAllowedChangesRemoveDC(t *testing.T) {
	assert := assert.New(t)
	rcc, cc := HelperInitCluster(t, "cassandracluster-3DC.yaml")

	//Simulate old spec with nodes at 0
	var nb int32
	cc.Spec.Topology.DC[1].NodesPerRacks = &nb

	status := cc.Status.DeepCopy()
	rcc.updateCassandraStatus(cc, status)

	//Initial Topology
	assert.Equal(3, cc.GetDCSize())
	assert.Equal(4, cc.GetDCRackSize())
	assert.Equal(4, len(status.CassandraRackStatus))

	//Remove a dc at specified index
	cc.Spec.Topology.DC.Remove(1)

	res := rcc.CheckNonAllowedChanges(cc, status)

	//Change allowed because dc has no nodes
	assert.Equal(true, res)

	//Topology must have been updated
	assert.Equal(2, cc.GetDCSize())

	//Topology must have been restored
	assert.Equal(3, cc.GetDCRackSize())

	//Check that status is updated
	assert.Equal(3, len(status.CassandraRackStatus))
}

// TestCheckNonAllowedChangesScaleDown test that operator won't allowed a Scale Down to 0 if there are Pods in dc and
// still has datas replicated
//Uses K8s fake Client, & Jolokia Mock
func TestCheckNonAllowedChangesScaleDown(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := HelperInitCluster(t, "cassandracluster-3DC.yaml")
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
			Labels: map[string]string{
				"app":                                  "cassandracluster",
				"cassandracluster":                     "cassandra-demo",
				"cassandracluster-uid":                 "cassandra-test-uid",
				"cassandraclusters.db.orange.com.dc":   "dc2",
				"cassandraclusters.db.orange.com.rack": "rack1",
				"cluster":                              "k8s.pic",
				"dc-rack":                              "dc2-rack1",
			},
		},
	}
	pod.Status.Phase = v1.PodRunning
	pod.Spec.Hostname = "cassandra-demo2-dc2-rack1-0"
	pod.Spec.Subdomain = "cassandra-demo2-dc2-rack1"
	hostName := k8s.PodHostname(*pod)
	rcc.CreatePod(pod)

	//Mock Jolokia Call to NonLocalKeyspacesInDC
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	keyspacesDescribed := []string{}

	httpmock.RegisterResponder("POST", JolokiaURL(hostName, jolokiaPort),
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

	res := rcc.CheckNonAllowedChanges(cc, status)
	rcc.updateCassandraStatus(cc, status)
	//Change not allowed because DC still has nodes
	assert.Equal(true, res)

	//We have restore nodesperrack
	assert.Equal(int32(1), *cc.Spec.Topology.DC[1].NodesPerRacks)

	//Changes replicated keyspaces (remove demo1 and demo2 which still have replicated datas
	//allKeyspaces is a global test variable
	allKeyspaces = []string{"system", "system_auth", "system_schema", "something", "else"}
	cc.Spec.Topology.DC[1].NodesPerRacks = &nb

	res = rcc.CheckNonAllowedChanges(cc, status)

	//Change  allowed because there is no more keyspace with replicated datas
	assert.Equal(false, res)

	//Nodes Per Rack is still 0
	assert.Equal(int32(0), *cc.Spec.Topology.DC[1].NodesPerRacks)
}

func TestInitClusterWithDeletePVC(t *testing.T) {
	assert := assert.New(t)
	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")

	updateDeletePvcStrategy(cc)
	assert.Equal([]string{"kubernetes.io/pvc-to-delete"}, cc.Finalizers)

	cc.Spec.DeletePVC = false
	updateDeletePvcStrategy(cc)
	assert.Equal([]string{}, cc.Finalizers)
}

func TestHasChange(t *testing.T) {
	assert := assert.New(t)
	changelog := []diff.Change{
		{Type: diff.DELETE, Path: []string{"DC", "1", "Rack", "2", "Name"}},
		{Type: diff.DELETE, Path: []string{"DC", "1", "Rack", "2", "RollingRestart"}},
		{Type: diff.DELETE, Path: []string{"DC", "1", "Rack", "2", "RollingPartition"}},
		{Type: diff.UPDATE, Path: []string{"DC", "2", "Name"}},
		{Type: diff.UPDATE, Path: []string{"DC", "2", "Rack", "2", "Name"}},
		{Type: diff.UPDATE, Path: []string{"DC", "2", "Rack", "2", "RollingRestart"}},
		{Type: diff.UPDATE, Path: []string{"DC", "2", "Rack", "2", "RollingPartition"}},
	}

	assert.False(hasChange(changelog, diff.DELETE, "DC"))
	assert.True(hasChange(changelog, diff.DELETE))
	assert.False(hasChange(changelog, diff.DELETE, "DC", "DC.Rack"))
	assert.True(hasChange(changelog, diff.DELETE, "-DC", "DC.Rack"))
	assert.True(hasChange(changelog, diff.UPDATE, "DC", "DC.Rack"))
	assert.True(hasChange(changelog, diff.UPDATE, "DC"))
	assert.False(hasChange(changelog, diff.UPDATE, "-DC", "DC.Rack"))
	assert.False(hasChange(changelog, diff.CREATE))

	changelog = []diff.Change{
		{Type: diff.UPDATE, Path: []string{"DC", "1", "Rack", "2", "RollingRestart"}},
		{Type: diff.UPDATE, Path: []string{"DC", "1", "NodesPerRacks"}},
	}

	assert.False(hasChange(changelog, diff.UPDATE, "DC"))
	assert.False(hasChange(changelog, diff.UPDATE, "DC.Rack"))

}

// hostIDMap map[string]string, pod *v1.Pod, status *api.CassandraClusterStatus
func TestUpdateCassandraNodesStatusForPod(t *testing.T) {
	hostIDMap := make(map[string]string)
	defaultIP := "127.0.0.1"
	defaultHostID := "a1d1e7fa-8073-408c-94c1-e3678013f90f"

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	cc.Status.CassandraNodesStatus = make(map[string]api.CassandraNodeStatus)

	mkPod := func(podName string, podIp string, ccReady bool) *v1.Pod {
		return &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: podName},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:  cassandraContainerName,
						Ready: ccReady,
					},
					{
						Name:  cassandraContainerName + "B",
						Ready: !ccReady,
					},
				},
				PodIP: podIp,
			},
		}
	}

	var podList []*v1.Pod

	// Pod ready and with correspondance into HostIDMap
	dc1Rack10PodName := "dc1-rack1-0"
	dc1Rack10PodIP := "10.100.150.110"
	dc1Rack10HostID := "3528d662-e4a8-4fb6-88f6-3f21056df7ea"
	cc.Status.CassandraNodesStatus[dc1Rack10PodName] = api.CassandraNodeStatus{NodeIp: defaultIP, HostId: defaultHostID}
	hostIDMap[dc1Rack10PodIP] = dc1Rack10HostID
	podList = append(podList, mkPod(dc1Rack10PodName, dc1Rack10PodIP, true))

	// Pod ready and without correspondance into HostIDMap
	dc1Rack20PodName := "dc1-rack2-0"
	dc1Rack20PodIP := "10.100.150.100"
	cc.Status.CassandraNodesStatus[dc1Rack20PodName] = api.CassandraNodeStatus{NodeIp: defaultIP, HostId: defaultHostID}
	podList = append(podList, mkPod(dc1Rack20PodName, dc1Rack20PodIP, true))

	// Pod not ready and with correspondance into HostIDMap
	dc2Rack10PodName := "dc2-rack1-0"
	dc2Rack10PodIP := "10.100.150.109"
	dc2Rack10HostID := "fsdf6716-dc54-414d-ef27-sdzdgkds04bf"
	cc.Status.CassandraNodesStatus[dc2Rack10PodName] = api.CassandraNodeStatus{NodeIp: defaultIP, HostId: defaultHostID}
	hostIDMap[dc2Rack10PodIP] = dc2Rack10HostID
	podList = append(podList, mkPod(dc2Rack10PodName, dc2Rack10PodIP, false))

	// Pod ready and with correspondance into HostIDMap
	dc2Rack11PodName := "dc2-rack1-1"
	dc2Rack11PodIP := "10.100.140.111"
	dc2Rack11HostID := "5228d662-e4a8-4fb6-88f6-3f21056f7ger"
	cc.Status.CassandraNodesStatus[dc2Rack11PodName] = api.CassandraNodeStatus{NodeIp: defaultIP, HostId: defaultHostID}
	hostIDMap[dc2Rack11PodIP] = dc2Rack11HostID
	podList = append(podList, mkPod(dc2Rack11PodName, dc2Rack11PodIP, true))

	for _, pod := range podList {
		updateCassandraNodesStatusForPod(hostIDMap, pod, &cc.Status)
	}

	assert.Equal(t, cc.Status.CassandraNodesStatus[dc1Rack10PodName], api.CassandraNodeStatus{HostId: dc1Rack10HostID, NodeIp: dc1Rack10PodIP})
	assert.Equal(t, cc.Status.CassandraNodesStatus[dc1Rack20PodName], api.CassandraNodeStatus{HostId: defaultHostID, NodeIp: defaultIP})
	assert.Equal(t, cc.Status.CassandraNodesStatus[dc2Rack10PodName], api.CassandraNodeStatus{HostId: defaultHostID, NodeIp: defaultIP})
	assert.Equal(t, cc.Status.CassandraNodesStatus[dc2Rack11PodName], api.CassandraNodeStatus{HostId: dc2Rack11HostID, NodeIp: dc2Rack11PodIP})
}

func TestCheckPodCrossIpUseCaseForPodKey(t *testing.T) {
	hostIDMap := make(map[string]string)

	mkPod := func(podName string, podIp string, ccReady bool) *v1.Pod {
		return &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: podName},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:  cassandraContainerName,
						Ready: ccReady,
					},
					{
						Name:  cassandraContainerName + "B",
						Ready: !ccReady,
					},
				},
				PodIP: podIp,
			},
		}
	}

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	cc.Status.CassandraNodesStatus = make(map[string]api.CassandraNodeStatus)

	// Pod Ip change but the Ip is not in the Jolokia Hostid - Ip map.
	dc1Rack10PodName := "dc1-rack1-0"
	oldDc1Rack10PodIp := "20.300.150.110"
	dc1Rack10PodIp := "10.100.150.110"
	dc1Rack10HostId := "3528d662-e4a8-4fb6-88f6-3f21056df7ea"
	cc.Status.CassandraNodesStatus[dc1Rack10PodName] = api.CassandraNodeStatus{NodeIp: oldDc1Rack10PodIp, HostId: dc1Rack10HostId}
	hostIDMap[oldDc1Rack10PodIp] = dc1Rack10HostId
	podNotFound := mkPod(dc1Rack10PodName, dc1Rack10PodIp, true)

	pod, _ := checkPodCrossIpUseCaseForPod(hostIDMap, podNotFound, &cc.Status)
	assert.True(t, pod == nil)

	// Pod doesn't Ip hostId are the same
	dc1Rack20PodName := "dc1-rack2-0"
	dc1Rack20PodIp := "10.100.150.100"
	dc1Rack20HostId := "ca716bef-dc68-427d-be27-b4eeede1e072"
	cc.Status.CassandraNodesStatus[dc1Rack20PodName] = api.CassandraNodeStatus{NodeIp: dc1Rack20PodIp, HostId: dc1Rack20HostId}
	hostIDMap[dc1Rack20PodIp] = dc1Rack20HostId
	podNoChange := mkPod(dc1Rack20PodName, dc1Rack20PodIp, true)

	pod, _ = checkPodCrossIpUseCaseForPod(hostIDMap, podNoChange, &cc.Status)
	assert.True(t, pod == nil)

	// Pod ip change and hostId are not the same in cache.
	dc2Rack10PodName := "dc2-rack1-0"
	oldDc2Rack10PodIp := "10.160.150.109"
	dc2Rack10PodIp := "10.100.150.109"
	cachedHostId := "ca716bef-dc68-427d-be27-b4eeede1e072"
	dc2Rack10HostId := "fsdf6716-dc54-414d-ef27-sdzdgkds04bf"
	cc.Status.CassandraNodesStatus[dc2Rack10PodName] = api.CassandraNodeStatus{NodeIp: oldDc2Rack10PodIp, HostId: dc2Rack10HostId}
	hostIDMap[dc2Rack10PodIp] = cachedHostId
	podCrossIp := mkPod(dc2Rack10PodName, dc2Rack10PodIp, false)

	pod, _ = checkPodCrossIpUseCaseForPod(hostIDMap, podCrossIp, &cc.Status)
	assert.Equal(t, pod, podCrossIp)
}

func TestProcessingPods(t *testing.T) {
	hostIDMap := make(map[string]string)

	mkPod := func(podName string, podIp string, restartCount int32) *v1.Pod {
		return &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: podName},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:         cassandraContainerName,
						Ready:        true,
						RestartCount: restartCount,
					},
					{
						Name:         cassandraContainerName + "B",
						Ready:        true,
						RestartCount: 10000,
					},
				},
				PodIP: podIp,
			},
		}
	}

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	cc.Status.CassandraNodesStatus = make(map[string]api.CassandraNodeStatus)

	// Pod ip change and hostId are not the same in cache.
	dc2Rack10PodName := "dc2-rack1-0"
	oldDc2Rack10PodIp := "10.180.150.109"
	dc2Rack10PodIp := "10.100.150.109"
	cachedHostId := "ca716bef-dc68-427d-be27-b4eeede1e072"
	dc2Rack10HostId := "fsdf6716-dc54-414d-ef27-sdzdgkds04bf"
	hostIDMap[dc2Rack10PodIp] = cachedHostId

	// No enough restart
	returnedPod, _ := processingPods(hostIDMap, cc.Spec.RestartCountBeforePodDeletion,
		[]v1.Pod{*mkPod(dc2Rack10PodName, dc2Rack10PodIp, 1)}, &cc.Status)
	cc.Status.CassandraNodesStatus[dc2Rack10PodName] = api.CassandraNodeStatus{NodeIp: oldDc2Rack10PodIp, HostId: dc2Rack10HostId}
	assert.True(t, returnedPod == nil)
	// No enough restart
	returnedPod, _ = processingPods(hostIDMap, cc.Spec.RestartCountBeforePodDeletion,
		[]v1.Pod{*mkPod(dc2Rack10PodName, dc2Rack10PodIp, cc.Spec.RestartCountBeforePodDeletion)}, &cc.Status)
	cc.Status.CassandraNodesStatus[dc2Rack10PodName] = api.CassandraNodeStatus{NodeIp: oldDc2Rack10PodIp, HostId: dc2Rack10HostId}
	assert.True(t, returnedPod == nil)
	// Enough restart
	pod := mkPod(dc2Rack10PodName, dc2Rack10PodIp, 100)
	returnedPod, _ = processingPods(hostIDMap, cc.Spec.RestartCountBeforePodDeletion,
		[]v1.Pod{*pod}, &cc.Status)
	cc.Status.CassandraNodesStatus[dc2Rack10PodName] = api.CassandraNodeStatus{NodeIp: oldDc2Rack10PodIp, HostId: dc2Rack10HostId}
	assert.Equal(t, returnedPod, pod)

	// Test with option disabled
	cc.Spec.RestartCountBeforePodDeletion = 0

	// No enough restart
	returnedPod, _ = processingPods(hostIDMap, cc.Spec.RestartCountBeforePodDeletion,
		[]v1.Pod{*mkPod(dc2Rack10PodName, dc2Rack10PodIp, 1)}, &cc.Status)
	cc.Status.CassandraNodesStatus[dc2Rack10PodName] = api.CassandraNodeStatus{NodeIp: oldDc2Rack10PodIp, HostId: dc2Rack10HostId}
	assert.True(t, returnedPod == nil)
	// No enough restart
	returnedPod, _ = processingPods(hostIDMap, cc.Spec.RestartCountBeforePodDeletion,
		[]v1.Pod{*mkPod(dc2Rack10PodName, dc2Rack10PodIp, cc.Spec.RestartCountBeforePodDeletion)}, &cc.Status)
	cc.Status.CassandraNodesStatus[dc2Rack10PodName] = api.CassandraNodeStatus{NodeIp: oldDc2Rack10PodIp, HostId: dc2Rack10HostId}
	assert.True(t, returnedPod == nil)
	// Enough restart
	returnedPod, _ = processingPods(hostIDMap, cc.Spec.RestartCountBeforePodDeletion,
		[]v1.Pod{*mkPod(dc2Rack10PodName, dc2Rack10PodIp, 100)}, &cc.Status)
	cc.Status.CassandraNodesStatus[dc2Rack10PodName] = api.CassandraNodeStatus{NodeIp: oldDc2Rack10PodIp, HostId: dc2Rack10HostId}
	assert.True(t, returnedPod == nil)
}