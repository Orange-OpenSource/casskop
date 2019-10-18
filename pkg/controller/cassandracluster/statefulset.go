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
	"strings"
	"time"

	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/k8s"
	"github.com/allamand/godebug/pretty"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	//patch "github.com/banzaicloud/k8s-objectmatcher/patch"
)

var (
	retryInterval = time.Second
	timeout       = time.Second * 5
)

//GetStatefulSet return the Statefulset name from the cluster in the namespace
func (rcc *ReconcileCassandraCluster) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {

	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return ss, rcc.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, ss)
}

func (rcc *ReconcileCassandraCluster) DeleteStatefulSet(namespace, name string) error {

	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return rcc.client.Delete(context.TODO(), ss)
}

//CreateStatefulSet create a new statefulset ss
func (rcc *ReconcileCassandraCluster) CreateStatefulSet(statefulSet *appsv1.StatefulSet) error {
	err := rcc.client.Create(context.TODO(), statefulSet)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("statefulset already exists: %cc", err)
		}
		return fmt.Errorf("failed to create cassandra statefulset: %cc", err)
		//return err
	}
	return nil
}

//UpdateStatefulSet updates an existing statefulset ss
func (rcc *ReconcileCassandraCluster) UpdateStatefulSet(statefulSet *appsv1.StatefulSet) error {
	err := rcc.client.Update(context.TODO(), statefulSet)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("statefulset already exists: %cc", err)
		}
		return fmt.Errorf("failed to update cassandra statefulset: %cc", err)
	}
	//Check that the new revision of statefulset has been taken into account
	err = wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		newSts, err := rcc.GetStatefulSet(statefulSet.Namespace, statefulSet.Name)
		if err != nil && !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("failed to get cassandra statefulset: %cc", err)
		}
		if statefulSet.ResourceVersion != newSts.ResourceVersion {
			logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name, "statefulset": statefulSet.Name}).Info(
				"Statefulset has new revision, we continue")
			return true, nil
		}
		logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name, "statefulset": statefulSet.Name}).Info(
			"Waiting for new version of statefulset")
		return false, nil
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name, "statefulset": statefulSet.Name}).Info(
			"Error Waiting for sts change")
	}
	return nil
}

// sts1 = stored statefulset and sts2 = new generated statefulset
func statefulSetsAreEqual(sts1, sts2 *appsv1.StatefulSet) bool {

	//updates to statefulset spec for fields other than 'replicas', 'template', and 'updateStrategy' are forbidden.

	//Things we won't check :
	//TODO: we should rely on library to ease the comparison between 2 statefulset and to know If we need to apply
	// change or not cf: https://github.com/banzaicloud/k8s-objectmatcher

	/*
		patchResult, err := patch.DefaultPatchMaker.Calculate(sts1, sts2)
		if err != nil {
			logrus.Infof("Template is different: " + pretty.Compare(sts1.Spec, sts2.Spec))
			return false
		}
		if !patchResult.IsEmpty() {
			logrus.Infof("Template is different: " + pretty.Compare(sts1.Spec, sts2.Spec))
			return false
		}
	*/
	sts1.Spec.Template.Spec.SchedulerName = sts2.Spec.Template.Spec.SchedulerName
	sts1.Spec.Template.Spec.DNSPolicy = sts2.Spec.Template.Spec.DNSPolicy // ClusterFirst
	for i := 0; i < len(sts1.Spec.Template.Spec.Containers); i++ {
		sts1.Spec.Template.Spec.Containers[i].LivenessProbe = sts2.Spec.Template.Spec.Containers[i].LivenessProbe
		sts1.Spec.Template.Spec.Containers[i].ReadinessProbe = sts2.Spec.Template.Spec.Containers[i].ReadinessProbe

		sts1.Spec.Template.Spec.Containers[i].TerminationMessagePath = sts2.Spec.Template.Spec.Containers[i].TerminationMessagePath
		sts1.Spec.Template.Spec.Containers[i].TerminationMessagePolicy = sts2.Spec.Template.Spec.Containers[i].TerminationMessagePolicy
	}

	for i := 0; i < len(sts1.Spec.Template.Spec.InitContainers); i++ {
		sts1.Spec.Template.Spec.InitContainers[i].LivenessProbe = sts2.Spec.Template.Spec.InitContainers[i].LivenessProbe
		sts1.Spec.Template.Spec.InitContainers[i].ReadinessProbe = sts2.Spec.Template.Spec.InitContainers[i].ReadinessProbe

		sts1.Spec.Template.Spec.InitContainers[i].TerminationMessagePath = sts2.Spec.Template.Spec.InitContainers[i].TerminationMessagePath
		sts1.Spec.Template.Spec.InitContainers[i].TerminationMessagePolicy = sts2.Spec.Template.Spec.InitContainers[i].TerminationMessagePolicy
	}

	//some defaultMode changes make falsepositif, so we bypass this, we already have check on configmap changes
	sts1.Spec.VolumeClaimTemplates = sts2.Spec.VolumeClaimTemplates
	sts1.Spec.PodManagementPolicy = sts2.Spec.PodManagementPolicy
	sts1.Spec.RevisionHistoryLimit = sts2.Spec.RevisionHistoryLimit

	if !apiequality.Semantic.DeepEqual(sts1.Spec, sts2.Spec) {
		logrus.WithFields(logrus.Fields{"statefulset": sts1.Name,
			"namespace": sts1.Namespace}).Info("Template is different: " + pretty.CompareAndPrintDiff(sts1.Spec, sts2.Spec))

		return false
	}

	return true
}

