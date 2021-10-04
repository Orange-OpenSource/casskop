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
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"

	"time"

	api "github.com/Orange-OpenSource/casskop/api/v2"
	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

type finalizedOp struct {
	err           error
	dcRackName    string
	pod           v1.Pod
	operationName string
}

type op struct {
	Action     func(*CassandraClusterReconciler, string, *api.CassandraCluster, string, v1.Pod) error
	Monitor    func(*JolokiaClient) (bool, error)
	PostAction func(*CassandraClusterReconciler, *api.CassandraCluster, string, v1.Pod) error
}

type operationMode string

const (
	NORMAL operationMode = "NORMAL"
	LEAVING = "LEAVING"
	DECOMMISSIONED = "DECOMMISSIONED"
	UNKNOWN = "UNKNOWN"
)

var podOperationMap = map[string]op{
	api.OperationCleanup:         {(*CassandraClusterReconciler).runCleanup,
		(*JolokiaClient).hasCleanupCompactions, nil},
	api.OperationRebuild:         {(*CassandraClusterReconciler).runRebuild,
		(*JolokiaClient).hasStreamingSessions, nil},
	api.OperationUpgradeSSTables: {(*CassandraClusterReconciler).runUpgradeSSTables,
		(*JolokiaClient).hasUpgradeSSTablesCompactions, nil},
	api.OperationRemove:          {(*CassandraClusterReconciler).runRemove,
		(*JolokiaClient).hasLeavingNodes,(*CassandraClusterReconciler).postRunRemove}}

const breakResyncLoop    = true
const continueResyncLoop = false
const monitorSleepDelay  = 10 * time.Second
const deletedPvcTimeout  = 30 * time.Second

var chanRunningOp = make(chan finalizedOp, 100)

func randomPodOperationKey() string {
	r := rand.Intn(len(podOperationMap))
	for k := range podOperationMap {
		if r == 0 {
			return k
		}
		r--
	}
	return "" // will never happen but make the compiler happy ¯\_(ツ)_/¯
}

//handlePodOperation will ensure that all Pod Operations which needed to be performed are done accordingly.
//It may return a breakResyncloop order meaning that the Operator won't update the statefulset until
//PodOperations are finishing gracefully.
func (rcc *CassandraClusterReconciler) handlePodOperation(cc *api.CassandraCluster, dcName, rackName string,
	status *api.CassandraClusterStatus, statefulsetIsReady bool) (bool, error) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	dcRackStatus := status.CassandraRackStatus[dcRackName]
	var breakResyncLoopSwitch = false
	var err error

	// If we ask a ScaleDown, We can't update the Statefulset before the decommission is done
	if rcc.weAreScalingDown(dcRackStatus) {
		//If a Decommission is Ongoing, we want to break the Resyncloop until the Decommission is succeed
		breakResyncLoopSwitch, err = rcc.ensureDecommission(cc, dcName, rackName, status, statefulsetIsReady)
		if err != nil {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc": dcName, "rack": rackName,
				"err": err}).Error("Error with decommission")
		}
		return breakResyncLoopSwitch, err
	}

	podsList, err := rcc.ListCassandraClusterPods(cc)
	if err != nil {
		return true, err
	}
	firstPod, err := GetLastOrFirstPodReady(podsList, false)
	if err != nil {
		return true, err
	}

	hostName := k8s.PodHostname(*firstPod)
	jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc, cc.Spec.ImageJolokiaSecret, cc.Namespace)

	hasJoiningNodes, err := jolokiaClient.hasJoiningNodes()
	if err != nil {
		return true, err
	}
	if hasJoiningNodes {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "dc": dcName, "rack": rackName,
			"err": err}).Error("Can't continue cause some nodes are joining the cluster")
		return hasJoiningNodes, nil
	}

	// If LastClusterAction was a ScaleUp and It is Done then
	// Execute Cleanup On labeled Pods
	if status.LastClusterActionStatus == api.StatusDone {
		// If I enable test on ScaleUp then it may be too restrictive :
		// we won't be able to label pods to execute an action outside of a scaleup
		// && status.LastClusterAction == api.ActionScaleUp {

		// Finalize operations that are done
		rcc.finalizeOperations(cc)

		// We run approximately a different operation each time
		rcc.ensureOperation(cc, dcName, rackName, status, randomPodOperationKey())
	}

	return breakResyncLoopSwitch, err
}

