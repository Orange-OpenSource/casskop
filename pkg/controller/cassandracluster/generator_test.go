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
	"testing"

	"github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/k8s"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestCreateNodeAffinity(t *testing.T) {
	assert := assert.New(t)

	nodeAffinity := createNodeAffinity(map[string]string{
		"A": "value1",
		"B": "value2",
		"C": "value3",
		"D": "value4",
		"E": "value5",
	})

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key, "A")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0], "value1")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Key, "B")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Values[0], "value2")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Key, "C")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Values[0], "value3")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Key, "D")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Values[0], "value4")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Key, "E")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Values[0], "value5")
}

func TestCreateNodeAffinitySort(t *testing.T) {
	assert := assert.New(t)

	//unsort labels gives sorted result
	nodeAffinity := createNodeAffinity(map[string]string{
		"B": "value2",
		"A": "value1",
		"D": "value4",
		"E": "value5",
		"C": "value3",
	})

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key, "A")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0], "value1")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Key, "B")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Values[0], "value2")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Key, "C")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Values[0], "value3")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Key, "D")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Values[0], "value4")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Key, "E")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Values[0], "value5")
}

func TestCreatePodAntiAffinityHard(t *testing.T) {
	assert := assert.New(t)

	labels := map[string]string{
		"label1": "value1",
		"label2": "value2",
		"label3": "value3",
	}
	podAntiAffinityHard := createPodAntiAffinity(true, labels)

	assert.Equal(podAntiAffinityHard.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey, hostnameTopologyKey)
	assert.Equal(podAntiAffinityHard.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels, labels)
}

func TestDeleteVolumeMount(t *testing.T) {

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	volumeMounts := generateCassandraVolumeMount(cc)

	assert.Equal(t, 4, len(volumeMounts))
	assert.Equal(t, "/extra-lib", volumeMounts[getPos(volumeMounts, "extra-lib")].MountPath)
	assert.Equal(t, "/etc/cassandra", volumeMounts[getPos(volumeMounts, "bootstrap")].MountPath)

	volumeMounts = deleteVolumeMount(volumeMounts, "bootstrap")
	assert.Equal(t, 3, len(volumeMounts))
	volumeMounts = append(volumeMounts, v1.VolumeMount{Name: "bootstrap", MountPath: "/etc/else"})

	assert.Equal(t, 4, len(volumeMounts))
	assert.Equal(t, "/extra-lib", volumeMounts[getPos(volumeMounts, "extra-lib")].MountPath)
	assert.Equal(t, "/etc/else", volumeMounts[getPos(volumeMounts, "bootstrap")].MountPath)

}

func TestGenerateCassandraService(t *testing.T) {
	assert := assert.New(t)

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	selector := k8s.LabelsForCassandra(cc)
	svc := generateCassandraService(cc, selector, nil)

	assert.Equal(map[string]string{
		"app":              "cassandracluster",
		"cassandracluster": "cassandra-demo",
		"cluster":          "k8s.pic"},
		svc.Labels)
	assert.Equal(map[string]string{"external-dns.alpha.kubernetes.io/hostname": "my.custom.domain.com."},
		svc.Annotations)
}

func TestGenerateCassandraStatefulSet(t *testing.T) {
	assert := assert.New(t)

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	labels, nodeSelector := k8s.GetDCRackLabelsAndNodeSelectorForStatefulSet(cc, 0, 0)
	sts := generateCassandraStatefulSet(cc, &cc.Status, "dc1", "dc1-rack1", labels, nodeSelector, nil)

	assert.Equal(map[string]string{
		"app":                                  "cassandracluster",
		"cassandracluster":                     "cassandra-demo",
		"cassandraclusters.db.orange.com.dc":   "dc1",
		"cassandraclusters.db.orange.com.rack": "rack1",
		"dc-rack":                              "dc1-rack1",
		"cluster":                              "k8s.pic"},
		sts.Labels)
	assert.Equal("my.custom.annotation", sts.Spec.Template.Annotations["exemple.com/test"])
}
