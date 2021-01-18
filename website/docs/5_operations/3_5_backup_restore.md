---
id: 3_5_backup_restore
title: Backup and restore
sidebar_label: Backup and restore
---
**Tip**: For a full working example step by step, please check also this [well written article](https://cscetbon.medium.com/casskop-1-0-1-backup-and-restore-ba92f01c00df). This also explain more deeply how Casskop Backup & Restore works in background

In order to provide Backup/Restore abilities we use InstaCluster's [cassandra-sidecar project](https://github.com/instaclustr/cassandra-sidecar) and add it to each Cassandra node to spawn. We want to thant Instaclustr for the modifications they made to make it work with CassKop!

## Backup

It is possible to backup keyspaces or tables from a cluster managed by Casskop. To start or schedule a backup, you 
create an object of type [CassandraBackup](/casskop/docs/6_references/5_cassandra_backup):

```yaml
apiVersion: db.orange.com/v1alpha1
kind: CassandraBackup
metadata:
  name: nightly-cassandra-backup
  labels:
    app: cassandra
spec:
  cassandraCluster: test-cluster
  datacenter: dc1
  storageLocation: s3://cassie
  snapshotTag: SnapshotTag2
  secret: cloud-backup-secrets
  schedule: "@midnight"
  entities: k1.t1,k2.t3
```

If there is no schedule defined, the backup will start as soon as it's created and won't be start again with that object.
You can always delete the object and recreate it though.

### Supported storage

The following storage options for storing the backups are:

- s3 (as in the example above)
- gcp
- azure
- oracle cloud

More details can be found on [Instaclustr's Cassandra backup page](https://github.com/instaclustr/cassandra-backup)

### Life cycle of the CassandraBackup object

When this object gets created, CassKop does a few checks to ensure:

- The specified Cassandra cluster exists
- If there is a secret that it has the expected parameters depending on the chosen backend
- If there is a schedule that its format is correct ([Cron expressions](https://godoc.org/gopkg.in/robfig/cron.v3#hdr-CRON_Expression_Format),
[Predefined schedules](https://godoc.org/gopkg.in/robfig/cron.v3#hdr-Predefined_schedules) or [Intervals](https://godoc.org/gopkg.in/robfig/cron.v3#hdr-Intervals))

Then, if all those checks pass, it triggers the backup if there is no schedule, or creates a Cron task with the specified schedule.

When this object gets deleted, if there is a scheduled task, it is unscheduled.

When this object gets updated, and the change is located in the spec section, CassKop unschedules the existing task and schedules a new one with the new parameters provided.

## Restore

Following the same logic, a [CassandraRestore](/casskop/docs/6_references/6_cassandra_restore) object must be created to trigger a restore, and it must refer to an
existing [CassandraBackup](/casskop/docs/6_references/5_cassandra_backup) object in K8S:

```yaml
apiVersion: db.orange.com/v1alpha1
kind: CassandraRestore
metadata:
  name: nightly-cassandra-backup
  labels:
    app: cassandra
spec:
  cassandraBackup: nightly-cassandra-backup
  cassandraCluster: test-cluster
  entities: k1.t1
```

### Entities

In the restore phase, you can specify a subset of the entities specified in the backup. For instance, you can backup 2
tables and only restore one.