//addPodOperationLabels will add Pod Labels labels on all Pod in the Current dcRackName
func (rcc *CassandraClusterReconciler) addPodOperationLabels(cc *api.CassandraCluster, dcName string,
	rackName string, labels map[string]string) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	//Select all Pods in the Rack
	selector := k8s.MergeLabels(k8s.LabelsForCassandraDCRack(cc, dcName, rackName))

	podsList, err := rcc.ListPods(cc.Namespace, selector)

	if err != nil || len(podsList.Items) < 1 {
		return
	}

	for _, pod := range podsList.Items {
		if pod.Status.Phase != v1.PodRunning || pod.DeletionTimestamp != nil {
			continue
		}

		newlabels := k8s.MergeLabels(pod.GetLabels(), labels)

		pod.SetLabels(newlabels)
		err = rcc.UpdatePod(&pod)
		if err != nil {
			logrus.Errorf("[%s][%s]:[%s] UpdatePod Error: %v", cc.Name, dcRackName, pod.Name, err)
		}

		logrus.Infof("[%s][%s]:[%s] UpdatePod Labels: %v", cc.Name, dcRackName, pod.Name, labels)

	}
}

// initOperation finds pods waiting for operation to run
func (rcc *CassandraClusterReconciler) initOperation(cc *api.CassandraCluster, status *api.CassandraClusterStatus,
	dcName, rackName, operationName string) []v1.Pod {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	selector := k8s.MergeLabels(k8s.LabelsForCassandraDCRack(cc, dcName, rackName),
		map[string]string{"operation-name": operationName,
			"operation-status": api.StatusToDo})

	podsList, err := rcc.ListPods(cc.Namespace, selector)
	now := metav1.Now()

	podLastOperation := &status.CassandraRackStatus[dcRackName].PodLastOperation

	if err != nil || len(podsList.Items) < 1 {

		if podLastOperation.Name == operationName && podLastOperation.Status == api.StatusOngoing && len(podLastOperation.Pods) < 1 {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
				"operation": strings.Title(operationName)}).Debug("Set podLastOperation to Done as there is no more Pod to work on")
			podLastOperation.Status = api.StatusDone
			podLastOperation.EndTime = &now

			//We want dynamic view of status on CassandraCluster
			rcc.updateCassandraStatus(cc, status)
		}
		return nil
	}

	if podLastOperation.Status != api.StatusOngoing {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"operation": strings.Title(operationName)}).Debug("Reset podLastOperation attributes")
		podLastOperation.Name = operationName
		podLastOperation.Status = api.StatusOngoing
		podLastOperation.StartTime = &now
		podLastOperation.EndTime = nil
		podLastOperation.PodsOK = []string{}
		podLastOperation.PodsKO = []string{}
		podLastOperation.Pods = []string{}

		//We want dynamic view of status on CassandraCluster
		rcc.updateCassandraStatus(cc, status)
	}

	return func(podsList *v1.PodList) []v1.Pod {
		podsSlice := make([]v1.Pod, 0)
		for _, pod := range podsList.Items {
			if pod.Status.Phase != v1.PodRunning || pod.DeletionTimestamp != nil {
				continue
			}
			podsSlice = append(podsSlice, pod)
		}
		return podsSlice
	}(podsList)
}

func (rcc *CassandraClusterReconciler) startOperation(cc *api.CassandraCluster, status *api.CassandraClusterStatus,
	pod v1.Pod, dcRackName, operationName string) error {
	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
		"operation": strings.Title(operationName)}).Info("Start operation")
	labels := map[string]string{"operation-status": api.StatusOngoing,
		"operation-start": k8s.LabelTime(), "operation-end": ""}

	err := rcc.UpdatePodLabel(&pod, labels)
	if err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"pod": pod.Name, "err": err.Error(), "labels": labels}).Debug("Failed to add labels to pod")
		return err
	}

	podLastOperation := &status.CassandraRackStatus[dcRackName].PodLastOperation
	podLastOperation.Pods = append(podLastOperation.Pods, pod.Name)
	podLastOperation.PodsOK = k8s.RemoveString(podLastOperation.PodsOK, pod.Name)
	podLastOperation.PodsKO = k8s.RemoveString(podLastOperation.PodsKO, pod.Name)

	rcc.updateCassandraStatus(cc, status)

	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
		"pod": pod.Name, "operation": strings.Title(operationName),
		"podLastOperation.OperatorName": podLastOperation.OperatorName,
		"podLastOperation.Pods":         podLastOperation.Pods}).Debug("Display information about pods")
	return nil
}

