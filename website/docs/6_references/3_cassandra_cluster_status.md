---
id: 3_cassandra_cluster_status
title: Cassandra cluster Status
sidebar_label: Cassandra cluster Status
---

## CassandraClusterStatus

[Check documentation for more informations](/casskop/docs/3_tasks/2_configuration_deployment/11_cassandra_cluster_status)

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|phase|string| Indicates the state this Cassandra cluster jumps in. Phase goes as one way as below: Initial -> Running <-> updating.|Yes| - |
|lastClusterAction|string|Is the Last Action at the Cluster level|Yes| - |
|lastClusterActionStatus|string|Is the Last Action Status at the Cluster level|Yes|-|
|seedlist|\[ \]string|it is the Cassandra SEED List used in the Cluster.|Yes|-|
|cassandraNodeStatus|map\[string\][CassandraNodeStatus](#cassandranodestatus)|represents a map of (hostId, Ip Node) couple for each Pod in the Cluster.|Yes| - |
|cassandraRackStatus|map\[string\][CassandraRackStatus](#cassandrarackstatus)|represents a map of statuses for each of the Cassandra Racks in the Cluster|Yes|-|

## CassandraNodeStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|hostId|string|The cassandra node's hostId.|Yes| - |
|nodeIp|string|The cassandra node's ip|Yes| - |

## CassandraRackStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|phase|string| Indicates the state this Cassandra cluster jumps in. Phase goes as one way as below: Initial -> Running <-> updating.|Yes| - |
|cassandraLastAction|[CassandraLastAction](#cassandralastaction)| Is the set of Cassandra State & Actions: Active, Standby..|Yes| - |
|podLastOperation|[PodLastOperation](#podlastoperation)| manage status for Pod Operation (nodetool cleanup, upgradesstables..).|Yes| - |

## CassandraLastAction

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|status|string|Action is the specific actions that can be done on a Cassandra Cluster such as cleanup, upgradesstables..|Yes| - |
|name|string|Type of action to perform : UpdateVersion, UpdateBaseImage, UpdateConfigMap.. |Yes| - |
|startTime|[Time](https://godoc.org/github.com/ericchiang/k8s/apis/meta/v1#Time)| |Yes| - |
|endTime|[Time](https://godoc.org/github.com/ericchiang/k8s/apis/meta/v1#Time)| |Yes| - |
|updatedNodes|\[ \]string | PodNames of updated Cassandra nodes. Updated means the Cassandra container image version matches the spec's version.|Yes| - |

## PodLastOperation

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|status|string||Yes| - |
|name|string|Name of the Operation |Yes| - |
|startTime|[Time](https://godoc.org/github.com/ericchiang/k8s/apis/meta/v1#Time)| |Yes| - |
|endTime|[Time](https://godoc.org/github.com/ericchiang/k8s/apis/meta/v1#Time)| |Yes| - |
|pods|\[ \]string | List of pods running an operation|Yes| - |
|podsOK|\[ \]string | List of pods that run an operation successfully|Yes| - |
|podsKO|\[ \]string | List of pods that fail to run an operation|Yes| - |
|OperatorName|string |Name of operator |Yes| - |
