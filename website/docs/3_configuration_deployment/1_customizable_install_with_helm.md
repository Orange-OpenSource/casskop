---
id: 1_customizable_install_with_helm
title: Customizable install with Helm
sidebar_label: Customizable install with Helm
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

This Helm chart installs CassKop to create/configure/manage Cassandra
clusters in a Kubernetes Namespace.
It will use a Custom Ressource Definition(CRD): `cassandraclusters.db.orange.com`
which implements a `CassandraCluster` object in Kubernetes.

## Introduction

### Configuration

The following tables lists the configurable parameters of the Cassandra Operator Helm chart and their default values.

| Parameter                        | Description                                      | Default                                   |
|----------------------------------|--------------------------------------------------|-------------------------------------------|
| `image.repository`               | Image                                            | `orangeopensource/casskop`                |
| `image.tag`                      | Image tag                                        | `v2.0.1-release`                          |
| `image.pullPolicy`               | Image pull policy                                | `Always`                                  |
| `image.imagePullSecrets.enabled` | Enable the use of secret for docker image        | `false`                                   |
| `image.imagePullSecrets.name`    | Name of the secret to connect to docker registry | -                                         |
| `createCustomResource`           | If true, create & deploy the CRD                 | `true`
| `rbacEnable`                     | If true, create & use RBAC resources             | `true`                                    |
| `resources`                      | Pod resource requests & limits                   | `{requests: {cpu: 10m, memory: 50Mi}, limits: {cpu: 1,memory: 512Mi}`               |
| `metricService`                  | deploy service for metrics                       | `false`                                   |
| `debug.enabled`                  | activate DEBUG log level  and enable shareProcessNamespace (allowing ephemeral container usage)              | `false`                                   |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
helm install --name casskop incubator/cassandra-operator -f values.yaml
```

### Installing the Chart


<Tabs
  defaultValue="dryrun"
  values={[
    { label: 'dry run', value: 'dryrun', },
    { label: 'release name', value: 'rn', },
  ]
}>
<TabItem value="dryrun">

```bash
helm install --dry-run \
    --debug.enabled orange-incubator/casskop \
    --set debug.enabled=true \
    --name casskop
```

</TabItem>
<TabItem value="rn">

```bash
helm install casskop orange-incubator/casskop
```

</TabItem>

</Tabs>

> the `-replace` flag allow you to reuse a charts release name

### Listing deployed charts

```bash
helm list
```

### Get status for the helm deployment

```bash
helm status casskop
```

## Install another version of the operator

To install another version of the operator use:

```bash
helm install --name=casskop --namespace=cassandra --set operator.image.tag=x.y.z orange-incubator/casskop`
```

where x.y.z is the version you want.

## Uninstaling the Charts

If you want to delete the operator from your Kubernetes cluster, the operator deployment should be deleted.

```bash
helm uninstall casskop
```

The command removes all the Kubernetes components associated with the chart and deletes the helm release.

> The CRDs created by the chart are not removed by default and should be manually cleaned up (if required)

Manually delete the CRDs:

```bash
kubectl delete crd cassandraclusters.db.orange.com
kubectl delete crd cassandrabackups.db.orange.com
kubectl delete crd cassandrarestores.db.orange.com
```

:::warning
If you delete the CRDs then it will delete **ALL** Clusters that has been created using these CRDs!!!
Please never delete CRDs without very very good care
:::

## Troubleshooting

### Install of the CRD

By default, the chart will install the Casskop CRDs if there are not yet installed. If you want to upgrade or downgrade to another charts version you will need
to delete the CRDs BEFORE installing the new chart. If you don't want to install CRDs with the chart using Helm, you can skip this step by adding `--skip-crds` as described
in [Helm 3 official documentation](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/).

```bash
helm install casskop orange-incubator/cassandra-operator --skip-crds
```
