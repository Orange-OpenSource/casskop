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

package k8s

import (
	"testing"
	"time"

	"fmt"

	"github.com/ghodss/yaml"

	"github.com/stretchr/testify/assert"
	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
)

func TestLabelTime(t *testing.T) {
	label := LabelTime()
	if !ReLabelTime.MatchString(label) {
		t.Errorf("Label returned is not well formatted")
	}
}

func TestLabelTime2Time(t *testing.T) {
	t1, err := LabelTime2Time("20180621T150405")
	if err != nil || t1 != time.Date(
		2018, 06, 21, 15, 04, 05, 0, time.UTC) {
		t.Errorf("Label containing time cannot be converted to time")
	}
}

func TestGetDCRackLabelsAndNodeSelectorForStatefulSet_WithTopology(t *testing.T) {
	assert := assert.New(t)

	var data = `
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-demo
  labels:
    cluster: k8s.pic 
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
      - name: online
        labels:
          location.dfy.orange.com/site : mts
        rack:
          - name: rack1
            labels: 
              location.dfy.orange.com/street : street1
          - name: rack2
            labels: 
              location.dfy.orange.com/street : street2
      - name: stats
        nodesPerRacks: 2
        labels: 
          location.dfy.orange.com/site : mts
        rack:
          - name: rack1
            labels: 
              location.dfy.orange.com/street : street3
`

	var cc api.CassandraCluster
	err := yaml.Unmarshal([]byte(data), &cc)
	if err != nil {
		fmt.Printf("error: %v", err)
	}

	var dc int = 0
	var rack int = 0

	labels, nodeSelector := GetDCRackLabelsAndNodeSelectorForStatefulSet(&cc, dc, rack)

	assert.Equal("latest", cc.Spec.Version)
	assert.Equal("orangeopensource/cassandra-image", cc.Spec.BaseImage)
	assert.Equal(cc.Spec.Topology.DC[dc].Name, labels["cassandraclusters.db.orange.com.dc"])
	assert.Equal(cc.Spec.Topology.DC[dc].Rack[rack].Name, labels["cassandraclusters.db.orange.com.rack"])
	assert.Equal(2, len(nodeSelector))
	assert.Equal("street1", nodeSelector["location.dfy.orange.com/street"])

	rack = 1
	labels, nodeSelector = GetDCRackLabelsAndNodeSelectorForStatefulSet(&cc, dc, rack)
	assert.Equal(cc.Spec.Topology.DC[dc].Name, labels["cassandraclusters.db.orange.com.dc"])
	assert.Equal(cc.Spec.Topology.DC[dc].Rack[rack].Name, labels["cassandraclusters.db.orange.com.rack"])
	assert.Equal("street2", nodeSelector["location.dfy.orange.com/street"])
}

func TestGetDCRackLabelsAndNodeSelectorForStatefulSet_WithoutTopology(t *testing.T) {
	assert := assert.New(t)

	var data = `
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-street10
  labels:
    cluster: k8s.kaas
spec:
  nodes: 1
  baseImage: "orangeopensource/cassandra-image"
  version: latest
  rollingPartition: 0
  dataCapacity: 10Gi
  dataStorageClass: local-storage
  hardAntiAffinity: true
  resources:
    requests:
      cpu: '4'
      memory: 8Gi
    limits:
      cpu: '4'
      memory: 8Gi
`

	var cc api.CassandraCluster
	err := yaml.Unmarshal([]byte(data), &cc)
	if err != nil {
		fmt.Printf("error: %v", err)
	}

	var dc int = 0
	var rack int = 0

	labels, nodeSelector := GetDCRackLabelsAndNodeSelectorForStatefulSet(&cc, dc, rack)

	assert.Equal("latest", cc.Spec.Version)
	assert.Equal("orangeopensource/cassandra-image", cc.Spec.BaseImage)
	assert.Equal(api.DefaultCassandraDC, labels["cassandraclusters.db.orange.com.dc"])
	assert.Equal(api.DefaultCassandraRack, labels["cassandraclusters.db.orange.com.rack"])
	assert.Equal(0, len(nodeSelector))

	rack = 1
	labels, nodeSelector = GetDCRackLabelsAndNodeSelectorForStatefulSet(&cc, dc, rack)
	assert.Equal(api.DefaultCassandraDC, labels["cassandraclusters.db.orange.com.dc"])
	assert.Equal(api.DefaultCassandraRack, labels["cassandraclusters.db.orange.com.rack"])
	assert.Equal(0, len(nodeSelector))
}

func TestContain(t *testing.T) {
	assert := assert.New(t)

	var a = []string{"1", "2", "3", "4", "5"}
	var b = []string{"1", "2", "3", "4", "5", "6"}

	assert.Equal(true, Contains(a, "2"))
	assert.Equal(true, Contains(b, "6"))
	assert.Equal(false, Contains(a, "6"))

}

func TestContainSlice(t *testing.T) {
	assert := assert.New(t)

	var a = []string{"1", "2", "3", "4", "5"}
	var b = []string{"1", "2", "3", "4", "5", "6"}

	assert.Equal(false, ContainSlice(a, b))
	assert.Equal(true, ContainSlice(b, a))

}

func TestMergeSlice(t *testing.T) {
	assert := assert.New(t)

	var a = []string{"1", "2", "3", "4", "5"}
	var b = []string{"1", "2", "3", "4", "5", "6"}
	var want = []string{"1", "2", "3", "4", "5", "6"}

	result := MergeSlice(a, b)
	assert.Equal(want, result)

	a = []string{"1", "3", "4", "5"}
	b = []string{"1", "2", "3", "4", "5", "6"}
	want = []string{"1", "3", "4", "5", "2", "6"}

	result = MergeSlice(a, b)
	assert.Equal(want, result)

	a = []string{"5", "3", "2"}
	b = []string{"1", "2", "3", "4", "5", "6"}
	want = []string{"5", "3", "2", "1", "4", "6"}

	result = MergeSlice(a, b)
	assert.Equal(want, result)
}
