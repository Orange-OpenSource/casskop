


# CassKop: Cassandra Kubernetes operator Description



<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->
**Table of Contents**

- [CassKop: Cassandra Kubernetes operator Description](#casskop-cassandra-kubernetes-operator-description)
    - [Overview](#overview)
        - [What is a Kubernetes operator](#what-is-a-kubernetes-operator)
        - [CassKop is our cassandra operator](#casskop-is-our-cassandra-operator)
        - [Deploying CassKop to Kubernetes](#deploying-casskop-to-kubernetes)
- [Deployment configuration](#deployment-configuration)
    - [Cassandra cluster configuration](#cassandra-cluster-configuration)
    - [Cassandra configuration](#cassandra-configuration)
        - [Cassandra docker image](#cassandra-docker-image)
        - [NodesPerRacks](#nodesperracks)
        - [Configuration override using configMap](#configuration-override-using-configmap)
        - [Configuration pre-run.sh script](#configuration-pre-runsh-script)
        - [JVM options](#jvm-options)
            - [Memory](#memory)
            - [GarbageCollector output](#garbagecollector-output)
        - [Authentication and authorizations](#authentication-and-authorizations)
    - [Cassandra storage](#cassandra-storage)
        - [Configuration](#configuration)
        - [Persistent volume claim](#persistent-volume-claim)
    - [Kubernetes objects](#kubernetes-objects)
        - [Services](#services)
        - [Statefulset](#statefulset)
    - [CPU and memory resources](#cpu-and-memory-resources)
        - [Resource limits and requests](#resource-limits-and-requests)
            - [Resource requests](#resource-requests)
            - [Resource limits](#resource-limits)
            - [Supported CPU formats](#supported-cpu-formats)
            - [Supported memory formats](#supported-memory-formats)
        - [Configuring resource requests and limits](#configuring-resource-requests-and-limits)
    - [Cluster topology: Cassandra rack aware deployments](#cluster-topology-cassandra-rack-aware-deployments)
        - [Quick overview](#quick-overview)
        - [Kubernetes nodes labels](#kubernetes-nodes-labels)
        - [Configuring pod scheduling](#configuring-pod-scheduling)
            - [Cassandra node placement in the Kubernetes cluster](#cassandra-node-placement-in-the-kubernetes-cluster)
                - [Node affinity](#node-affinity)
                - [Pod anti affinity](#pod-anti-affinity)
                - [Using dedicated nodes](#using-dedicated-nodes)
            - [Configuring hard antiAffinity in Cassandra cluster](#configuring-hard-antiaffinity-in-cassandra-cluster)
        - [Cassandra notion of dc and racks](#cassandra-notion-of-dc-and-racks)
        - [Configure the CassandraCluster CRD for dc & rack](#configure-the-cassandracluster-crd-for-dc--rack)
        - [How CassKop configures dc and rack in Cassandra](#how-casskop-configures-dc-and-rack-in-cassandra)
    - [Implementation architecture](#implementation-architecture)
        - [1 Statefulset for each racks](#1-statefulset-for-each-racks)
        - [Sequences](#sequences)
        - [Naming convention of created objects](#naming-convention-of-created-objects)
            - [List of resources created as part of the Cassandra cluster](#list-of-resources-created-as-part-of-the-cassandra-cluster)
    - [Advanced configuration](#advanced-configuration)
        - [Docker login for private registry](#docker-login-for-private-registry)
        - [Management of allowed Cassandra nodes disruption](#management-of-allowed-cassandra-nodes-disruption)
    - [Cassandra nodes management](#cassandra-nodes-management)
        - [HealthChecks](#healthchecks)
        - [Pod lifeCycle](#pod-lifecycle)
            - [PreStop](#prestop)
        - [Prometheus metrics export](#prometheus-metrics-export)
    - [CassandraCluster Status](#cassandracluster-status)
    - [Cassandra cluster CRD definition version 0.3.0](#cassandra-cluster-crd-definition-version-030)

<!-- markdown-toc end -->

## Overview

The Cassandra Kubernetes Operator (CassKop) makes it easy to run Apache Cassandra on Kubernetes. Apache Cassandra is
a popular free, open-source distributed wide column store NoSQL database management system.


CassKop will allow to easily create and manage Rack aware Cassandra Clusters.

### What is a Kubernetes operator

Kubernetes Operators are first-class citizens of a Kubernetes cluster and are 
application-specific controllers that extends Kubernetes to create, configure, and manage instances of complex applications.

We have choosen to use a [Custom Resource Definition (CRD)](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/) 
which creates a new Object Kind named **CassandraCluster** in Kubernetes which allow us to :

- Store object and state directly into Kubernetes
- Manage declarative inputs to set up the Cluster
- Declarative Updates - performing workflow actions on an existing CassandraCluster is 
straightforward. Updating a cluster amounts to updating the required declarative attributes 
in the CRD Yaml with new data and re-applying the CRD using kubectl.
CassKop's diff-based reconciliation logic ensures that only the required changes are made to a CassandraCluster.
- CassKop monitors create/update events for the CRD and performs the required actions.
- CassKop runs in a loop and reacts to events as they happen to reconcile a desired 
state with the actual state of the system.

### CassKop is our cassandra operator

CassKop will define a new Kubernetes object named `CassandraCluster` which 
will be used to describe and instantiate a Cassandra Cluster in Kubernetes. Example: 
- [cluster definition](../samples/cassandracluster-pic-test-acceptance-3.yaml)

CassKop is a Kubernetes custom controller which will loop over events on 
`CassandraCluster` objects and reconcile with kubernetes resources needed to create a valid 
Cassandra Cluster deployment.

CassKop is listening only in the Kubernetes namespace it is deployed in, and is able to manage several Cassandra Clusters within this namespace. 

When receiving a CassandraCluster object, CassKop will start creating required Kubernetes
resources, such as Services, Statefulsets, and so on.

Every time the desired CassandraCluster resource is updated by the user, CassKop performs
corresponding updates on the Kubernetes resources, so that the Cassandra cluster reflects the state of the desired cluster resource. Such updates might trigger a rolling update of the pods.

Finally, when the desired resource is deleted, CassKop starts to undeploy the cluster and delete all related Kubernetes resources. 


### Deploying CassKop to Kubernetes

See [Deploy the Cassandra Operator in the cluster](/#deploy-the-cassandra-operator-in-the-cluster)


# Deployment configuration

This chapter describes how to configure different aspects of the Cassandra Clusters.


## Cassandra cluster configuration

The full schema of the `CassandraCluster` resource is described in the [Cassandra Cluster CRD Definition](#cassandra-cluster-crd-definition-version-020).
All labels that are applied to the desired `CassandraCluster` resource will also be applied to the Kubernetes resources
making up the Cassandra cluster. This provides a convenient mechanism for those resources to be labelled in whatever way
the user requires.

## Cassandra configuration

### Cassandra docker image

CassKop relies on specific Docker Cassandra Image, where we provide specific startup script
that will be used to make live configuration on the Cassandra node we deploy.

The actual associated Docker image can be found in the gitlab repo :
[Cassandra for k8s](https://github.com/Orange-OpenSource/cassandra-image/).


You can manage the url and version of the docker image using the `spec.baseImage` and `spec.version` parameters of the
cassandracluster CRD.

> The Cassandra Image tag name must follow a specific semantic to allow CassKop to detect major
> changes in Cassandra version which may need to perform specific actions such as `nodetool upgradesstables`.

The actual naming of the Cassandra Image is as follow :

`<cassandra_version>-<java_version>-<git-tag-name>`

Example :
`3.11.3-8u201-0.3.0`

CassKop may use (TODO: in future) the first part of the tag which represents the Cassandra version to detect
what to do.


When Changing one of thoses fields, CassKop will triger an
[UpdateDockerImage](../documentation/operations.md#updatedockerimage)

### NodesPerRacks

One of the requirements for CassKop is to always keep the same number of nodes in each of it's racks, per Cassandra
DCs. The number of nodes used for the Cassandra Cluster is configured using the `CassandraCluster.spec.nodesPerRacks`
property.

> If you have not specify a Cluster Topology, then you'll have a default datacenter named `dc1`and a default rack named
> `rack1` 

It is a good practice for Cassandra to keep the same number of nodes in each Cassandra Racks. CassKop guarantees
that, but you can define different numbers of replicas for racks in different Cassandra DataCenters. 

This is done using the `nodesPerRacks` property in `CassandraCluster.spec.topology.dc[<idx>].nodesPerRacks`. If
specified on the datacenter level, this parameter takes priority over the global `CassandraCluster.spec.nodesPerRacks`.

Example:
example to scale up the nodesPerRacks in DC2 :

```yaml
  topology:
    dc:
      - name: dc1
        rack:
          - name: rack1
          - name: rack2
      - name: dc2
        nodesPerRacks: 3        <--- We increase by one this value
        rack:
          - name: rack1
```

> The number of Cassandra nodes will be the multiplication of the number of racks * the nodesPerRacks value.

If we changes on of theses properties then CassKop will triger either a [ScaleUp](../documentation/operations.md#scaleup)
or a [ScaleDown](../documentation/operations.md#scaledown) operation.

### Configuration override using configMap

CassKop allows you to customize the configuration of Apache Cassandra nodes by specifying a dedicated `ConfigMap`
name in the `CassandraCluster.spec.configMapName` containing configuration files to be overwritten above the default
configuration files of the dedicated Docker Image.

We have a specific Cassandra Docker image startup script that will overwrite each files in the directory
`/etc/cassandra`, from the one specified in the configMap if they exists.


>You can surcharge any files in the docker image `/etc/cassandra` with the ConfigMap.

Typical overwriting files may be :
- cassandra.yaml
- jvm.options
- specifying a pre-run.sh script

You can find an example with [this example](../samples/cassandra-configmap-v1.yaml)
```
$ kubectl apply -f config/cassandra-configmap-v1.yaml
configmap "cassandra-configmap-v1" created
```

Now you can add the `configMapName: cassandra-configmap-v1` to the Spec section of your CassandraCluster definition
[example](../samples/cassandracluster.yaml)

If you edit the ConfigMap it won't be detected neither by CassKop nor by the statefulsets/pods (unless you reboot the
pods). 
It is recommanded for configuration changes, to version the configmap and to create apply new configmap in the CRD, this
will trigger a rollingRestart of the whole cluster applying the new configuration.


> **IMPORTANT:** each time you specify a new configMap CassKop will start a `rollingUpdate` of all nodes
> in the cluster. more info on [UpdateConfigMap](../documentation/operations.md#updateconfigmap)

> **IMPORTANT:** At this time CassKop won't allow you to specify only excerpt of the configurations files, your
> ConfigMap **MUST** contain valid and complete configuration files 



### Configuration pre-run.sh script

In case you need to make some specific actions on a particular node, such as make uses of the **CASSANDRA_REPLACE_NODE**
variable, you can uses the pre-run.sh script in the ConfigMap. If present, the cassandra docker will execute this script
prior to the `run.sh` script from the docker image.

example of a configMap with the pre-run.sh script :

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cassandra-configmap-pre-run
data:
  pre_run.sh: |-
    echo "** this is a pre-scrip for run.sh that can be edit with configmap"
    test "$(hostname)" == 'cassandra-demo-dc1-rack1-0' && export CASSANDRA_REPLACE_NODE=10.233.93.174
    echo "** end of pre_run.sh script, continue with run.sh"
```

> **IMPORTANT:** In the case you use the configmap for one-time specific action, don't forget to edit again to remove
> the specific treatment once it is no more needed.

### JVM options

#### Memory

Apache Cassandra are running inside of a Java Virtual Machine (JVM). JVM has many configuration options to optimize the
performance for different platforms and architectures. 

CassKop allows configuring theses values by adding a `jvm.options` in the user `ConfigMap`.

The default value used for `-Xmx` depends on whether there is a memory request configured for the container :
- If there is a memory request, the JVM's maximum memory must be set to a value corresponding to the limit.
- When there is no memory request, CassKop will limit it to "2048M".

- set the memory request and the memory limit to the same value, so that the pod is in guarantee mode

> CassKop will automatically compute the env var CASSANDRA_MAX_HEAP which is used to define `-Xms` and `-Xmx` in the
> `/run.sh` docker image script, from 1/4 of container Memory Limit.

#### GarbageCollector output

We have a specific parameter in the CRD `spec.gcStdout: true/false` which specify if we wants to send the JVM garbage collector logs
in the stdout of the container, or inside a specific file in the container.

Default value is true, so it sends GC logs in stdout along with cassandra's logs.

### Authentication and authorizations

CassKop uses Jolokia from the cassandra-image to communicate. We can add
authentication on Jolokia by defining a secret :

Example:
```console
apiVersion: v1
kind: Secret
metadata:
  name: jolokia-auth
type: Opaque
data:
  password: TTBucDQ1NXcwcmQ=
  username: am9sb2tpYS11c2Vy
```

and in the CRD you wille define `spec.imageJolokiaSecret`

```console
...
  imageJolokiaSecret:
    name: jolokia-auth
...
```

CassKop will propagate the secrets in Cassandra so that it can configure
Jolokia, and uses it to connect.



## Cassandra storage

### Configuration

Cassandra is a stateful application. It needs to store data on disks. CassKop allows you to configure the type of
storage you want to use.

Storage can be configured using the `storage` property in `CassandraCluster.spec`

> **Important:** Once the Cassandra cluster is deployed, the storage cannot be changed.

Persistent storage uses Persistent Volume Claims to provision persistent volumes for storing data.
The `PersistentVolumes` are acquired using a `PersistentVolumeClaim` which is managed by CassKop. The
`PersistentVolumeClaim` can use a `StorageClass` to trigger automatic volume provisioning.

> It is recommended to uses local-storage with quick ssd disk access for low latency. We have only tested the
> `local-storage` storage class within CassKop.

CassandraCluster fragment of persistent storage definition :

```
# ...
  dataCapacity: "300Gi"
  dataStorageClass: "local-storage"
  deletePVC: true
# ...
```

- `dataCapacity` (required): Defines the size of the persistent volume claim, for example, "1000Gi".
- `dataStorageClass`(optional): Define the type of storage to uses (or use
  default one). We recommand to uses local-storage for better performances but
  it can be any storage with high ssd througput.
- `deletePVC`(optional): Boolean value which specifies if the Persistent Volume Claim has to be deleted when the cluster
  is deleted. Default is `false`.

> **WARNING**: Resizing persistent storage for existing CassandraCluster is not currently supported. You must decide the
> necessary storage size before deploying the cluster.

The above example asks that each nodes will have 300Gi of data volumes to persist the Cassandra data's using the
local-storage storage class provider.
The parameter deletePVC is used to control if the data storage must persist when the according statefulset is deleted.

> **WARNING:** If we don't specify dataCapacity, then CassKop will uses the Docker Container ephemeral storage, and
> all data will be lost in case of a cassandra node reboot.


### Persistent volume claim

When the persistent storage is used, it will create PersistentVolumeClaims with the following names:

`data-<cluster-name>-<dc-name>-<rack-name>-<idx>`

Persistent Volume Claim for the volume used for storing data to the cluster `<cluster-name>` for the Cassandra DC
`<dc-name>` and the rack `<rack-name>` for the Pod with ID `<idx>`.

> **IMPORTANT**: Note that with local-storage the PVC object makes a link between the Pod and the Node. While this
> object is existing the Pod will be sticked to the node chosen by the scheduler. In the case you want to move the
> Cassandra node to a new kubernetes node, you will need at some point to manually delete the associate PVC so that the
> scheduler can choose another Node for scheduling. This is cover in the Operation document.


## Kubernetes objects

### Services

Cassandra Pods will be accessible via Kubernetes headless services. CassKop will create a service for each
Cassandra DC define in the Topology section.

Service will be used by application to connect to the Cassandra Cluster.
Service will also be used for Cassandra to find others SEEDS nodes in the cluster.

### Statefulset

- **Statefulsets** is a powerful entity in Kubernetes to manage Pods, associated with some essential conventions :
    - Pod name: pods are created sequentially, starting with the name of the statefulset and ending with zero : 
    `<statefulset-name>-<ordinal-index>`. 
    - Network address: the statefulset uses a headless service to control the domain name of its pods. As each pod is
      created, it gets a matching DNS subdomain
    `<pod-name>.<service-name>.<namespace>`.


## CPU and memory resources

For every deployed container, CassKop allows you to specify the resources which should be reserved for it
and the maximum resources that can be consumed by it. We support two types of resources:

- Memory
- CPU

CassKop is using the Kubernetes syntax for specifying CPU and memory resources.

### Resource limits and requests

Resource limits and requests can be configured using the `resources` property in `CassandraCluster.spec.resources`.

#### Resource requests

Requests specify the resources that will be reserved for a given container. Reserving the resources will ensure that
they are always available. 

> **Important:** If the resource request is for more than the available free resources on the scheduled kubernetes node,
> the pod will remain stuck in "pending" state until the required resources become available. 


```
# ...
resources:
  requests:
    cpu: 12
    memory: 64Gi
# ...
```

#### Resource limits

Limits specify the maximum resources that can be consumed by a given container. The limit is not reserved and might not
be always available. The container can use the resources up to the limit only when they are available. The resource
limits should be always higher than the resource requests.  


```
# ...
resources:
  limits:
    cpu: 12
    memory: 64Gi
# ...
```

#### Supported CPU formats

CPU requests and limits are supported in the following formats:
- Number of CPU cores as integer (`5` CPU core) or decimal (`2.5`CPU core).
- Number of millicpus / millicores (`100m`) where 1000 millicores is the same as `1` CPU core.

For more details about CPU specification, refer to 
[kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu)

#### Supported memory formats

Memory requests and limits are specified in megabytes, gigabytes, mebibytes, gibibytes.
- to specify memory in megabytes, uses the `M` suffix. For example `1000M`.
- to specify memory in gigabytes, uses the `G` suffix. For example `1G`.
- to specify memory in mebibytes, uses the `Mi` suffix. For example `1000Mi`.
- to specify memory in gibibytes, uses the `Gi` suffix. For example `1Gi`.

For more details about CPU specification, refer to 
[kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-memory)

### Configuring resource requests and limits

the resources requests and limits for CPU and memory will be applied to all Cassandra Pods deployed in the Cluster.

It is configured directly in the `CassandraCluster.spec.resources`: 

```
  resources:
    requests:
      cpu: '2'
      memory: 2Gi
    limits:
      cpu: '2'
      memory: 2Gi
```

Depending on the values specified, Kubernetes will define 3 levels for QoS : (BestEffort < Burstable < Guaranteed).

- BestEffort: if no resources are specified
- Burstable: if limits > requests. if a system needs more resources, thoses pods can be terminated if they use more than
  requested and if there is no more BrstEffor Pods to terminated
- Guaranteed: request=limits. It is the recommanded configuration for cassandra pods.

When updating the crd resources, this will trigger an [UpdateResources](../documentation/operations.md#updateresources)) action.
 
## Cluster topology: Cassandra rack aware deployments

CassKop rack awareness feature helps to spread the Cassandra nodes replicas among different racks in the
kubernetes infrastructure. Enabling rack awareness helps to improve availability of Cassandra nodes and the data they
are hosting, through correct use of Cassandra's replication factor. 
 
> **Note:** Rack might represent an availability zone, data center or an actual physical rack in your data center.

### Quick overview

CassKop awareness can be configured in the `CassandraCluster.spec.topology` section. 

If the `topology` section is missing then CassKop will create a default Cassandra DC `dc1` and a default Rack
`rack1`. 
In this case, CassKop will not use specific kubernetes nodes labels for placement and in consequence the cluster is
not rack aware. 

In the `topology` section we can declare all the Cassandra Datacenters (DCs) and Racks, and for each of them we provide
labels which will need to match the labels assigned to the Kubernetes cluster nodes. 

Each of our rack targets Kubernetes nodes having the combination of labels defined in the `dc` section and in the `rack`
section. The labels are used in Kubernetes when scheduling the Cassandra pods to kubernetes nodes.
This has the effect of spreading the Cassandra nodes across physical zones.

 > **IMPORTANT:** CassKop doesn't rely on specific labels and will adapt to any topology you may define in your
 > datacenter or cloud provider. 

### Kubernetes nodes labels

Cassandra will run on Kubernetes Nodes, which may already have some labels representing their geographic topology.

Example :

![](../docs/assets/kubernetes-operators/topology-custom-example.png)

Example of labels for node001 in our dc:

```
location.myorg.com/bay=1
location.myorg.com/building=building
location.myorg.com/label=SC_K08_-_KUBERNETES
location.myorg.com/room=Salle_1
location.myorg.com/site=SiteName
location.myorg.com/street=Rue_3
```

In the cloud the labels used for topology may better look like :

![](../docs/assets/kubernetes-operators/topology-custom-example.png)

```
beta.kubernetes.io/fluentd-ds-ready=true
failure-domain.beta.kubernetes.io/region=europe-west1
kubernetes.io/hostname=gke-demo-default-pool-0c404f82-0100
beta.kubernetes.io/arch=amd64
failure-domain.beta.kubernetes.io/zone=europe-west1-d
beta.kubernetes.io/os=linux
cloud.google.com/gke-os-distribution=cos
beta.kubernetes.io/instance-type=n1-standard-4
cloud.google.com/gke-nodepool=default-pool
```

The idea is to use the Kubernetes nodes labels which refer to their physical position in the datacenters, to allow or
not Cassandra Pods placement.

Because CassKop manages its Cassandra Node Pods through 
[statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/), 
each Pod will inherit the placement configuration of its parent
statefulset. In order to place Pods on different racks, we need to associate a new statefulset with specialized nodes
placement constraints for each Cassandra Rack we define.


>Because the AZ feature was not yet available for statefulsets when we began to develop CassKop, we chose to
>implement 1 statefulset for 1 Cassandra Rack. Hopefully, we will be able to benefit of improvements about the
>topology-aware dynamic provisioning feature proposed in futur version of K8S (more informations here :
>https://kubernetes.io/blog/2018/10/11/topology-aware-volume-provisioning-in-kubernetes/
>)

See Example of configuration with topology : [cassandracluster-demo-gke.yaml](samples/cassandracluster-demo-gke.yaml)

### Configuring pod scheduling

When two applications are scheduled to the same Kubernetes node, both applications might use the same resources like
disk I/O and impact performances. It may be recommended to schedule Cassandra Pods in a way that avoids sharing nodes
with other critical workloads. Using the right nodes or dedicated a set of nodes only for cassandra are the best ways to
avoid such problems.

[Placement of Pods](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) in a Statefulsets are done 
using **NodeAffinity** and **PodAntiAffinity** : 

- [nodeAffinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) will be used to select specific
  nodes which have specific targeted labels.
- [podAntiAffinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity) can
  be used to ensure that critical applications are never scheduled on the same node.
  
#### Cassandra node placement in the Kubernetes cluster

##### Node affinity

To target a Specific Kubernetes group of Nodes CassKop needs to define a specific **nodeAffinity**
section in the targeted `dc-rack` statefulset to match specific kubernetes nodes labels.

Example. If we want to deploy our statefulset only on nodes which have theses labels :

```
location.myorg.com/site=Valbonne
location.myorg.com/building=HT2
location.myorg.com/room=Salle_1
location.myorg.com/street=Rue_11
```

CassKop need to add this section in the Statefulset definition :

```
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - key: location.myorg.com/building
            operator: In
            values:
            - HT2
          - key: location.myorg.com/room
            operator: In
            values:
            - Salle_1
          - key: location.myorg.com/site
            operator: In
            values:
            - Valbonne
          - key: location.myorg.com/street
            operator: In
            values:
            - Rue_9
```

All Pods that will be created from this statefulset will target only nodes which have theses 4 labels.
If there is no kubernetes nodes available with theses labels, then the statefulset will remain stuck in a pending state,
until we correct either labels on nodes, or deployment constraints definition of the statefulset. 

We can also use specific labels to dedicate some kubernetes nodes to some type of work, for instance dedicated some
nodes for Cassandra. 

You can define custom labels on kubernetes nodes :

```
kubectl label node <your-node> <label-name>=<label-value>
```

> This is done automatically by combining the labels you specify in the Cassandra CRD definition in the `Topology`
> section. 

##### Pod anti affinity

If we loose a Kubernetes node, then we may want to limit the impact to loosing only one Cassandra node. Otherwise,
depending on the replication factor, we may have a data loss. 

In our CassandraCluster, the statefulset will target a pool of Kubernetes nodes using it's NodeSelector we just saw
above.
All theses Pods will inherit the specific labels from the Statefulset. 

To implement the limitation "one Cassandra Pod per Kubernetes node", we use the pod definition section
`podAntiAffinity`. This tells kubernetes that it can't deploy to a Kubernetes node, if a pod having the same labels
already exists. 

```
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - podAffinityTerm:
          labelSelector:
            matchLabels:
              app: cassandracluster
              cassandracluster: cassandra
              cluster: k8s.kaas
          topologyKey: kubernetes.io/hostname
        weight: 100
```

This Tells that we **Require** not deploy to a node if a pod already exists with theses existing labels. 

> This is configured by default by CassKop

##### Using dedicated nodes

Cluster administrators can mark selected kubernetes nodes as tainted. Nodes with taints are excluded from regular
scheduling and normal pods will not be scheduled to run on them. Only services which can tolerate the taint set on the
node can be scheduled on it. The only other services running on such nodes will be kubernetes system services such  as
log collectors or software  defined networks

Taints can be used to create dedicated nodes. Running Cassandra on dedicated nodes can have many advantages. There will
be no other applications running on the same nodes which could cause disturbance or consume the resources needed  for
Cassandra. That can lead to improved performance and stability.

Example of tainting a node :

```
kubectl taint node <your-node> dedicated=Cassandra:NoSchedule
```

Additionally, add a label to the selected nodes as well

```
kubectl label node <your-node> dedicated=Cassandra
```

The `toleration` must be applied on the statefulset at the same level as `affinity`.

```
...
    tolerations:
      - key: "dedicated"
        operator: "Equal"
        value: "Cassandra"
        effect: "NoSchedule"
...
```

> **IMPORTANT:** toleration must be used with node affinity on the same labels
> TODO: actually the toleration is not implemented in CassKop


#### Configuring hard antiAffinity in Cassandra cluster

In development environment, we may have other concern than in production and we may allow our cluster to deploy several
nodes on the same kubernetes nodes.

The boolean `hardAntiAffinity` parameter in `CassandraCluster.spec` will define if we want the constraint to be
**Required** or **Preferred**


If `hardAntiAffinity=false` then the podAntiAffinity will be **preferred** instead of **required** and then kubernetes
will try to not put the cassandra node on the kubernetes node **BUT** it will allow to do it if it has no other choices.

### Cassandra notion of dc and racks

As we previously see, the Cassandra rack awareness is defined using several Cassandra datacenters `dc`s and `rack`s.
The `CassandraCluster.spec.topology` section allows us to define the virtual notion of DC & Rack.
For each we will define Kubernetes `labels` that will be used for pod placement.

> The name and numbers of labels used to define a DC & rack is not defined in advance and will be defined in each
> CassandraCluster manifests, depending on labels presents on Kubernetes Nodes and the required topology

### Configure the CassandraCluster CRD for dc & rack

If the section topology is missing: 
- then CassKop will deploy without label constraints to any available kubernetes nodes in the cluster
- there will be only one DC defined named by default **dc1**
- there will be only one Rack defined named by default **rack1**

If the **topology** section is defined, this will enable to create specific Cassandra dcs and racks: this creates a
specific statefulset for each rack and deploys on Kubernetes nodes matching the desired Node Labels (the concatenation
of DC labels + Rack labels for each statefulset)

Example of topology section:

```
...
  nodesPerRacks: 3
...
  topology:
    dc:
      - name: dc1
        labels:
          location.k8s.myorg.com/site : Valbonne
          location.k8s.myorg.com/building : HT2
        rack:
          - name: rack1
            labels:
              location.k8s.myorg.com/room : Salle_1
              location.k8s.myorg.com/street : Rue_9
          - name: rack2
            labels:
              location.k8s.myorg.com/room : Salle_1
              location.k8s.myorg.com/street : Rue_10
      - name: dc2
        nodesPerRacks: 4
        labels:
          location.k8s.myorg.com/site : Valbonne
          location.k8s.myorg.com/building : HT2
        rack:
          - name: rack1
            labels:
              location.k8s.myorg.com/room : Salle_1
              location.k8s.myorg.com/street : Rue_11
```

- This will create 2 Cassandra DC (`dc1` & `dc2`)
    - For DC `dc1` it will create 2 Racks : `rack1` and `rack2`
      - In each of theses Racks there will be 3 Cassandra nodes
      - The `dc1` will have 6 nodes
    - For DC `dc2` it will create 1 Rack : `rack1`
      - the `dc2` overwrite the global parameter `nodesPerRacks=3` with a value of `4`. 
      - The `dc2` will have 4 nodes

> **Important:** We want to have the same numbers of Cassandra nodes in each Racks for a dedicated
> Datacenter. We can still have different values for different datacenters. 


The NodeSelectors labels for each Rack will be the aggregation of labels of the DC and the labels for the Racks :

For instance with this example, NodeSelectors labels for `dc1` / `rack2` will be :
 ```
 location.k8s.myorg.com/site : Valbonne
 location.k8s.myorg.com/building : HT2
 location.k8s.myorg.com/room : Salle_1
 location.k8s.myorg.com/street : Rue_10
 ```

> The `dc` / `rack` topology definition is generic and does not rely on particular labels. It uses the ones
> corresponding to your needs.

> **Note:** The names for the dc and rack must be lowercase and respect Kubernetes DNS naming which follow [RFC 1123
> definition](http://tools.ietf.org/html/rfc1123#section-2) which can be expressed with this regular expression :
> `[a-z0-9]([-a-z0-9]*[a-z0-9])?`


### How CassKop configures dc and rack in Cassandra

CassKop will add 2 specific labels on each created Pod to tell them in witch Cassandra DC and Rack they belong :

Example :
```
cassandraclusters.db.orange.com.dc=dc1
cassandraclusters.db.orange.com.rack=rack1
```

Using the Kubernetes DownwardAPI, CassKop will inject into the Cassandra Image 2 environment variables, from theses
2 labels. Excerpt from the Statefulset template :

```
...
v1.EnvVar{
	Name: "CASSANDRA_DC",
	ValueFrom: &v1.EnvVarSource{
		FieldRef: &v1.ObjectFieldSelector{
			FieldPath: "metadata.labels['cassandraclusters.db.orange.com.dc']",
		},
	},
},
v1.EnvVar{
	Name: "CASSANDRA_RACK",
	ValueFrom: &v1.EnvVarSource{
		FieldRef: &v1.ObjectFieldSelector{
			FieldPath: "metadata.labels['cassandraclusters.db.orange.com.rack']",
		},
	},
},
...
```

In order to allow configuring Cassandra with the DC and Rack information, we use a specific [Cassandra
Image](https://github.com/Orange-OpenSource/cassandra-image/), which has a startup script that will retrieve
theses environment variables, and configure the Cassandra `cassandra-rackdc.properties` file with the values for dc and
rack.

The Cassandra Image makes us of the `GossipingPropertyFileSnitch` Cassandra Snitch, so that both Kubernetes and
Cassandra are aware of the chosen topology. 


## Implementation architecture

### 1 Statefulset for each racks

CassKop will create a dedicated statefulset and service for each couple `dc-rack` defined in the
`topology`section. This is done to ensure we'll always have the same amounts of cassandra nodes in each rack for a
specified DC.


![architecture](http://www.plantuml.com/plantuml/proxy?src=https://raw.github.com/Orange-OpenSource/cassandra-k8s-operator/master/documentation/uml/architecture.puml)

### Sequences

CassKop will works in sequence for each DC-Rack it has created which are different statefulsets kubernetes objects.
Each time we request a change on the cassandracluster CRD which implies rollingUpdate of the statefulset, CassKop will
perform the update on the first dc-rack.

> CassKop will then wait for the operation to complete before starting the upgrade on the next dc-rack!!

If you play with `spec.topology.dc[].rack[].rollingPartition` with value greater than 0, then the rolling update of the rack
won't end and CassKop won't update the next one. In order to allow a statefulset to upgrade completely the rollingPartition must be set to 0 (default).


### Naming convention of created objects

When declaring a new `CassandraCluster`, we need to specify its Name, and all its configuration.

Here is an excerpt of a CassandraCluster CRD definition:

```
apiVersion: "db.orange.com/v1alpha1"
kind: "CassandraCluster"
metadata:
  name: cassandra-demo
  labels:
    cluster: optional-label 
spec:
  ...
  nodesPerRacks: 3
  
  topology:
    dc:
      - name: dc1
        labels:
          location.myorg.com/site : mts
        rack:
          - name: rack1
            labels: 
              location.myorg.com/street : street1
          - name: rack2
            labels: 
              location.myorg.com/street : street2
      - name: dc2
        nodesPerRacks: 4
        labels: 
          location.myorg.com/site : mts
        rack:
          - name: rack1
            labels: 
              location.myorg.com/street : street3
```

A complete example can be found [here](../samples/cassandracluster-pic.yaml)

Kubernetes objects created by CassKop are named according to : 
- `<CassandraClusterName>-<DCName>-<RackName>` 

> **IMPORTANT:** All elements must be in lowerCase according to Kubernetes DNS naming constraints

#### List of resources created as part of the Cassandra cluster

- `<cluster-name>`
    - PodDisruptionBudget: this is checked by Kubernetes and by CassKop and allows only 1 pod disrupted
      on the whole cluster. CassKop won't update statefulset in case there is a disruption.
- `<cluster-name>-<dc-name>`
    - [Headless service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) at dc level
      used as client applications entry point to contact all nodes in a Cassandra DC.
- `<cluster-name>-<dc-name>-<rack-name>`
    - Statefulset which is in charge of managing Cassandra Pods for dc-name and rack-name
    - Service headless used for Seeds discovery
- `<cluster-name>-<dc-name>-<rack-name>-<idx>`
    - Pods Names in the Statefulset for dc-name and rack-name with ordinal index.
- `data-<cluster-name>-<dc-name>-<rack-name>-<idx>`
    - PersistentVolumeClaim representing the data for the associated Cassandra pod.
- `<cluster-name>-<dc-name>-<rack-name>-exporter-jmx`
    - Service Name for the exporter JMX for dc-name and rack-name
- `<cluster-name>`


With the previous example:
- The CassandraCluster name is `cassandra-demo`
- the first DC is named `dc1`
    - the first rack is named `rack1`
    - the second rack is named `rack2`
- the second DC is named `dc2`
    - the first rack is named `rack1`

Example for DC `dc1-rack1` :
- the statefulsets is named : `cassandra-demo-dc1-rack1`,`cassandra-demo-dc1-rack2`,`cassandra-demo-dc2-rack1`
    - the statefulset Pods name will add the ordinal number suffix :
      `cassandra-demo-dc1-rack1-0`,..,`cassandra-demo-dc1-rack1-n` for each dc-racks
- The services will be names : `cassandra-demo-dc1` and `cassandra-demo-dc2` 
- the associated service for Prometheus metrics export will be named :
  `cassandra-demo-dc1-exporter-jmx`,`cassandra-demo-dc2-exporter-jmx`  
- the PVC (Persistent Volume Claim) of each pods will be named **data-<podName>** ex: `data-cassandra-demo-dc1-rack1-0`
  for each dc-racks
- the [PodDisruptionBudget (PDB)](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) will be named :
  `cassandra-demo` and will target all pods of the cluster
  
> **Note:**: usually the PDB is only used when dealing with pod eviction (draining a kubernetes node). But CassKop
> also checks the PDB to know if it is allowed to make some actions on the cluster (restart Pod, apply changes..). If
> the PDB won't allow CassKop to make the change, it will wait until the PDB rule is satisfied. (We won't be able
> to make any change on the cluster, but it cassandra will continue to work underneeth).

    
## Advanced configuration

### Docker login for private registry

If you need to use a docker registry with authentication, then you will need to create a specific kubernetes secret with
theses informations.
Then you will configure the CRD with the secret name, so that it provides the data to each Statefulsets, which in
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


### Management of allowed Cassandra nodes disruption

CassKop makes uses of the kubernetes PodDisruptionBudget objetc to specify how many cassandra nodes disruption is
allowed on the cluster. By default, we only tolerate 1 disrupted pod at a time and will prevent to makes actions if
there is aloready an ongling disruption on the cluster.

In some edge cases it can be useful to make force the operator to continue it's actions even if there is already a
disruption ongoing. We can tune this by updating the `spec.maxPodUnavailable` parameter of the cassandracluster CRD.

> **IMPORTANT:** it is recommanded to not touch this parameter unless you know what you are doing.

## Cassandra nodes management

CassKop in duo with the Cassandra docker Image is responsible of the lifecycle of the Cassandra nodes.

### HealthChecks

Healthchecks are periodical tests which verify Cassandra's health. When the healthcheck fails, Kubernetes can assume
that the application is not healthy and attempt to fix it. Kubernetes supports two types of Healthcheck probes : 
- Liveness probes
- Readiness probes.

You can find more details in the [Kubernetes
documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/).

Both `livenessProbe` and `readinessProbe` support two additional options:
- `initialDelaySeconds`: defines the initial delay before the probe is tried for the first time. Default is 15 seconds
- `timeoutSeconds`: defines the timeout of the probe. CassKop uses 20 seconds.
- `periodSeconds`: the period to wait between each call to a probe: CassKop uses 40 seconds.

> TODO: This is actually not configurable by CassKop: [Issue102](https://github.com/Orange-OpenSource/cassandra-k8s-operator/issues/102)


### Pod lifeCycle

The Kubernetes Pods allows user to defines specific hooks to be executed at some times

#### PreStop

CassKop uses the PreStop hook to execute some commands before the pod is going to be killed.
In first iteration we were executing a `nodetool drain` and it used to make some unpredictible behaviours.
At the time of writing this document, there is no `PreStop` action executed. 


### Prometheus metrics export

We currently uses the CoreOS Prometheus Operator to export the Cassandra nodes metrics. We must create a serviceMonitor
object in the prometheus namespaces, pointing to the exporter-prometheus-service which is created by CassKop:


```yaml
$ cat samples/prometheus-cassandra-service-monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: prometheus-cassandra-jmx
  labels:
    k8s-apps: cassandra-k8s-jmx
    prometheus: kube-prometheus
    component: cassandra
    app: cassandra
spec:
  jobLabel: kube-prometheus-cassandra-k8s-jmx
  selector:
    matchLabels:
      k8s-app: exporter-cassandra-jmx
  namespaceSelector:
      matchNames:
      - cassandra
      - cassandra-demo
  endpoints:
  - port: http-promcassjmx
    interval: 15s
```

Actually the Cassandra nodes uses the work of Oleg Glusahak https://github.com/oleg-glushak/cassandra-prometheus-jmx but
this may change in the futur.

## CassandraCluster Status

You can request kubernetes Object `cassandracluster` representing the Cassandra cluster to retrieve information about
it's status :

```
$ kubectl describe cassandracluster cassandra
...
status:
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
- **CassandraRackStatus** represents a map of statuses for each of the Cassandra Racks in the Cluster
  - **<Cassandra DC-Rack Name>**
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
  
> When Status=Done for each Racks, then there is no specific action ongoing on the cluster and the
> lastClusterActionStatus will turn also to Done.


## Cassandra cluster CRD definition version 0.3.0

The CRD Type is how we want to declare a CassandraCluster Object into Kubernetes.

To achieve this, we update the CRD to manage both :

- The new topology section
- The new CassandraRack

![architecture](http://www.plantuml.com/plantuml/proxy?src=https://raw.github.com/Orange-OpenSource/cassandra-k8s-operator/master/documentation/uml/crd.puml)
