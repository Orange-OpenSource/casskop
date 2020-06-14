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
	"testing"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Orange-OpenSource/casskop/pkg/k8s"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestCreateNodeAffinity(t *testing.T) {
	assert := assert.New(t)

	nodeAffinity := createNodeAffinity(map[string]string{
		"A": "value1",
		"B": "value2",
		"C": "value3",
		"D": "value4",
		"E": "value5",
	})

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key, "A")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0], "value1")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Key, "B")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Values[0], "value2")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Key, "C")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Values[0], "value3")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Key, "D")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Values[0], "value4")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Key, "E")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Values[0], "value5")
}

func TestCreateNodeAffinitySort(t *testing.T) {
	assert := assert.New(t)

	//unsort labels gives sorted result
	nodeAffinity := createNodeAffinity(map[string]string{
		"B": "value2",
		"A": "value1",
		"D": "value4",
		"E": "value5",
		"C": "value3",
	})

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key, "A")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0], "value1")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Key, "B")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[1].Values[0], "value2")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Key, "C")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[2].Values[0], "value3")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Key, "D")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[3].Values[0], "value4")

	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Key, "E")
	assert.Equal(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[4].Values[0], "value5")
}

func TestCreatePodAntiAffinityHard(t *testing.T) {
	assert := assert.New(t)

	labels := map[string]string{
		"label1": "value1",
		"label2": "value2",
		"label3": "value3",
	}
	podAntiAffinityHard := createPodAntiAffinity(true, labels)

	assert.Equal(podAntiAffinityHard.RequiredDuringSchedulingIgnoredDuringExecution[0].TopologyKey, hostnameTopologyKey)
	assert.Equal(podAntiAffinityHard.RequiredDuringSchedulingIgnoredDuringExecution[0].LabelSelector.MatchLabels, labels)
}

func TestVolumeMounts(t *testing.T) {
	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

	volumeMounts := generateContainerVolumeMount(cc, initContainer)
	assert.Equal(t, 1, len(volumeMounts))
	assert.Equal(t, "/bootstrap", volumeMounts[getPos(volumeMounts, "bootstrap")].MountPath)

	volumeMounts = generateContainerVolumeMount(cc, bootstrapContainer)
	assert.Equal(t, 3, len(volumeMounts))
	assert.Equal(t, "/etc/cassandra", volumeMounts[getPos(volumeMounts, "bootstrap")].MountPath)
	assert.Equal(t, "/extra-lib", volumeMounts[getPos(volumeMounts, "extra-lib")].MountPath)
	assert.Equal(t, "/opt/bin", volumeMounts[getPos(volumeMounts, "tools")].MountPath)

	volumeMounts = generateContainerVolumeMount(cc, cassandraContainer)
	assert.Equal(t, 5, len(volumeMounts))
	assert.Equal(t, "/etc/cassandra", volumeMounts[getPos(volumeMounts, "bootstrap")].MountPath)
	assert.Equal(t, "/extra-lib", volumeMounts[getPos(volumeMounts, "extra-lib")].MountPath)
	assert.Equal(t, "/opt/bin", volumeMounts[getPos(volumeMounts, "tools")].MountPath)
	assert.Equal(t, "/tmp", volumeMounts[getPos(volumeMounts, "tmp")].MountPath)
	assert.Equal(t, "/var/lib/cassandra", volumeMounts[getPos(volumeMounts, "data")].MountPath)
}

func TestGenerateCassandraService(t *testing.T) {
	assert := assert.New(t)

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	selector := k8s.LabelsForCassandra(cc)
	svc := generateCassandraService(cc, selector, nil)

	assert.Equal(map[string]string{
		"app":              "cassandracluster",
		"cassandracluster": "cassandra-demo",
		"cluster":          "k8s.pic"},
		svc.Labels)
	assert.Equal(map[string]string{"external-dns.alpha.kubernetes.io/hostname": "my.custom.domain.com."},
		svc.Annotations)
}

