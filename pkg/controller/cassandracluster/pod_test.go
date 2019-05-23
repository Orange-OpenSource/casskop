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

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/runtime/schema"
	kubetesting "k8s.io/client-go/testing"
)

var (
	podsGroup = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
)

func newPodUpdateAction(ns string, pod *v1.Pod) kubetesting.UpdateActionImpl {
	return kubetesting.NewUpdateAction(podsGroup, ns, pod)
}

func newPodGetAction(ns, name string) kubetesting.GetActionImpl {
	return kubetesting.NewGetAction(podsGroup, ns, name)
}

func newPodCreateAction(ns string, pod *v1.Pod) kubetesting.CreateActionImpl {
	return kubetesting.NewCreateAction(podsGroup, ns, pod)
}

func assertPodIsNotRunning(t *testing.T, err error) {
	if !reflect.DeepEqual(err, fmt.Errorf("Pod is not running")) {
		t.Errorf("Pod found is supposed to not be running")
	}
}
func TestGetLastOrFirstPod(t *testing.T) {
	mkPod := func(id int, running bool) *v1.Pod {
		podStatus := v1.PodFailed
		if running {
			podStatus = v1.PodRunning
		}
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("ns-nm-%02d", id)},
			Status: v1.PodStatus{Phase: podStatus}}
	}
	mkList := func(pods ...*v1.Pod) *v1.PodList {
		list := &v1.PodList{ListMeta: metav1.ListMeta{ResourceVersion: "1"}}
		for _, pod := range pods {
			list.Items = append(list.Items, *pod)
		}
		return list
	}
	podlist := mkList(mkPod(10, true), mkPod(1, true),
		mkPod(3, true), mkPod(20, true), mkPod(12, true))

	pod, err := GetLastOrFirstPod(podlist, first)

	if err != nil || pod.ObjectMeta.Name != "ns-nm-01" {
		t.Errorf("Not first pod returned")
	}

	pod, err = GetLastOrFirstPod(podlist, last)

	if err != nil || pod.ObjectMeta.Name != "ns-nm-20" {
		t.Errorf("Not last pod returned")
	}

	podlist = mkList(mkPod(10, true), mkPod(1, false))

	pod, err = GetLastOrFirstPod(podlist, first)
	assertPodIsNotRunning(t, err)

	podlist.Items[0].Status = v1.PodStatus{Phase: v1.PodRunning}
	ts := metav1.Now()
	podlist.Items[0].DeletionTimestamp = &ts

	pod, err = GetLastOrFirstPod(podlist, first)
	assertPodIsNotRunning(t, err)
}

/*
** Here the Kubernetes Client Mock does not help because we uses the SDK
** i'll come back here when https://github.com/operator-framework/operator-sdk/issues/284
** will be implemented
 */
/*
func TestPodServiceGetCreateOrUpdate(t *testing.T) {
	testPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "testpod1",
			ResourceVersion: "10",
		},
	}

	testns := "testns"

	tests := []struct {
		name            string
		pod             *v1.Pod
		getPodResult    *v1.Pod
		errorOnGet      error
		errorOnCreation error
		expActions      []kubetesting.Action
		expErr          bool
	}{
		{
			name:            "A new pod should create a new pod.",
			pod:             testPod,
			getPodResult:    nil,
			errorOnGet:      kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation: nil,
			expActions: []kubetesting.Action{
				newPodGetAction(testns, testPod.ObjectMeta.Name),
				newPodCreateAction(testns, testPod),
			},
			expErr: false,
		},
		{
			name:            "A new pod should error when create a new pod fails.",
			pod:             testPod,
			getPodResult:    nil,
			errorOnGet:      kubeerrors.NewNotFound(schema.GroupResource{}, ""),
			errorOnCreation: errors.New("wanted error"),
			expActions: []kubetesting.Action{
				newPodGetAction(testns, testPod.ObjectMeta.Name),
				newPodCreateAction(testns, testPod),
			},
			expErr: true,
		},
		{
			name:            "An existent pod should update the pod.",
			pod:             testPod,
			getPodResult:    testPod,
			errorOnGet:      nil,
			errorOnCreation: nil,
			expActions: []kubetesting.Action{
				newPodGetAction(testns, testPod.ObjectMeta.Name),
				newPodUpdateAction(testns, testPod),
			},
			expErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock.
			mcli := &kubernetes.Clientset{}

			mcli.AddReactor("get", "pods", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, test.getPodResult, test.errorOnGet
			})
			mcli.AddReactor("create", "pods", func(action kubetesting.Action) (bool, runtime.Object, error) {
				return true, nil, test.errorOnCreation
			})

			err := CreateOrUpdatePod(testns, "testpod1", test.pod)

			if test.expErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				// Check calls to kubernetes.
				assert.Equal(test.expActions, mcli.Actions())
			}
		})
	}
}
*/
