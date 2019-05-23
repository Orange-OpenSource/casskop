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
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/k8s"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func preventClusterDeletion(cc *api.CassandraCluster, value bool) {
	if value {
		cc.SetFinalizers([]string{"kubernetes.io/pvc-to-delete"})
		return
	}
	cc.SetFinalizers([]string{})
}

func updateDeletePvcStrategy(cc *api.CassandraCluster) {
	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "deletePVC": cc.Spec.DeletePVC,
		"finalizers": cc.Finalizers}).Debug("updateDeletePvcStrategy called")
	// Remove Finalizers if DeletePVC is not enabled
	if !cc.Spec.DeletePVC && len(cc.Finalizers) > 0 {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Info("Won't delete PVCs when nodes are removed")
		preventClusterDeletion(cc, false)
	}
	// Add Finalizer if DeletePVC is enabled
	if cc.Spec.DeletePVC && len(cc.Finalizers) == 0 {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Info("Will delete PVCs when nodes are removed")
		preventClusterDeletion(cc, true)
	}
}

// CheckDeletePVC checks if DeletePVC is updated and update DeletePVC strategy
func (rcc *ReconcileCassandraCluster) CheckDeletePVC(cc *api.CassandraCluster) error {
	var oldCRD api.CassandraCluster
	if cc.Annotations[api.AnnotationLastApplied] == "" {
		return nil
	}

	//We retrieved our last-applied-configuration stored in the CRD
	err := json.Unmarshal([]byte(cc.Annotations[api.AnnotationLastApplied]), &oldCRD)
	if err != nil {
		logrus.Errorf("[%s]: Can't get Old version of CRD", cc.Name)
		return nil
	}

	if cc.Spec.DeletePVC != oldCRD.Spec.DeletePVC {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Debug("DeletePVC has been updated")
		updateDeletePvcStrategy(cc)
		return rcc.client.Update(context.TODO(), cc)
	}

	return nil
}

// CheckNonAllowedChanged - checks if there are some changes on CRD that are not allowed on statefulset
// If a non Allowed Changed is Find we won't Update associated kubernetes objects, but we will put back the old value
// and Patch the CRD with correct values
func (rcc *ReconcileCassandraCluster) CheckNonAllowedChanged(cc *api.CassandraCluster,
	status *api.CassandraClusterStatus) bool {
	var oldCRD api.CassandraCluster
	if cc.Annotations[api.AnnotationLastApplied] == "" {
		return false
	}

	if lac, _ := cc.ComputeLastAppliedConfiguration(); string(lac) == cc.Annotations[api.AnnotationLastApplied] {
		//there are no changes to take care about
		return false
	}

	//We retrieved our last-applied-configuration stored in the CRD
	err := json.Unmarshal([]byte(cc.Annotations[api.AnnotationLastApplied]), &oldCRD)
	if err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Error("Can't get Old version of CRD")
		return false
	}

	//Global scaleDown to 0 is forbidden
	if cc.Spec.NodesPerRacks == 0 {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the changed on NodesPerRack=0 restore to OldValue[%d]", oldCRD.Spec.NodesPerRacks)
		cc.Spec.NodesPerRacks = oldCRD.Spec.NodesPerRacks
		needUpdate = true
	}
	//DataCapacity change is forbidden
	if cc.Spec.DataCapacity != oldCRD.Spec.DataCapacity {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the changed on DataCapacity from [%s] to NewValue[%s]", oldCRD.Spec.DataCapacity, cc.Spec.DataCapacity)
		cc.Spec.DataCapacity = oldCRD.Spec.DataCapacity
		needUpdate = true
	}
	//DataStorage
	if cc.Spec.DataStorageClass != oldCRD.Spec.DataStorageClass {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the changed on DataStorageClass from [%s] to NewValue[%s]", oldCRD.Spec.DataStorageClass, cc.Spec.DataStorageClass)
		cc.Spec.DataStorageClass = oldCRD.Spec.DataStorageClass
		needUpdate = true
	}

	if needUpdate {
		status.LastClusterAction = api.ActionCorrectCRDConfig
		return true
	}

	if needUpdate, updateStatus := CheckNonAllowedRemoveDC(rcc, cc, status, &oldCRD); needUpdate {
		if updateStatus != "" {
			status.LastClusterAction = updateStatus
		}
		return true
	}

	if needUpdate, updateStatus := rcc.CheckNonAllowedScaleDown(cc, status, &oldCRD); needUpdate {
		if updateStatus != "" {
			status.LastClusterAction = updateStatus
		}
		return true
	}

	//What if we ask to changes Pod ressources ?
	// It is authorized, but the operator needs to detect it to prevent multiple statefulsets updates in the same time
	// the operator must handle thoses update sequentially, so we flag each dcrackname with this information
	if !reflect.DeepEqual(cc.Spec.Resources, oldCRD.Spec.Resources) {
		logrus.Infof("[%s]: We ask to Change Pod Resources from %v to %v", cc.Name, oldCRD.Spec.Resources, cc.Spec.Resources)

		for dc := 0; dc < cc.GetDCSize(); dc++ {
			dcName := cc.GetDCName(dc)
			for rack := 0; rack < cc.GetRackSize(dc); rack++ {

				rackName := cc.GetRackName(dc, rack)
				dcRackName := cc.GetDCRackName(dcName, rackName)
				dcRackStatus := status.CassandraRackStatus[dcRackName]

				logrus.Infof("[%s][%s]: Update Rack Status UpdateResources=Ongoing", cc.Name, dcRackName)
				dcRackStatus.CassandraLastAction.Name = api.ActionUpdateResources
				dcRackStatus.CassandraLastAction.Status = api.StatusToDo
				now := metav1.Now()
				status.CassandraRackStatus[dcRackName].CassandraLastAction.StartTime = &now
				status.CassandraRackStatus[dcRackName].CassandraLastAction.EndTime = nil
			}
		}

	}

	return false
}

