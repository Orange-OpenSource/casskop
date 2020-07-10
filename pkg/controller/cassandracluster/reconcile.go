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
	"regexp"
	"strings"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/r3labs/diff"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type gaugeVec struct {
	*prometheus.GaugeVec
}

var (
	ClusterPhaseMetric = gaugeVec{prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_phase",
			Help: "Current phase of a cluster",
		},
		[]string{"cluster"},
	)}

	ClusterActionMetric = gaugeVec{prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_action",
			Help: "Actions done on a cluster",
		},
		[]string{"cluster"},
	)}
)

const topologyChangeRefused = "The Operator has refused the Topology change. "

func (metric gaugeVec) set(phase api.ClusterStateInfo, clusterName string) {
	metric.With(
		prometheus.Labels{"cluster": clusterName},
	).Set(phase.Id)
}

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

// CheckNonAllowedChanges - checks if there are some changes on CRD that are not allowed on statefulset
// If a non Allowed Changed is Find we won't Update associated kubernetes objects, but we will put back the old value
// and Patch the CRD with correct values
func (rcc *ReconcileCassandraCluster) CheckNonAllowedChanges(cc *api.CassandraCluster,
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
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).
			Warningf("The Operator has refused the change on NodesPerRack=0 restore to OldValue[%d]",
				oldCRD.Spec.NodesPerRacks)
		cc.Spec.NodesPerRacks = oldCRD.Spec.NodesPerRacks
		needUpdate = true
	}

	for dc := 0; dc < cc.GetDCSize(); dc++ {
		dcName := cc.GetDCName(dc)
		//DataCapacity change is forbidden
		if cc.GetDataCapacityForDC(dcName) != oldCRD.GetDataCapacityForDC(dcName) {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dcName": dcName}).
				Warningf("The Operator has refused the change on DataCapacity from [%s] to NewValue[%s]",
					oldCRD.GetDataCapacityForDC(dcName), cc.GetDataCapacityForDC(dcName))
			cc.Spec.DataCapacity = oldCRD.Spec.DataCapacity
			cc.Spec.Topology.DC[dc].DataCapacity = oldCRD.Spec.Topology.DC[dc].DataCapacity
			needUpdate = true
		}
		//DataStorage
		if cc.GetDataStorageClassForDC(dcName) != oldCRD.GetDataStorageClassForDC(dcName) {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dcName": dcName}).
				Warningf("The Operator has refused the change on DataStorageClass from [%s] to NewValue[%s]",
					oldCRD.GetDataStorageClassForDC(dcName), cc.GetDataStorageClassForDC(dcName))
			cc.Spec.DataStorageClass = oldCRD.Spec.DataStorageClass
			cc.Spec.Topology.DC[dc].DataStorageClass = oldCRD.Spec.Topology.DC[dc].DataStorageClass
			needUpdate = true
		}
	}

	if needUpdate {
		status.LastClusterAction = api.ActionCorrectCRDConfig.Name
		ClusterActionMetric.set(api.ActionCorrectCRDConfig, cc.Name)
		return true
	}

	var updateStatus string
	if needUpdate, updateStatus = CheckTopologyChanges(rcc, cc, status, &oldCRD); needUpdate {
		if updateStatus != "" {
			status.LastClusterAction = updateStatus
		}
		if updateStatus == api.ActionCorrectCRDConfig.Name {
			cc.Spec.Topology = (&oldCRD).Spec.Topology
			ClusterActionMetric.set(api.ActionCorrectCRDConfig, cc.Name)
		}

		return true
	}

	if updateStatus == api.ActionDeleteRack.Name {
		ClusterActionMetric.set(api.ActionDeleteRack, cc.Name)
		return true
	}

	if needUpdate = rcc.CheckNonAllowedScaleDown(cc, status, &oldCRD); needUpdate {
		status.LastClusterAction = api.ActionCorrectCRDConfig.Name
		ClusterActionMetric.set(api.ActionCorrectCRDConfig, cc.Name)
		return true
	}

	//What if we ask to changes Pod ressources ?
	// It is authorized, but the operator needs to detect it to prevent multiple statefulsets updates in the same time
	// the operator must handle thoses updates sequentially, so we flag each dcrackname with this information
	if !reflect.DeepEqual(cc.Spec.Resources, oldCRD.Spec.Resources) {
		logrus.Infof("[%s]: We ask to Change Pod Resources from %v to %v", cc.Name, oldCRD.Spec.Resources, cc.Spec.Resources)

		for dc := 0; dc < cc.GetDCSize(); dc++ {
			dcName := cc.GetDCName(dc)
			for rack := 0; rack < cc.GetRackSize(dc); rack++ {

				rackName := cc.GetRackName(dc, rack)
				dcRackName := cc.GetDCRackName(dcName, rackName)
				dcRackStatus := status.CassandraRackStatus[dcRackName]

				logrus.Infof("[%s][%s]: Update Rack Status UpdateResources=Ongoing", cc.Name, dcRackName)
				dcRackStatus.CassandraLastAction.Name = api.ActionUpdateResources.Name
				ClusterActionMetric.set(api.ActionUpdateResources, cc.Name)
				dcRackStatus.CassandraLastAction.Status = api.StatusToDo
				now := metav1.Now()
				status.CassandraRackStatus[dcRackName].CassandraLastAction.StartTime = &now
				status.CassandraRackStatus[dcRackName].CassandraLastAction.EndTime = nil
			}
		}
	}

	return false
}

