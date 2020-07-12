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

	"github.com/Orange-OpenSource/casskop/pkg/k8s"

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

// PodContainersReady returns true if all container in the Pod are ready
func PodContainersReady(pod *v1.Pod) bool {
	if pod.Status.ContainerStatuses != nil && len(pod.Status.ContainerStatuses) > 0 {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Ready == false {
				return false
			}
		}
		return true
	}
	return false
}

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
	return GetLastOrFirstPodItem(podsList.Items, last)
}

// GetLastOrFirstPodReady returns the las or first pod that is ready
func GetLastOrFirstPodReady(podsList []v1.Pod, last bool) (*v1.Pod, error) {
	var readyPods []v1.Pod
	for _, pod := range podsList {
		if cassandraPodIsReady(&pod) {
				readyPods = append(readyPods, pod)
		}
	}
	return GetLastOrFirstPodItem(readyPods, last)
}

// GetLastOrFirstPod returns the last or first pod.
func GetLastOrFirstPodItem(podsList []v1.Pod, last bool) (*v1.Pod, error) {
	nb := len(podsList)

	if nb < 1 {
		return nil, fmt.Errorf("there is no pod")
	}

	idx := 0
	if last {
		idx = nb - 1
	}

	items := podsList[:]

	// Sort pod list using ending number in field ObjectMeta.Name
	sort.Slice(items, func(i, j int) bool {
		id1, _ := strconv.Atoi(reEndingNumber.FindString(items[i].ObjectMeta.Name))
		id2, _ := strconv.Atoi(reEndingNumber.FindString(items[j].ObjectMeta.Name))
		return id1 < id2
	})

	pod := podsList[idx]

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

// GetFirstPod returns the first pod satisfying the selector, being in the namespace and being ready
func (rcc *ReconcileCassandraCluster) GetFirstPodReady(namespace string, selector map[string]string) (*v1.Pod, error) {
	podsList, err := rcc.ListPods(namespace, selector)
	if err != nil {
		return nil, fmt.Errorf("failed to get cassandra's pods: %v", err)
	}
	return GetLastOrFirstPodReady(podsList.Items, first)
}

// GetLastPod returns the last pod satisfying the selector and being in the namespace
func (rcc *ReconcileCassandraCluster) GetLastPodReady(namespace string, selector map[string]string) (*v1.Pod, error) {
	podsList, err := rcc.ListPods(namespace, selector)
	if err != nil {
		return nil, fmt.Errorf("failed to get cassandra's pods: %v", err)
	}
	return GetLastOrFirstPodReady(podsList.Items, last)
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

//hasUnschedulablePod goal is to detect if Pods are unschedulable
// - for lake of resources cpu/memory
// - with bad docker image (imagepullbackoff)
// - or else to add
func (rcc *ReconcileCassandraCluster) hasUnschedulablePod(namespace string, dcName, rackName string) bool {
	podsList, err := rcc.ListPods(rcc.cc.Namespace, k8s.LabelsForCassandraDCRack(rcc.cc, dcName, rackName))
	if err != nil || len(podsList.Items) < 1 {
		return false
	}
	for _, pod := range podsList.Items {
		if pod.Status.Phase != v1.PodRunning && pod.Status.Conditions != nil {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Ready == false && cs.State.Waiting != nil && cs.State.Waiting.Reason == "ImagePullBackOff" {
					//TODO: delete Pod in this case so that it can be scheduled again if image in spec and image in
					// pod have changd
					return true
				}
			}
			for _, cond := range pod.Status.Conditions {
				if (cond.Reason == v1.PodReasonUnschedulable) ||
					//try catch non ready pods
					(cond.Type == v1.PodReady && cond.Status == v1.ConditionFalse && cond.Reason == "ContainersNotReady") {
					return true
				}
			}
		}
	}
	return false
}

func (rcc *ReconcileCassandraCluster) ListPods(namespace string, selector map[string]string) (*v1.PodList, error) {

	clientOpt := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(selector),
	}

	opt := []client.ListOption{
		clientOpt,
	}

	pl := &v1.PodList{}
	return pl, rcc.client.List(context.TODO(), pl, opt...)
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
