---
id: 1_getting_started
title: Getting Started
sidebar_label: Getting Started
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

The operator uses the standard Cassandra image (tested up to Version 3.11), and can run on Kubernetes 1.13.3+.

:::info
The operator supports Cassandra 3.11.0+
:::

As a pre-requisite it needs :

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) version v1.13.3+.
- [Helm](https://helm.sh/) version v2.12.2+.
- Access to a Kubernetes v1.13.3+ cluster.
- Cassandra needs fast local storage (we have tested with local storage provisioner, GKE ssd storage, and Rancher local-path-provisioner). Fast remote storage should work but is not tested yet.

## CassKop

### Kubernetes preparation

First, we need to create a `namespace` in order to host our operator & cluster (here we call it cassandra)

```
kubectl create namespace cassandra
```

We recommend using a **custom StorageClass** to leverage the volume binding mode `WaitForFirstConsumer`

```bash
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: exampleStorageclass
parameters:
  type: pd-standard
provisioner: kubernetes.io/gce-pd
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

:::tip
Remember to set your CassandraCluster CR properly to use the newly created StorageClass.
:::

### Installing CassKop with Helm

You can (should) deploy CassKop using this [Helm chart](https://github.com/Orange-Opensource/casskop/tree/master/helm):

First we add the repo:


```bash
helm repo add orange-incubator https://orange-kubernetes-charts-incubator.storage.googleapis.com/
```

Then the chart itself depending on your installed version of Helm:

<Tabs
  defaultValue="helm3"
  values={[
    { label: 'helm 3', value: 'helm3', },
    { label: 'helm previous', value: 'helm', },
  ]
}>
<TabItem value="helm3">

```bash
# You have to create the namespace before executing following command
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/casskop/master/deploy/crds/db.orange.com_cassandraclusters_crd.yaml
helm install casskop --namespace=cassandra orange-incubator/casskop
```

</TabItem>
<TabItem value="helm">

```bash
# Install helm 
helm init --history-max 200
kubectl create serviceaccount tiller --namespace kube-system
kubectl create -f tiller-clusterrolebinding.yaml
helm init --service-account tiller --upgrade

# Deploy operator
helm install --name=casskop --namespace=cassandra orange-incubator/casskop
```
</TabItem>
</Tabs>

:::note
To install the an other version of the operator use `helm install --name=casskop --namespace=cassandra --set operator.image.tag=x.y.z orange-incubator/casskop`
:::


You can find more information in the [Customizable install with helm](/casskop/docs/2_setup/3_install/1_customizable_install_with_helm)
If you have problem you can see [Troubleshooting](/casskop/docs/7_troubleshooting/1_operations_issues) section


### Deploy a Cassandra cluster

#### Deploy a ConfigMap

Before we can deploy our cluster, we need to create a configmap.
This configmap will enable us to customize Cassandra's behaviour.
More details on this can be found [here](https://github.com/Orange-OpenSource/casskop/blob/master/documentation/description.md#configuration-override-using-configmap)

But for our example we will use the simple example: 

```
kubectl apply -f samples/cassandra-configmap-v1.yaml
```

#### Deploy CassandraCluster resource

Once the operator is deployed inside a Kubernetes cluster, a new API will be accessible, so 
you'll be able to create, update and delete cassandraclusters.


In order to deploy a new cassandra cluster a [specification](https://github.com/Orange-OpenSource/casskop/blob/master/samples/cassandracluster.yaml) has to be created. As an example :

```
kubectl apply -f samples/cassandracluster.yaml
```

### Monitoring

We can quickly setup monitoring for our deployed Cassandra nodes using
Prometheus operator.

#### Deploy Prometheus

You can deploy the CoreOs Prometheus operator on your cluster:
You can find example [helm value.yaml](https://github.com/Orange-OpenSource/casskop/blob/master/samples/prometheus-values.yaml) to configure the Prometheus operator:

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
      - cassandra-demo
      - default
  endpoints:
  - port: promjmx
    interval: 15s
```

Your namespace need to be listed in the `namespaceSelector` section.

#### Add Grafana dashboard for Cassandra

You can import this [dashboard](https://github.com/Orange-OpenSource/casskop/blob/master/samples/prometheus-grafana-cassandra-dashboard.json) to retrieve metrics about your Cassandra cluster.

## MultiCasskop

### Pre-requisite

In order to have a working Multi-CassKop operator we need to have at least 2 k8s clusters: k8s-cluster-1 and k8s-cluster-2

- k8s >=v1.15 installed on each site, with kubectl configure to access both of thems
- The pods of each site must be able to reach pods on other sites, this is outside of the scope of Multi-Casskop and can
  be achieve by different solutions such as:
  - in our on-premise cluster, we leverage [Calico](https://www.projectcalico.org/why-bgp/) routable IP pool in order to make this possible
  - this can also be done using mesh service such as istio
  - there may be other solutions as well
- having CassKop installed (with its ConfigMap) in each namespace see [CassKop installation](#casskop)
- having [External-DNS](https://github.com/kubernetes-sigs/external-dns) with RFC2136 installed in each namespace to
  manage your DNS sub zone. see [Install external dns](#install-external-dns)
- You need to create secrets from targeted k8s clusters in current k8S cluster (see [Bootstrap](#bootstrap-api-access-to-k8s-cluster-2-from-k8s-cluster-1))
- You may need to create network policies for Multi-Casskop inter-site communications to k8s apis, if using so.

:::warning
We have only tested the configuration with Calico routable IP pool & external DNS with RFC2136 configuration.
:::

#### Bootstrap API access to k8s-cluster-2 from k8s-cluster-1

Multi-Casskop will be deployed in k8s-cluster-1, change your kubectl context to point to this cluster.

In order to allow our Multi-CassKop controller to have access to k8s-cluster-2 from k8s-cluster-1, we are going to use
[kubemcsa](https://github.com/admiraltyio/multicluster-service-account/releases/tag/v0.6.1) from
[Admiralty](https://admiralty.io/) to be able to export secret from k8s-cluster-2 to k8s-cluster1

```sh
kubemcsa export --context=cluster2 --namespace cassandra-e2e cassandra-operator --as k8s-cluster2 | kubectl apply -f -
```

:::tips
This will create in current k8s (k8s-cluster-1) the k8s secret associated to the
**cassandra-operator** service account of namespace **cassandra-e2e** in k8s-cluster2.
/!\ The Secret will be created with the name **k8s-cluster2** and this name must be used when starting Multi-CassKop and
in the MultiCassKop CRD definition see below
:::

#### Install CassKop

CassKop must be deployed on each targeted Kubernetes clusters.

### Install External-DNS

[External-DNS](https://github.com/kubernetes-sigs/external-dns) must be installed in each Kubernetes clusters.
Configure your external DNS with a custom values pointing to your zone and deploy it in your namespace 

```console
helm install -f /private/externaldns-values.yaml --name casskop-dns external-dns 
```

### Install Multi-CassKop

Proceed with Multi-CassKop installation only when [Pre-requisites](#pre-requisites) are fulfilled.

Deployment with Helm. Multi-CassKop and CassKop shared the same github/helm repo and semantic version.

<Tabs
  defaultValue="helm3"
  values={[
    { label: 'helm 3', value: 'helm3', },
    { label: 'helm previous', value: 'helm', },
  ]
}>
<TabItem value="helm3">

```bash
# You have to create the namespace before executing following command
kubectl apply -f https://github.com/Orange-OpenSource/casskop/blob/master/multi-casskop/deploy/crds/multicluster_v1alpha1_cassandramulticluster_crd.yaml
helm install  multi-casskop --namespace=cassandra orange-incubator/multi-casskop --set k8s.local=k8s-cluster1 --set k8s.remote={k8s-cluster2}
```

</TabItem>
<TabItem value="helm">

```bash
# Install helm 
helm init --history-max 200
kubectl create serviceaccount tiller --namespace kube-system
kubectl create -f tiller-clusterrolebinding.yaml
helm init --service-account tiller --upgrade

# Deploy operator
helm install --name=multi-casskop --namespace=cassandra orange-incubator/multi-casskop --set k8s.local=k8s-cluster1 --set k8s.remote={k8s-cluster2}
```
</TabItem>
</Tabs>

When starting Multi-CassKop, we need to give some parameters:
- k8s.local is the name of the k8s-cluster we want to refere to when talking to this cluster.
- k8s.remote is a list of other kubernetes we want to connect to.

:::info
Names used there should map with the name used in the MultiCassKop CRD definition)
the Names in `k8s.remote` must match the names of the secret exported with the [kubemcsa](#bootstrap-api-access-to-k8s-cluster-2-from-k8s-cluster-1) command
:::

### Create the MultiCassKop CRD

You can deploy a MultiCassKop CRD instance.

You can create the Cluster with the following example [multi-casskop/samples/multi-casskop.yaml](https://github.com/Orange-OpenSource/casskop/tree/master/multi-casskop/samples/multi-casskop.yaml) file :

```
kubectl apply -f multi-casskop/samples/multi-casskop.yaml
```

This is the sequence of operations:
- MultiCassKop first creates the CassandraCluster in k8s-cluster1. 
- Then local CassKop starts to creates the associated Cassandra Cluster.
  - When CassKop has created its Cassandra cluster, it updates CassandraCluster object's status with the phase=Running meaning that
  all is ok
- Then MultiCassKop start creating the other CassandraCluster in k8s-cluster2
- Then local CassKop started to creates the associated Cassandra Cluster.
  - Thanks to the routable seed-list configured with external dns names, Cassandra pods are started by connecting to
    already existings Cassandra nodes from k8s-cluster1 with the goal to form a uniq Cassandra Ring.

In resulting, We can see that each clusters have the required pods.

If we go in one of the created pods, we can see that nodetool see pods of both clusters:

```
cassandra@cassandra-e2e-dc1-rack2-0:/$ nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address         Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.100.146.150  93.95 KiB  256          49.8%             cfabcef2-3f1b-492d-b028-0621eb672ec7  rack2
UN  10.100.146.108  108.65 KiB  256          48.3%             d1185b37-af0a-42f9-ac3f-234e541f14f0  rack1
Datacenter: dc2
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address         Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.100.151.38   69.89 KiB  256          51.4%             ec9003e0-aa53-4150-b4bb-85193d9fa180  rack5
UN  10.100.150.34   107.89 KiB  256          50.5%             a28c3c59-786f-41b6-8eca-ca7d7d14b6df  rack4
```