// ensureOperation goal is to find pods with Labels :
//  - operation-name=xxxx and operation-status=To-Do
func (rcc *CassandraClusterReconciler) ensureOperation(cc *api.CassandraCluster, dcName, rackName string,
	status *api.CassandraClusterStatus, operationName string) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	podsSlice, checkOnly := rcc.getPodsToWorkOn(cc, dcName, rackName, status, operationName)

	// For each pod where we need to run the operation on
	for _, pod := range podsSlice {
		hostName := k8s.PodHostname(pod)
		// We check if an operation is running
		if checkOnly {
			go rcc.monitorOperation(hostName, cc, dcRackName, pod, operationName)
			continue
		}
		// Add the operatorName to the last pod operation in case the operator pod is replaced
		status.CassandraRackStatus[dcRackName].PodLastOperation.OperatorName = os.Getenv("POD_NAME")
		err := rcc.startOperation(cc, status, pod, dcRackName, operationName)
		if err != nil {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
				"pod": pod.Name, "err": err}).Debug("Failed to start operation on pod")
			continue
		}
		go rcc.runOperation(operationName, hostName, cc, dcRackName, pod, status)
	}
}

func (rcc *CassandraClusterReconciler) finalizeOperations(cc *api.CassandraCluster) {
	// Finalize all operations here to avoid update conflicts
	for chanOp := 0; chanOp < len(chanRunningOp); chanOp++ {
		op := <-chanRunningOp
		rcc.finalizeOperation(op.err, cc, op.dcRackName, op.pod, &rcc.cc.Status,
			strings.Title(op.operationName))
	}
}

func (rcc *CassandraClusterReconciler) runOperation(operationName, hostName string, cc *api.CassandraCluster, dcRackName string, pod v1.Pod,
	status *api.CassandraClusterStatus) {
	err := podOperationMap[operationName].Action(rcc, hostName, cc, dcRackName, pod)

	// If there is an error we finalize the operation but skip any existing post action
	if err != nil {
		chanRunningOp <- finalizedOp{err, dcRackName, pod, operationName}
		return
	}
	postAction := podOperationMap[operationName].PostAction
	if postAction != nil {
		err = postAction(rcc, cc, dcRackName, pod)
	}
	chanRunningOp <- finalizedOp{err, dcRackName, pod, operationName}
}

/* ensureDecommission will ensure that the Last Pod of the StatefulSet will be decommissionned
	- If pod.status=To-DO then executeDecommission in the Pod and flag pod.status as **Ongoing**
	- If pod.status=Ongoing then if pod is not running then flag its status as **Done**
	- If pod.status=Done then delete Pod PVC and ChangeActionStatus to **Continue**

  it return breakResyncloop=true is we need to bypass update of the Statefulset.
  it return breakResyncloop=false if we want to call the ensureStatefulset method. */
