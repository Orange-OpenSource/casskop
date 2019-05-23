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

package v1alpha1

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func helperInitCluster(t *testing.T, name string) *CassandraCluster {
	var cc CassandraCluster
	err := yaml.Unmarshal(helperLoadBytes(t, name), &cc)
	if err != nil {
		//log.Fatal("error: %v", err)
		log.Fatal("error: helpInitCluster")
	}
	cc.InitCassandraRackList()
	return &cc
}

func TestGetNodesPerRacks_NoTopo(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-NoTopo.yaml")

	nodesPerRack := cc.GetNodesPerRacks("dc-rack1")

	assert.Equal(int32(8), nodesPerRack)
}

func TestGetNodesPerRacks_1DC(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-1DC.yaml")

	nodesPerRack := cc.GetNodesPerRacks("online-rack1")
	assert.Equal(int32(7), nodesPerRack)

	nodesPerRack = cc.GetNodesPerRacks("online-rack2")
	assert.Equal(int32(7), nodesPerRack)
}

func TestGetNodesPerRacks_2DC(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	nodesPerRack := cc.GetNodesPerRacks("online-rack1")
	assert.Equal(int32(6), nodesPerRack)

	nodesPerRack = cc.GetNodesPerRacks("online-rack2")
	assert.Equal(int32(6), nodesPerRack)

	nodesPerRack = cc.GetNodesPerRacks("stats-rack1")
	assert.Equal(int32(2), nodesPerRack)

	nodesPerRack = cc.GetNodesPerRacks("toto-toto")
	assert.Equal(int32(6), nodesPerRack)

}

func TestGetNumTokensPerRacks_NoTopo(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-NoTopo.yaml")

	nodesPerRack := cc.GetNumTokensPerRacks("dc-rack1")

	assert.Equal(int32(256), nodesPerRack)

}
func TestGetNumTokensPerRacks_2DC(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	numTokens := cc.GetNumTokensPerRacks("online-rack1")
	assert.Equal(int32(200), numTokens)

	numTokens = cc.GetNumTokensPerRacks("online-rack2")
	assert.Equal(int32(200), numTokens)

	numTokens = cc.GetNumTokensPerRacks("stats-rack1")
	assert.Equal(int32(32), numTokens)

	numTokens = cc.GetNumTokensPerRacks("toto-toto")
	assert.Equal(int32(256), numTokens)

}

func TestGetDCSize(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	nb := cc.GetDCSize()
	assert.Equal(int(2), nb)

}

func TestGetDCName(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	name := cc.GetDCName(0)
	assert.Equal("online", name)

	name = cc.GetDCName(1)
	assert.Equal("stats", name)

}

func TestGetRackSize(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	nb := cc.GetRackSize(0)
	assert.Equal(int(2), nb)

	nb = cc.GetRackSize(1)
	assert.Equal(int(2), nb)

}

func TestGetRackName(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	name := cc.GetRackName(0, 0)
	assert.Equal("rack1", name)

	name = cc.GetRackName(0, 1)
	assert.Equal("rack2", name)

	name = cc.GetRackName(1, 0)
	assert.Equal("rack1", name)

}

func TestGetDCRackName(t *testing.T) {
	assert := assert.New(t)

	var cc CassandraCluster
	assert.Equal("online-rack1", cc.GetDCRackName("online", "rack1"))

	//must log error on matchname DNS-1035
	assert.Equal("", cc.GetDCRackName("online", "Rack1"))
}

func TestInitCassandraRackList_NoTopo(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-NoTopo.yaml")

	cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(1, len(cc.Status.CassandraRackStatus))

	assert.Equal(1, cc.GetDCSize())
	assert.Equal(1, cc.GetRackSize(0))

}

func TestInitCassandraRackList_TopoDCNoRack(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-TopoDCNoRack.yaml")

	cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)
	assert.Equal(1, len(cc.Status.CassandraRackStatus))

	assert.Equal(1, cc.GetDCSize())
	assert.Equal(1, cc.GetRackSize(0))

}

func TestInitCassandraRackList_2DC(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack2"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack2"].CassandraLastAction.Name)
	assert.Equal(4, len(cc.Status.CassandraRackStatus))

}

