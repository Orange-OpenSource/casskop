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

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodsSlice(t *testing.T) {
	assert := assert.New(t)

	rcc, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	status := &cc.Status

	operatorName := "new-name"
	oldOperatorName := "old-name"
	operationName := "cleanup"
	dcRackName := "dc1-rack1"
	podLastOperation := &status.CassandraRackStatus[dcRackName].PodLastOperation
	podLastOperation.Name = operationName

	// Conditions to return checkOnly set to true with an empty podsSlice
	podLastOperation.Status = api.StatusOngoing
	podLastOperation.OperatorName = oldOperatorName

	podsSlice, checkOnly := rcc.podsSlice(cc, status, *podLastOperation, dcRackName, operationName, operatorName)

	assert.Equal(len(podsSlice), 0)
	assert.Equal(checkOnly, true)

	// Missing condition sets checkOnly to false
	podLastOperation.Status = api.StatusDone

	podsSlice, checkOnly = rcc.podsSlice(cc, status, *podLastOperation, dcRackName, operationName, operatorName)

	assert.Equal(len(podsSlice), 0)
	assert.Equal(checkOnly, false)

	//Create a pod to have something to put in podsSlice
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cassandra-demo-dc1-rack1-0",
			Namespace: "ns",
			Labels:    map[string]string{"app": "cassandracluster"},
		},
	}
	pod.Status.Phase = v1.PodRunning
	rcc.CreatePod(pod)

	podLastOperation.Status = api.StatusOngoing
	podLastOperation.Pods = []string{pod.GetName()}
	// Set the operator name to a different value than the current operator name
	podLastOperation.OperatorName = oldOperatorName

	podsSlice, checkOnly = rcc.podsSlice(cc, status, *podLastOperation, dcRackName, operationName, operatorName)
	assert.Equal(podsSlice, []v1.Pod{*pod})
	assert.Equal(checkOnly, true)
}