func (rcc *CassandraClusterReconciler) ensureDecommission(cc *api.CassandraCluster, dcName, rackName string,
	status *api.CassandraClusterStatus, statefulsetIsReady bool) (bool, error) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	podLastOperation := &status.CassandraRackStatus[dcRackName].PodLastOperation

	if podLastOperation.Name != api.OperationDecommission {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"lastOperation": podLastOperation.Name}).Warnf("Another operation than decommission was asked")
		return continueResyncLoop, nil
	}

	switch podLastOperation.Status {

	case api.StatusToDo, api.StatusContinue:

		return rcc.ensureDecommissionToDo(cc, dcName, rackName, status)

	case api.StatusFinalizing:
		lastPod, err := rcc.GetPod(cc.Namespace, podLastOperation.Pods[0])
		if err != nil {
			if apierrors.IsNotFound(err) {
				return rcc.deletePodPVC(cc, dcName, rackName, status, lastPod, statefulsetIsReady)
			}
			return breakResyncLoop, fmt.Errorf("Failed to get pod %s: %v", podLastOperation.Pods[0], err)
		}

		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"lastPod": lastPod.Name}).Infof("Statefulset is scaling down, waiting..")
		return breakResyncLoop, nil

	case api.StatusOngoing:

		if podLastOperation.Pods == nil || podLastOperation.Pods[0] == "" {
			return breakResyncLoop, fmt.Errorf("Status is Ongoing, we should have a PodLastOperation Pods item")
		}

		lastPod, err := rcc.GetPod(cc.Namespace, podLastOperation.Pods[0])
		if err != nil {
			return breakResyncLoop, fmt.Errorf(
				"Failed to get last pod '%s': %v", podLastOperation.Pods[0], err)
		}

		hostName := k8s.PodHostname(*lastPod)
		jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc, cc.Spec.ImageJolokiaSecret, cc.Namespace)

		if err != nil {
			return breakResyncLoop, err
		}

		operationMode, err := jolokiaClient.NodeOperationMode()

		if err != nil {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
				"hostName": hostName, "err": err}).Error("Jolokia call failed")
			return breakResyncLoop, err
		}

		if operationMode == NORMAL {
			t, err := k8s.LabelTime2Time(lastPod.Labels["operation-start"])
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"operation-start": lastPod.Labels["operation-start"],
				}).Debugf("Can't parse time")
			}
			now, _ := k8s.LabelTime2Time(k8s.LabelTime())

			if t.Add(api.DefaultDelayWaitForDecommission * time.Second).After(now) {
				logrus.WithFields(logrus.Fields{
					"cluster": cc.Name, "rack": dcRackName, "pod": lastPod.Name,
					"operationMode": operationMode,
					"DefaultDelayWaitForDecommission": api.DefaultDelayWaitForDecommission,
				}).Info("Decommission was applied less than {DefaultDelayWaitForDecommission} seconds, waiting")
			} else {
				logrus.WithFields(logrus.Fields{
					"cluster": cc.Name, "rack": dcRackName, "pod": lastPod.Name, "operationMode": operationMode,
				}).Info("Seems that decommission has not correctly been applied, trying again..")
				status.CassandraRackStatus[dcRackName].PodLastOperation.Status = api.StatusToDo
			}
			return breakResyncLoop, nil
		}

		if operationMode == DECOMMISSIONED || operationMode == UNKNOWN {
			logrus.WithFields(logrus.Fields{
				"cluster": cc.Name, "rack": dcRackName, "lastPod": lastPod.Name, "operationMode": operationMode,
			}).Infof("Node has left the ring, waiting for statefulset Scaledown")
			podLastOperation.Status = api.StatusFinalizing
			if err := rcc.updateCassandraStatus(cc, status); err != nil {
				return continueResyncLoop, err
			}
			return continueResyncLoop, nil
		}

		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": lastPod.Name,
			"operationMode": operationMode}).Info("Cassandra node is decommissioning, we need to wait")
		return breakResyncLoop, nil

	//Set LastAction to Continue in case more decommissions are needed
	case api.StatusDone:
		if podLastOperation.PodsOK == nil || podLastOperation.PodsOK[0] == "" {
			return breakResyncLoop, fmt.Errorf("For Status Done we should have a PodLastOperation.PodsOK item")
		}
		status.CassandraRackStatus[dcRackName].CassandraLastAction.Status = api.StatusContinue
		return breakResyncLoop, nil

	default:
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"status": podLastOperation.Status}).Errorf("Error this should not happened: unknown status")
	}

	return continueResyncLoop, nil
}