func generatePaths(s string) []string {
	return strings.Split(s, ".")
}

// lookForFilter checks if filters are found in path and add the information to filtersFound if that's the case
func lookForFilter(path []string, filters [][]string, filtersFound *map[string]bool) {
	for _, filter := range filters {
		if 2*len(filter)+1 == len(path) {
			currentPath := path[0]
			for i := 2; i < len(path)-1; i += 2 {
				currentPath += "." + path[i]
			}
			if currentPath == strings.Join(filter, ".") {
				if _, ok := (*filtersFound)[currentPath]; !ok {
					(*filtersFound)[currentPath] = true
				}
			}
		}
	}
}

// hasChange returns if there is a change with the type provided and matching all paths
// paths can be prepended with a - to specify  that it should not be found
// for instance ('DC', '-DC.Rack') means a DC change without a DC.Rack change
// changes of property NodesPerRacks are skipped
func hasChange(changelog diff.Changelog, changeType string, paths ...string) bool {
	regexPath := regexp.MustCompile("^\\-([^\\+]*)$")
	if len(changelog) == 0 {
		return false
	}
	noPaths := len(paths) == 0
	includeFilters := [][]string{}
	excludeFilters := [][]string{}
	for _, path := range paths {
		if match := regexPath.FindStringSubmatch(path); len(match) > 0 {
			excludeFilters = append(excludeFilters, generatePaths(match[1]))
			continue
		}
		includeFilters = append(includeFilters, generatePaths(path))
	}
	idx := "-1"
	var includedFiltersFound, excludedFiltersFound map[string]bool
	for _, cl := range changelog {
		// Only scan changes on Name/NumTokens
		if cl.Type == changeType &&
			// DC Changes
			(cl.Path[2] == "Name" || cl.Path[2] == "NumTokens" ||
				// Rack changes
				(len(cl.Path) > 4 && cl.Path[4] == "Name")) {
			if noPaths {
				return true
			}

			// We reset counters when it's a new index
			if cl.Path[1] != idx {
				idx = cl.Path[1]
				includedFiltersFound = map[string]bool{}
				excludedFiltersFound = map[string]bool{}
			}

			// We look for all matching filters
			lookForFilter(cl.Path, includeFilters, &includedFiltersFound)

			// We look for all excluding filters
			lookForFilter(cl.Path, excludeFilters, &excludedFiltersFound)

			if len(includedFiltersFound) == len(includeFilters) && len(excludedFiltersFound) == 0 {
				return true
			}
		}
	}
	return false
}

