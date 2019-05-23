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
	"fmt"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
)

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

func TestUpdateStatusIfSeedListHasChanged(t *testing.T) {
	assert := assert.New(t)

	var cc api.CassandraCluster
	err := yaml.Unmarshal([]byte(cc2Dcs), &cc)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	cc.InitCassandraRackList()

	assert.Equal(api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack1"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["online-rack2"].CassandraLastAction.Name)
	assert.Equal(api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["stats-rack1"].CassandraLastAction.Name)
	assert.Equal(3, len(cc.Status.CassandraRackStatus))

	cc.Status.SeedList = cc.InitSeedList()

	var a = []string{"cassandra-demo-online-rack1-0.cassandra-demo-online-rack1.ns",
		"cassandra-demo-online-rack1-1.cassandra-demo-online-rack1.ns",
		"cassandra-demo-online-rack2-0.cassandra-demo-online-rack2.ns",
		"cassandra-demo-stats-rack1-0.cassandra-demo-stats-rack1.ns",
		"cassandra-demo-stats-rack1-1.cassandra-demo-stats-rack1.ns"}

	assert.Equal(5, len(cc.Status.SeedList))

	assert.Equal(true, reflect.DeepEqual(a, cc.Status.SeedList))

}
