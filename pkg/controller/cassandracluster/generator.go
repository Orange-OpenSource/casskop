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

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"

	"github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/k8s"

	"sort"
	"strconv"
	"strings"
)

/*JvmMemory sets the maximium size of the heap*/
type JvmMemory struct {
	maxHeapSize string
}

/*Bunch of different constants*/
const (
	cassandraContainerName = "cassandra"
	defaultJvmMaxHeap      = "2048M"
	hostnameTopologyKey    = "kubernetes.io/hostname"
	JolokiaPort            = 8778
	ThriftPort             = 9160
	CqlPort                = 9042
	JmxPort                = 7199

	passwordAuthenticator = "PasswordAuthenticator"
	passwordAuthorizer    = "CassandraAuthorizer"

	livenessInitialDelaySeconds int32 = 120
	livenessHealthCheckTimeout  int32 = 20
	livenessHealthCheckPeriod   int32 = 10

	readinessInitialDelaySeconds int32 = 60
	readinessHealthCheckTimeout  int32 = 10
	readinessHealthCheckPeriod   int32 = 10
)

func generateCassandraService(cc *api.CassandraCluster, dcName string, labels map[string]string, ownerRefs []metav1.OwnerReference) *v1.Service {

	//	labels = util.MergeLabels(labels, generateLabels(cassandraRoleName, cc.Name))

	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            cc.GetName() + "-" + dcName,
			Namespace:       cc.GetNamespace(),
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: v1.ClusterIPNone,
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Port:     cassandraPort,
					Protocol: v1.ProtocolTCP,
					Name:     cassandraPortName,
				},
				v1.ServicePort{
					Port:     cassandraThrift,
					Protocol: v1.ProtocolTCP,
					Name:     cassandraThriftName,
				},
				v1.ServicePort{
					Port:     exporterPort,
					Protocol: v1.ProtocolTCP,
					Name:     exporterPortName,
				},
			},
			Selector: labels,
		},
	}
}

func generateCassandraExporterService(cc *api.CassandraCluster, dcRackName string, labels map[string]string, ownerRefs []metav1.OwnerReference) *v1.Service {
	name := cc.GetName()
	namespace := cc.Namespace

	mlabels := k8s.MergeLabels(labels, map[string]string{"k8s-app": "exporter-cassandra-jmx"})

	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s-exporter-jmx", name, dcRackName),
			Namespace:       namespace,
			Labels:          mlabels,
			OwnerReferences: ownerRefs,
		},
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: v1.ClusterIPNone,
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Port:     exporterCassandraJmxPort,
					Protocol: v1.ProtocolTCP,
					Name:     exporterCassandraJmxPortName,
				},
			},
			Selector: labels,
		},
	}
}

func generateCassandraVolumes(cc *api.CassandraCluster) []v1.Volume {
	var v []v1.Volume

	if cc.Spec.ConfigMapName != "" {
		v = append(v, v1.Volume{
			Name: "cassandra-config",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: cc.Spec.ConfigMapName,
					},
					DefaultMode: func(i int32) *int32 { return &i }(493), //493 is base10 to 0755 base8
				},
			},
		})
	}

	return v
}

func generateCassandraVolumeMount(cc *api.CassandraCluster) []v1.VolumeMount {
	var vm []v1.VolumeMount

	if cc.Spec.DataCapacity != "" {
		vm = append(vm, v1.VolumeMount{
			Name:      "data",
			MountPath: "/var/lib/cassandra",
		})
	}

	if cc.Spec.ConfigMapName != "" {
		vm = append(vm, v1.VolumeMount{
			Name:      "cassandra-config",
			MountPath: "/tmp/cassandra/configmap",
		})
	}
	return vm
}

func generateVolumeClaimTemplate(cc *api.CassandraCluster, labels map[string]string) []v1.PersistentVolumeClaim {

	var pvc []v1.PersistentVolumeClaim

	if cc.Spec.DataCapacity != "" {

		pvc = []v1.PersistentVolumeClaim{
			v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "data",
					Labels: labels,
				},
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{
						v1.ReadWriteOnce,
					},

					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"storage": generateResourceQuantity(cc.Spec.DataCapacity),
						},
					},
				},
			},
		}
	} else {
		logrus.Warnf("[%s]: No Spec.DataCapacity was specified -> You Cluster WILL NOT HAVE PERSISTENT DATA!!!!!", cc.Name)
	}

	if cc.Spec.DataStorageClass != "" {
		pvc[0].Spec.StorageClassName = &cc.Spec.DataStorageClass
	}

	return pvc
}

