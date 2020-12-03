---
id: 5_cassandra_backup
title: Cassandra backup
sidebar_label: Cassandra backup
---

```yaml
apiVersion: db.orange.com/v1alpha1
kind: CassandraBackup
metadata:
  name: backup-demo-scheduled
spec:
  cassandraCluster: cluster-demo
  datacenter: dc1
  storageLocation: s3://cscetbon-lab
  secret: aws-backup-secrets
  entities: k1.standard1
  snapshotTag: second
# I don't really expect you to run backups every minute ;)
  schedule: "@every 1m"
```

## CassandraBackup

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects users must create.|No|-||
|spec|[CassandraBackupSpec](#cassandrabackupspec)|defines the desired state of CassandraBackup.|No|nil|
|status|[CassandraBackupStatus](#cassandrabackupstatus)|defines the observed state of CassandraBackup.|No|nil|

## CassandraBackupSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|bandwidth|string|Specify the bandwidth to not exceed when uploading files to the cloud. Format supported is \d+[KMG] case insensitive. You can use values like 10M (meaning 10MB), 1024, 1024K, 2G, etc...|no|-|
|cassandraCluster|string|Name of the CassandraCluster to backup|Yes|-|
|concurrentConnections|int32|Maximum number of threads used to download files from the cloud. Defaults to 10|No|-|
|datacenter|string|Cassandra DC name to back up, used to find the cassandra nodes in the CassandraCluster|Yes|-|
|duration|string|Specify a duration the backup should try to last. See https://golang.org/pkg/time/#ParseDuration for an exhaustive list of the supported units. You can use values like .25h, 15m, 900s all meaning 15 minutes|No|-|
|entities|string|Database entities to backup, it might be either only keyspaces or only tables prefixed by their respective keyspace, e.g. 'k1,k2' if one wants to backup whole keyspaces or 'ks1.t1,ks2.t2' if one wants to restore specific tables. These formats are mutually exclusive so 'k1,k2.t2' is invalid. An empty field will backup all keyspaces|No|-|
|schedule|string|Specify a schedule to assigned to the backup. The schedule doesn't enforce anything so if you schedule multiple backups around the same time they would conflict. See https://godoc.org/github.com/robfig/cron for more information regarding the supported formats|No|-|
|secret|string|Name of Secret to use when accessing cloud storage providers|No|-|
|snapshotTag|string|name of snapshot to make so this snapshot will be uploaded to storage location. If not specified, the name of snapshot will be automatically generated and it will have name 'autosnap-milliseconds-since-epoch'|Yes|-|
|storageLocation|string|URI for the backup target location e.g. s3 bucket, filepath|Yes|-|

## CassandraBackupStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|condition|[Condition](#condition)|BackRestCondition describes the observed state of a Restore at a certain point|Yes|-|
|coordinatorMember|string|Name of the pod the restore operation is executed on|Yes|-|
|id|string|unique identifier of an operation, a random id is assigned to each operation after a request is submitted, from caller's perspective, an id is sent back as a response to his request so he can further query state of that operation, referencing id, by operations/{id} endpoint|Yes|-|
|progress|string|Progress is a percentage, 100% means the operation is completed, either successfully or with errors|Yes|-|
|timeCompleted|string| |Yes|-|
|timeCreated|string| |Yes|-|
|timeStarted|string| |Yes|-|

### Condition

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|failureCause|[[]items](#items)| |Yes|nil
|lastTransitionTime|string| |Yes|-
|type|string| |Yes|-


### Items

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|message|string|message explaining the error|Yes|-|
|source|string|hostame of a node where this error has occurred|Yes|-|
