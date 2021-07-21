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
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	v1 "k8s.io/api/core/v1"

	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/sirupsen/logrus"
)

//hasNoPodDisruption return true if there is no Disruption in the Pods of the cassandra Cluster
func (rcc *ReconcileCassandraCluster) hasNoPodDisruption() bool {
	return rcc.storedPdb.Status.DisruptionsAllowed > 0
}

//weAreScalingDown return true if we are Scaling Down the provided dc-rack
func (rcc *ReconcileCassandraCluster) weAreScalingDown(dcRackStatus *api.CassandraRackStatus) bool {
	if dcRackStatus.CassandraLastAction.Name == api.ActionScaleDown.Name &&
		(dcRackStatus.CassandraLastAction.Status == api.StatusToDo ||
			dcRackStatus.CassandraLastAction.Status == api.StatusOngoing ||
			dcRackStatus.CassandraLastAction.Status == api.StatusContinue) {
		return true
	}
	return false
}

func cassandraPodIsReady(pod *v1.Pod) bool {
	cassandraContainerStatus := getCassandraContainerStatus(pod)

	if cassandraContainerStatus != nil && cassandraContainerStatus.Name == cassandraContainerName &&
		pod.Status.Phase == v1.PodRunning && cassandraContainerStatus.Ready {
		return true
	}
	return false
}

func getCassandraContainerStatus(pod *v1.Pod) *v1.ContainerStatus{

	for i := range pod.Status.ContainerStatuses {
		if pod.Status.ContainerStatuses[i].Name == cassandraContainerName {
			return &pod.Status.ContainerStatuses[i]
		}
	}
	return nil
}

func cassandraPodRestartCount(pod *v1.Pod) int32 {
	for idx := range pod.Status.ContainerStatuses {
		if pod.Status.ContainerStatuses[idx].Name == cassandraContainerName {
			return pod.Status.ContainerStatuses[idx].RestartCount
		}
	}
	return 0
}



// DeletePVC deletes persistentvolumes of nodes in a rack
func (rcc *ReconcileCassandraCluster) DeletePVCs(cc *api.CassandraCluster, dcName string, rackName string) {
	lpvc, err := rcc.ListPVC(cc.Namespace, k8s.LabelsForCassandraDCRack(cc, dcName, rackName))
	if err != nil {
		logrus.Errorf("failed to get cassandra's PVC: %v", err)
	}
	for _, pvc := range lpvc.Items {
		err := rcc.deletePVC(&pvc)

		if err != nil {
			logrus.Errorf("[%s]: Error Deleting PVC[%s], Please make manual Actions..", cc.Name, pvc.Name)
		} else {
			logrus.Infof("[%s]: Delete PVC[%s] OK", cc.Name, pvc.Name)
		}
	}
}