//Implements Checks to see if the Operator will accept or refused the CRD Changed
//Return true imply that we want to update the status
func CheckNonAllowedRemoveDC(rcc *ReconcileCassandraCluster, cc *api.CassandraCluster,
	status *api.CassandraClusterStatus, oldCRD *api.CassandraCluster) (bool, string) {
	//Check if we ask to remove DC/Rack and If we are allow to do it
	if cc.GetDCRackSize() < oldCRD.GetDCRackSize() {
		//We ask a decrease in the amount of DCRack
		dcsize := cc.GetDCSize()
		olddcsize := oldCRD.GetDCSize()
		if dcsize < olddcsize-1 {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the Topology changed. You can only remove 1 DC at a time, "+
				"not only a Rack: %v restored to %v", cc.Spec.Topology, oldCRD.Spec.Topology)
			cc.Spec.Topology = oldCRD.Spec.Topology
			return true, api.ActionCorrectCRDConfig
		}

		//If we ask to remove only a rack, then it is not authorized
		if dcsize == olddcsize {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the Topology changed. You can only remove an entire DC, "+
				"not only a Rack: %v restored to %v", cc.Spec.Topology, oldCRD.Spec.Topology)
			cc.Spec.Topology = oldCRD.Spec.Topology
			return true, api.ActionCorrectCRDConfig
		}

		//Here we have asked to remove a DC. Check that the nbNodesPerRack is 0 else refuse the modification
		//Which DC we need to remove ?
		dcName := cc.GetRemovedDCName(oldCRD)
		//We need to check how many nodes were in the old CRD (before the user delete it)
		found, nbNodes := oldCRD.GetDCNodesPerRacksFromName(dcName)

		//if dc not found it's ok to remove
		if found && nbNodes > 0 {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the Topology changed. "+
				"You must scale down the dc %s to 0 before deleting the dc", dcName)
			cc.Spec.Topology = oldCRD.Spec.Topology
			return true, api.ActionCorrectCRDConfig
		}

		if cc.Status.LastClusterAction == api.ActionScaleDown &&
			cc.Status.LastClusterActionStatus != api.StatusDone {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the Topology changed. "+
				"You must wait to the end of ScaleDown to 0 before deleting the dc %s", dcName)
			cc.Spec.Topology = oldCRD.Spec.Topology
			return true, api.ActionCorrectCRDConfig

		}

		//If We have come here, we are allowed to remove the DC
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("We asked to remove dc %s", dcName)
		//We apply this change to the Cluster status

		return rcc.deleteDCObjects(cc, status, oldCRD)

	}
	return false, ""
}

