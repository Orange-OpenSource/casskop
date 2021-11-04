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

package k8s

import (
	"context"
	"fmt"
	"regexp"
	"time"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Regex to extract date from label
var ReLabelTime = regexp.MustCompile(`(?P<y>\d{4})(?P<m>\d{2})(?P<d>\d{2})T(?P<hh>\d{2})(?P<mm>\d{2})(?P<ss>\d{2})`)

// addOwnerRefToObject appends the desired OwnerReference to the object
func AddOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}

// labelsForCassandra returns the labels for selecting the resources
// belonging to the given name.
func LabelsForCassandraDCRack(cc *api.CassandraCluster, dcName string, rackName string) map[string]string {
	m := map[string]string{
		"app":                                  "cassandracluster",
		"cassandracluster":                     cc.GetName(),
		"dc-rack":                              cc.GetDCRackName(dcName, rackName),
		"cassandraclusters.db.orange.com.dc":   dcName,
		"cassandraclusters.db.orange.com.rack": rackName,
	}
	return MergeLabels(cc.GetLabels(), m)
}

func LabelsForCassandraDC(cc *api.CassandraCluster, dcName string) map[string]string {
	m := map[string]string{
		"app":                                "cassandracluster",
		"cassandracluster":                   cc.GetName(),
	}

	if len(dcName) > 0{
		m["cassandraclusters.db.orange.com.dc"] = dcName
	}
	return MergeLabels(cc.GetLabels(), m)
}

func LabelsForCassandra(cc *api.CassandraCluster) map[string]string {
	m := map[string]string{
		"app":              "cassandracluster",
		"cassandracluster": cc.GetName(),
	}
	return MergeLabels(cc.GetLabels(), m)
}

func RemoveString(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//ContainSlice return true if each element of n exists in ref
func ContainSlice(ref []string, n []string) bool {
	for i := range n {
		if !Contains(ref, n[i]) {
			return false
		}
	}
	return true
}

//MergeSlice will add ad the end of old any elements of new which is missing
//we want to keep the order of elements in old
func MergeSlice(old []string, new []string) []string {
	var result []string

	//result start from old but we don't add if elem don't exists in new
	for i := range old {
		if Contains(new, old[i]) {
			result = append(result, old[i])
		}
	}
	//Do we need to add more elements from new ?
	for i := range new {
		if !Contains(old, new[i]) {
			result = append(result, new[i])
		}
	}

	return result
}

// MergeLabels merges all the label maps received as argument into a single new label map.
func MergeLabels(allLabels ...map[string]string) map[string]string {
	res := map[string]string{}

	for _, labels := range allLabels {
		for k, v := range labels {
			res[k] = v
		}
	}
	return res
}

// asOwner returns an owner reference set as the cassandra cluster CRD
func AsOwner(cc *api.CassandraCluster) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       "CassandraCluster",
		Name:       cc.Name,
		UID:        cc.UID,
		Controller: &trueVar,
	}
}

// LabelTime returns a supported label string containing the current date and time
func LabelTime() string {
	t := metav1.Now()
	return fmt.Sprintf("%d%02d%02dT%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// LabelTime2Time converts a label string containing a time into a Time
func LabelTime2Time(label string) (time.Time, error) {
	reformattedLabel := ReLabelTime.ReplaceAllString(label, `${y}-${m}-${d}T${hh}:${mm}:${ss}`)
	return time.Parse("2006-01-02T15:04:05", reformattedLabel)
}

// DCRackLabelsAndNodeSelectorForStatefulSet function return a map with the labels DC & Rack to deploy
// on the statefulset.
// dc and int are the indice of respectively the dc and the rack in the CassandraCluster configuration
func DCRackLabelsAndNodeSelectorForStatefulSet(cc *api.CassandraCluster, dc int, rack int) (map[string]string, map[string]string) {
	var dcName, rackName string
	var nodeSelector = map[string]string{}

	dcsize := len(cc.Spec.Topology.DC)
	if dcsize < 1 || dc > dcsize-1 {
		dcName = api.DefaultCassandraDC
		rackName = api.DefaultCassandraRack
	} else {
		nodeSelector = MergeLabels(cc.Spec.Topology.DC[dc].Labels)
		dcName = cc.Spec.Topology.DC[dc].Name
		racksize := len(cc.Spec.Topology.DC[dc].Rack)
		if racksize < 1 || rack > racksize-1 {
			rackName = "Rack-1"
		} else {
			nodeSelector = MergeLabels(nodeSelector, cc.Spec.Topology.DC[dc].Rack[rack].Labels)
			rackName = cc.Spec.Topology.DC[dc].Rack[rack].Name
		}
	}

	labels := MergeLabels(LabelsForCassandraDCRack(cc, dcName, rackName), map[string]string{
		"cassandraclusters.db.orange.com.dc":   dcName,
		"cassandraclusters.db.orange.com.rack": rackName,
	})

	return labels, nodeSelector
}

// LookupCassandra Cluster returns the running cluster instance based on its name and namespace
func LookupCassandraCluster(client runtimeClient.Client, clusterName,
	clusterNamespace string) (cluster *api.CassandraCluster, err error) {
	cluster = &api.CassandraCluster{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: clusterName, Namespace: clusterNamespace}, cluster)
	return
}

func LookupCassandraBackup(client runtimeClient.Client, backupName,
	backupNamespace string) (backup *api.CassandraBackup, err error) {
	backup = &api.CassandraBackup{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: backupName, Namespace: backupNamespace}, backup)
	return
}

// IsMarkedForDeletion determines if the object is marked for deletion
func IsMarkedForDeletion(m metav1.ObjectMeta) bool {
	return m.GetDeletionTimestamp() != nil
}

// PodHostname returns hostname of a pod
func PodHostname(pod v1.Pod) string {
	return fmt.Sprintf("%s.%s", pod.Spec.Hostname, pod.Spec.Subdomain)
}

func PodByName(podList *v1.PodList, podName string) *v1.Pod {
	for _, pod := range podList.Items {
		if pod.Name == podName {
			return &pod
		}
	}
	return nil
}