//CheckTopologyChanges checks to see if the Operator accepts or refuses the CRD changes
func CheckTopologyChanges(rcc *ReconcileCassandraCluster, cc *api.CassandraCluster,
	status *api.CassandraClusterStatus, oldCRD *api.CassandraCluster) (bool, string) {

	changelog, _ := diff.Diff(oldCRD.Spec.Topology, cc.Spec.Topology)

	if hasChange(changelog, diff.UPDATE) ||
		hasChange(changelog, diff.DELETE, "DC.Rack", "-DC") ||
		hasChange(changelog, diff.CREATE, "DC.Rack", "-DC") {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf(
			topologyChangeRefused+"No change other than adding/removing a DC can happen: %v restored to %v",
			cc.Spec.Topology, oldCRD.Spec.Topology)
		return true, api.ActionCorrectCRDConfig.Name
	}

	if cc.GetDCSize() < oldCRD.GetDCSize()-1 {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf(
			topologyChangeRefused+"You can only remove 1 DC at a time, "+
				"not only a Rack: %v restored to %v", cc.Spec.Topology, oldCRD.Spec.Topology)
		return true, api.ActionCorrectCRDConfig.Name
	}

	if cc.GetDCRackSize() < oldCRD.GetDCRackSize() {

		if cc.Status.LastClusterAction == api.ActionScaleDown.Name &&
			cc.Status.LastClusterActionStatus != api.StatusDone {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).
				Warningf(topologyChangeRefused +
					"You must wait to the end of ScaleDown to 0 before deleting a DC")
			return true, api.ActionCorrectCRDConfig.Name
		}

		dcName := cc.GetRemovedDCName(oldCRD)

		//We need to check how many nodes were in the old CRD (before the user delete it)
		if found, nbNodes := oldCRD.GetDCNodesPerRacksFromName(dcName); found && nbNodes > 0 {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).
				Warningf(topologyChangeRefused+
					"You must scale down the DC %s to 0 before deleting it", dcName)
			return true, api.ActionCorrectCRDConfig.Name
		}

		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf("Removing DC %s", dcName)

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
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackNameToDelete}).Warnf(
					"Can't Delete Statefulset: %v", err)
			}
			names := []string{
				cc.Name + "-" + cc.GetDCFromDCRackName(dcRackNameToDelete),                   //name-dc
				cc.Name + "-" + dcRackNameToDelete,                                           //name-dc-rack
				cc.Name + "-" + cc.GetDCFromDCRackName(dcRackNameToDelete) + "-exporter-jmx", //name-dc-exporter-jmx
			}
			for i := range names {
				err = rcc.DeleteService(cc.Namespace, names[i])
				if err != nil && !apierrors.IsNotFound(err) {
					logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackNameToDelete}).Warnf(
						"Can't Delete Service: %v", err)
				}
			}

		}
		ClusterActionMetric.set(api.ActionDeleteDC, cc.Name)
		return true, api.ActionDeleteDC.Name
	}
	return false, ""
}