func generateCassandraStatefulSet(cc *api.CassandraCluster, status *api.CassandraClusterStatus, dcRackName string,
	labels map[string]string, nodeSelector map[string]string, ownerRefs []metav1.OwnerReference) *appsv1.StatefulSet {
	name := cc.GetName()
	namespace := cc.Namespace

	spec := cc.Spec
	cassandraImage := k8s.GetCassandraImage(cc)
	resources := getCassandraResources(spec)

	volumes := generateCassandraVolumes(cc)
	volumemounts := generateCassandraVolumeMount(cc)

	volumeClaimTemplate := generateVolumeClaimTemplate(cc, labels)

	for _, pvc := range volumeClaimTemplate {
		k8s.AddOwnerRefToObject(&pvc, k8s.AsOwner(cc))
	}

	nodeAffinity := createNodeAffinity(nodeSelector)

	nodesPerRacks := cc.GetNodesPerRacks(dcRackName)
	numTokensPerRacks := cc.GetNumTokensPerRacks(dcRackName)

	rollingPartition := cc.GetRollingPartitionPerRacks(dcRackName)

	//in statefulset.go we surcharge this value with conditions
	seedList := cc.GetSeedList(&status.SeedList)

	terminationPeriod := int64(api.DefaultTerminationGracePeriodSeconds)

	ss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name + "-" + dcRackName,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: name + "-" + dcRackName,
			Replicas:    &nodesPerRacks,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: "RollingUpdate",
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					Partition: &rollingPartition,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Affinity: &v1.Affinity{
						NodeAffinity:    nodeAffinity,
						PodAntiAffinity: createPodAntiAffinity(cc.Spec.HardAntiAffinity, k8s.LabelsForCassandra(cc)),
					},
					SecurityContext: &v1.PodSecurityContext{
						RunAsUser:    cc.Spec.RunAsUser,
						RunAsNonRoot: func(b bool) *bool { return &b }(true),

						FSGroup: func(i int64) *int64 { return &i }(1),
					},

					Containers: []v1.Container{
						v1.Container{
							Name:            cassandraContainerName,
							Image:           cassandraImage,
							ImagePullPolicy: cc.Spec.ImagePullPolicy,
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									Name:          "intra-node",
									ContainerPort: 7000,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									Name:          "tls-intra-node",
									ContainerPort: 7001,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									Name:          "jmx",
									ContainerPort: JmxPort,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									Name:          "cql",
									ContainerPort: CqlPort,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									Name:          "promjmx",
									ContainerPort: 1234,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									Name:          "jolokia",
									ContainerPort: JolokiaPort,
									Protocol:      v1.ProtocolTCP,
								},
								v1.ContainerPort{
									Name:          "thrift",
									ContainerPort: ThriftPort,
									Protocol:      v1.ProtocolTCP,
								},
							},

							SecurityContext: &v1.SecurityContext{
								Capabilities: &v1.Capabilities{
									Add: []v1.Capability{
										"IPC_LOCK",
									},
								},
								ProcMount: func(s v1.ProcMountType) *v1.ProcMountType { return &s }(v1.DefaultProcMount),
							},

							Lifecycle: &v1.Lifecycle{
								PreStop: &v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"/bin/bash",
											"-c",
											"/etc/cassandra/pre_stop.sh",
										},
									},
								},
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "CASSANDRA_MAX_HEAP",
									Value: defineJvmMemory(resources).maxHeapSize,
								},
								v1.EnvVar{
									Name:  "CASSANDRA_SEEDS",
									Value: seedList,
								},
								v1.EnvVar{
									Name:  "CASSANDRA_CLUSTER_NAME",
									Value: name,
								},
								v1.EnvVar{
									Name:  "CASSANDRA_AUTO_BOOTSTRAP",
									Value: "true",
								},
								v1.EnvVar{
									Name:  "CASSANDRA_AUTHENTICATOR",
									Value: passwordAuthenticator,
								},
								v1.EnvVar{
									Name:  "CASSANDRA_AUTHORIZER",
									Value: passwordAuthorizer,
								},
								v1.EnvVar{
									Name: "POD_IP",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "status.podIP",
										},
									},
								},
								v1.EnvVar{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
								v1.EnvVar{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
								v1.EnvVar{
									Name:  "SERVICE_NAME",
									Value: name + "-" + dcRackName,
								},
								v1.EnvVar{
									Name:  "CASSANDRA_GC_STDOUT",
									Value: strconv.FormatBool(cc.Spec.GCStdout),
								},
								v1.EnvVar{
									Name:  "CASSANDRA_NUM_TOKENS",
									Value: strconv.Itoa(int(numTokensPerRacks)),
								},
								v1.EnvVar{
									Name: "NODE_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
								v1.EnvVar{
									Name: "CASSANDRA_DC",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.labels['cassandraclusters.db.orange.com.dc']",
										},
									},
								},
								v1.EnvVar{
									Name: "CASSANDRA_RACK",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.labels['cassandraclusters.db.orange.com.rack']",
										},
									},
								},
							},
							ReadinessProbe: &v1.Probe{
								InitialDelaySeconds: readinessInitialDelaySeconds,
								TimeoutSeconds:      readinessHealthCheckTimeout,
								PeriodSeconds:       readinessHealthCheckPeriod,
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"/bin/bash",
											"-c",
											"/ready-probe.sh",
										},
									},
								},
							},
							LivenessProbe: &v1.Probe{
								InitialDelaySeconds: livenessInitialDelaySeconds,
								TimeoutSeconds:      livenessHealthCheckTimeout,
								PeriodSeconds:       livenessHealthCheckPeriod,
								Handler: v1.Handler{
									Exec: &v1.ExecAction{
										Command: []string{
											"/bin/bash",
											"-c",
											"nodetool status",
										},
									},
								},
							},
							VolumeMounts: volumemounts,
							//on utilise la command par defaut de l'image
							/* we remove this to use the cmd specified in the docker image*/
							/*
								Command: []string{
									//"/sbin/dumb-init",
									//"/bin/bash",
									//"/run.sh",
									//debug: keep container running in case of problem
									"sh",
									"-c",
									"tail -f /dev/null",
								},
							*/
							/**/
							Resources: resources,
						},
					},
					Volumes:                       volumes,
					RestartPolicy:                 v1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &terminationPeriod,
				},
			},
			VolumeClaimTemplates: volumeClaimTemplate,
		},
	}

	//Add secrets

	if (cc.Spec.ImagePullSecret != v1.LocalObjectReference{}) {
		ss.Spec.Template.Spec.ImagePullSecrets = []v1.LocalObjectReference{cc.Spec.ImagePullSecret}
	}

	if (cc.Spec.ImageJolokiaSecret != v1.LocalObjectReference{}) {
		for idx, container := range ss.Spec.Template.Spec.Containers {
			if container.Name == cassandraContainerName {
				ss.Spec.Template.Spec.Containers[idx].Env = append(container.Env,
					v1.EnvVar{
						Name: "JOLOKIA_USER",
						ValueFrom: &v1.EnvVarSource{
							SecretKeyRef: &v1.SecretKeySelector{
								LocalObjectReference: cc.Spec.ImageJolokiaSecret,
								Key:                  "username",
							},
						},
					},
					v1.EnvVar{
						Name: "JOLOKIA_PASSWORD",
						ValueFrom: &v1.EnvVarSource{
							SecretKeyRef: &v1.SecretKeySelector{
								LocalObjectReference: cc.Spec.ImageJolokiaSecret,
								Key:                  "password",
							},
						},
					},
					v1.EnvVar{
						Name:  "CASSANDRA_AUTH_JOLOKIA",
						Value: "true"})
			}
		}
	}

	return ss
}

