---
id: 6_cassandra_restore
title: Cassandra restore
sidebar_label: Cassandra restore
---

`CassandraRestoreSpec` defines the specification for a restore of a Cassandra backup.

```yaml
apiVersion: db.orange.com/v2
kind: CassandraRestore
metadata:
  name: restore-demo
spec:
  cassandraCluster: cluster-demo
  cassandraBackup: backup-demo
  entities: k1.standard1
```

## CassandraRestore

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects users must create.|No|nil|
|spec|[CassandraRestoreSpec](#cassandrarestorespec)|defines the desired state of CassandraRestore.|No|nil|
|status|[CassandraRestoreStatus](#cassandrarestorestatus)|defines the observed state of CassandraRestore.|No|nil|

## CassandraRestoreSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|cassandraBackup|string|Name of the [CassandraBackup](/casskop/docs/6_references/5_cassandra_backup) to restore|Yes|-|
|cassandraCluster|string|Name of the CassandraCluster the restore belongs to|Yes|-|
|cassandraDirectory|string|Directory of Cassandra where data folder resides. Defaults to /var/lib/cassandra|No|-|
|datacenter|string|Cassandra DC name to restore to, a restore will truncate tables but restore only to this datacenter if specified|No|-|
|concurrentConnection|int32|Maximum number of threads used to download files from the cloud. Defaults to 10|No|-|
|entities|string|Database entities to restore, it might be either only keyspaces or only tables prefixed by their respective keyspace, e.g. 'k1,k2' if one wants to backup whole keyspaces or 'ks1.t1,ks2.t2' if one wants to restore specific tables. These formats are mutually exclusive so 'k1,k2.t2' is invalid. An empty field will restore all keyspaces|No|-|
|exactSchemaVersion|boolean|When set a running node's schema version must match the snapshot's schema version. There might be cases when we want to restore a table for which its CQL schema has not changed but it has changed for other table / keyspace but a schema for that node has changed by doing that. Defaults to False|No|false|
|noDeleteTruncates|boolean|When set do not delete truncated SSTables after they've been restored during CLEANUP phase. Defaults to false|No|false|
|schemaVersion|string|Version of the schema to restore from. Upon backup, a schema version is automatically appended to a snapshot name and its manifest is uploaded under that name. In case we have two snapshots having same name, we might distinguish between the two of them by using the schema version. If schema version is not specified, we expect a unique backup taken with respective snapshot name. This schema version has to match the version of a Cassandra node we are doing restore for (hence, by proxy, when global request mode is used, all nodes have to be on exact same schema version). Defaults to False|No|-|
|secret|string|Name of Secret to use when accessing cloud storage providers|No|-|

## CassandraRestoreStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|condition|[Condition](#condition)|BackRestCondition describes the observed state of a Restore at a certain point|No|-|
|coordinatorMember|string|Name of the pod the restore operation is executed on|No|-|
|id|string|unique identifier of an operation, a random id is assigned to each operation after a request is submitted, from caller's perspective, an id is sent back as a response to his request so he can further query state of that operation, referencing id, by operations/{id} endpoint|No|-|
|progress|string|Progress is a percentage, 100% means the operation is completed, either successfully or with errors|No|-|
|timeCompleted|string| |No|-|
|timeCreated|string| |No|-|
|timeStarted|string| |No|-|

### Condition

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|failureCause|[[]items](#items)| |No|nil
|lastTransitionTime|string| |No|-
|type|string| |No|-


### Items

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|message|string|message explaining the error|No|-|
|source|string|hostame of a node where this error has occurred|No|-|
