---
id: 11_cassandra_cluster_status
title: CassandraCluster Status
sidebar_label: CassandraCluster Status
---

You can request kubernetes Object `cassandracluster` representing the Cassandra cluster to retrieve information about
it's status :

```
$ kubectl describe cassandracluster cassandra
...
status:
   Cassandra Node Status:
     cassandra-demo-dc1-rack1-0:
      Host Id:  ca716bef-dc68-427d-be27-b4eeede1e072
      Node Ip:  10.100.150.51
     cassandra-demo-dc1-rack1-1:
      Host Id:  3528d662-e4a8-4fb6-88f6-3f21056df7ea
      Node Ip:  10.100.150.39
     cassandra-demo-dc1-rack1-2:
      Host Id:  a1d1e7fa-8073-408c-94c1-e3678013f90f
      Node Ip:  10.100.150.38
     cassandra-demo-dc1-rack2-0:
      Host Id:  83ea3410-db00-47fe-9051-e9f877ce5e63
      Node Ip:  10.100.150.111
     cassandra-demo-dc1-rack2-1:
      Host Id:  200bf115-5caf-4218-8e84-e804296c5026
      Node Ip:  10.100.150.108
     cassandra-demo-dc1-rack2-2:
      Host Id:  27ee7414-a695-4744-bf39-41db9d23ddb2
      Node Ip:  10.100.150.110
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: ScaleUp
        endTime: 2018-07-12T14:10:28Z
        startTime: 2018-07-12T14:09:34Z
        status: Done
      phase: Running
      podLastOperation:
        Name: cleanup
        endTime: 2018-07-12T14:07:35Z
        podsOK:
        - cassandra-demo-dc1-rack1-0
        - cassandra-demo-dc1-rack1-1
        - cassandra-demo-dc1-rack1-2
        startTime: 2018-07-12T14:06:22Z
        status: Done
    dc1-rack2:
      cassandraLastAction:
        Name: ScaleUp
        endTime: 2018-07-12T14:10:58Z
        startTime: 2018-07-12T14:10:28Z
        status: Done
      phase: Running
      podLastOperation:
        Name: cleanup
        endTime: 2018-07-12T14:08:16Z
        podsOK:
        - cassandra-demo-dc1-rack2-0
        - cassandra-demo-dc1-rack2-1
        - cassandra-demo-dc1-rack2-2
        startTime: 2018-07-12T14:08:09Z
        status: Done
  lastClusterAction: ScaleUp
  lastClusterActionStatus: Done        
...
  phase: Running
  seedlist:
  - cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.cassandra-demo.svc.kaas-prod-priv-sph
```

The CassandraCluster prints out it's whole status.

- **seedlist**: it is the Cassandra SEED List used in the Cluster.
- **Phase** : it's the global state for the cassandra cluster which can have different values :
    - **Initialization**, we just launched a new cluster, and waiting for its requested state
    - **Running**, the cluster is running normally
    - **Pending**, the number of Nodes requested has changed, waiting for reconciliation
- **lastClusterAction** Is the Last Action at the Cluster level
- **lastClusterActionStatus** Is the Last Action Status at the Cluster level
- **CassandraNodeStatus**: represents a map of (hostId, Ip Node) couple for each Pod in the Cluster
  - **${Cassandra node pod's name}**
    - **HostId**: the cassandra node's hostId
    - **IpNode**: the cassandra node's ip
- **CassandraRackStatus** represents a map of statuses for each of the Cassandra Racks in the Cluster
  - **${Cassandra DC-Rack Name}**
    - **Cassandra Last Action**: it's an action which is ongoing on the Cassandra cluster :
        - **Name**: name of the Action
            - **UpdateConfigMap** a new ConfigMap has been submitted to the cluster
            - **UpdateDockerImage** a new Docker Image has been submitted to the cluster
            - **UpdateSeedList** a new SeedList must be deployed on the cluster
            - **UpdateResources** CassKop must apply new resources values for it's statefulsets            
            - **RollingRestart** CassKop performs a rollingrestart on the target statefulset
            - **ScaleUp** a scale Up has been requested
            - **ScaleDown** a scale Down has been requested.
            - **UpdateStatefulset** a change has been submitted to the statefulset, but CassKop doesn't know exactly
              which one.              
        - **Status**: status of the Action
            - **Configuring**: Only used for UpdateSeedList, we need to synchronise all statefulset with this operation before starting it
            - **ToDo**: an action is scheduled
            - **Ongoing**: an action is ongoing, see Start Time
            - **Continue**: the action may be continuing (used for ScaleDown)
            - **Done**: the action is Done, see End Time
        - **Start Time**: time of start of the operation
        - **End Time**: time of end of the operation
    - **Pod Last Operation**: it's an operation done at Pod Level
        - **Name**: Name of the Operation
            - **decommissioning**: a nodetool decommissioning must be performed on a pod
            - **cleanup**: a nodetool cleanup must be performed on a pod
            - **rebuild**: a nodetool rebuild must be performed on a pod
            - **upgradesstables**: a nodetool upgradesstables must be performed on a pod            
        - **Status**:
            - **Manual**: an operation is recommended to be scheduled by a human
            - **ToDo**: an operation is scheduled    
            - **Ongoing**: an operation is ongoing, see start time
            - **Done**: an operation is done, see end time
        - **Pods**: list of Pods on which the operation is ongoing
        - **PodsOK**: list of Pods on which the operation is done
        - **PodsKO**: list of Pods on which the operation has not been completed correctly
        - **Start Time**: time of start for an operation
        - **End Time**: time of end for an operation        
  
> When Status=Done for each Rack, then there is no specific action ongoing on the cluster and the
> lastClusterActionStatus will turn also to Done.