---
id: 1_getting_started
title: Getting Started
sidebar_label: Getting Started
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

CassKop uses the standard Cassandra image (tested up to Version 3.11), and can run on Kubernetes 1.13.3+.


As a pre-requisite it needs :

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) version v1.13.3+.
- [Helm](https://helm.sh/) version v2.12.2+.
- Access to a Kubernetes v1.13.3+ cluster.
- Cassandra needs fast local storage (we have tested with local storage provisioner, GKE ssd storage, and Rancher local-path-provisioner). Fast remote storage should work but is not tested yet.

## CassKop deployment

### Kubernetes preparation

First, we need to create a `namespace` in order to host our operator & cluster (here we call it cassandra)

```bash
kubectl create namespace cassandra
```

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
# You have to create the namespace before executing following commands
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/casskop/master/deploy/crds/db.orange.com_cassandraclusters_crd.yaml 
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/casskop/master/deploy/crds/db.orange.com_cassandrabackups_crd.yaml 
kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/casskop/master/deploy/crds/db.orange.com_cassandrarestores_crd.yaml

helm install casskop --namespace=cassandra orange-incubator/cassandra-operator
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
helm install --name=casskop --namespace=cassandra orange-incubator/cassandra-operator
```

</TabItem>
</Tabs>

You can find more information in the [Customizable install with helm](/casskop/docs/3_configuration_deployment/1_customizable_install_with_helm).

If you have problem you can see [Troubleshooting](/casskop/docs/7_troubleshooting/1_operations_issues) section

### Deploy a Cassandra cluster

#### Deploy a ConfigMap

Before we can deploy our cluster, we need to create a configmap.
This configmap will enable us to customize Cassandra's behaviour.
More details on this can be found [here](/casskop/docs/3_configuration_deployment/2_cassandra_configuration#configuration-override-using-configmap)

But for our example we will use the simple example:

```bash
kubectl apply -f samples/cassandra-configmap-v1.yaml
```

#### Deploy CassandraCluster resource

Once the operator is deployed inside a Kubernetes cluster, a new API will be accessible, so
you'll be able to create, update and delete cassandraclusters.

In order to deploy a new cassandra cluster a [specification](https://github.com/Orange-OpenSource/casskop/blob/master/samples/cassandracluster.yaml) has to be created. As an example :

``` bash
kubectl apply -f samples/cassandracluster.yaml
```

### Monitoring

We can quickly setup monitoring for our deployed Cassandra nodes using
Prometheus operator.

#### Deploy Prometheus

You can deploy the CoreOs Prometheus operator on your cluster:
You can find example [helm value.yaml](https://github.com/Orange-OpenSource/casskop/blob/master/samples/prometheus-values.yaml) to configure the Prometheus operator:

```console
kubectl create namespace monitoring
helm install --namespace monitoring prometheus-monitoring stable/prometheus-operator \
    --set prometheusOperator.createCustomResource=false \
    --set grafana.image.tag=7.0.1 \
    --set grafana.plugins="{briangann-gauge-panel,grafana-clock-panel,grafana-piechart-panel,grafana-polystat-panel,savantly-heatmap-panel,vonage-status-panel}"
```

#### Add ServiceMonitor for Cassandra

Then you have to create ServiceMonitor objects to monitor Cassandra nodes and CassKop. You can update this to specify 
which namespace to monitor (Your namespace need to be listed in the `namespaceSelector` section.)

```kubectl  apply -f monitoring/servicemonitor/```


#### Add Grafana dashboard for Cassandra

You can use our dashboard that monitors both Cassandra nodes and CassKop by running:
```kubectl  apply -f monitoring/dashboards/```