//CheckNonAllowedScaleDown goal is to discard the scaleDown to 0 is there is still replicated data towards the
// corresponding DC
func (rcc *ReconcileCassandraCluster) CheckNonAllowedScaleDown(cc *api.CassandraCluster,
	status *api.CassandraClusterStatus,
	oldCRD *api.CassandraCluster) bool {

	if ok, dcName, dc := cc.FindDCWithNodesTo0(); ok {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Infof("Ask ScaleDown to 0 for dc %s", dcName)

		//We take the first Rack
		rackName := cc.GetRackName(dc, 0)

		selector := k8s.MergeLabels(k8s.LabelsForCassandraDCRack(cc, dcName, rackName))
		podsList, err := rcc.ListPods(cc.Namespace, selector)
		if err != nil || len(podsList.Items) < 1 {
			if err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf(
					"The Operator has refused the ScaleDown (no pod found). "+
						"topology %v restored to %v", cc.Spec.Topology, oldCRD.Spec.Topology)
				cc.Spec.Topology = oldCRD.Spec.Topology
				return true
			}
			//else there is already no pods so it's ok
			return false
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
				keyspacesWithData, err = jolokiaClient.NonLocalKeyspacesInDC(dcName)
			}
			if err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf(
					"The Operator has refused the ScaleDown (NonLocalKeyspacesInDC failed %s). ", err)
				cc.Spec.Topology = oldCRD.Spec.Topology
				return true
			}
			if len(keyspacesWithData) != 0 {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf(
					"The Operator has refused the ScaleDown. Keyspaces still having data %v", keyspacesWithData)
				cc.Spec.Topology = oldCRD.Spec.Topology
				return true
			}
			logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Warningf(
				"Cassandra has no more replicated data on dc %s, we can scale Down to 0", dcName)
			return false
		}
	}
	return false
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
				return fmt.Errorf("Name used for DC and/or Rack are not good")
			}

			//If we have added a dc/rack to the CRD, we add it to the Status
			if _, ok := status.CassandraRackStatus[dcRackName]; !ok {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Infof("DC-Rack(%s-%s) does not exist, "+
					"initialize it in status", dcName, rackName)
				ClusterPhaseMetric.set(api.ClusterPhaseInitial, cc.Name)
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
				logrus.WithFields(logrus.Fields{"cluster": cc.Name,
					"dc-rack": dcRackName}).Infof("failed to get cassandra's statefulset (%s) %v", Name, err)
			} else {

				//Update CassandraClusterPhase
				rcc.UpdateCassandraRackStatusPhase(cc, dcName, rackName, storedStatefulSet, status)

				//Find if there is an Action to execute or to end
				rcc.getNextCassandraClusterStatus(cc, dc, rack, dcName, rackName, storedStatefulSet, status)

				//If not Initializing cluster execute pod operations queued
				if dcRackStatus.Phase != api.ClusterPhaseInitial.Name {
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
						if dcRackStatus.Phase != api.ClusterPhaseRunning.Name ||
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
							"LastActionStatus": dcRackStatus.CassandraLastAction.Status}).Warning(
							"Should Not see this message ;)" +
								" Waiting Rack to be running before continuing, we loop on Next Rack, maybe we don't want that")
						continue

					}
				}

			}
			if err = rcc.ensureCassandraService(cc); err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Errorf("ensureCassandraService Error: %v", err)
			}

			if err = rcc.ensureCassandraServiceMonitoring(cc, dcName); err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name,
					"dc-rack": dcRackName}).Errorf("ensureCassandraServiceMonitoring Error: %v", err)
			}

			breakLoop, err := rcc.ensureCassandraStatefulSet(cc, status, dcName, dcRackName, dc, rack)
			if err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name,
					"dc-rack": dcRackName}).Errorf("ensureCassandraStatefulSet Error: %v", err)
			}
			if cc.Spec.UnlockNextOperation {
				//If we enter specific change we remove unlockNextOperation from Spec
				cc.Spec.UnlockNextOperation = false
				needUpdate = true
			}
			if breakLoop {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name,
					"dc-rack": dcRackName}).Debug("We just update Statefulset " +
					"we break ReconcileRack")
				return nil
			}

			//If the Phase is not running then we won't check on Next Racks so we return
			//We don't want to make any changes in 2 racks at the same time
			if dcRackStatus.Phase != api.ClusterPhaseRunning.Name ||
				(dcRackStatus.CassandraLastAction.Status == api.StatusOngoing ||
					dcRackStatus.CassandraLastAction.Status == api.StatusFinalizing) {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name,
					"dc-rack": dcRackName}).Infof("Waiting Rack to be running before continuing, " +
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
			dcRackStatus, exist := status.CassandraRackStatus[dcRackName]
			if !exist {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Infof("the DC(%s) and Rack(%s) does not exist, "+
					"the rack status will be updated in next reconcile", dcName, rackName)
				continue
			}

			// If there is a lastAction ongoing in a Rack we update cluster lastaction accordingly
			if dcRackStatus.CassandraLastAction.Status != api.StatusDone {
				status.LastClusterActionStatus = dcRackStatus.CassandraLastAction.Status
				status.LastClusterAction = dcRackStatus.CassandraLastAction.Name
				setLastClusterActionStatus = true
			}

			//If a rack is not running we return
			if dcRackStatus.Phase != api.ClusterPhaseRunning.Name {
				status.Phase = dcRackStatus.Phase

				if _, ok := cc.Status.CassandraRackStatus[dcRackName]; !ok ||
					cc.Status.CassandraRackStatus[dcRackName].Phase != dcRackStatus.Phase {
					logrus.WithFields(logrus.Fields{"cluster": cc.Name,
						"dc-rack": dcRackName}).Infof("Update Rack Status: %s", dcRackStatus.Phase)
				}
				return
			}

		}

	}

	//If there is no more action in racks, we update cluster
	if !setLastClusterActionStatus &&
		status.LastClusterActionStatus != api.StatusDone {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Infof("Action %s is done!", status.LastClusterAction)
		status.LastClusterActionStatus = api.StatusDone
		status.Phase = api.ClusterPhaseRunning.Name
		ClusterPhaseMetric.set(api.ClusterPhaseRunning, cc.Name)
	}

	//If cluster phase id not running, we update it
	if status.Phase != api.ClusterPhaseRunning.Name && status.LastClusterActionStatus == api.StatusDone {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name}).Infof("Cluster is running")
		status.Phase = api.ClusterPhaseRunning.Name
		ClusterPhaseMetric.set(api.ClusterPhaseRunning, cc.Name)
	}

	return
}