//ensureDecommissionToDo
// State To-DO -> Ongoing
// set podLastOperation.Pods and label targeted pod (lastPod)
func (rcc *CassandraClusterReconciler) ensureDecommissionToDo(cc *api.CassandraCluster, dcName, rackName string,
	status *api.CassandraClusterStatus) (bool, error) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	var list []string
	podLastOperation := &status.CassandraRackStatus[dcRackName].PodLastOperation

	// We Get LastPod From StatefulSet
	lastPod, err := rcc.GetLastPod(cc.Namespace, k8s.LabelsForCassandraDCRack(cc, dcName, rackName))
	if err != nil {
		return breakResyncLoop, fmt.Errorf("Failed to get last cassandra's pods: %v", err)
	}
	//If Pod is unschedulable, we bypass decommission (cassandra is not running)
	if lastPod.Status.Phase == v1.PodPending &&
		lastPod.Status.Conditions != nil &&
		lastPod.Status.Conditions[0].Reason == "Unschedulable" {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"pod": lastPod.Name}).Warn("ScaleDown detected on a pending Pod. we don't launch decommission")
		podLastOperation.Status = api.StatusFinalizing
		podLastOperation.PodsOK = []string{}
		podLastOperation.Pods = append(list, lastPod.Name)
		podLastOperation.PodsKO = []string{}
		status.CassandraRackStatus[dcRackName].CassandraLastAction.Status = api.StatusContinue
		return continueResyncLoop, nil
	}
	if lastPod.Status.Phase != v1.PodRunning || lastPod.DeletionTimestamp != nil {
		return breakResyncLoop, fmt.Errorf("Pod is not running")
	}
	logrusFields := logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": lastPod.Name}
	logrus.WithFields(logrusFields).Info("ScaleDown detected, we launch decommission")

	//Ensure node is not leaving or absent from the ring
	hostName := k8s.PodHostname(*lastPod)
	jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc, cc.Spec.ImageJolokiaSecret, cc.Namespace)

	if err != nil {
		return breakResyncLoop, err
	}

	operationMode, err := jolokiaClient.NodeOperationMode()

	if err != nil {
		logrusFields["err"] = err
		logrusFields["hostname"] = hostName
		logrus.WithFields(logrusFields).Error("Jolokia call failed")
		return breakResyncLoop, err
	}

	if operationMode == DECOMMISSIONED || operationMode == UNKNOWN || operationMode == LEAVING {
		logrus.WithFields(logrusFields).Info("Node is leaving or has already been decommissioned")
		return breakResyncLoop, nil
	}

	if err = rcc.UpdatePodLabel(lastPod,
		map[string]string{
			"operation-status": api.StatusOngoing,
			"operation-start":  k8s.LabelTime(),
			"operation-name":   api.OperationDecommission,
		}); err != nil {
		logrusFields["err"] = err
		logrus.WithFields(logrusFields).Debug("Error updating pod")
	}

	podLastOperation.Status = api.StatusOngoing
	podLastOperation.Pods = append(list, lastPod.Name)
	podLastOperation.PodsOK = []string{}
	podLastOperation.PodsKO = []string{}

	logrus.WithFields(logrusFields).Debug("Decommissioning cassandra node")

	go func() {
		logrus.WithFields(logrusFields).Debug("Node decommission starts")
		err = jolokiaClient.NodeDecommission(cc.Spec.ServerVersion >= "4.0")
		logrus.WithFields(logrusFields).Debug("Node decommission ended")
		if err != nil {
			logrusFields["err"] = err
			logrus.WithFields(logrusFields).Debug("Node decommission failed")
		}
	}()

	return breakResyncLoop, nil
}

//deletePodPVC
// State To-DO -> Ongoing
func (rcc *CassandraClusterReconciler) deletePodPVC(cc *api.CassandraCluster, dcName, rackName string,
	status *api.CassandraClusterStatus, lastPod *v1.Pod, statefulsetIsReady bool) (bool, error) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	podLastOperation := &status.CassandraRackStatus[dcRackName].PodLastOperation

	pvcName := "data-" + podLastOperation.Pods[0]
	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
		"pvc": pvcName}).Info("Decommission done -> we delete PVC")
	if pvc, err := rcc.GetPVC(cc.Namespace, pvcName); err == nil {
		if rcc.deletePVC(pvc) != nil {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
				"pvc": pvcName}).Error("Error deleting PVC, Please make manual Actions..")
		} else {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
				"pvc": pvcName}).Info("PVC deleted")
		}
	} else if !apierrors.IsNotFound(err) {
		// Error when looking for the PVC let's retry
		return breakResyncLoop, nil
	}

	SetStatusForMoreDecommissions(podLastOperation, !statefulsetIsReady)

	podLastOperation.PodsOK = []string{lastPod.Name}
	now := metav1.Now()
	podLastOperation.EndTime = &now
	podLastOperation.Pods = []string{}
	return breakResyncLoop, nil
}

func SetStatusForMoreDecommissions(podLastOperation *api.PodLastOperation, moreDecommisions bool) {
	if moreDecommisions {
		podLastOperation.Status = api.StatusContinue
	} else {
		podLastOperation.Status = api.StatusDone
	}
}

