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
	"math"
	"sort"
	"strings"
	"time"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"k8s.io/apimachinery/pkg/util/wait"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/allamand/godebug/pretty"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
	return ss, rcc.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, ss)
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
	return rcc.Client.Delete(context.TODO(), ss)
}

//CreateStatefulSet create a new statefulset ss
func (rcc *ReconcileCassandraCluster) CreateStatefulSet(statefulSet *appsv1.StatefulSet) error {
	err := rcc.Client.Create(context.TODO(), statefulSet)
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
	revision := statefulSet.ResourceVersion
	if err := rcc.Client.Update(context.TODO(), statefulSet); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("statefulset already exists: %cc", err)
		}
		return fmt.Errorf("failed to update cassandra statefulset: %cc", err)
	}
	//Check that the new revision of statefulset has been taken into account
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		newSts, err := rcc.GetStatefulSet(statefulSet.Namespace, statefulSet.Name)
		if err != nil && !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("failed to get cassandra statefulset: %cc", err)
		}
		if revision != newSts.ResourceVersion {
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

func sortPVCs(pvcs *[]v1.PersistentVolumeClaim) {
	sort.Slice(*pvcs, func(i, j int) bool {
		return (*pvcs)[i].Name < (*pvcs)[j].Name
	})
}

func sortContainers(containers *[]v1.Container) {
	sort.Slice(*containers, func(i, j int) bool {
		return (*containers)[i].Name < (*containers)[j].Name
	})
}

func statefulSetsAreEqual(sts1, sts2 *appsv1.StatefulSet) bool {

	//updates to statefulset spec for fields other than 'replicas', 'template', and 'updateStrategy' are forbidden.
	sts1Spec := &sts1.Spec.Template.Spec
	sts2Spec := &sts2.Spec.Template.Spec

	sts1Spec.SchedulerName = sts2Spec.SchedulerName
	sts1Spec.DNSPolicy = sts2Spec.DNSPolicy // ClusterFirst

	containersSpecSts1Len := len(sts1Spec.Containers)
	containersSpecSts2Len := len(sts2Spec.Containers)
	if containersSpecSts1Len != containersSpecSts2Len {
		logrus.WithFields(logrus.Fields{"statefulset": sts1.Name,
			"namespace": sts1.Namespace}).Info(
			fmt.Sprintf("Template is different, the number of containers are not the same: len(sts1.Spec.Template.Spec.Containers) = %d, "+
				"len(sts2.Spec.Template.Spec.Containers) = %d",
				containersSpecSts1Len, containersSpecSts2Len))
		return false
	}

	pvcsSpecSts1Len := len(sts1.Spec.VolumeClaimTemplates)
	pvcsSpecSts2Len := len(sts2.Spec.VolumeClaimTemplates)
	if pvcsSpecSts1Len != pvcsSpecSts2Len {
		logrus.WithFields(logrus.Fields{"statefulset": sts1.Name,
			"namespace": sts1.Namespace}).Info(
			fmt.Sprintf("Template is different, the number of pvcs are not the same: len(sts1.Spec.VolumeClaimTemplates) = %d, "+
				"len(sts2.Spec.Template.Spec.VolumeClaimTemplates) = %d",
				pvcsSpecSts1Len, pvcsSpecSts2Len))
		return false
	}

	sortPVCs(&sts1.Spec.VolumeClaimTemplates)
	sortPVCs(&sts2.Spec.VolumeClaimTemplates)

	sortContainers(&sts1Spec.Containers)
	sortContainers(&sts2Spec.Containers)

	for i := 0; i < len(sts1.Spec.VolumeClaimTemplates); i++ {
		sts2.Spec.VolumeClaimTemplates[i].TypeMeta = sts1.Spec.VolumeClaimTemplates[i].TypeMeta
		sts2.Spec.VolumeClaimTemplates[i].Status = sts1.Spec.VolumeClaimTemplates[i].Status
		if sts2.Spec.VolumeClaimTemplates[i].Spec.VolumeMode == nil {
			sts2.Spec.VolumeClaimTemplates[i].Spec.VolumeMode = sts1.Spec.VolumeClaimTemplates[i].Spec.VolumeMode
		}
	}

	sts2.Status.Replicas = sts1.Status.Replicas

	if patchResult, err := patch.DefaultPatchMaker.Calculate(sts1, sts2); err != nil || !patchResult.IsEmpty() {
		logrus.Debug("Template is different: " + pretty.Compare(sts1.Spec, sts2.Spec))
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
	// if there is an existing pod disruptions
	// Or if we are not scaling Down the current statefulset
	if !rcc.hasNoPodDisruption() {
		if rcc.cc.Spec.UnlockNextOperation {
			logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name, "dc-rack": dcRackName}).Warn(
				"Cluster has a disruption but we have unlock the next operation")
		} else {
			logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name, "dc-rack": dcRackName}).Info(
				"Cluster has a disruption, waiting before applying any changes to statefulset")
			return api.ContinueResyncLoop, nil
		}
	}

	// Already exists, need to Update.
	statefulSet.ResourceVersion = rcc.storedStatefulSet.ResourceVersion
	// We grab the existing labels and add them back to the generated StatefulSet
	statefulSet.Spec.Template.SetLabels(rcc.storedStatefulSet.Spec.Template.GetLabels())

	//If UpdateSeedList=Ongoing, we allow the new SeedList to be propagated into the Statefulset
	//and change the status to Finalizing (it starts a RollingUpdate)
	if dcRackStatus.CassandraLastAction.Name == api.ActionUpdateSeedList.Name &&
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

	//Scaling operations must be done one node at a time to fully support decommissions and bootstraps
	replicaCountDiff := float64(*statefulSet.Spec.Replicas - *rcc.storedStatefulSet.Spec.Replicas)
	if math.Abs(replicaCountDiff) > 1 {
		scaleOperation := "up"
		incrementValue := int32(1)
		if replicaCountDiff < 0 {
			scaleOperation = "down"
			incrementValue *= -1
		}
		logrus.WithFields(logrus.Fields{"cluster": rcc.cc.Name,
			"dc-rack": dcRackName}).Debugf("Must scale %s one node at a time. Update stfs " +
			"replicas to %d instead of %d for now",
			scaleOperation, *rcc.storedStatefulSet.Spec.Replicas + incrementValue, *statefulSet.Spec.Replicas)
		*statefulSet.Spec.Replicas = *rcc.storedStatefulSet.Spec.Replicas + incrementValue
	}

	if dcRackStatus.CassandraLastAction.Name == api.ActionRollingRestart.Name &&
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
		dcRackStatus.CassandraLastAction.Name = api.ActionUpdateStatefulSet.Name
		dcRackStatus.CassandraLastAction.StartTime = &now
		dcRackStatus.CassandraLastAction.EndTime = nil
	}

	return api.BreakResyncLoop, rcc.UpdateStatefulSet(statefulSet)

}

func getBootstrapContainerFromStatefulset(sts *appsv1.StatefulSet) *v1.Container {
	for _, container := range sts.Spec.Template.Spec.InitContainers {
		if container.Name == "bootstrap" {
			return &container
		}
	}
	return nil
}

func getStoredSeedList(storedStatefulSet *appsv1.StatefulSet) []string {
	container := getBootstrapContainerFromStatefulset(storedStatefulSet)
	if container != nil {
		for _, env := range container.Env {
			if env.Name == "CASSANDRA_SEEDS" {
				return strings.Split(env.Value, ",")
			}
		}
	}
	return []string{}
}

func isStatefulSetNotReady(storedStatefulSet *appsv1.StatefulSet) bool {
	if storedStatefulSet.Status.ReadyReplicas != *storedStatefulSet.Spec.Replicas {
		return true
	}
	return false
}
