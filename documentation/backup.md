# Backup or restore your data

In order to provide Backup/Restore abilities we use InstaCluster's [cassandra-sidecar project]
(https://github.com/instaclustr/cassandra-sidecar) and add it to each Cassandra node to spawn.

We have a designated controller called controller_cassandrabackup that is in charge of managing backups and interacts with the corresponding cassandra sidecars. Depending on the operation you wanna do you create a specific kubernetes object. We'll probably update our kubectl plugin to provide some of those features, access the list of backups etc..

## Backup operation

In order to backup, you create an object of type CassandraBackup

```yaml
apiVersion: db.orange.com/v2
kind: CassandraBackup
metadata:
  name: test-cassandra-backup
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

### Life cycle of the CassandraBackup object

When this object gets created, CassKop does a few checks to ensure :
- the specified Cassandra cluster exists
- if there is a secret that it has the expected parameters depending on the chosen backend
- if there is a schedule that its format is correct

Then, if all those checks pass, it triggers the backup if there is no schedule, or creates a Cron task with the specified schedule. 

When this object gets deleted, if there is a scheduled task, it is unscheduled.

When this object gets updated, and the change is located in the spec section, CassKop unschedules the existing task and schedules a new one with the new parameters provided.