func (rcc *ReconcileCassandraCluster) deleteDCObjects(cc *api.CassandraCluster,
	status *api.CassandraClusterStatus, oldCRD *api.CassandraCluster) (bool, string) {

	dcRackNameToDeleteList := cc.FixCassandraRackList(status)

	if len(dcRackNameToDeleteList) > 0 {

		for _, dcRackNameToDelete := range dcRackNameToDeleteList {

			err := rcc.DeleteStatefulSet(cc.Namespace, cc.Name+"-"+dcRackNameToDelete)
			if err != nil && !apierrors.IsNotFound(err) {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackNameToDelete}).Warnf("Can't Delete Statefulset: %v", err)
			}
			names := []string{
				cc.Name + "-" + cc.GetDCFromDCRackName(dcRackNameToDelete),                   //name-dc
				cc.Name + "-" + dcRackNameToDelete,                                           //name-dc-rack
				cc.Name + "-" + cc.GetDCFromDCRackName(dcRackNameToDelete) + "-exporter-jmx", //name-dc-exporter-jmx
			}
			for i := range names {
				err = rcc.DeleteService(cc.Namespace, names[i])
				if err != nil && !apierrors.IsNotFound(err) {
					logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackNameToDelete}).Warnf("Can't Delete Service: %v", err)
				}
			}

		}
		return true, api.ActionDeleteDC
	}
	return false, ""
}

//CheckNonAllowedScaleDown goal is to discard the scaleDown to 0 is there is still replicated data towards the
// corresponding DC
func (rcc *ReconcileCassandraCluster) CheckNonAllowedScaleDown(cc *api.CassandraCluster,
	status *api.CassandraClusterStatus,
	oldCRD *api.CassandraCluster) (bool, string) {

	if ok, dcName, dc := cc.FindDCWithNodesTo0(); ok {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Infof("Ask ScaleDown to 0 for dc %s", dcName)

		//We take the first Rack
		rackName := cc.GetRackName(dc, 0)

		selector := k8s.MergeLabels(k8s.LabelsForCassandraDCRack(cc, dcName, rackName))
		podsList, err := rcc.ListPods(cc.Namespace, selector)
		if err != nil || len(podsList.Items) < 1 {
			if err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the ScaleDown (no pod found). "+
					"topology %v restored to %v", cc.Spec.Topology, oldCRD.Spec.Topology)
				cc.Spec.Topology = oldCRD.Spec.Topology
				return true, api.ActionCorrectCRDConfig
			}
			//else there is already no pods so it's ok
			return false, ""
		}

		//We take the first available Pod
		for _, pod := range podsList.Items {
			if pod.Status.Phase != v1.PodRunning || pod.DeletionTimestamp != nil {
				continue
			}
			hostName := fmt.Sprintf("%s.%s", pod.Spec.Hostname, pod.Spec.Subdomain)
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Debugf("The Operator will ask node %s", hostName)
			jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc,
				cc.Spec.ImageJolokiaSecret, cc.Namespace)
			var keyspacesWithData []string
			if err == nil {
				keyspacesWithData, err = jolokiaClient.HasDataInDC(dcName)
			}
			if err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the ScaleDown (HasDataInDC failed %s). ", err)
				cc.Spec.Topology = oldCRD.Spec.Topology
				return true, api.ActionCorrectCRDConfig
			}
			if len(keyspacesWithData) != 0 {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("The Operator has refused the ScaleDown. Keyspaces still having data %v", keyspacesWithData)
				cc.Spec.Topology = oldCRD.Spec.Topology
				return true, api.ActionCorrectCRDConfig
			}
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("Cassandra has no more replicated data on dc %s, we can scale Down to 0", dcName)
			return false, ""
		}
	}
	return false, ""
}