//CreateOrUpdateStatefulSet Create statefulset if not existing, or update it if existing.
func (rcc *ReconcileCassandraCluster) CreateOrUpdateStatefulSet(statefulSet *appsv1.StatefulSet,
	status *api.CassandraClusterStatus, dcRackName string) (bool, error) {
	dcRackStatus := status.CassandraRackStatus[dcRackName]
	var err error
	now := metav1.Now()

	rcc.storedStatefulSet, err = rcc.GetStatefulSet(statefulSet.Namespace, statefulSet.Name)
	if err != nil {
		// If no resource we need to create.
		if apierrors.IsNotFound(err) {
			return api.BreakResyncLoop, rcc.CreateStatefulSet(statefulSet)
		}
		return api.ContinueResyncLoop, err
	}

	//We will not Update the Statefulset
	// if there is existing disruptions on Pods
	// Or if we are not scaling Down the current statefulset
	if rcc.thereIsPodDisruption() {
		if rcc.weAreScalingDown(dcRackStatus) && rcc.hasOneDisruptedPod() {
			logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name,
				"dc-rack": dcRackName}).Info("Cluster has 1 Pod Disrupted" +
				"but that may be normal as we are decommissioning")
		} else if rcc.cc.Spec.UnlockNextOperation {
			logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name,
				"dc-rack": dcRackName}).Warn("Cluster has 1 Pod Disrupted" +
				"but we have unlock the next operation")
		} else {
			logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name,
				"dc-rack": dcRackName}).Info("Cluster has Disruption on Pods, " +
				"we wait before applying any change to statefulset")
			return api.ContinueResyncLoop, nil
		}
	}

	// Already exists, need to Update.
	statefulSet.ResourceVersion = rcc.storedStatefulSet.ResourceVersion
	// We grab the existing labels and add them back to the generated StatefulSet
	statefulSet.Spec.Template.SetLabels(rcc.storedStatefulSet.Spec.Template.GetLabels())

	//If UpdateSeedList=Ongoing, we allow the new SeedList to be propagated into the Statefulset
	//and change the status to Finalizing (it start a RollingUpdate)
	if dcRackStatus.CassandraLastAction.Name == api.ActionUpdateSeedList &&
		dcRackStatus.CassandraLastAction.Status == api.StatusToDo {
		logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name, "dc-rack": dcRackName}).Info("Update SeedList on Rack")
		dcRackStatus.CassandraLastAction.Status = api.StatusOngoing
		dcRackStatus.CassandraLastAction.StartTime = &now
	} else {

		//We need to keep the SeedList from the stored statefulset
		//we retrieve it in the Env CASSANDRA_SEEDS of the bootstrap container
		ic := getBootstrapContainerFromStatefulset(statefulSet)
		oldIc := getBootstrapContainerFromStatefulset(rcc.storedStatefulSet)
		for i, env := range ic.Env {
			if env.Name == "CASSANDRA_SEEDS" {
				for _, oldenv := range oldIc.Env {
					if oldenv.Name == "CASSANDRA_SEEDS" && env.Value != oldenv.Value {
						ic.Env[i].Value = oldenv.Value
					}
				}
			}
		}
	}

	//Hack for ScaleDown:
	//because before applying a scaledown at Kubernetes (statefulset) level we need to execute a cassandra decommission
	//we want the statefulset to only perform one scaledown at a time.
	//we have some call which will block the call of this method until the decommission is not OK, so here
	//we just need to change the scaledown value if more than 1 at a time.
	if *rcc.storedStatefulSet.Spec.Replicas-*statefulSet.Spec.Replicas > 1 {
		*statefulSet.Spec.Replicas = *rcc.storedStatefulSet.Spec.Replicas - 1
	}

	if dcRackStatus.CassandraLastAction.Name == api.ActionRollingRestart &&
		dcRackStatus.CassandraLastAction.Status == api.StatusToDo {
		statefulSet.Spec.Template.SetLabels(k8s.MergeLabels(statefulSet.Spec.Template.GetLabels(), map[string]string{
			"rolling-restart": k8s.LabelTime()}))
		dcRackStatus.CassandraLastAction.Status = api.StatusOngoing
		dcRackStatus.CassandraLastAction.StartTime = &now
	}

	//Except for RollingRestart we check If Statefulset has changed
	if !rcc.cc.Spec.NoCheckStsAreEqual &&
		statefulSetsAreEqual(rcc.storedStatefulSet.DeepCopy(), statefulSet.DeepCopy()) {
		logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name,
			"dc-rack": dcRackName}).Debug("Statefulsets Are Equal: No Update")
		return api.ContinueResyncLoop, nil
	}

	//If the Status is To-Do, then the Action will be Ongoing once we update the statefulset
	if dcRackStatus.CassandraLastAction.Status == api.StatusToDo {
		dcRackStatus.CassandraLastAction.Status = api.StatusOngoing
		dcRackStatus.CassandraLastAction.StartTime = &now
		dcRackStatus.CassandraLastAction.EndTime = nil
	}

	if !rcc.cc.Spec.NoCheckStsAreEqual &&
		dcRackStatus.CassandraLastAction.Status == api.StatusDone {
		logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name,
			"dc-rack": statefulSet.Labels["dc-rack"]}).Debug("Start Updating Statefulset")
		dcRackStatus.CassandraLastAction.Status = api.StatusOngoing
		dcRackStatus.CassandraLastAction.Name = api.ActionUpdateStatefulSet
		dcRackStatus.CassandraLastAction.StartTime = &now
		dcRackStatus.CassandraLastAction.EndTime = nil
	}

	return api.BreakResyncLoop, rcc.UpdateStatefulSet(statefulSet)

}

func getBootstrapContainerFromStatefulset(sts *appsv1.StatefulSet) *v1.Container {
	for _, ic := range sts.Spec.Template.Spec.InitContainers {
		if ic.Name == "bootstrap" {
			return &ic
		}
	}
	return nil
}

func getStoredSeedListTab(storedStatefulSet *appsv1.StatefulSet) []string {
	ic := getBootstrapContainerFromStatefulset(storedStatefulSet)
	if ic != nil {
		for _, env := range ic.Env {
			if env.Name == "CASSANDRA_SEEDS" {
				return strings.Split(env.Value, ",")
			}
		}
	}
	return []string{}
}

func isStatefulSetNotReady(storedStatefulSet *appsv1.StatefulSet) bool {
	if storedStatefulSet.Status.Replicas != *storedStatefulSet.Spec.Replicas ||
		storedStatefulSet.Status.ReadyReplicas != *storedStatefulSet.Spec.Replicas {
		return true
	}
	return false
}
