---
id: 0_customizable_install_with_helm
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
| `image.repository`               | Image                                            | `orangeopensource/casskop` |
| `image.tag`                      | Image tag                                        | `0.5.1-master`                            |
| `image.pullPolicy`               | Image pull policy                                | `Always`                                  |
| `image.imagePullSecrets.enabled` | Enable the use of secret for docker image        | `false`                                   |
| `image.imagePullSecrets.name`    | Name of the secret to connect to docker registry | -                                         |
| `rbacEnable`                     | If true, create & use RBAC resources             | `true`                                    |
| `resources`                      | Pod resource requests & limits                   | `{}`                                      |
| `metricService`                  | deploy service for metrics                       | `false`                                   |
| `debug.enabled`                  | activate DEBUG log level                         | `false`                                   |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
helm install --name casskop incubator/cassandra-operator -f values.yaml
```

### Installing the Chart

:::important Helm 3 users
You need to manually install the crds beforehand

```console
kubectl apply -f https://github.com/Orange-OpenSource/casskop/blob/master/deploy/crds/db.orange.com_cassandraclusters_crd.yaml
```
:::

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
helm install --name casskop orange-incubator/casskop
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
helm delete casskop
```

The command removes all the Kubernetes components associated with the chart and deletes the helm release.

> The CRD created by the chart are not removed by default and should be manually cleaned up (if required)

Manually delete the CRD:

```bash
kubectl delete crd cassandraclusters.dfy.orange.com
```

:::warning
If you delete the CRD then :
It will delete **ALL** Clusters that has been created using this CRD!!!
Please never delete a CRD without very very good care
:::

Helm always keeps records of what releases were installed. Need to see the deleted releases? `helm list --deleted`
shows those, and `helm list --all` shows all of the releases (deleted and currently deployed, as well as releases that
failed)

Because Helm keeps records of deleted releases, a release name cannot be re-used. (If you really need to re-use a
release name, you can use the `--replace` flag, but it will simply re-use the existing release and replace its
resources.)

Note that because releases are preserved in this way, you can rollback a deleted resource, and have it re-activate.

To purge a release

```console
helm delete --purge casskop
```

## Troubleshooting

### Install of the CRD

By default, the chart will install the Casskop CRD via a helm hook but this installation is global for the whole
cluster.You may then deploy a chart with an existing CRD already deployed. In that case you can get an error like :

```bash
$ helm install --name casskop ./helm/cassandra-operator
Error: customresourcedefinitions.apiextensions.k8s.io "cassandraclusters.db.orange.com" already exists
```

In this case there si a parameter to say not to use the hook to install the CRD :

```bash
helm install --name casskop ./helm/cassandra-operator --no-hooks
```
