---
id: 3_multi_casskop
title: Multi-CassKop
sidebar_label: Multi-CassKop
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

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


```bash
helm install  multi-casskop orange-incubator/multi-casskop --set k8s.local=k8s-cluster1 --set k8s.remote={k8s-cluster2}
```

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