media: ./assets
local: ./assets/kubernetes-operators
layout: true
gitlab: https://gitlab.si.francetelecom.fr/dfyarchicloud/slides-presentations-docker
editpath: /edit/master/Slides-integration-kubernetes-gitlab.md
<!-- /!\ pas de commentaires avant les parametres ci-dessus -->
<!-- /assets fait partie du viewer global a toutes les pres -->
<!-- /asset est le rep image en local de cette pres -->

.footnote[
    [
    ![edit]({{media}}/edit-50.png)]({{gitlab}}{{editpath}})
    ![orange]({{media}}/orange-50x50.png)
    ![dfy]({{media}}/dfy-50x53.png)
    ]
---
name: CassKop
class: center, middle

# CassKop

## Cassandra Kubernetes Operator

### Demo

<!--
![:scale 50%]({{local}}/CassandraOperator2.png)
-->


<br><br><br><br><br><br><br><br>
.right[Sébastien Allamand

Orange Digital Factory - 2019 - Version 1.0
]

---

# Agenda

- [Introduction](#3)
- [Demo](#11)
    - [CassKop deployment](#12)
    - [Deployment of a C* cluster (rack or AZ aware)](#14)
    - [Scaling up the cluster](#16)
    - [Pods Operations](#17)
    - [Scaling down the cluster](#18)
    - [Adding a Cassandra DC](#20)
    - [Removing a Cassandra DC](#23)
    - [Setting and modifying configuration files](#25)
    - [Setting and modifying configuration parameters](#27)
    - [Update of the Cassandra docker image](#28)
    - [Rolling restart of a cassandra rack](#29)
    - [Stopping a K8S node for maintenance](#30)
        - [Make a remove node ](#32)
        - [Make a replace address ](#34)
    - [Make reapairs with Cassandra Reaper](#36)
- [Where are we ?](#37)
- [Annexes](#39)

---

class: center, middle, big

# Introduction

---

# CassKop's CRD: cassandracluster

CassKop define its own Custom Ressource Definition named **cassandracluster**
This new k8s objects allow to describe the Cassandra cluster a user wants to manage.
We can interract directly with this new objects using kubectl.

### List, describe, create, delete  deployed clusters:
```console
$ kubectl get cassandracluster
NAME                     AGE
demo-cassandra-cluster   3d
```

```console
$ kubectl describe CassandraCluster demo-cassandra-cluster
Name:         demo-cassandra-cluster
Namespace:    cassandra
Labels:       app=cassandra-cluster
...
```

### CassKop also ships with a specific plugin to ease user interraction  with the managed cassandra cluster.

```console
$ kubectl casskop -h                                                                                                                                                                           10:28:22 
usage: kubectl-casskop <command> [<args>]

The available commands are:
   cleanup
   upgradesstables
   rebuild
   remove
```

---


# CassandraCluster CRD Manifest

```yaml
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-demo
  labels:
    cluster: demo
spec:
  nodesPerRacks: 2
  baseImage: orangeopensource/cassandra-image
  version: 3.11.4-8u212-0.3.1-cqlsh
  configMapName: cassandra-configmap-v1
  dataCapacity: "3Gi"
  imagepullpolicy: IfNotPresent
  checkStatefulsetsAreEqual: true
  hardAntiAffinity: true
  deletePVC: true
  autoPilot: false
  gcStdout: true
  autoUpdateSeedList: true
  maxPodUnavailable: 1
  resources:
    requests:
      cpu: '8'
      memory: 32Gi
    limits:
      cpu: '8'
      memory: 32Gi
```

> If no topology section is defined then CassKop create default dc1 / rack1

You can change your storageclass by adding the `spec.dataStorageClass:
"yourclass"` parameter to the CRD.


---


# CassKop Cassandra Cluster Topology

In order to ensure hight availability of datas it's important to store them on
nodes with different physical location.
With CassKop, it's all about mapping Cassandra's DC and Racks onto specific labels sets on k8s
nodes in CRD `spec.topology` section.

For example we could have different k8s node labelling model : 

.left-equal-column[
![:scale 100%]({{local}}/topology-custom-example.png)
]


.right-equal-column[
![:scale 100%]({{local}}/topology-gke-example.png)
]


---


# Cassandra DC/Rack Awareness

.left-equal-column[
<small>
According to the cassandra cluster manifest configuration, CassKop will create severals Kubernetes objects (statefulsets, services, volume claims, podDisruptionBudget...).

At this time, CassKop will create as many statefulsets as it is asked to deploy differents racks:
  - 1 Statefulset per Racks
  - 1 service per DC used by client to discover the cluster nodes.
  - the number of nodes per rack is configured per DC so we have equity between racks of a DC.
  - use of Label selector to associate cassandra nodes in a rack onto different physycal location of k8s nodes.
  - configurable Anti-affinity so that we get only one cassandra pod on each k8s node.
</small>
]


.right-equal-column[
![:scale 78%]({{local}}/cassandra-sts-uml.png)
]

---

# CassandraCluster CRD Topology

```yaml
...
  topology:
    dc:
      - name: dc1
        labels:
          failure-domain.beta.kubernetes.io/region: "europe-west1"        
        rack:
          - name: rack1
            labels:
              failure-domain.beta.kubernetes.io/zone: "europe-west1-b"            
          - name: rack2
            labels:
              failure-domain.beta.kubernetes.io/zone: "europe-west1-c"            
      - name: dc2
        nodesPerRacks: 2
        labels:
          failure-domain.beta.kubernetes.io/region: "europe-west1"
        rack:
          - name: rack1
            labels:
              failure-domain.beta.kubernetes.io/zone: "europe-west1-d"                          
```

The topology section allows to make association between cassandra pods and specific kubernetes nodes.
This allows to spread Cassandra pods for different DCs and Racks onto phisically different k8s nodes.

> If k8s nodes are correctly labelized and topology section filled this guarantee the high availability of datas


---

# Cassandra own the same topology

```console
$ k exec -ti <pod> nodetool status
Datacenter: dc1
==================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.72.243  358.5 GiB  32           53.0%             ee57995a-1bd1-4794-b375-0d15b8285d82  rack1
UN  172.18.68.71   318.22 GiB  32           47.0%             c61a92dc-a352-4d7d-b8e6-62183e8e15a1  rack1
UN  172.18.69.31   342.27 GiB  32           50.6%             16177fb4-a25c-4559-a466-71324d0d000a  rack2
UN  172.18.64.220  334.03 GiB  32           49.4%             5a4d659c-5870-44a6-9091-b89f955a5790  rack2
Datacenter: dc2
=================
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.67.17   581.45 GiB  32           43.6%             7f4d803c-e8b2-47c7-b620-43505645128f  rack1
UN  172.18.70.68   752.88 GiB  32           56.4%             c7f552f4-daa8-4a2d-a932-f5fdee05d043  rack1
UN  172.18.72.245  577.66 GiB  32           57.1%             affa060a-d875-4772-9ffa-7ed3396ef08b  rack2
UN  172.18.68.72   576.93 GiB  32           42.9%             0b0f0b8e-9a2a-4dca-b066-41c52abb537c  rack2
```

> Cassandra and nodetool works with IP while Kubernetes and CassKop work with names

```console
kubectl get pods -o wide
NAME                                     READY     STATUS        RESTARTS   AGE       IP              NODE      
cassandra-bench-dc1-rack1-0              1/1       Running       0          28d       172.18.68.71    node006   
cassandra-bench-dc1-rack1-1              1/1       Running       0          28d       172.18.72.243   node009   
cassandra-bench-dc1-rack2-0              1/1       Running       2          28d       172.18.69.31    node005   
cassandra-bench-dc1-rack2-1              1/1       Running       0          28d       172.18.64.220   node008   
cassandra-bench-dc2-rack1-0              1/1       Running       2          28d       172.18.70.68    node007   
cassandra-bench-dc2-rack1-1              1/1       Running       0          28d       172.18.67.17    node004   
cassandra-bench-dc2-rack2-0              1/1       Running       1          28d       172.18.68.72    node006   
cassandra-bench-dc2-rack2-1              1/1       Running       0          28d       172.18.72.245   node009
```


---

# CassandraCluster status management

CassKop manage a status section in the CRD objects:


```console
$ k get cassandracluster cassandra-demo -o yaml
...
status:
  cassandraRackStatus:
    dc1-rack1:
      cassandraLastAction:
        Name: Initializing
        endTime: 2018-10-09T09:48:40Z
        status: Done
      phase: Running
      podLastOperation:
        Name: cleanup
        operatorName: operator-goland
        pods:
        - demo-cassandra-cluster-dc1-rack1-0
        - demo-cassandra-cluster-dc1-rack1-0
        startTime: 2018-10-09T12:06:13Z
        status: Ongoing
    dc1-rack2:
      cassandraLastAction:
        Name: Initializing
        status: Ongoing
      phase: Initializing
      podLastOperation: {}
  lastClusterAction: Initializing
  lastClusterActionStatus: Done
  phase: Running
  seedlist:
  - demo-cassandra-cluster-dc1-rack1-0.demo-cassandra-cluster-dc1-rack1.cassandra.svc.cluster.local
```

---


class: center, middle, big

<!--
![:scale 80%](images/demo.png)
-->

# Demo

---

# Pre-requisite

## Helm

We are using helm version v2.12.2 you can install it from
https://github.com/helm/helm/releases/tag/v2.12.2

```console
$ helm init
```

> On GKE we need to increase RBAC rights for your user:
```console
kubectl create clusterrolebinding cluster-admin-binding \
--clusterrole cluster-admin --user [USER_ACCOUNT]
```
and for tiller:
```console
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'
```

---


# CassKop deployment

## Creation and configuration of kubernetes namespace

Create a namespace and swith on it.

```console
kubectl create namespace cassandra-demo
kubens cassandra-demo
```

## Optional: PSP

CassKop need specific capabilities in order to work : `IPC_LOCK` and the flag `allowPrivilegeEscalation: true`

If yout cluster is using PSP (Pod Security Policies) then you need to allow your
namespace with the use of thoses capabilities:

The PSP RoleBinding can be retrieved from  the cassandra namespace

```console
kubectl apply -f deploy/psp-cassie.yaml
kubectl apply -f deploy/psp:sa:cassie.yaml
kubectl apply -f deploy/clusterRole-cassie.yaml
```



---

## Deploy the CRD :

If the CRD don't already exists in the cluster we need to deploy it :

```console
k apply -f deploy/crds/db_v1alpha1_cassandracluster_crd.yaml
```

## CassKop's operator deployment :

```console
$ helm install --name cassandra-demo ./helm/cassandra-k8s-operator
```

check operator's logs: 

```console
$ k logs -l app=cassandra-k8s-operator --tail=10
```

---

# Deployment of a C* cluster (rack or AZ aware)

First, we deploy the configmap which contains the Cassandra configuration override files.
[samples/cassandra-configmap-v1.yaml](https://github.com/Orange-OpenSource/cassandra-k8s-operator/blob/master/samples/cassandra-configmap-v1.yaml).
Then the cassandra cluster :


Depending on your k8s cluster topology you may choose or adapt one of the
following configurations files to deploy a cluster with a 3 nodes ring in 3
different racks with anti-affinity 
- [samples/cassandracluster.yaml (with no labels)](https://github.com/Orange-OpenSource/cassandra-k8s-operator/blob/master/samples/cassandracluster.yaml)
- [samples/cassandracluster-demo.yaml (with basic labels)](https://github.com/Orange-OpenSource/cassandra-k8s-operator/blob/master/samples/cassandracluster-demo.yaml)
- [samples/cassandracluster-demo-gke.yaml (with europe-west gke labels)](https://github.com/Orange-OpenSource/cassandra-k8s-operator/blob/master/samples/cassandracluster-demo-gke.yaml)


> On GKE I deploy, using dedicated storagecless for ssd local-storage on
> regional cluster with 3 zones but CassKop can adapt to any configuration

```console
k patch storageclasses.storage.k8s.io standard —patch 'metadata:\n  annotations:\n     storageclass.beta.kubernetes.io/is-default-class: "false"'
kubectl apply -f gke-storage-standard-wait.yaml
or if you have ssd
kubectl apply -f gke-storage-ssd-wait.yaml
```

Example for démo on gke in europe-west (you can adapt the demo file according to
your cluster):
```console
kubectl apply -f samples/cassandra-configmap-v1.yaml
kubectl apply -f samples/cassandracluster-demo-gke.yaml
```



---

## CassKop's status

We can see the operator's status directly in it's kubernetes object representation

```console
kubectl get -o yaml cassandracluster
```


## Cassandra's status

```console
$ k exec -ti cassandra-demo-dc1-rack1-0 nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address       Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.72.46  108.59 KiB  256          64.8%             b8714011-430d-442f-8149-30af6ffa84dd  rack1
UN  172.18.70.28  107.89 KiB  256          66.3%             7c266c21-c463-4249-a092-af18df69baf1  rack3
UN  172.18.68.29  88.89 KiB  256          69.0%             e47a6333-4045-4631-a7ae-7e37887457c5  rack2
```

We have in Cassandra the same topology has asked by the manifest.


---


# Scaling up the cluster

We want to scale the cluster by specifying 2 nodesPerRacks instead of 1.
As we have 3 racks in our DC, we will add 3 cassandra nodes so a cassandra ring of 6.
In test environment we can directly edit the kubernetes object :

```console
$ k edit cassandraclusters.db.orange.com cassandra-demo
```

```console
  topology:
    dc:
      - name: dc1
        nodesPerRacks: 2
```

When the scale up is done on the Cassandra ring, we must apply a **cleanup** on all nodes in the ring.

- CassKop will label pods with [pod operations](#17), and can automatically triggered if `autoPilot: true`
```console
operation-name=cleanup
operation-status=ToDo
```
- Else it will label the Pod telling a human operator that it needs to launch the operation manually.
```console
operation-name=cleanup
operation-status=Manual
```

---

# Pods Operations

At any time, If there is no ongoing actions on the cassandra cluster, we can trigger operations on Pods.
Casskop watch Pod labels in order to detect if it is requested to execute some actions.

We also have a plugin to ease the management of operations like a cassandra cleanup. The plugin simply set some labels
on Pods, and CassKop will watch for thoses labels changed and will triggered associated operations.

We can check the evolution of labels on Pods while the operator execute thems.

```console
for x in `seq 1 3`; do
 echo cassandra-demo-dc1-rack$x;
 kubectl label pod cassandra-demo-dc1-rack$x-0 --list | grep operation ; echo ""
 kubectl label pod cassandra-demo-dc1-rack$x-1 --list | grep operation ; echo ""
done
```

We can start a cleanup on all nodes in dc1 with 

```console
$ kubectl casskop cleanup --prefix cassandra-demo-dc1
Namespace cassandra-demo
Trigger cleanup on pod cassandra-demo-dc1-rack1-0
Trigger cleanup on pod cassandra-demo-dc1-rack1-1
Trigger cleanup on pod cassandra-demo-dc1-rack2-0
Trigger cleanup on pod cassandra-demo-dc1-rack2-1
Trigger cleanup on pod cassandra-demo-dc1-rack3-0
Trigger cleanup on pod cassandra-demo-dc1-rack3-1
```

> the plugin can trigger lot of more operation we will see that in next steps

---

# Scaling down the cluster

Now we are going to make a ScaleDown (set nodesPerRacks=1), so the ring will go from 6 nodes to 3. One requirement of CassKop is to always
have the same number of cassandra nodes in each rack of a cassandra DC.


```console
$ k edit cassandraclusters.db.orange.com cassandra-demo
```
Instantally CassKop update it's status, and will manage 3 ScaleDown sequentially, 1 in each rack.

> Prior to scale down a node at k8s level CassKop will run a decommission of the node from the Cassandra ring.

```console
  status:
    cassandraRackStatus:
      dc1-rack1:
        cassandraLastAction:
          Name: ScaleDown
          startTime: 2019-02-28T19:20:51Z
          status: Ongoing
        phase: Pending
        podLastOperation:
          Name: decommission
          operatorName: cassandra-demo2-cassandra-k8s-operator-56d48f9d47-cbf5x
          pods:
          - cassandra-demo-dc1-rack1-1
          startTime: 2019-02-28T19:19:45Z
          status: Finalizing
...          
    lastClusterAction: ScaleDown
    lastClusterActionStatus: ToDo
```

---

# Add some Datas

## Insert some datas

We can uses cassandra-stress to insert some datas with RF=3 on both DCs:

```console
make cassandra-stress small
```

If you install prometheus, see in [Annexes](#41)

We can also view the Cassandra Metrics in the [Grafana Dashboard](http://localhost:8001/api/v1/namespaces/monitoring/services/http:prometheus-grafana:80/proxy/d/000000086/cassandra-dfy-by-pod?orgId=1)


---


# Adding a Cassandra DC

CassKop can add at any time a cassandra DC in an existing cluster.

Uncomment the 2d DC in the Topology section in the example manifest, and apply again the manifest

```console
$ k apply -f samples/cassandracluster-kaas-demo2.yaml
```

CassKop va will create the 2d DC and racks, statefulsets, services..

> The status of the CassandraCluster becoma again Initializing

## Check Cassandra Status

```console
$ k exec -ti cassandra-demo-dc1-rack1-0 nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address       Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.68.35  237.36 KiB  256          34.5%             e47a6333-4045-4631-a7ae-7e37887457c5  rack2
UN  172.18.70.33  231.8 KiB  256          34.1%             7c266c21-c463-4249-a092-af18df69baf1  rack3
UN  172.18.72.53  221.92 KiB  256          31.5%             b8714011-430d-442f-8149-30af6ffa84dd  rack1
Datacenter: dc2
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address       Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.18.67.64  103.87 KiB  256          0.0%              ef834816-9b08-49ca-90c2-4c60f85bbc8e  rack3
UN  172.18.69.39  142.05 KiB  256          0.0%              48747a0e-9d04-4303-bad0-7b5c30cba7d8  rack1
UN  172.18.71.61  103.63 KiB  256          0.0%              8de1b2b1-2395-486a-bee1-452e1fdfbb38  rack2
```

---

## Update of the SeedList

If `autoUpdateSeedList:true` CassKop will ensure that we always have 3 cassandra
nodes to be in the SeedList in each DCs. If the Topology evolves in a way that the SeedList is changed, then CassKop
will trigger a rolling update of the cluster to apply the new SeedList.

We can set this parameter to false to disable the compute of the SeedList by CassKop, then you can configure it 
manually in the status section of the CRD:

```console
    seedlist:
    - cassandra-demo-dc1-rack1-0.cassandra-demo-dc1-rack1.cassandra-demo
    - cassandra-demo-dc1-rack2-0.cassandra-demo-dc1-rack2.cassandra-demo
    - cassandra-demo-dc1-rack3-0.cassandra-demo-dc1-rack3.cassandra-demo
    - cassandra-demo-dc2-rack1-0.cassandra-demo-dc2-rack1.cassandra-demo
    - cassandra-demo-dc2-rack2-0.cassandra-demo-dc2-rack2.cassandra-demo
```

## Change Replication Factor

We add replication to the new DC

```console
k exec -ti cassandra-demo-dc1-rack1-0 -- cqlsh -u cassandra -p cassandra -e "
ALTER KEYSPACE bench WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3, 'dc2': 3}; 
ALTER KEYSPACE system_auth WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3, 'dc2': 3}; 
ALTER KEYSPACE system_distributed WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3, 'dc2': 3}; 
ALTER KEYSPACE system_traces WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3, 'dc2': 3}; 
"
```

---

## Rebuild new DC

```console
$ k casskop rebuild --prefix cassandra-demo-dc2 dc1
Namespace cassandra-demo
Trigger rebuild on pod cassandra-demo-dc2-rack1-0
Trigger rebuild on pod cassandra-demo-dc2-rack2-0
Trigger rebuild on pod cassandra-demo-dc2-rack3-0
```

CassKop will update the labels on pods while performing
```console
for x in `seq 1 3`; do
 echo cassandra-demo-dc2-rack$x; 
 kubectl label pod cassandra-demo-dc2-rack$x-0 --list | grep operation ; echo ""
done
```

CassKop will update it's status while performing the operation

```yaml
      dc2-rack1:
        cassandraLastAction:
          Name: Initializing
          endTime: 2019-02-28T20:10:14Z
          status: Done
        phase: Running
        podLastOperation:
          Name: rebuild
          endTime: 2019-02-28T20:19:25Z
          operatorName: cassandra-demo2-cassandra-k8s-operator-56d48f9d47-cbf5x
          podsOK:
          - cassandra-demo-dc2-rack1-0
          startTime: 2019-02-28T20:19:11Z
          status: Done
      dc2-rack2:
```

---

# Removing a Cassandra DC

## Protections

Pre-requisites:
- Prior to delete a DC we must ScaleDown to 0 it's nodes
- Prior to ScaleDown to 0, all Replication Factors of each keyspace must be updated to not have data replication to the DC

To delete a DC simply remove the lines in the topology section. Try without make a scaledown.

```console
$ k edit cassandraclusters.db.orange.com cassandra-demo
```

The cassandracluster status is :

```console
$ k get cassandraclusters.db.orange.com cassandra-demo -o yaml
...
    lastClusterAction: CorrectCRDConfig
    lastClusterActionStatus: Done
...
```

In CassKop's logs:

```
time="2019-02-28T20:37:08Z" level=warning msg="[cassandra-demo] The Operator has refused the Topology changed. You must scale down the dc dc2 to 0 before deleting the dc"
```

---

## ScaleDown to 0

In CassKop's logs:

```console
time="2019-02-28T20:39:51Z" level=info msg="[cassandra-demo]: Ask ScaleDown to 0 for dc dc2"
time="2019-02-28T20:39:52Z" level=warning msg="[cassandra-demo]The Operator has refused the ScaleDown. Keyspaces still
having data [system_distributed system_auth system_traces]"
```

We need to alter the keyspace to change the RF :

```console
k exec -ti cassandra-demo-dc1-rack1-0 -- cqlsh -u cassandra -p cassandra -e "
ALTER KEYSPACE system_distributed WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3}; 
ALTER KEYSPACE system_auth WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3}; 
ALTER KEYSPACE system_traces WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3};
ALTER KEYSPACE bench WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3};"
```

Once the replication has been disable on the dc2 we can restart the scale down to 0

## Removing the DC

Once the ScaleDown to 0 is Done, we can delete the DC

> if `spec.autoUpdateSeedList` is true, deleting the DC from the CRD will
> trigger a RollingUpdate of the cluster to apply the new SeedList

---


# Setting and modifying configuration files

By default CassKop will replace files from the docker Image `/etc/cassandra` with the files specified in the
configmap.

## Create a new ConfigMap

In this example, we create a new ConfigMap with the modification change we want to apply

```
compaction_throughput_mb_per_sec: 16
becomes
compaction_throughput_mb_per_sec: 0

and

# stream_throughput_outbound_megabits_per_sec: 200
becomes
stream_throughput_outbound_megabits_per_sec: 10
```

[samples/cassandra-configmap-v2.yaml](https://gitlab.si.francetelecom.fr/kubernetes/cassandra-k8s-operator/blob/master/samples/cassandra-configmap-v2.yaml)

```console
k apply -f samples/cassandra-configmap-v2.yaml
```

Now I need to update the cassandracluster with the new configMap name.

replace the line:
```
configMapName: cassandra-configmap-v1
becomes
configMapName: cassandra-configmap-v2
```

---

## We apply the Update
```console
$ k apply -f samples/cassandracluster-demo.yaml
```

Because we ask CassKop to change the configmap it starts a RollingUpdate of the whole cluster

> You can control the rolling restart of each nodes racks with the rollingPartition flag.


We can see the diff in the statefulset objects:

```yaml
     ConfigMap: {
      LocalObjectReference: {
-      Name: "cassandra-configmap-v1",
+      Name: "cassandra-configmap-v2",
      },
      Items: [
      ],
```

---

# Setting and modifying configuration parameters

We can uses the configmap in order to modify some parameters to the whole cluster or on dedicated pods.
This is done through the use of the `pre_run.sh` script which can be defined in the configmap.
This script is executed when the Pods start.

```yaml
k edit configmaps cassandra-configmap-v2
```

We ask to change a parameter on the pod cassandra-demo-dc1-rack2-0:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cassandra-jvm-configmap-full-v2
data:
  pre_run.sh: |-
    test "$(hostname)" == 'cassandra-demo-dc1-rack2-0' && echo "update param" && sed -i 's/windows_timer_interval: 1/windows_timer_interval: 2/' /etc/cassandra/cassandra.yaml
  cassandra.yaml: |
  ...
```

As we just have edited the configmap, and not apply a new configmap, there is no
rolling restart. We need to manually delete the targeted pod in order it can
restart with the new configuration

```console
$ k delete pod cassandra-demo-dc1-rack2-0
```
We can check that this parameter has been updated in the targeted pod:
```console
$ k exec -ti cassandra-demo-dc1-rack2-0 -- grep windows_timer /etc/cassandra/cassandra.yaml
windows_timer_interval: 2
```


---

# Update of the Cassandra docker image

The docker image version used for Cassandra is defined in the CRD by `version: 3.11.4-8u212-0.3.1-cqlsh`

Update this value with `version: latest-cqlsh` and apply again the manigest

CassKop start a rolling update of the entire cluster. 

> You can control the rolling upgrade on each racks by using the `rollingPartition` parameter

We can see the ongoing action in CassKop's status: 

```yaml
    lastClusterAction: UpdateDockerImage
    lastClusterActionStatus: Ongoing
```

---



# Rolling restart of a cassandra rack

We set the `rollingRestart: true` in the rack section we want to restart.

we can do this using the plugin (not yet implemented)

```console
$ k CassKop rollingrestart 
```

or by addind the key in the racks you want, example for rolling restart of the rack2: 

```yaml
...
        rack:
          - name: rack1
            labels:
              location.physical/rack : "1"
          - name: rack2
            rollingRestart: true
            labels:
              location.physical/rack : "2"
...
```

> Once the Operator start the rolling Restart, it will remove the parameter from the cassandracluster manifest.

> Note: When a Pod restart, it may (will) whange it's IP adress but will keep the same name. The Operator works only
> with name to avoid problem with IP changes

---

# Stopping a K8S node for maintenance

If a human operator needs to stop a k8s node for a maintenance, it will first ask Pods on the node to shutdown.
This is done using the `kubectl drain` command.

> Note: The PDB of the Cassandra Cluster will refuse the drain if there is already a disruption ongoing on the cluster

When a k8s node is done there will be several options for recovery:

- The stop is temporary and the node can be quickly comes back in the cluster: nothing to do - the Cassandra pod will
  reboot normally once the node is schedulable again.
- There is a problem in the node and it can't come back quickly into the cluster, We cant to move the pod on another
  node. We have 2 options:
    - Make a **RemoveNode** using CassKop: it will execute a removenode of the old Pod, then delete the PVC, finally
      k8s will reschedule a new Pod with the same name on another node.
    - Make a **ReplaceNode** using CassKop: it will configure the pre_run.sh script to boot the pod with a specific
      option to replace the old one. Then it delete the PVC so that the Pod can be recreate on another node.

> Depending on the storage class you are using there is some actions that will
> difere. For instance with local-storage each pod is associated with a dedicated k8s
> instance, and we need to delete the PVC to enable a pod to be re-scheduled on
> another node. That is not the case for GCP default storage.

---

## Cluster before stoping the node

View of the cluster before we make actions:

```console
NAME                         READY   STATUS      RESTARTS   AGE     IP             NODE  
cassandra-demo-dc1-rack1-0   1/1     Running     0          2d      172.31.182.7   node004
cassandra-demo-dc1-rack2-0   1/1     Running     0          2d      172.31.181.9   node005
cassandra-demo-dc1-rack3-0   1/1     Running     0          2d      172.31.183.9   node006
cassandra-demo-dc2-rack1-0   1/1     Running     0          2d      172.31.180.8   node007
cassandra-demo-dc2-rack2-0   1/1     Running     0          2d      172.31.179.23  node008
cassandra-demo-dc2-rack3-0   1/1     Running     0          2d      172.31.184.10  node009
```


```console
k exec -ti cassandra-demo-dc1-rack1-0 nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.31.181.9   428.1 GiB  256          63.2%             df0aa0a4-3f94-4598-90b0-00e3243bb239  rack2
UN  172.31.183.9   457.03 GiB  256          67.6%             3e697ed7-d06c-4851-bea8-bf10314a4f90  rack3
UN  172.31.182.7   407.92 GiB  256          69.2%             fa856a6e-5b26-45d9-8b5c-99ecea4229f6  rack1
Datacenter: dc2
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.31.184.10  453.43 GiB  256          67.2%             89d8cb59-4df7-49a2-9726-e9d2330a03ec  rack3
UN  172.31.180.8   408.07 GiB  256          67.0%             1f2880c0-a2d0-41f4-8b98-8cc50d079656  rack1
UN  172.31.179.23  411.47 GiB  256          65.8%             e96343ba-307b-4d64-aaf9-025c4bf2b6b4  rack2
```

---

# Make a remove node 

## We drain a node for maintenance

```console
$ k drain node008 --ignore-daemonsets
```

The Cassandra pod on this node becomes pending:

```console
NAME                        READY   STATUS      RESTARTS   AGE     IP              NODE   
cassandra-demo-dc2-rack2-0  0/1     Pending     0          20s     <none>          <none> 
```

## RemoveNode with pod name

We ask the plugin to remove the Cassandra node.

```console
k casskop remove --pod cassandra-demo-dc1-rack2-0 --from-pod cassandra-demo-dc1-rack1-0
```

> If there is no IP associated to the Pod, we need to provide the old IP of the Pod to the plugin.

```console
ERRO[1955] Operation failed cluster=cassandra-demo error="Can't find an IP assigned to pod cassandra-demo-dc2-rack2-0.
You need to provide its old IP to remove it from the cluster" operation=Remove pod=cassandra-demo-dc1-rack1-0 rack=dc1-rack1

```

---

## RemoveNode with pod name and IP

If the targeted pod don't have anymore it's IP adress we need to specify it in the command

```console
k casskop remove --pod cassandra-demo-dc1-rack2-0 --previous-ip 172.24.64.156 --from-pod cassandra-demo-dc1-rack1-0
```

The old node status change to LEAVING and it will stay while all range tokens of this node hasn't been streamed on
others replicas.

```console
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.31.181.9   428.1 GiB  256          63.2%             df0aa0a4-3f94-4598-90b0-00e3243bb239  rack2
UN  172.31.183.9   457.03 GiB  256          67.6%             3e697ed7-d06c-4851-bea8-bf10314a4f90  rack3
UN  172.31.182.7   409.8 GiB  256          69.2%             fa856a6e-5b26-45d9-8b5c-99ecea4229f6  rack1
Datacenter: dc2
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.31.184.10  453.82 GiB  256          67.2%             89d8cb59-4df7-49a2-9726-e9d2330a03ec  rack3
UN  172.31.180.8   409.32 GiB  256          67.0%             1f2880c0-a2d0-41f4-8b98-8cc50d079656  rack1
DL  172.31.179.23  411.47 GiB  256          65.8%             e96343ba-307b-4d64-aaf9-025c4bf2b6b4  rack2
```

> In some case the streaming phase can hang cf [this jira issue](https://issues.apache.org/jira/browse/CASSANDRA-6542)

Once the streaming is ended, the operator will delete the associated PVC and k8s will reschedule the pod on another
node. The new Pod will be in JOINING state and Cassandra will rebalance datas in the cluster.


---

# Make a replace address

The other option (quicker) is to create a new Cassandra Pod with the replace_adress option so that it will replace the
dead one.

```console
k get pods
...
cassandra-demo-dc2-rack1-0  0/1     Pending   0          12m
```

```console
k exec -ti cassandra-demo-dc1-rack1-0 nodetool status
...
Datacenter: dc2
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
UN  172.31.184.14  301.74 KiB  256          34.1%             2bc88841-05b1-4a54-a634-7ab55455bbff  rack3
UN  172.31.181.12  300.31 KiB  256          33.1%             54aab669-7ec0-48a2-83e1-53a6ba8d7682  rack2
DN  172.31.180.11  340.23 KiB  256          34.6%             7e72f678-a262-4dbc-9ca5-ac33699f38ee  rack1
```

We need to configure the **pre_run.sh** script from the ConfigMap and adding a
line with replace adresse instructions for our targeted pod. (change <pod_ip>)

Example:

```console
...
  pre_run.sh: |-
    test "$(hostname)" == 'cassandra-demo-dc2-rack1-0' && echo "-Dcassandra.replace_address_first_boot=<pod_ip>" >
    /etc/cassandra/jvm.options
...
```

---

## Applying the replace of the node

Once the ConfigMap edited, we can delete the PVC and the Pod so that k8s can rechedule on another available node:

```console
$ k delete pvc data-cassandra-demo-dc2-rack1-0
```

```console
$ k delete pod cassandra-demo-dc2-rack1-0
```

The Pod will boot and activate the flag **replace_address_first_boot**. The Pod will start streaming from Replicas so
that it can retrieve all datas from the previous dead Pod

> Note: The streaming time may take several hours depending on the volume of datas

---



# Make reapairs with Cassandra Reaper

We currently delegates the repairs to the Cassandra Reaper tool which must be
installed independantly :


```console
helm install --namespace cassandra-demo -n cassandra-reaper incubator/cassandra-reaper
```

Acess to the UI through k Proxy:

http://localhost:8001/api/v1/namespaces/cassandra-demo/services/cassandra-reaper:8080/proxy/webui/

Then we just to add the cluster by entering one of our seeds to start working with reaper.

---

# Where are we ?

## In progress

- Network encryption
- Monitoring (Prometheus/Grafana)
- Plugin
- Documentation
- Tests

# Still in roadmap

- Backup/restore
- Multi-regions

---

class: center, middle

# Thanks


![]({{local}}/minions.jpeg)

---

class: center, middle

# Annexes

---

# Deploy Prometheus Operator

We currently uses the prometheus Operator with a grafana dashboard to retrieve the metrics of our clusters

```console
$ kubectl create namespace monitoring
$ helm install --namespace monitoring --name prometheus stable/prometheus-operator
```

If you don't have ingress you can uses:


```console 
k proxy
```

- [grafana](http://localhost:8001/api/v1/namespaces/monitoring/services/http:prometheus-grafana:80/proxy/)
  (admin / prom-operator)
- [prometheus](http://localhost:8001/api/v1/namespaces/monitoring/services/prometheus-operated:9090/proxy/)
- [alertmanager](http://localhost:8001/api/v1/namespaces/monitoring/services/kube-prometheus-alertmanager:9093/proxy/)


## Creation os ServiceMonitor for C* Metrics

```console
k apply -f samples/prometheus-cassandra-service-monitor.yaml
```

## Add Grafana dashboard for Cassandra

You can import this [dashboard](https://github.com/Orange-OpenSource/cassandra-k8s-operator/blob/master/samples/prometheus-grafana-cassandra-dashboard.json) to retrieve metrics about your cassandra cluster.


---


# Availability tests



| Action                                              | Effect                                   |
|-----------------------------------------------------|------------------------------------------|
| 1.  Delete a pod                                    | k8s restart the pod                      |
| 2.  Connect in a pod and kill the cassandra process | k8s restart the pod                      |
| 3.  Drain un k8s node and kill a pod on this node   | the pod is pending it can't be scheduled |
| 3.1 Uncordon the node                               | k8s restart the pod                      |
| 3.2 We delete the associated PVC                    | k8s move the pod on another k8s node     |


> K8S node actions are reserved to the k8s administrators

> Check the impact of the PodDisruptionBudget on allowed operations.



## Cheat sheet

```console
$ k get pods -o wide            // Watch des pods avec ip et nodes
$ k delete pod                  // delete du pod
$ k exec -ti <pod> -- kill 0    // kill du cassandra dans le containeur
$ k drain node <node>           // the node can't be schedulable
$ k delete pvc <pvc>            // the pvc is the link between a Pod and the k8s node
$ k edit pdb <name>             // editer the pdb
$ k get events -w               // Check for kubernetes Events in the current namespace
```
