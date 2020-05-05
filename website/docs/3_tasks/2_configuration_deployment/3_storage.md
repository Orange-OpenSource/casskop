---
id: 3_storage
title: Storage
sidebar_label: Storage
---

## Configuration

Cassandra is a stateful application. It needs to store data on disks. CassKop allows you to configure the type of
storage you want to use.

Storage can be configured using the `storage` property in `CassandraCluster.spec` for global Data Centers configuration, or can be overrided at `CassandraCluster.spec.topology.dc` level. 

:::important
Once the Cassandra cluster is deployed, the storage cannot be changed.
:::

Persistent storage uses Persistent Volume Claims to provision persistent volumes for storing data.
The `PersistentVolumes` are acquired using a `PersistentVolumeClaim` which is managed by CassKop. The
`PersistentVolumeClaim` can use a `StorageClass` to trigger automatic volume provisioning.

> It is recommended to use local-storage with quick ssd disk access for low latency. We have only tested the
> `local-storage` storage class within CassKop.

CassandraCluster fragment of persistent storage definition :

```
...
  # Global configuration
  dataCapacity: "300Gi"
  dataStorageClass: "local-storage"
  deletePVC: true
  ...
  topology:
     dc:
       - name: dc1
         # DC level configuration
         dataCapacity: "10Gi"
         dataStorageClass: "test-storage"
         ...
       - name: dc2
         ...
  ...
...

```

- `dataCapacity` (required): Defines the size of the persistent volume claim, for example, "1000Gi".
- `dataStorageClass`(optional): Define the type of storage to use (or use
  default one). We recommand to use local-storage for better performances but
  it can be any storage with high ssd througput.
- `deletePVC`(optional): Boolean value which specifies if the Persistent Volume Claim has to be deleted when the cluster
  is deleted. Default is `false`.
  
In this example, all statefulsets related to the `dc2` will have the default configuration for the `data` PV :

- `dataCapacity` : 300Gi
- `dataStorageClass`: local-storage

All statefulsets related to the `dc1` will have the specific configuration for the `data` PV :

- `dataCapacity` : 10Gi
- `dataStorageClass` : test-storage

:::warning
Resizing persistent storage for existing CassandraCluster is not currently supported. You must decide the
necessary storage size before deploying the cluster.
:::

The above example asks that each node will have 300Gi of data volumes to persist the Cassandra data's using the
local-storage storage class provider.
The parameter deletePVC is used to control if the data storage must persist when the according statefulset is deleted.

:::warning
If we don't specify dataCapacity, then CassKop will use the Docker Container ephemeral storage, and
all data will be lost in case of a cassandra node reboot.
:::

## Persistent volume claim

When the persistent storage is used, it will create PersistentVolumeClaims with the following names:

`data-<cluster-name>-<dc-name>-<rack-name>-<idx>`

Persistent Volume Claim for the volume used for storing data to the cluster `<cluster-name>` for the Cassandra DC
`<dc-name>` and the rack `<rack-name>` for the Pod with ID `<idx>`.

:::important
Note that with local-storage the PVC object makes a link between the Pod and the Node. While this
object is existing the Pod will be sticked to the node chosen by the scheduler. In the case you want to move the
Cassandra node to a new kubernetes node, you will need at some point to manually delete the associate PVC so that the
scheduler can choose another Node for scheduling. This is cover in the Operation document.
:::

## Additionnals storages configuration

For extra needed not covered by the defaults volumes managed through the CassandraCluster CR, we are allowing you to define your own storage configurations.
To do this, you will configure the `storageConfigs` property in `CassandraCluster.Spec`.

CassandraCluster fragment for dynamic persistent storage definition : 

```yaml
# ...
     storageConfigs:
        - mountPath: "/var/lib/cassandra/log"
          name: "gc-logs"
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            storageClassName: local-storage
            resources:
              requests:
                storage: 5Gi
        - mountPath: "/var/log/cassandra"
          name: "cassandra-logs"
          pvcSpec:
            accessModes:
              - ReadWriteOnce
            storageClassName: local-storage
            resources:
              requests:
                storage: 10Gi
# ...
```

- `storageConfigs` *(required)* : Defines the list of storage config object, which will instantiate `Persitence Volume Claim` and associate volume to pod of cassandra node.
    - `mountPath` *(required)* : Defines the path into `cassandra container` where the volume will be mounted.
    - `name` *(required)* : Used to define the `PVC` and `VolumeMount` names.
    - `pvcSpec` *(required)* : pvcSpec describes the PVC used for the mountPath described above it requires a kubernetes PVC spec.
    
With the above configuration, the following configuration will be added to the `rack statefulset` definition : 

```yaml
# ...
  volumeMounts:
  #...
  - mountPath: /var/lib/cassandra/log
    name: gc-logs
  - mountPath: /var/log/cassandra
    name: cassandra-logs
  #...
# ...
  volumeClaimTemplates:
  #...
  - metadata:
      name: gc-logs
      labels:
        app: cassandracluster
        cassandracluster: cassandra-demo
        cassandraclusters.db.orange.com.dc: dcsts
        cassandraclusters.db.orange.com.rack: rack1
        cluster: casskop
        dc-rack: dcsts-rack1
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 5Gi
      storageClassName: local-storage
      volumeMode: Filesystem
  - metadata:
      name: cassandra-logs
      labels:
        app: cassandracluster
        cassandracluster: cassandra-demo
        cassandraclusters.db.orange.com.dc: dcsts
        cassandraclusters.db.orange.com.rack: rack1
        cluster: casskop
        dc-rack: dcsts-rack1
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 5Gi
      storageClassName: local-storage
      volumeMode: Filesystem
  #...
# ...
```