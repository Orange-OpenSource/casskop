---
id: 9_advanced_configuration
title: Advanced Configuration
sidebar_label: Advanced Configuration
---

## Docker login for private registry

If you need to use a docker registry with authentication, then you will need to create a specific kubernetes secret with
this information.
Then you will configure the CRD with the secret name, so that it provides the data to each Statefulset, which in
turn propagate it to each created Pod.

Create the secret :

```
kubectl create secret docker-registry yoursecretname \
  --docker-server=yourdockerregistry
  --docker-username=yourlogin \
  --docker-password=yourpass \
  --docker-email=yourloginemail
```

Then we will add a **imagePullSecrets** parameter in the CRD definition with value the name of the 
previously created secret. You can give several secrets :

```
imagePullSecrets:
  name: yoursecretname
```


## Management of allowed Cassandra nodes disruption

CassKop makes use of the kubernetes PodDisruptionBudget objetc to specify how many cassandra nodes disruption is
allowed on the cluster. By default, we only tolerate 1 disrupted pod at a time and will prevent to makes actions if
there is aloready an ongling disruption on the cluster.

In some edge cases it can be useful to make force the operator to continue it's actions even if there is already a
disruption ongoing. We can tune this by updating the `spec.maxPodUnavailable` parameter of the cassandracluster CRD.

:::important
it is recommended to not touch this parameter unless you know what you are doing.
:::

## Cross Ip Management

### Global mecanism

Cassandra works on IPs and not on hostname, so in a case where two cassandra cross their Ips, no Cassandra will be able to run properly and will loop on the following error :

```log
cassandra Exception (java.lang.RuntimeException) encountered during startup: A node with address /10.100.150.35 already exists, cancelling join. Use cassandra.replace_address if you want to replace this node.
cassandra java.lang.RuntimeException: A node with address /10.100.150.35 already exists, cancelling join. Use cassandra.replace_address if you want to replace this node.
cassandra     at org.apache.cassandra.service.StorageService.checkForEndpointCollision(StorageService.java:577)
cassandra     at org.apache.cassandra.service.StorageService.prepareToJoin(StorageService.java:823)
cassandra     at org.apache.cassandra.service.StorageService.initServer(StorageService.java:683)
cassandra     at org.apache.cassandra.service.StorageService.initServer(StorageService.java:632)
cassandra     at org.apache.cassandra.service.CassandraDaemon.setup(CassandraDaemon.java:388)
cassandra     at org.apache.cassandra.service.CassandraDaemon.activate(CassandraDaemon.java:620)
cassandra     at org.apache.cassandra.service.CassandraDaemon.main(CassandraDaemon.java:732)
cassandra ERROR [main] 2020-02-21 08:29:44,398 CassandraDaemon.java:749 - Exception encountered during startup
cassandra java.lang.RuntimeException: A node with address /10.100.150.35 already exists, cancelling join. Use cassandra.replace_address if you want to replace this node.
cassandra     at org.apache.cassandra.service.StorageService.checkForEndpointCollision(StorageService.java:577) ~[apache-cassandra-3.11.4.jar:3.11.4]
cassandra     at org.apache.cassandra.service.StorageService.prepareToJoin(StorageService.java:823) ~[apache-cassandra-3.11.4.jar:3.11.4]
cassandra     at org.apache.cassandra.service.StorageService.initServer(StorageService.java:683) ~[apache-cassandra-3.11.4.jar:3.11.4]
cassandra     at org.apache.cassandra.service.StorageService.initServer(StorageService.java:632) ~[apache-cassandra-3.11.4.jar:3.11.4]
cassandra     at org.apache.cassandra.service.CassandraDaemon.setup(CassandraDaemon.java:388) [apache-cassandra-3.11.4.jar:3.11.4]
cassandra     at org.apache.cassandra.service.CassandraDaemon.activate(CassandraDaemon.java:620) [apache-cassandra-3.11.4.jar:3.11.4]
cassandra     at org.apache.cassandra.service.CassandraDaemon.main(CassandraDaemon.java:732) [apache-cassandra-3.11.4.jar:3.11.4]
```

Following the [issue #170](https://github.com/Orange-OpenSource/casskop/issues/170), at least using Kubernetes and [Project Calico](https://docs.projectcalico.org/v3.9/getting-started/kubernetes/), we may fall into this issue, 
for example using a fixed [ip pool](https://docs.projectcalico.org/v3.9/reference/resources/ippool) size.

To manage this case we introduced the `restartCountBeforePodDeletion` CassandraCluster spec field which takes an `int32` as value.

:::note
If you set it with a value lower or equals to 0, or if you omit it, no action will be performed
:::

In setting this field, the cassandra operator will check for each `CassandraCluster` if a pod is in a restart situation (based on restart count of the cassandra container inside the pod).
In the case where the restartCount of the pod is greater than the value of `restartCountBeforePodDeletion` field and if we are in a Ip cross situation, we will delete the pod, which will be recreated by the Statefulset. 
In the case of Project Calico usage, this force the pod to get another available IP, which fixes our bug. 

## Ip Cross situation detection

To detect that we are in a Ip cross situation, we add a new status field `CassandraNodeStatus` which will maintain a cache about the map of *Ip node* and his *hostId*, 
for all ready pods.

:::note
to have more information about this status field, you can check [CassandraCluster Status](#cassandracluster-status)
:::

So when we check pods, we perform a Jolokia call to get a map of the cluster nodes IPs with their corresponding HostId.
If a pod is failing with the constraints described above, we compare the hostId associated to the Pod's IP, and the hostId
associated to the Pod name stored into the `CassandraNodeStatus` : 

- if they match, or there are no match for the pod Ip into the map returned by Jolokia, we are not in a Ip cross situation,
- if they mismatch, we are in a Ip cross situation.