//ReconcileRack will try to reconcile cassandra for each of the couple DC/Rack defined in the topology
func (rcc *ReconcileCassandraCluster) ReconcileRack(cc *api.CassandraCluster,
	status *api.CassandraClusterStatus) (err error) {

	for dc := 0; dc < cc.GetDCSize(); dc++ {
		dcName := cc.GetDCName(dc)
		for rack := 0; rack < cc.GetRackSize(dc); rack++ {

			rackName := cc.GetRackName(dc, rack)
			dcRackName := cc.GetDCRackName(dcName, rackName)
			if dcRackName == "" {
				return fmt.Errorf("Name uses for DC and/or Rack are not good")

			}

			//If we have added a dc/rack in the CRD, we add it to the Status
			if _, ok := status.CassandraRackStatus[dcRackName]; !ok {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Infof("the DC(%s) and Rack(%s) does not exist, initialize it in status", dcName, rackName)
				cc.InitCassandraRackinStatus(status, dcName, rackName)
				//Return will stop operator reconcile loop until next one
				//used here to write CassandraClusterStatus properly
				return nil
			}
			dcRackStatus := status.CassandraRackStatus[dcRackName]

			if cc.DeletionTimestamp != nil && cc.Spec.DeletePVC {
				rcc.DeletePVCs(cc, dcName, rackName)
				//Go to next rack
				continue
			}
			Name := cc.Name + "-" + dcRackName
			storedStatefulSet, err := rcc.GetStatefulSet(cc.Namespace, Name)
			if err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName}).Infof("failed to get cassandra's statefulset (%s) %v", Name, err)
			} else {

				//Update CassandraClusterPhase
				rcc.UpdateCassandraRackStatusPhase(cc, dcName, rackName, storedStatefulSet, status)

				//Find if there is an Action to execute or to end
				rcc.getNextCassandraClusterStatus(cc, dc, rack, dcName, rackName, storedStatefulSet, status)

				//If Not in +Initial State
				// Find if we have some Pod Operation to Execute, and execute thees
				if dcRackStatus.Phase != api.ClusterPhaseInitial {
					breakResyncloop, err := rcc.executePodOperation(cc, dcName, rackName, status)
					if err != nil {
						logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName,
							"err": err}).Error("Executing pod operation failed")
					}
					//For some Operations, we must NOT update the statefulset until Done.
					//So we block until OK
					if breakResyncloop {
						// If an Action is ongoing on the current Rack,
						// we don't want to check or start actions on Next Rack
						if dcRackStatus.Phase != api.ClusterPhaseRunning ||
							dcRackStatus.CassandraLastAction.Status == api.StatusToDo ||
							dcRackStatus.CassandraLastAction.Status == api.StatusOngoing ||
							dcRackStatus.CassandraLastAction.Status == api.StatusContinue {
							logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName,
								"err": err}).Debug("Waiting Rack to be running before continuing, " +
								"we break ReconcileRack Without Updating Statefulset")
							return nil
						}
						logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName,
							"LastActionName":   dcRackStatus.CassandraLastAction.Name,
							"LastActionStatus": dcRackStatus.CassandraLastAction.Status}).Warning("Should Not see this message ;)" +
							" Waiting Rack to be running before continuing, we loop on Next Rack, maybe we don't want that")
						continue

					}
				}
			}

			if err = rcc.ensureCassandraService(cc, dcName, rackName); err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName}).Errorf("ensureCassandraService Error: %v", err)
			}

			if err = rcc.ensureCassandraDCService(cc, dcName); err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName}).Errorf("ensureCassandraDCService Error: %v", err)
			}

			if err = rcc.ensureCassandraServiceMonitoring(cc, dcName); err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName}).Errorf("ensureCassandraServiceMonitoring Error: %v", err)
			}

			if err = rcc.ensureCassandraStatefulSet(cc, status, dcRackName, dc, rack); err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName}).Errorf("ensureCassandraStatefulSet Error: %v", err)
			}

			//If the Phase is not running Then we won't check on Next Racks so we return
			//We don't want to make change in 2 racks in a same time
			if dcRackStatus.Phase != api.ClusterPhaseRunning ||
				(dcRackStatus.CassandraLastAction.Status == api.StatusOngoing ||
					dcRackStatus.CassandraLastAction.Status == api.StatusFinalizing) {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName}).Infof("Waiting Rack to be running before continuing, " +
					"we break ReconcileRack after updated statefulset")
				return nil
			}
		}

	}

	//If cluster is deleted and DeletePVC is set, we can now stop preventing the cluster from being deleted
	//cause PVCs have been deleted
	if cc.DeletionTimestamp != nil && cc.Spec.DeletePVC {
		preventClusterDeletion(cc, false)
		return rcc.client.Update(context.TODO(), cc)
	}

	return nil
}

