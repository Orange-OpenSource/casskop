---
id: 3_5_backup_restore
title: Backup and restore
sidebar_label: Backup and restore
---

In order to provide Backup/Restore abilities we use InstaCluster's [cassandra-sidecar project](https://github.com/instaclustr/cassandra-sidecar) and add it to each Cassandra node to spawn. We want to thant Instaclustr for the modifications they made to make it work with CassKop!

## Backup

It is possible to backup keyspaces or tables from cluster managed by Casskop.

To start or schedule a backup, you create an object of type CassandraBackup:

```yaml
apiVersion: db.orange.com/v1alpha1
kind: CassandraBackup
metadata:
  name: nightly-cassandra-backup
  labels:
    app: cassandra
spec:
  cassandracluster: test-cluster
  datacenter: dc1
  storageLocation: s3://cassie
  snapshotTag: SnapshotTag2
  secret: cloud-backup-secrets
  schedule: "@midnight"
  entities: k1.t1,k2.t3
```

### Supported storage

The following storage options for storing the backups are:

- S3 compatible (as in the example above)
- ???

### Life cycle of the CassandraBackup object

When this object gets created, CassKop does a few checks to ensure :

- the specified Cassandra cluster exists
- if there is a secret that it has the expected parameters depending on the chosen backend
- if there is a schedule that its format is correct (crontab like)

Then, if all those checks pass, it triggers the backup if there is no schedule, or creates a Cron task with the specified schedule.

When this object gets deleted, if there is a scheduled task, it is unscheduled.

When this object gets updated, and the change is located in the spec section, CassKop unschedules the existing task and schedules a new one with the new parameters provided.

## Restore

Following the same logic, a CassandraRestore object must be created to trigger a restore:
A restore must refer to a Backup object in K8S

```yaml
apiVersion: db.orange.com/v1alpha1
kind: CassandraRestore
metadata:
  name: nightly-cassandra-backup
  labels:
    app: cassandra
spec:
  backup:
    name: nightly-cassandra-backup
  cluster: test-cluster
```