func TestInitCassandraRackinStatus(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack2"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack2"].CassandraLastAction.Name)
	assert.Equal(4, len(cc.Status.CassandraRackStatus))
	//Add new DC from existing RackStatus
	cc.InitCassandraRackinStatus(&cc.Status, "foo", "bar")

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["foo-bar"].CassandraLastAction.Name)
	assert.Equal(5, len(cc.Status.CassandraRackStatus))
}

func TestGetStatusDCRackSize(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)

	assert.Equal(2, cc.GetDCSize())
	assert.Equal(2, cc.GetRackSize(0))
	assert.Equal(2, cc.GetRackSize(1))

	nb := cc.GetStatusDCRackSize()
	assert.Equal(int(4), nb)

	nb = cc.GetDCRackSize()
	assert.Equal(int(4), nb)

	//Remove 1 rack/dc
	cc.Spec.Topology.DC.Remove(1)

	cc.InitCassandraRackList()

	assert.Equal(1, cc.GetDCSize())
	assert.Equal(2, cc.GetRackSize(0))

	nb = cc.GetStatusDCRackSize()
	assert.Equal(int(2), nb)
	nb = cc.GetDCRackSize()
	assert.Equal(int(2), nb)
}

//Test that a reinit keep history of changes in the status
func TestGetStatusDCRackSize_KeepChanges(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	status := cc.Status.DeepCopy()

	nb := cc.GetStatusDCRackSize()
	assert.Equal(int(4), nb)
	assert.Equal(nb, cc.GetDCRackSize())

	//add info in status
	cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name = ActionUpdateSeedList
	assert.Equal(cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name, ActionUpdateSeedList)

	//Remove 1 dc
	cc.Spec.Topology.DC.Remove(1)

	deleteRackList := cc.FixCassandraRackList(status)

	sort.Strings(deleteRackList)
	assert.Equal("stats-rack1", deleteRackList[0])
	assert.Equal("stats-rack2", deleteRackList[1])

	nb = len(status.CassandraRackStatus)
	assert.Equal(2, nb)
	assert.Equal(nb, cc.GetDCRackSize())
	assert.Equal(cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name, ActionUpdateSeedList)
	assert.Equal(2, len(status.CassandraRackStatus))
}

func TestInitSeedList_2DC(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	//cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack2"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack2"].CassandraLastAction.Name)

	cc.Status.SeedList = cc.InitSeedList()

	assert.Equal("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns", cc.Status.SeedList[0])
	assert.Equal("cassandra-demo-online-rack1-1.cassandra-demo-online-rack1.ns", cc.Status.SeedList[1])
	assert.Equal("cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns", cc.Status.SeedList[2])
	assert.Equal("cassandra-demo-stats-rack1-0.cassandra-demo-stats-rack1.ns", cc.Status.SeedList[3])
	assert.Equal("cassandra-demo-stats-rack1-1.cassandra-demo-stats-rack1.ns", cc.Status.SeedList[4])
	assert.Equal("cassandra-demo-stats-rack2-0.cassandra-demo-stats-rack2.ns", cc.Status.SeedList[5])
	assert.Equal(6, len(cc.Status.SeedList))
}

func TestInitSeedList_NoTopo(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-NoTopo.yaml")

	//cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(1, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	assert.Equal("cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.ns", cc.Status.SeedList[0])
	assert.Equal("cassandra-demo-dc1-rack1-1.cassandra-demo-dc1-rack1.ns", cc.Status.SeedList[1])
	assert.Equal("cassandra-demo-dc1-rack1-2.cassandra-demo-dc1-rack1.ns", cc.Status.SeedList[2])
	assert.Equal(3, len(cc.Status.SeedList))

}

