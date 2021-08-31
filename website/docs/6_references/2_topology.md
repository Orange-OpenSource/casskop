---
id: 2_topology
title: Topology
sidebar_label: Topology
---

## Topology

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|dc|\[ \][DC](#dc)|List of DC defined in the CassandraCluster.|Yes| - |

## DC

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|Name of the DC|Yes|dc1|
|labels|map\[string\]string|Labels used to target Kubernetes nodes|No||
|rack|\[ \][Rack](#rack)|List of Racks defined in the Cassandra DC|Yes|-|
|config|map|Configuration used by the config builder to generated cassandra.yaml and other configuration files|No||
|nodesPerRacks|int32|Number of nodes to deploy for a Cassandra deployment in each Racks.|Optional, if not filled, used value define in [CassandraClusterSpec](/casskop/docs/6_references/1_cassandra_cluster#cassandraclusterspec)|1|
|dataCapacity|string|Define the Capacity for Persistent Volume Claims in the local storage. [Check documentation for more informations](/casskop/docs/3_configuration_deployment/3_storage#configuration)|Optional, if not filled, used value define in [CassandraClusterSpec](/casskop/docs/6_references/1_cassandra_cluster#cassandraclusterspec)||
|dataStorageClass|string|Define StorageClass for Persistent Volume Claims in the local storage. [Check documentation for more informations](/casskop/docs/3_configuration_deployment/3_storage#configuration)|Optional, if not filled, used value define in [CassandraClusterSpec](/casskop/docs/6_references/1_cassandra_cluster#cassandraclusterspec)||

## Rack

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|name|string|Name of the Rack|Yes|rack1|
|labels|map\[string\]string|Labels used to target Kubernetes nodes|No|-|
|config|map|Configuration used by the config builder to generated cassandra.yaml and other configuration files|No||
|rollingRestart|bool|Flag to tell the operator to trigger a rolling restart of the Rack|Yes|false|
|rollingPartition|int32|The Partition to control the Statefulset Upgrade|Yes|0|