func (rcc *CassandraClusterReconciler) podsSlice(cc *api.CassandraCluster, status *api.CassandraClusterStatus,
	podLastOperation api.PodLastOperation, dcRackName, operationName, operatorName string) ([]v1.Pod, bool) {
	checkOnly := false
	podsSlice := make([]v1.Pod, 0)
	// Operator is different from when the previous operation was started
	// Set checkOnly to restart the monitoring function to wait until the operation is done
	if podLastOperation.Name == operationName && podLastOperation.Status == api.StatusOngoing &&
		podLastOperation.OperatorName != "" && podLastOperation.OperatorName != operatorName {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"podLastOperation.OperatorName": podLastOperation.OperatorName, "operatorName": operatorName,
			"operation": strings.Title(operationName)}).Info("Operator's name is different, we enable checking routines")
		podLastOperation.OperatorName = operatorName

		for _, podName := range podLastOperation.Pods {
			p, err := rcc.GetPod(cc.Namespace, podName)
			if err != nil || p.Status.Phase != v1.PodRunning || p.DeletionTimestamp != nil {
				continue
			}
			podsSlice = append(podsSlice, *p)
		}
		checkOnly = true
		return podsSlice, checkOnly
	}
	dcName, rackName := cc.GetDCNameAndRackNameFromDCRackName(dcRackName)
	podsSlice = rcc.initOperation(cc, status, dcName, rackName, operationName)
	return podsSlice, checkOnly
}

// Get pods that need an operation to run on
// Returns if checking is needed (can happen if the operator has been killed during an operation)
func (rcc *CassandraClusterReconciler) getPodsToWorkOn(cc *api.CassandraCluster, dcName, rackName string,
	status *api.CassandraClusterStatus, operationName string) ([]v1.Pod, bool) {
	dcRackName := cc.GetDCRackName(dcName, rackName)
	var checkOnly bool
	podsSlice := make([]v1.Pod, 0)

	operatorName := os.Getenv("POD_NAME")
	if len(operatorName) == 0 {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName}).Info("POD_NAME is not defined and is mandatory")
		return podsSlice, checkOnly
	}

	// Every time we update this variable we have to run updateCassandraStatus
	podLastOperation := &status.CassandraRackStatus[dcRackName].PodLastOperation

	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
		"podLastOperation.OperatorName": podLastOperation.OperatorName,
		"podLastOperation.Pods":         podLastOperation.Pods}).Debug("Display information about pods")

	podsSlice, checkOnly = rcc.podsSlice(cc, status, *podLastOperation, dcRackName, operationName, operatorName)

	if checkOnly {
		if len(podsSlice) == 0 {
			// If previous running pods are done or cannot be found, we update the operator status
			podLastOperation.Status = api.StatusDone
			now := metav1.Now()
			podLastOperation.EndTime = &now
		}
		rcc.updateCassandraStatus(cc, status)
	}
	return podsSlice, checkOnly
}

func (rcc *CassandraClusterReconciler) updatePodLastOperation(clusterName, dcRackName, podName, operation string,
	status *api.CassandraClusterStatus, err error) {
	podLastOperation := &status.CassandraRackStatus[dcRackName].PodLastOperation
	if err != nil {
		// We set the operation-status to Error on failing pods
		logrus.WithFields(logrus.Fields{"cluster": clusterName, "rack": dcRackName, "pod": podName,
			"operation": operation, "err": err.Error()}).Error("Error in updatePodLastOperation")
		podLastOperation.PodsKO = append(podLastOperation.PodsKO, podName)
	} else {
		podLastOperation.PodsOK = append(podLastOperation.PodsOK, podName)
	}
	// We remove the pod from the list of pods running the operation
	podLastOperation.Pods = k8s.RemoveString(podLastOperation.Pods, podName)
}

/* finalizeOperation sets the labels on the pod where ran an operation depending on the error status
   It also updates status.CassandraRackStatus[dcRackName].PodLastOperation
*/
func (rcc *CassandraClusterReconciler) finalizeOperation(err error, cc *api.CassandraCluster, dcRackName string,
	pod v1.Pod, status *api.CassandraClusterStatus, operationName string) {
	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
		"status": status, "operation": operationName}).Debug("Finalize operation")
	labels := map[string]string{"operation-status": api.StatusDone, "operation-end": k8s.LabelTime()}

	if err != nil {
		labels["operation-status"] = api.StatusError
	}

	ccRefreshed := cc.DeepCopy()

	rcc.updatePodLastOperation(cc.Name, dcRackName, pod.Name, strings.Title(operationName), status, err)

	for {
		if err = rcc.UpdatePodLabel(&pod, labels); err != nil {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
				"labels": labels, "error": err}).Error("Can't update labels")
			continue
		}
		if err = rcc.updateCassandraStatus(ccRefreshed, status); err == nil {
			break
		}
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
			"status": status, "error": err}).Debug("Got an error. Getting a new version of Cassandra Cluster")
		if rcc.Client.Get(context.TODO(), types.NamespacedName{Name: cc.Name, Namespace: cc.Namespace}, ccRefreshed) == nil {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
				"status": status}).Debug("Got a new version of Cassandra Cluster")
			continue
		}
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
			"status": status}).Debug("Can't get new version of Cassandra Cluster. Will try again")
		time.Sleep(retryInterval)
	}
}