func TestInitSeedList_1DC(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-1DC.yaml")

	//cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack2"].CassandraLastAction.Name)
	assert.Equal(2, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	assert.Equal("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns", cc.Status.SeedList[0])
	assert.Equal("cassandra-demo-online-rack1-1.cassandra-demo-online-rack1.ns", cc.Status.SeedList[1])
	assert.Equal("cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns", cc.Status.SeedList[2])
	assert.Equal(3, len(cc.Status.SeedList))

	assert.Equal("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns,cassandra-demo-online-rack1-1.cassandra-demo-online-rack1.ns,cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns", cc.GetSeedList(&cc.Status.SeedList))

}

func TestInitSeedList_2DC5R(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC5R.yaml")

	//cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack2"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack3"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack4"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack5"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack2"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack3"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack4"].CassandraLastAction.Name)
	assert.Equal(9, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	assert.Equal("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns", cc.Status.SeedList[0])
	assert.Equal("cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns", cc.Status.SeedList[1])
	assert.Equal("cassandra-demo-online-rack3-0.cassandra-demo-online-rack3.ns", cc.Status.SeedList[2])
	assert.Equal("cassandra-demo-stats-rack1-0.cassandra-demo-stats-rack1.ns", cc.Status.SeedList[3])
	assert.Equal("cassandra-demo-stats-rack2-0.cassandra-demo-stats-rack2.ns", cc.Status.SeedList[4])
	assert.Equal("cassandra-demo-stats-rack3-0.cassandra-demo-stats-rack3.ns", cc.Status.SeedList[5])
	assert.Equal(6, len(cc.Status.SeedList))

	assert.Equal("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns,cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns,cassandra-demo-online-rack3-0.cassandra-demo-online-rack3.ns,cassandra-demo-stats-rack1-0.cassandra-demo-stats-rack1.ns,cassandra-demo-stats-rack2-0.cassandra-demo-stats-rack2.ns,cassandra-demo-stats-rack3-0.cassandra-demo-stats-rack3.ns", cc.GetSeedList(&cc.Status.SeedList))
}

func TestInitSeedList_1DC1R1P(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-1DC1R1P.yaml")

	//cc.InitCassandraRackList()

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack2"].CassandraLastAction.Name)
	assert.Equal(2, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	assert.Equal("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns", cc.Status.SeedList[0])
	assert.Equal("cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns", cc.Status.SeedList[1])
	assert.Equal(2, len(cc.Status.SeedList))

	assert.Equal("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns,cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns", cc.GetSeedList(&cc.Status.SeedList))
}

func TestIsPodInSeedList(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-1DC1R1P.yaml")

	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)
	assert.Equal(ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack2"].CassandraLastAction.Name)
	assert.Equal(2, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	assert.Equal("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns", cc.Status.SeedList[0])
	assert.Equal("cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns", cc.Status.SeedList[1])
	assert.Equal(2, len(cc.Status.SeedList))

	assert.Equal("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns,cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns", cc.GetSeedList(&cc.Status.SeedList))

	assert.Equal(true, cc.IsPodInSeedList("cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns"))
	assert.Equal(true, cc.IsPodInSeedList("cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns"))

	assert.Equal(false, cc.IsPodInSeedList("cassandra-demo-online-rack3-0.cassandra-demo-online-rack2.ns"))

}

//Test that a reinit keep history of changes in the status
func TestComputeLastAppliedConfiguration(t *testing.T) {
	assert := assert.New(t)

	cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	lastAppliedConfiguration, _ := cc.ComputeLastAppliedConfiguration()
	result := `{"kind":"CassandraCluster","apiVersion":"db.orange.com/v1alpha1","metadata":{"name":"cassandra-demo","namespace":"ns","creationTimestamp":null,"labels":{"cluster":"k8s.pic"}},"spec":{"nodesPerRacks":6,"baseImage":"orangeopensource/cassandra-image","version":"latest","imagepullpolicy":"","runAsUser":null,"resources":{"requests":{"cpu":"1","memory":"2Gi"},"limits":{"cpu":"1","memory":"2Gi"}},"deletePVC":true,"autoPilot":true,"maxPodUnavailable":0,"dataCapacity":"3Gi","dataStorageClass":"local-storage","imagePullSecret":{},"imageJolokiaSecret":{},"topology":{"dc":[{"name":"online","labels":{"location.dfy.orange.com/site":"mts"},"rack":[{"name":"rack1","labels":{"location.dfy.orange.com/street":"street1"}},{"name":"rack2","labels":{"location.dfy.orange.com/street":"street2"}}],"numTokens":200},{"name":"stats","labels":{"location.dfy.orange.com/site":"mts"},"rack":[{"name":"rack1","labels":{"location.dfy.orange.com/street":"street3"}},{"name":"rack2","labels":{"location.dfy.orange.com/street":"street4"}}],"nodesPerRacks":2,"numTokens":32}]}},"status":{}}`

	//add info in status
	assert.Equal(result, string(lastAppliedConfiguration))

}
