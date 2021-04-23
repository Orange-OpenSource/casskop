---
id: 1_cassandra_cluster
title: Cassandra cluster
sidebar_label: Cassandra cluster
---

`CassandraCluster` describes the desired state of the Cassandra cluster we want to setup through the operator.

```yaml
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-demo
  labels:
    cluster: k8s.kaas
spec:
  cassandraImage: cassandra:3.11
  bootstrapImage: orangeopensource/cassandra-bootstrap:0.1.4
  configMapName: cassandra-configmap-v1
  dataCapacity: "200Mi"
  dataStorageClass: "local-storage"
  imagepullpolicy: IfNotPresent  
  hardAntiAffinity: false           # Do we ensure only 1 cassandra on each node ?
  deletePVC: true
  autoPilot: false
  config:
    jvm-options:
      log_gc: "true"
  autoUpdateSeedList: false
  maxPodUnavailable: 1
  runAsUser: 999
  shareProcessNamespace: true
  resources:         
    requests:
      cpu: '1'
      memory: 2Gi
    limits:
      cpu: '1'
      memory: 2Gi
  topology:
    dc:
      - name: dc1
        nodesPerRacks: 1
        rack:
          - name: rack1
          - name: rack2
          - name: rack3
```

## CassandraCluster

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects users must create.|No|nil|
|spec|[CassandraClusterSpec](/casskop/docs/6_references/1_cassandra_cluster#cassandraclusterspec)|defines the desired state of CassandraCluster.|No|nil|
|status|[CassandraClusterStatus](/casskop/docs/6_references/3_cassandra_cluster_status#cassandraclusterstatus)|defines the observed state of CassandraCluster.|No|nil|

## CassandraClusterSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|nodesPerRacks|int32|Number of nodes to deploy for a Cassandra deployment in each Racks. If NodesPerRacks = 2 and there is 3 racks, the cluster will have 6 Cassandra Nodes|Yes|1|
|cassandraImage|string|Image + version to use for Cassandra|Yes|cassandra:3.11.6|
|imagepullpolicy|[PullPolicy](https://godoc.org/k8s.io/api/core/v1#PullPolicy)|Define the pull policy for C* docker image|Yes|[PullAlways](https://godoc.org/k8s.io/api/core/v1#PullPolicy)|
|bootstrapImage|string|Image used for bootstrapping cluster (use the form : base:version)|Yes|orangeopensource/cassandra-bootstrap:0.1.4|
|initContainerImage|string|Image used in the initContainer (use the form : base:version)|Yes|cassandra:3.11.6|
|initContainerCmd|string|Command to execute in the initContainer in the targeted image|Yes|cp -vr /etc/cassandra/* /bootstrap|
|runAsUser|int64|Define the id of the user to run in the Cassandra image|Yes|999|
|readOnlyRootFilesystem|Make the pod as Readonly|bool|Yes|true|
|resources|[Resources](#https://godoc.org/k8s.io/api/core/v1#ResourceRequirements)|Define the Requests & Limits resources spec of the "cassandra" container|Yes|-|
|hardAntiAffinity|bool|HardAntiAffinity defines if the PodAntiAffinity of the statefulset has to be hard (it's soft by default)|Yes|false|
|pod|[PodPolicy](#podpolicy)||No|-|
|service|[ServicePolicy](#servicepolicy)||No|-|
|deletePVC|bool|Defines if the PVC must be deleted when the cluster is deleted|Yes|false|
|debug|bool|Is used to surcharge Cassandra pod command to not directly start cassandra but starts an infinite wait to allow user to connect a bash into the pod to make some diagnoses.|Yes|false|
|shareProcessNamespace|bool|When process namespace sharing is enabled, processes in a container are visible to all other containers in that pod. [Check documentation for more informations](https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/)|Yes|false|
|autoPilot|bool|Defines if the Operator can fly alone or if we need human action to trigger actions on specific Cassandra nodes. [Check documentation for more informations](/casskop/docs/5_operations/2_pods_operations)|Yes|false|
|noCheckStsAreEqual|bool||Yes|false|
|autoUpdateSeedList|bool| Defines if the Operator automatically update the SeedList according to new cluster CRD topology|Yes|false|
|maxPodUnavailable|int32|Number of MaxPodUnavailable used in the [PodDisruptionBudget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/#specifying-a-poddisruptionbudget)|Yes|1|
|restartCountBeforePodDeletion|int32|defines the number of restart allowed for a cassandra container allowed before deleting the pod  to force its restart from scratch. if set to 0 or omit, no action will be performed based on restart count. [Check documentation for more informations](/casskop/docs/3_configuration_deployment/9_advanced_configuration#ip-cross-situation-detection)|Yes|0|
|unlockNextOperation|bool|Very special Flag to hack CassKop reconcile loop - use with really good Care|Yes|false|
|dataCapacity|string|Define the Capacity for Persistent Volume Claims in the local storage. [Check documentation for more informations](/casskop/docs/3_configuration_deployment/3_storage#configuration)|Yes||
|dataStorageClass|string|Define StorageClass for Persistent Volume Claims in the local storage. [Check documentation for more informations](/casskop/docs/3_configuration_deployment/3_storage#configuration)|Yes||
|storageConfigs|\[  \][StorageConfig](#storageconfig)|Defines additional storage configurations. [Check documentation for more informations](/casskop/docs/3_configuration_deployment/3_storage#additionnals-storages-configuration)|No| - |
|sidecarConfigs|\[  \][Container](https://godoc.org/k8s.io/api/core/v1#Container)|Defines additional sidecar configurations. [Check documentation for more informations](/casskop/docs/3_configuration_deployment/5_sidecars)|No| - |
|configMapName|string|Name of the ConfigMap for Cassandra configuration (cassandra.yaml). If this is empty, operator will uses default cassandra.yaml from the baseImage. If this is not empty, operator will uses the cassandra.yaml from the Configmap instead. [Check documentation for more informations](/casskop/docs/3_configuration_deployment/2_cassandra_configuration#configuration-override-using-configmap)|No| - |
|imagePullSecret|[LocalObjectReference](https://godoc.org/k8s.io/api/core/v1#LocalObjectReference)|Name of the secret to uses to authenticate on Docker registries. If this is empty, operator do nothing. If this is not empty, propagate the imagePullSecrets to the statefulsets|No| - |
|imageJolokiaSecret|[LocalObjectReference](https://godoc.org/k8s.io/api/core/v1#LocalObjectReference)|JMX Secret if Set is used to set JMX_USER and JMX_PASSWORD|No| - |
|topology|[Topology](/casskop/docs/6_references/2_topology#topology)|To create Cassandra DC and Racks and to target appropriate Kubernetes Nodes|Yes| - |
|livenessInitialDelaySeconds|int32|Defines initial delay for the liveness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|120|
|livenessHealthCheckTimeout|int32|Defines health check timeout for the liveness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|20|
|livenessHealthCheckPeriod|int32|Defines health check period for the liveness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|10|
|livenessFailureThreshold|int32|Defines failure threshold for the liveness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|(value set by kubernetes cluster)|
|livenessSuccessThreshold|int32|Defines success threshold for the liveness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|(value set by kubernetes cluster)|
|readinessInitialDelaySeconds|int32|Defines initial delay for the readiness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|60|
|readinessHealthCheckTimeout|int32|Defines health check timeout for the readiness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|10|
|readinessHealthCheckPeriod|int32|Defines health check period for the readiness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|10|
|readinessFailureThreshold|int32|Defines failure threshold for the readiness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|(value set by kubernetes cluster)|
|readinessSuccessThreshold|int32|Defines success threshold for the readiness probe of the main. [Configure liveness Readiness startup probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#configure-probes)|Yes|(value set by kubernetes cluster)|

## PodPolicy

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|annotations|map\[string\]string|Annotations specifies the annotations to attach to headless service the CassKop operator creates|No|-|
|tolerations|[Toleration](https://godoc.org/k8s.io/api/core/v1#Toleration)|Tolerations specifies the tolerations to attach to the pods the CassKop operator creates|No| - |

## ServicePolicy

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|annotations|map\[string\]string|Annotations specifies the annotations to attach to headless service the CassKop operator creates|No|-|

## StorageConfig

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|Name of the storage config, used to name PV to reuse into sidecars for example.|Yes| - |
|mountPath|string|Path where the volume will be mount into the main cassandra container inside the pod.|Yes| - |
|pvcSpec|[PersistentVolumeClaimSpec](https://godoc.org/k8s.io/api/core/v1#PersistentVolumeClaimSpec)|Kubernetes PVC spec. [create-a-persistentvolumeclaim](https://kubernetes.io/docs/tasks/configure-pod-container/configure-persistent-volume-storage/#create-a-persistentvolumeclaim).|Yes| - |