func generateResourceQuantity(qs string) resource.Quantity {
	q, _ := resource.ParseQuantity(qs)
	return q
}

func defineJvmMemory(resources v1.ResourceRequirements) JvmMemory {

	var mhs string

	if resources.Limits.Memory().IsZero() == false {
		m := float64(resources.Limits.Memory().Value()) * float64(0.25) // Maxheapsize should be 1/4 of container Memory Limit
		mi := int(m / float64(1048576))
		mhs = strings.Join([]string{strconv.Itoa(mi), "M"}, "")

	} else {
		mhs = defaultJvmMaxHeap
	}

	return JvmMemory{
		maxHeapSize: mhs,
	}
}

func generatePodDisruptionBudget(name string, namespace string, labels map[string]string, ownerRefs metav1.OwnerReference, maxUnavailable intstr.IntOrString) *policyv1beta1.PodDisruptionBudget {
	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{ownerRefs},
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
}

func getCassandraResources(spec api.CassandraClusterSpec) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: getRequests(spec.Resources),
		Limits:   getLimits(spec.Resources),
	}
}

func getLimits(resources api.CassandraResources) v1.ResourceList {
	return generateResourceList(resources.Limits.CPU, resources.Limits.Memory)
}

func getRequests(resources api.CassandraResources) v1.ResourceList {
	return generateResourceList(resources.Requests.CPU, resources.Requests.Memory)
}

func generateResourceList(cpu string, memory string) v1.ResourceList {
	resources := v1.ResourceList{}
	if cpu != "" {
		resources[v1.ResourceCPU], _ = resource.ParseQuantity(cpu)
	}
	if memory != "" {
		resources[v1.ResourceMemory], _ = resource.ParseQuantity(memory)
	}
	return resources
}

// createNodeAffinity creates NodeAffinity section for the statefulset.
// the selectors will be sorted byt the key name of the labels map before creating the statefulset
func createNodeAffinity(labels map[string]string) *v1.NodeAffinity {

	if len(labels) == 0 {
		return &v1.NodeAffinity{}
	}

	var nodeSelectors []v1.NodeSelectorRequirement

	//we make a new map in order to sort becaus a map is random by design
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys) //sort by key
	for _, key := range keys {
		selector := v1.NodeSelectorRequirement{
			Key:      key,
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{labels[key]},
		}
		nodeSelectors = append(nodeSelectors, selector)
	}

	return &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: nodeSelectors,
				},
			},
		},
	}
}

func createPodAntiAffinity(hard bool, labels map[string]string) *v1.PodAntiAffinity {
	if hard {
		// Return a HARD anti-affinity (no same pods on one node)
		return &v1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
				v1.PodAffinityTerm{
					TopologyKey: hostnameTopologyKey,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: labels,
					},
				},
			},
		}
	}

	// Return a SOFT anti-affinity
	return &v1.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
			v1.WeightedPodAffinityTerm{
				Weight: 100,
				PodAffinityTerm: v1.PodAffinityTerm{
					TopologyKey: hostnameTopologyKey,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: labels,
					},
				},
			},
		},
	}
}