//FlipCassandraClusterUpdateSeedListStatus checks if all racks has the status UpdateSeedList=To-do
//It then sets UpdateSeedList to Ongoing to start the operation
func FlipCassandraClusterUpdateSeedListStatus(cc *api.CassandraCluster, status *api.CassandraClusterStatus) {

	//if global status is not yet  "Configuring", we skip this one
	if status.LastClusterAction == api.ActionUpdateSeedList.Name &&
		status.LastClusterActionStatus == api.StatusConfiguring {
		var setOperationOngoing = true

		//All racks must be configuring the seed list or in initializing mode
		for dc := 0; dc < cc.GetDCSize(); dc++ {
			dcName := cc.GetDCName(dc)
			for rack := 0; rack < cc.GetRackSize(dc); rack++ {

				rackName := cc.GetRackName(dc, rack)
				dcRackName := cc.GetDCRackName(dcName, rackName)
				dcRackStatus := status.CassandraRackStatus[dcRackName]

				if !(dcRackStatus.CassandraLastAction.Name == api.ActionUpdateSeedList.Name &&
					dcRackStatus.CassandraLastAction.Status == api.StatusConfiguring) {
					if dcRackStatus.CassandraLastAction.Name != api.ClusterPhaseInitial.Name {
						setOperationOngoing = false
					}

					break
				}
			}
		}

		//If all racks are in "configuring" state, we set all status to ToDo to trigger the operator actions
		if setOperationOngoing {
			for dc := 0; dc < cc.GetDCSize(); dc++ {
				dcName := cc.GetDCName(dc)
				for rack := 0; rack < cc.GetRackSize(dc); rack++ {

					rackName := cc.GetRackName(dc, rack)
					dcRackName := cc.GetDCRackName(dcName, rackName)
					dcRackStatus := status.CassandraRackStatus[dcRackName]

					logrus.WithFields(logrus.Fields{"cluster": cc.Name,
						"dc-rack": dcRackName}).Infof("Update Rack Status UpdateSeedList=ToDo")
					dcRackStatus.CassandraLastAction.Name = api.ActionUpdateSeedList.Name
					ClusterActionMetric.set(api.ActionUpdateSeedList, cc.Name)
					dcRackStatus.CassandraLastAction.Status = api.StatusToDo
				}
			}
		}
	}
	return
}

func (rcc *ReconcileCassandraCluster) CheckPodsState(cc *api.CassandraCluster,
	status *api.CassandraClusterStatus) (err error) {

	podsList, err := rcc.ListCassandraClusterPods(cc)
	if err != nil {
		return err
	}

	if status.CassandraNodesStatus == nil {
		status.CassandraNodesStatus = make(map[string]api.CassandraNodeStatus)
	}

	if len(podsList) < 1 {
		return nil
	}

	logrus.WithFields(logrus.Fields{"cluster": cc.Name,
		"err": err}).Debug("Get first available pod")

	firstPod, err := GetLastOrFirstPodReady(podsList, true)
	if err != nil {
		return err
	}

	hostName := fmt.Sprintf("%s.%s", firstPod.Spec.Hostname, firstPod.Spec.Subdomain)

	logrus.WithFields(logrus.Fields{"cluster": cc.Name,
		"err": err}).Info(fmt.Sprintf("We will request : %s to catch hostIdMap", hostName))

	jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc, cc.Spec.ImageJolokiaSecret, cc.Namespace)
	if err != nil {
		return err
	}

	hostIDMap, err := jolokiaClient.hostIDMap()
	if err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name,
			"err": err}).Errorf("Failed to call %s to get hostIdMap", hostName)
		return err
	}

	podToDelete, err := processingPods(hostIDMap, cc.Spec.RestartCountBeforePodDeletion, podsList, status)
	if err != nil {
		return err
	}
	if podToDelete != nil {
		return rcc.client.Delete(context.TODO(), podToDelete)
	}

	return nil
}

