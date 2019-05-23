# CassKop - Cassandra Kubernetes operator

[![CircleCI](https://circleci.com/gh/Orange-OpenSource/cassandra-k8s-operator.svg?style=svg&circle-token=480ca5c31a9e9ef9b893151dd2d7c15eaf0e94d0)](https://circleci.com/gh/Orange-OpenSource/cassandra-k8s-operator)

## Project overview

The CassKop Cassandra Kubernetes operator makes it easy to run Apache Cassandra on Kubernetes. Apache Cassandra is a popular, 
free, open-source, distributed wide column store, NoSQL database management system. 
The operator allows to easily create and manage racks and data centers aware Cassandra clusters.

The Cassandra operator is based on the CoreOS
[operator-sdk](https://github.com/operator-framework/operator-sdk) tools and APIs.

> **NOTE**: This is an alpha-status project. We do regular tests on the code and functionality, but we can not assure a
> production-ready stability at this time.
> Our goal is to make it run in production as quickly as possible.


CassKop creates/configures/manages Cassandra clusters atop Kubernetes and is by default **space-scoped** which means
that :
- CassKop is able to manage X Cassandra clusters in one Kubernetes namespace.
- You need X instances of CassKop to manage Y Cassandra clusters in X different namespaces (1 instance of CassKop
  per namespace).

> This adds security between namespaces with a better isolation, and less work for each operator.

## TL;DR - CassKop presentation

We have some slides for a [CassKop demo](https://orange-opensource.github.io/cassandra-k8s-operator/index.html?slides=Slides-CassKop-demo.md#1)


## CassKop features


At this time of the project, the goal is to deploy a Cassandra cluster in 1 Kubernetes datacenter, but this will 
change in next versions to deal with Kubernetes in multi-datacenters. 

The following features are supported by CassKop:
- [x] Deployment of a C* cluster (rack or AZ aware)
- [x] Scaling up the cluster (with cleanup)
- [x] Scaling down the cluster (with decommission prior to Kubernetes scale down)
- [x] Pods operations (removenode, upgradesstable, cleanup, rebuild..)
- [x] Adding a Cassandra DC
- [x] Removing a Cassandra DC
- [x] Setting and modifying configuration files
- [x] Setting and modifying configuration parameters
- [x] Update of the Cassandra docker image
- [x] Rolling update of a Cassandra cluster
    - [x] Update of Cassandra version (including upgradesstable in case of major upgrade)
    - [x] Update of JVM
    - [x] Update of configuration
- [x] Rolling restart of a Cassandra rack
- [x] Stopping a Kubernetes node for maintenance
    - [x] Process a remove node (and create new Cassandra node on another Kubernetes node)
    - [x] Process a replace address (of the old Cassandra node on another Kubernetes node)
- [x] Manage operations on pods through CassKop plugin (cleanup, rebuild, upgradesstable, removenode..)
- [ ] Monitoring (using Instaclustr Prometheus exporter to Prometheus/Grafana)
- [ ] Performing live backup of Cassandra datas (using Instaclustr sidecar)
- [ ] Performing live restore of datas (using Instaclustr sidecar)
- [x] Performing live Cassandra repairs through the use of [Cassandra reaper](http://cassandra-reaper.io/)
- [ ] Pause/Restart operations through CassKoP plugin.

> CassKop doesn't use nodetool but invokes operations through authenticated JMX/Jolokia call


## Pre-requisites

### For developers

Operator SDK is part of the operator framework provided by RedHat & CoreOS. The goal 
is to provide high-level abstractions that simplifies creating Kubernetes operators.

The quick start guide walks through the process of building the Cassandra operator 
using the SDK CLI, setting up the RBAC, deploying the operator and creating a 
Cassandra cluster.

You can find this in the [Developer section](documentation/development.md)

### For users

Users should only need Kubectl & helm cli

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) version v1.13.3+.
- [Helm](https://helm.sh/) version v2.12.2+.
- Access to a Kubernetes v1.13.3+ cluster.
- Cassandra needs fast local storage (we have tested with local storage
  provisioner, GKE ssd storage, and Rancher local-path-provisioner)

#### Install CassKop kubectl plugin

You can install the plugin by copying the [file](plugins/kubectl-casskop) into your PATH.


## Build pipelines

We uses CircleCI as our CI tool to build and test the operator.

### Build image

To accelerate build phases we have created a custom [build-image](docker/circleci/Dockerfile) used by the CircleCI pipeline:

https://cloud.docker.com/u/orangeopensource/repository/docker/orangeopensource/casskop-build

You can find more info in the [developer Section](documentation/development.md)



### Cassandra operator

The Cassandra operator image is automatically built and stored on [Docker Hub](https://cloud.docker.com/u/orangeopensource/repository/docker/orangeopensource/cassandra-k8s-operator)

[![CircleCI](https://circleci.com/gh/Orange-OpenSource/cassandra-k8s-operator.svg?style=svg&circle-token=480ca5c31a9e9ef9b893151dd2d7c15eaf0e94d0)](https://circleci.com/gh/Orange-OpenSource/cassandra-k8s-operator)

### Associated Cassandra image

The operator works with specific Docker Cassandra image which is build in the repository:
https://github.com/Orange-OpenSource/cassandra-image

This table shows compatibility between CassKop and associated Cassandra image


| Operator      | Cassandra-k8s         |
| ---------     | -----------           |
| 0.3.1-release | >= 3.11.4-8u212-0.3.1 |



> docker image: orangeopensource/cassandra-image:3.11.4-8u212-0.3.1

## Deploy the Cassandra operator in the cluster

First, we need to create a Kubernetes `namespace` in order to host our operator & cluster

```
kubectl create namespace cassandra
```

### First deploy the CassandraCluster CRD definition

Before deploying the operator, we need to create the CRD (CassandraCluster Custom Resource Definition).

Deploy RBAC and CassKop custom resource definition
```
kubectl apply -f deploy/crds/db_v1alpha1_cassandracluster_crd.yaml
```

Check that CRD is deployed

```
$ kubectl get crd
NAME                              AGE
...
cassandraclusters.db.orange.com   1h
...
```

### With Helm

To ease the use of the Cassandra operator, a [Helm](https://helm.sh/) charts has been 
created

> We are looking where to store our helm in the future


Deploy CassKop:

```console
$ helm install --name casskop ./helm/cassandra-k8s-operator                                                                                                                                15:34:24 
NAME:   casskop
LAST DEPLOYED: Thu May 23 15:34:27 2019
NAMESPACE: cassandra-demo
STATUS: DEPLOYED

RESOURCES:
==> v1/ServiceAccount
NAME                    SECRETS  AGE
cassandra-k8s-operator  1        0s

==> v1beta1/Role
NAME                    AGE
cassandra-k8s-operator  0s

==> v1/RoleBinding
NAME                    AGE
cassandra-k8s-operator  0s

==> v1/Deployment
NAME                            DESIRED  CURRENT  UP-TO-DATE  AVAILABLE  AGE
casskop-cassandra-k8s-operator  1        1        1           0          0s

==> v1/Pod(related)
NAME                                            READY  STATUS             RESTARTS  AGE
casskop-cassandra-k8s-operator-78786b9bf-cjggg  0/1    ContainerCreating  0
0s
```

> You can find more information in the [Cassandra operator Helm readme](helm/cassandra-k8s-operator/readme.md)


This creates a Kubernetes Deployment for the operator, with RBAC settings.

Once deployed, you may find the Pods created by the Charts. If you deploy a release
named `casskop`, then your pod will have a name similar to :

```
$ kubectl get pod
NAME                                 READY     STATUS    RESTARTS   AGE
casskop-cassandra-k8s-operator-78786b9bf-cjggg   1/1       Running   0          1h
```

You can view the CassKop logs using 

```
$ kubectl logs -f cassandra-cassandra-k8s-operator-78786b9bf-cjggg
```


## Deploy a Cassandra cluster


### From local yaml spec

Once the operator is deployed inside a Kubernetes cluster, a new API will be accessible, so 
you'll be able to create, update and delete cassandraclusters.

In order to deploy a new cassandra cluster a [specification](samples/cassandracluster.yaml) 
has to be created:

For example :

```
kubectl apply -f samples/cassandracluster.yaml
```

see pods coming into life :

```
kubectl get pods -w
```

You can watch the status updates in real time on your CassandraCluster object :

```
watch 'kubectl describe cassandracluster cassandra-demo | tail -20'
```

## Cassandra cluster status

You can find mode information on the `CassandraCluster.status` in [this section](documentation/description.md#cassandracluster-status)

<!--
If you need mode information you can check the [algorithm](documentation/algorithms.md) page
-->

## Make operation on the cluster

You can do a lot of [operations](documentation/operations.md) on your Cassandra cluster.

## Cassandra operator recovery

If the Cassandra operator restarts, it can recover its previous state thanks to the CRD objects 
`CassandraClusters` which stored directly in Kubernetes, description and state of the Cassandra cluster.


## Cleanup

If you want to delete the operator from your Kubernetes cluster, the operator deployment 
should be deleted.

Also, the CRD has to be deleted too:
```
kubectl delete crd cassandraclusters.dfy.orange.com
```

> **!!!!!!!!WARNING!!!!!!!!**
>
> If you delete the CRD then **!!!!!!WAAAARRRRNNIIIIINNG!!!!!!**
>
> It will delete **ALL** Clusters that has been created using this CRD!!!
>
> Please never delete a CRD without very very good care

### Operator SDK

CassKop is build using operator SDK:

- [operator-sdk](https://github.com/operator-framework/operator-sdk)
- [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager)

### Monitoring

We can quickly setup monitoring for our deployed Cassandra nodes using
Prometheus operator.

#### Deploy Prometheus

You can deploy the CoreOs Prometheus operator on your cluster:
You can find example [helm value.yaml](samples/prometheus-values.yaml) to configure the Prometheus operator:

```console
$ kubectl create namespace monitoring
$ helm install --namespace monitoring --name prometheus stable/prometheus-operator
```

#### Add ServiceMonitor for Cassandra

Then you have to define a ServiceMonitor object to monitor
cluster deployed by your cassandra operator (one time), update this to specify which namespace to monitor.

cassandra-service-monitor.yml
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kube-prometheus-cassandra-k8s-jmx
  labels:
    k8s-apps: cassandra-k8s-jmx
    prometheus: kube-prometheus
    component: cassandra
    release: prometheus    
spec:
  jobLabel: kube-prometheus-cassandra-k8s-jmx
  selector:
    matchLabels:
      k8s-app: exporter-cassandra-jmx
  namespaceSelector:
      matchNames:
      - cassandra
      - default
  endpoints:
  - port: http-promcassjmx
    interval: 15s
```


Your namespace need to be listed in the `namespaceSelector` section.

#### Add Grafana dashboard for Cassandra

You can import this [dashboard](samples/prometheus-grafana-cassandra-dashboard.json) to retrieve metrics about your Cassandra cluster.