func TestGenerateCassandraStatefulSet(t *testing.T) {
	assert := assert.New(t)
	dcName := "dc1"
	rackName := "rack1"
	dcRackName := fmt.Sprintf("%s-%s", dcName, rackName)

	_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")
	ccDefault := cc.DeepCopy()
	cc.CheckDefaults()
	labels, nodeSelector := k8s.GetDCRackLabelsAndNodeSelectorForStatefulSet(cc, 0, 0)
	sts, _ := generateCassandraStatefulSet(cc, &cc.Status, dcName, dcRackName, labels, nodeSelector, nil)

	assert.Equal(map[string]string{
		"app":                                  "cassandracluster",
		"cassandracluster":                     "cassandra-demo",
		"cassandraclusters.db.orange.com.dc":   "dc1",
		"cassandraclusters.db.orange.com.rack": "rack1",
		"dc-rack":                              "dc1-rack1",
		"cluster":                              "k8s.pic",
	}, sts.Labels)

	assert.Equal("my.custom.annotation", sts.Spec.Template.Annotations["exemple.com/test"])
	assert.Equal([]v1.Toleration{
		{
			Key:      "my_custom_taint",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}, sts.Spec.Template.Spec.Tolerations)

	checkVolumeClaimTemplates(t, labels, sts.Spec.VolumeClaimTemplates, "10Gi", "test-storage")
	checkLiveAndReadiNessProbe(t, sts.Spec.Template.Spec.Containers,
		1010, 201, 32, 7, 9, 1205, 151, 17, 50, 30)
	checkVolumeMount(t, sts.Spec.Template.Spec.Containers)
	checkVarEnv(t, sts.Spec.Template.Spec.Containers, cc, dcRackName)
	checkDefaultInitContainerResources(t, sts.Spec.Template.Spec.InitContainers)

	cc.Spec.StorageConfigs[0].PVCSpec = nil
	_, err := generateCassandraStatefulSet(cc, &cc.Status, dcName, dcRackName, labels, nodeSelector, nil)
	assert.NotEqual(t, err, nil)

	// Test default setup
	dcNameDefault := "dc2"
	rackNameDefault := "rack1"
	dcRackNameDefault := fmt.Sprintf("%s-%s", dcNameDefault, rackNameDefault)
	setupForDefaultTest(ccDefault)

	ccDefault.CheckDefaults()
	labelsDefault, nodeSelectorDefault := k8s.GetDCRackLabelsAndNodeSelectorForStatefulSet(ccDefault, 0, 0)
	stsDefault, _ := generateCassandraStatefulSet(ccDefault, &ccDefault.Status, dcNameDefault, dcRackNameDefault, labelsDefault, nodeSelectorDefault, nil)

	checkVolumeClaimTemplates(t, labels, stsDefault.Spec.VolumeClaimTemplates, "3Gi", "local-storage")
	checkLiveAndReadiNessProbe(t, stsDefault.Spec.Template.Spec.Containers,
		60, 10, 10, 0, 0, 120, 20, 10, 0, 0)
	checkDefaultInitContainerResources(t, stsDefault.Spec.Template.Spec.InitContainers)

}

func setupForDefaultTest(cc *api.CassandraCluster) {
	cc.Spec.LivenessFailureThreshold = nil
	cc.Spec.LivenessSuccessThreshold = nil
	cc.Spec.LivenessHealthCheckPeriod = nil
	cc.Spec.LivenessHealthCheckTimeout = nil
	cc.Spec.LivenessInitialDelaySeconds = nil
	cc.Spec.ReadinessHealthCheckPeriod = nil
	cc.Spec.ReadinessHealthCheckTimeout = nil
	cc.Spec.ReadinessInitialDelaySeconds = nil
	cc.Spec.ReadinessFailureThreshold = nil
	cc.Spec.ReadinessSuccessThreshold = nil
}

func checkLiveAndReadiNessProbe(t *testing.T, containers []v1.Container,
	readinessInitialDelaySecond,
	readinessTimeoutSeconds,
	readinessPeriodSeconds,
	readinessFailureThreshold,
	readinessSuccessThreshold,
	livenessInitialDelaySecond,
	livenessTimeoutSeconds,
	livenessPeriodSeconds,
	livenessFailureThreshold,
	livenessSuccessThreshold int32) {
	for _, c := range containers {
		if c.Name == cassandraContainerName {
			// Readiness Config check
			assert.Equal(t, readinessInitialDelaySecond, c.ReadinessProbe.InitialDelaySeconds)
			assert.Equal(t, readinessTimeoutSeconds, c.ReadinessProbe.TimeoutSeconds)
			assert.Equal(t, readinessPeriodSeconds, c.ReadinessProbe.PeriodSeconds)
			assert.Equal(t, readinessFailureThreshold, c.ReadinessProbe.FailureThreshold)
			assert.Equal(t, readinessSuccessThreshold, c.ReadinessProbe.SuccessThreshold)

			// Liveness Config check
			assert.Equal(t, livenessInitialDelaySecond, c.LivenessProbe.InitialDelaySeconds)
			assert.Equal(t, livenessTimeoutSeconds, c.LivenessProbe.TimeoutSeconds)
			assert.Equal(t, livenessPeriodSeconds, c.LivenessProbe.PeriodSeconds)
			assert.Equal(t, livenessFailureThreshold, c.LivenessProbe.FailureThreshold)
			assert.Equal(t, livenessSuccessThreshold, c.LivenessProbe.SuccessThreshold)
		}
	}
}

func checkVolumeClaimTemplates(t *testing.T, expectedlabels map[string]string, pvcs []v1.PersistentVolumeClaim,
	dataCapacity, dataClassStorage string) {
	assert.Equal(t, 3, len(pvcs))
	for _, pvc := range pvcs {
		switch pvc.Name {
		case "data":
			assert.Equal(t, generateExpectedDataStoragePVC(expectedlabels, dataCapacity, dataClassStorage), pvc)
		case "gc-logs":
			assert.Equal(t, generateExpectedGcLogsStoragePVC(expectedlabels), pvc)
		case "cassandra-logs":
			assert.Equal(t, generateExpectedCassandraLogsStoragePVC(expectedlabels), pvc)
		default:
			t.Errorf("unexpected pvc name: %s.", pvc.Name)
		}
	}
}

func generateExpectedDataStoragePVC(expectedlabels map[string]string, dataCapacity, dataClassStorage string) v1.PersistentVolumeClaim {

	expectedDataStorageQuantity, _ := resource.ParseQuantity(dataCapacity)

	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "data",
			Labels: expectedlabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},

			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": expectedDataStorageQuantity,
				},
			},
			StorageClassName: &dataClassStorage,
		},
	}
}