func (rcc *ReconcileCassandraCluster) ListCassandraClusterPods(cc *api.CassandraCluster) ([]v1.Pod, error) {
	var podsList []v1.Pod

	// We loop on each DC and Rack of the CassandraCluster
	for dc := 0; dc < cc.GetDCSize(); dc++ {
		dcName := cc.GetDCName(dc)
		for rack := 0; rack < cc.GetRackSize(dc); rack++ {
			rackName := cc.GetRackName(dc, rack)
			dcRackName := cc.GetDCRackName(dcName, rackName)
			if dcRackName == "" {
				return nil, fmt.Errorf("Name used for DC and/or Rack are not good")
			}

			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc-rack": dcRackName}).Debug("List available pods")
			pods, err := rcc.ListPods(cc.Namespace, k8s.LabelsForCassandraDCRack(cc, dcName, rackName))
			if err != nil {
				return nil, err
			}

			podsList = append(podsList, pods.Items...)
		}
	}
	return podsList, nil
}

func processingPods(hostIDMap map[string]string, restartCountBeforePodDeletion int32, podsList []v1.Pod,
	status *api.CassandraClusterStatus) (*v1.Pod, error) {

	// For each pod
	for _, pod := range podsList {
		cassandraPodRestartCount := cassandraPodRestartCount(&pod)
		// If the cassandra container has performed more restart than allowed
		if restartCountBeforePodDeletion > 0 && cassandraPodRestartCount > restartCountBeforePodDeletion {
			logrus.WithFields(logrus.Fields{"pod": pod.Name}).Debug(fmt.Sprintf("Pod found in restart status with %d restart", restartCountBeforePodDeletion))
			return checkPodCrossIpUseCaseForPod(hostIDMap, &pod, status)
		}
		updateCassandraNodesStatusForPod(hostIDMap, &pod, status)
	}

	return nil, nil
}

func updateCassandraNodesStatusForPod(hostIDMap map[string]string, pod *v1.Pod, status *api.CassandraClusterStatus) {

	// Update Pod, HostId, Ip couple cached into status
	hostId, keyFound := hostIDMap[pod.Status.PodIP]
	if keyFound == true && cassandraPodIsReady(pod) {
		status.CassandraNodesStatus[pod.Name] = api.CassandraNodeStatus{HostId: hostId, NodeIp: pod.Status.PodIP}
	}
}

func checkPodCrossIpUseCaseForPod(hostIDMap map[string]string, pod *v1.Pod, status *api.CassandraClusterStatus) (*v1.Pod, error) {

	// We compare the hostId associated to the pod (cached into the resource status) and the one associated
	// to the podIp into the cassandra cluster.
	hostId, keyFound := hostIDMap[pod.Status.PodIP]
	statusHostId := status.CassandraNodesStatus[pod.Name].HostId
	logrus.WithFields(logrus.Fields{"pod": pod.Name}).Debug(fmt.Sprintf("Test Keyfound %t, statusHostId : %s, hostId: %s", keyFound, statusHostId, hostId))

	// If the hostId associated to the pod in the status is not the same one
	// that the one associated into cassandra for the same IP, so we are in IP cross cases.
	if keyFound == true && statusHostId != hostId {
		logrus.WithFields(logrus.Fields{"pod": pod.Name}).
			Info(fmt.Sprintf("Pod %s, have a cross Ip situation. The pod have ip : %s, with hostId : %s, "+
				"but this ip is already associated to the hostId : %s. We force delete of the pod", pod.Name, pod.Status.PodIP, statusHostId, hostId))
		return pod, nil
	}
	return nil, nil
}