func (rcc *CassandraClusterReconciler) monitorOperation(hostName string, cc *api.CassandraCluster, dcRackName string,
	pod v1.Pod, operationName string) {
	// Wait until there are no more cleanup compactions
	for {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"pod": pod.Name, "host": hostName, "operation": operationName}).Info("Checking if operation is still running on node")
		jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc,
			cc.Spec.ImageJolokiaSecret, cc.Namespace)
		if err == nil {
			operationIsRunning, err := podOperationMap[operationName].Monitor(jolokiaClient)
			// When there is an error it returns true to try again during the next loop
			if err != nil {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
					"pod": pod.Name, "host": hostName, "operation": operationName, "err": err}).Error("Got an error from Jolokia")
				operationIsRunning = true
			}
			if operationIsRunning != true {
				break
			}
		}
		time.Sleep(monitorSleepDelay)
	}
	postAction := podOperationMap[operationName].PostAction
	var err error
	if postAction != nil {
		err = postAction(rcc, cc, dcRackName, pod)
	}
	chanRunningOp <- finalizedOp{err, dcRackName, pod, operationName}
}

func (rcc *CassandraClusterReconciler) runUpgradeSSTables(hostName string, cc *api.CassandraCluster, dcRackName string,
	pod v1.Pod) error {
	var err error
	operation := strings.Title(api.OperationUpgradeSSTables)

	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
		"hostName": hostName, "operation": operation}).Info("Operation start")

	jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc,
		cc.Spec.ImageJolokiaSecret, cc.Namespace)
	if err == nil {
		err = jolokiaClient.NodeUpgradeSSTables(0)
	}
	return err
}

func (rcc *CassandraClusterReconciler) runRebuild(hostName string, cc *api.CassandraCluster, dcRackName string, pod v1.Pod) error {
	var err error
	var keyspaces []string
	var rebuildFrom, labelSet = pod.GetLabels()["operation-argument"]
	operation := strings.Title(api.OperationRebuild)

	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
		"hostName": hostName, "operation": operation}).Info("Operation start")

	jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc,
		cc.Spec.ImageJolokiaSecret, cc.Namespace)

	if labelSet != true {
		err = errors.New("operation-argument is needed to get the datacenter name to rebuild from")
	} else if keyspaces, err = jolokiaClient.NonLocalKeyspacesInDC(rebuildFrom); err == nil && len(keyspaces) == 0 {
		err = fmt.Errorf("%s  has no keyspace to replicate data from", rebuildFrom)
	}

	// In case of an error set the status on the pod and skip it
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
		"datacenter": rebuildFrom, "operation": operation}).Info("Execute the Jolokia Operation")

	if err == nil {
		err = jolokiaClient.NodeRebuild(rebuildFrom)
	}
	return err
}

func (rcc *CassandraClusterReconciler) runRemove(hostName string, cc *api.CassandraCluster, dcRackName string, pod v1.Pod) error {
	operation := strings.Title(api.OperationRemove)

	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
		"hostName": hostName, "operation": operation}).Info("Operation start")

	var label, labelSet = pod.GetLabels()["operation-argument"]
	if labelSet != true {
		return errors.New("operation-argument is needed to get the pod name to remove from the cluster")
	}

	val := strings.Split(label, "_")
	podToRemove := val[0]
	var podIPToRemove string
	if len(val) == 2 {
		podIPToRemove = val[1]
	}

	if podToRemove == "" && podIPToRemove == "" {
		return fmt.Errorf("Expected format is `[Name][_IP]` with at least one value but none was found")
	}
	// Name can be omitted in case the pod has already been deleted but then IP must be provided
	// When an IP is provided it will be used by the removeNode operation
	if podIPToRemove != "" && net.ParseIP(podIPToRemove) == nil {
		return fmt.Errorf("%s is not an IP address", podIPToRemove)
	}

	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
		"nodeToRemove": podToRemove, "operation": operation}).Info("Execute the Jolokia Operation")

	var lostPod *v1.Pod
	var err error
	if podToRemove != "" {
		// We delete the pod that is no longer part of the cluster
		lostPod, err = rcc.GetPod(cc.Namespace, podToRemove)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("Failed to get pod '%s': %v", podToRemove, err)
			}
			// If we can't find it, it means it has already been deleted somehow. That's okay as long as we got its IP
			if podIPToRemove == "" {
				return fmt.Errorf("Pod %s not found. You need to provide its old IP to remove it from the cluster", podToRemove)
			}
		}
	}

	// If no IP is not provided, we grab it from the existing pod
	if podIPToRemove == "" {
		podIPToRemove = lostPod.Status.PodIP
		if podIPToRemove == "" {
			return fmt.Errorf("Can't find an IP assigned to pod %s. You need to provide its old IP to remove it from the cluster", podToRemove)
		}
	}

	jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc, cc.Spec.ImageJolokiaSecret, cc.Namespace)

	if err == nil {
		var hostIDMap map[string]string
		// Get hostID from internal map and pass it to removeNode function
		if hostIDMap, err = jolokiaClient.hostIDMap(); err == nil {
			if hostID, keyFound := hostIDMap[podIPToRemove]; keyFound != true {
				err = fmt.Errorf("Host with IP '%s' not found in hostIdMap", podIPToRemove)
			} else {
				logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
					"nodeToRemove": podToRemove, "operation": operation}).Info("Jolokia Remove node operation")
				err = jolokiaClient.NodeRemove(hostID)
			}
		}
	}

	return err
}