func generateExpectedGcLogsStoragePVC(expectedlabels map[string]string) v1.PersistentVolumeClaim {

	expectedDataStorageQuantity, _ := resource.ParseQuantity("10Gi")
	expectedDataStorageClassName := "standard-wait"

	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "gc-logs",
			Labels: expectedlabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},

			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": expectedDataStorageQuantity,
				},
			},
			StorageClassName: &expectedDataStorageClassName,
		},
	}
}

func generateExpectedCassandraLogsStoragePVC(expectedlabels map[string]string) v1.PersistentVolumeClaim {

	expectedDataStorageQuantity, _ := resource.ParseQuantity("10Gi")
	expectedDataStorageClassName := "standard-wait"

	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "cassandra-logs",
			Labels: expectedlabels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},

			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": expectedDataStorageQuantity,
				},
			},
			StorageClassName: &expectedDataStorageClassName,
		},
	}
}

func checkVolumeMount(t *testing.T, containers []v1.Container) {
	assert.Equal(t, len(containers), 3)
	for _, container := range containers {
		switch container.Name {
		case "cassandra":
			assert.Equal(t, len(container.VolumeMounts), 7)
		case "gc-logs":
			assert.Equal(t, len(container.VolumeMounts), 1)
		case "cassandra-logs":
			assert.Equal(t, len(container.VolumeMounts), 1)
		default:
			t.Errorf("unexpected container: %s.", container.Name)
		}

		_, cc := helperInitCluster(t, "cassandracluster-2DC.yaml")

		for _, volumeMount := range container.VolumeMounts {
			switch container.Name {
			case "cassandra":
				assert.True(t, volumesContains(append(generateContainerVolumeMount(cc, cassandraContainer),
					generateCassandraStorageConfigVolumeMounts()...), volumeMount))
			case "gc-logs":
				assert.True(t, volumesContains([]v1.VolumeMount{{Name: "gc-logs", MountPath: "/var/log/cassandra"}}, volumeMount))
			case "cassandra-logs":
				assert.True(t, volumesContains([]v1.VolumeMount{{Name: "cassandra-logs", MountPath: "/var/log/cassandra"}}, volumeMount))
			default:
				t.Errorf("unexpected container: %s.", container.Name)
			}
		}
	}
}

func checkDefaultInitContainerResources(t *testing.T, containers []v1.Container) {
	resources := api.CassandraResources{
		Limits:   api.CPUAndMem{Memory: defaultInitContainerLimitsMemory, CPU: defaultInitContainerLimitsCPU},
		Requests: api.CPUAndMem{Memory: defaultInitContainerRequestsMemory, CPU: defaultInitContainerRequestsCPU},
	}
	resourcesRequirements := v1.ResourceRequirements{
		Limits:   requests(resources),
		Requests: limits(resources),
	}

	for _, container := range containers {
		switch container.Name {
		case "bootstrap":
			assert.Equal(t, container.Resources, resourcesRequirements)
		case "init-config":
			assert.Equal(t, container.Resources, resourcesRequirements)
		default:
		}
	}
}

func volumesContains(vms []v1.VolumeMount, mount v1.VolumeMount) bool {
	for _, vm := range vms {
		if mount == vm {
			return true
		}
	}
	return false
}

func generateCassandraStorageConfigVolumeMounts() []v1.VolumeMount {
	var vms []v1.VolumeMount
	vms = append(vms, v1.VolumeMount{Name: "gc-logs", MountPath: "/var/lib/cassandra/log"})
	vms = append(vms, v1.VolumeMount{Name: "cassandra-logs", MountPath: "/var/log/cassandra"})

	return vms
}

func checkVarEnv(t *testing.T, containers []v1.Container, cc *api.CassandraCluster, dcRackName string) {
	cassieResources := cassandraResources(cc.Spec)
	envVar := bootstrapContainerEnvVar(cc, &cc.Status, cassieResources, dcRackName)

	cassandraEnvVars := map[string]string{
		cassandraMaxHeap:         defineJvmMemory(cassieResources).maxHeapSize,
		"CASSANDRA_SEEDS":        "",
		"CASSANDRA_CLUSTER_NAME": clusterName,
		"POD_IP":                 "",
		"CASSANDRA_GC_STDOUT":    "false",
		"CASSANDRA_NUM_TOKENS":   "256",
		"CASSANDRA_DC":           "",
		"CASSANDRA_RACK":         "",
	}

	for name, value := range cassandraEnvVars {
		assert.Equal(t, value, cassandraEnvVars[name])
	}

	for _, container := range containers {
		if container.Name != cassandraContainerName {
			for _, env := range envVar {
				assert.Contains(t, envVar, env)
			}
		}
	}
}
