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
	"regexp"
	"sort"
	"strconv"

	"github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/k8s"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	last  = true
	first = false
)

var reEndingNumber = regexp.MustCompile("[0-9]+$")

func (rcc *ReconcileCassandraCluster) GetPod(namespace, name string) (*v1.Pod, error) {

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return pod, rcc.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, pod)
}

// GetLastOrFirstPod returns the last or first pod satisfying the selector and being in the namespace
func GetLastOrFirstPod(podsList *v1.PodList, last bool) (*v1.Pod, error) {
	nb := len(podsList.Items)

	if nb < 1 {
		return nil, fmt.Errorf("there is no pod")
	}

	idx := 0
	if last {
		idx = nb - 1
	}

	items := podsList.Items[:]

	// Sort pod list using ending number in field ObjectMeta.Name
	sort.Slice(items, func(i, j int) bool {
		id1, _ := strconv.Atoi(reEndingNumber.FindString(items[i].ObjectMeta.Name))
		id2, _ := strconv.Atoi(reEndingNumber.FindString(items[j].ObjectMeta.Name))
		return id1 < id2
	})

	pod := podsList.Items[idx]

	if pod.Status.Phase != v1.PodRunning || pod.DeletionTimestamp != nil {
		return nil, fmt.Errorf("Pod is not running")
	}
	return &pod, nil
}

// GetFirstPod returns the first pod satisfying the selector and being in the namespace
func (rcc *ReconcileCassandraCluster) GetFirstPod(namespace string, selector map[string]string) (*v1.Pod, error) {
	podsList, err := rcc.ListPods(namespace, selector)
	if err != nil {
		return nil, fmt.Errorf("failed to get cassandra's pods: %v", err)
	}
	return GetLastOrFirstPod(podsList, first)
}

// GetLastPod returns the last pod satisfying the selector and being in the namespace
func (rcc *ReconcileCassandraCluster) GetLastPod(namespace string, selector map[string]string) (*v1.Pod, error) {
	podsList, err := rcc.ListPods(namespace, selector)
	if err != nil {
		return nil, fmt.Errorf("failed to get cassandra's pods: %v", err)
	}
	return GetLastOrFirstPod(podsList, last)
}

func (rcc *ReconcileCassandraCluster) UpdatePodLabel(pod *v1.Pod, label map[string]string) error {
	podToUpdate, err := rcc.GetPod(pod.Namespace, pod.Name)
	if err != nil {
		return err
	}
	labels := k8s.MergeLabels(podToUpdate.GetLabels(), label)
	podToUpdate.SetLabels(labels)
	return rcc.UpdatePod(podToUpdate)
}

func (rcc *ReconcileCassandraCluster) ListPods(namespace string, selector map[string]string) (*v1.PodList, error) {

	opt := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(selector),
		// HACK: due to a fake client bug, ListOptions.Raw.TypeMeta must be
		// explicitly populated for testing.
		//
		// See https://github.com/kubernetes-sigs/controller-runtime/issues/168
		// https://github.com/operator-framework/operator-sdk/blob/00ef545399db0e4b8fe4474188a342369874162a/test/test-framework/pkg/controller/memcached/memcached_controller.go#L157
		Raw: &metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				//Kind:       "CassandraCluster",
				//APIVersion: api.SchemeGroupVersion.String(),
				Kind:       "Pod",
				APIVersion: "v1",
			},
		},
	}
	pl := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
	return pl, rcc.client.List(context.TODO(), opt, pl)
}

func (rcc *ReconcileCassandraCluster) CreatePod(pod *v1.Pod) error {
	err := rcc.client.Create(context.TODO(), pod)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("pod already exists: %cc", err)
		}
		return fmt.Errorf("failed to create cassandra pod: %cc", err)
		//return err
	}
	return nil
}

func (rcc *ReconcileCassandraCluster) UpdatePod(pod *v1.Pod) error {
	err := rcc.client.Update(context.TODO(), pod)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("pod already exists: %cc", err)
		}
		return fmt.Errorf("failed to update cassandra pod: %cc", err)
		//return err
	}
	return nil
}

func (rcc *ReconcileCassandraCluster) CreateOrUpdatePod(namespace string, name string, pod *v1.Pod) error {
	storedPod, err := rcc.GetPod(namespace, pod.Name)
	if err != nil {
		// If no resource we need to create.
		if apierrors.IsNotFound(err) {
			return rcc.CreatePod(pod)
		}
		return err
	}

	// Already exists, need to Update.
	pod.ResourceVersion = storedPod.ResourceVersion
	return rcc.UpdatePod(pod)
}

//DeletePod delete a pod
func (rcc *ReconcileCassandraCluster) DeletePod(pod *v1.Pod) error {
	err := rcc.client.Delete(context.TODO(), pod)
	if err != nil {
		return fmt.Errorf("failed to delete Pod: %cc", err)
	}
	return nil
}

//ForceDeletePod delete a pod with a grace period of 0 seconds
func (rcc *ReconcileCassandraCluster) ForceDeletePod(pod *v1.Pod) error {
	err := rcc.client.Delete(context.TODO(), pod, client.GracePeriodSeconds(0))
	if err != nil {
		return fmt.Errorf("failed to delete Pod: %cc", err)
	}
	return nil
}