// UpdateCassandraClusterStatusPhase goal is to calculate the Cluster Phase according to StatefulSet Status.
func UpdateCassandraClusterStatusPhase(cc *api.CassandraCluster, status *api.CassandraClusterStatus) {
	var setLastClusterActionStatus bool
	for dc := 0; dc < cc.GetDCSize(); dc++ {
		dcName := cc.GetDCName(dc)
		for rack := 0; rack < cc.GetRackSize(dc); rack++ {

			rackName := cc.GetRackName(dc, rack)
			dcRackName := cc.GetDCRackName(dcName, rackName)
			dcRackStatus := status.CassandraRackStatus[dcRackName]

			// If there is a lastAction ongoing in a Rack we update cluster lastaction accordingly
			if dcRackStatus.CassandraLastAction.Status != api.StatusDone {
				status.LastClusterActionStatus = dcRackStatus.CassandraLastAction.Status
				status.LastClusterAction = dcRackStatus.CassandraLastAction.Name
				setLastClusterActionStatus = true
			}

			if dcRackStatus.Phase != api.ClusterPhaseRunning {
				status.Phase = dcRackStatus.Phase

				if _, ok := cc.Status.CassandraRackStatus[dcRackName]; !ok ||
					cc.Status.CassandraRackStatus[dcRackName].Phase != dcRackStatus.Phase {
					logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName}).Infof("Update Rack Status: %s", dcRackStatus.Phase)
				}
				return
			}

		}

	}
	//If there is no more action in racks, we update cluster
	if !setLastClusterActionStatus && status.LastClusterActionStatus != api.StatusDone {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Infof("Action %s is done!", status.LastClusterAction)
		status.LastClusterActionStatus = api.StatusDone
		status.Phase = api.ClusterPhaseRunning
	}
	return
}

//FlipCassandraClusterUpdateSeedListStatus checks if all racks has the status UpdateSeedList=To-do
//Then it update to UpdateSeedList=Ongoing to start the operation
func FlipCassandraClusterUpdateSeedListStatus(cc *api.CassandraCluster, status *api.CassandraClusterStatus) {

	//if global status is not yet  "Configuring", we skip this one
	if cc.Spec.AutoUpdateSeedList &&
		status.LastClusterAction == api.ActionUpdateSeedList &&
		status.LastClusterActionStatus == api.StatusConfiguring {
		var setOperationOngoing = true

		//Check if We need to start operation
		//all status of all racks must be "configuring"
		for dc := 0; dc < cc.GetDCSize(); dc++ {
			dcName := cc.GetDCName(dc)
			for rack := 0; rack < cc.GetRackSize(dc); rack++ {

				rackName := cc.GetRackName(dc, rack)
				dcRackName := cc.GetDCRackName(dcName, rackName)
				dcRackStatus := status.CassandraRackStatus[dcRackName]

				//If not all racks are in "configuring", then we don't flip status to to-do except for initializing rack
				if !(dcRackStatus.CassandraLastAction.Name == api.ActionUpdateSeedList &&
					dcRackStatus.CassandraLastAction.Status == api.StatusConfiguring) {
					//if rack is initializing we allow it to Flip
					if dcRackStatus.CassandraLastAction.Name != api.ClusterPhaseInitial {
						setOperationOngoing = false
					}

					break
				}
			}
		}

		//If all racks are in "configuring" state, we update all in to-do state so the operator can start actions
		if setOperationOngoing {
			for dc := 0; dc < cc.GetDCSize(); dc++ {
				dcName := cc.GetDCName(dc)
				for rack := 0; rack < cc.GetRackSize(dc); rack++ {

					rackName := cc.GetRackName(dc, rack)
					dcRackName := cc.GetDCRackName(dcName, rackName)
					dcRackStatus := status.CassandraRackStatus[dcRackName]

					logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName}).Infof("Update Rack Status UpdateSeedList=ToDo")
					dcRackStatus.CassandraLastAction.Name = api.ActionUpdateSeedList
					dcRackStatus.CassandraLastAction.Status = api.StatusToDo
				}
			}
		}
	}
	return
}