func (rcc *CassandraClusterReconciler) waitUntilPvcIsDeleted(namespace, pvcName string) error {
	err := wait.Poll(retryInterval, deletedPvcTimeout, func() (done bool, err error) {
		_, err = rcc.GetPVC(namespace, pvcName)
		if err != nil && apierrors.IsNotFound(err) {
			logrus.WithFields(logrus.Fields{"namespace": namespace,
				"pvc": pvcName}).Info("PVC no longer exists")
			return true, nil
		}
		logrus.WithFields(logrus.Fields{"namespace": namespace,
			"pvc": pvcName}).Info("Waiting for PVC to be deleted")
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (rcc *CassandraClusterReconciler) postRunRemove(cc *api.CassandraCluster, dcRackName string, pod v1.Pod) error {
	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name}).Info("Post operation start")

	var label, labelSet = pod.GetLabels()["operation-argument"]
	if labelSet != true {
		return errors.New("operation-argument is needed to get the pod name to remove from the cluster")
	}
	podToRemove := strings.Split(label, "_")[0]

	if podToRemove == "" {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name,
			"rack": dcRackName}).Info("RemoveNode done. No pod was provided so we're done'")
		return nil
	}

	// We delete the attached PVC
	pvcName := "data-" + podToRemove
	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
		"pvc": pvcName}).Info("RemoveNode done. We now delete its PVC")

	pvc, err := rcc.GetPVC(cc.Namespace, pvcName)
	if err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"pvc": pvcName}).Error("Cannot get PVC")
	} else {
		err = rcc.deletePVC(pvc)
		if err != nil && !apierrors.IsNotFound(err) {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
				"pvc": pvcName}).Error("Error deleting PVC, manual actions required...")
			return err
		}
		_ = rcc.waitUntilPvcIsDeleted(cc.Namespace, pvcName)
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"pvc": pvcName}).Info("PVC deleted")
	}

	// We delete the pod that is no longer part of the cluster
	lostPod, err := rcc.GetPod(cc.Namespace, podToRemove)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("Failed to get pod '%s': %v", podToRemove, err)
		}
	}
	err = rcc.ForceDeletePod(lostPod)

	if err != nil {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"pod": podToRemove}).Error("Error deleting Pod, manual actions required...")
	} else {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName,
			"pod": podToRemove}).Info("Pod deleted")
	}
	return err
}

func (rcc *CassandraClusterReconciler) runCleanup(hostName string, cc *api.CassandraCluster, dcRackName string, pod v1.Pod) error {
	var err error
	operation := strings.Title(api.OperationCleanup)

	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
		"hostName": hostName, "operation": operation}).Info("Operation start")

	// In case of an error set the status on the pod and skip it
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "rack": dcRackName, "pod": pod.Name,
		"operation": operation}).Info("Execute the Jolokia Operation")

	jolokiaClient, err := NewJolokiaClient(hostName, JolokiaPort, rcc,
		cc.Spec.ImageJolokiaSecret, cc.Namespace)

	if err == nil {
		err = jolokiaClient.NodeCleanup()
	}
	return err
}
