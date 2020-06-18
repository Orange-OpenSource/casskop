
# Multi-CassKop - Multi site Cassandra Kubernetes operator Helm chart

This Helm chart install Multi-CassKop the Orange's multi site Cassandra Kubernetes operator to create/configure/manage Cassandra 
clusters in a Kubernetes Namespace.

It will uses a Custom Ressource Definition CRD: `multicasscop.db.orange.com`, 
which implements a `MultiCasskop` kubernetes custom ressource definition.

This Operator main usage is to create `CassandraCluster` instances in several different Kubernetes cluster, so that you
can have a single Cassandra cluster spread on several Kubernetes clusters.

## Introduction


### Configuration

The following tables lists the configurable parameters of the Cassandra Operator Helm chart and their default values.


| Parameter                        | Description                                      | Default                                   |
|----------------------------------|--------------------------------------------------|-------------------------------------------|
| `image.repository`               | Image                                            | `orangeopensource/casskop` |
| `image.tag`                      | Image tag                                        | `0.3.1-master`                            |
| `image.pullPolicy`               | Image pull policy                                | `Always`                                  |
| `image.imagePullSecrets.enabled` | Enable tue use of secret for docker image        | `false`                                   |
| `image.imagePullSecrets.name`    | Name of the secret to connect to docker registry | -                                         |
| `rbacEnable`                     | If true, create & use RBAC resources             | `true`                                    |
| `resources`                      | Pod resource requests & limits                   | `{}`                                      |
| `metricService`                  | deploy service for metrics                       | `false`                                   |
| `debug.enabled`                  | activate DEBUG log level  and enable shareProcessNamespace (allowing ephemeral container usage)                        | `false`                                   |



Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

### Installing the Chart

Install a multi-casskop release :

```console
$ helm repo add orange-incubator https://orange-kubernetes-charts-incubator.storage.googleapis.com
$ helm install --name multi-casskop orange-incubator/multi-casskop
```

### Listing deployed charts

```
helm list
```

### Get Status for the helm deployment :

```
helm status multi-casskop

```

## Uninstaling the Charts

If you want to delete the operator from your Kubernetes cluster, the operator deployment 
should be deleted.

```
$ helm delete casskop
```
The command removes all the Kubernetes components associated with the chart and deletes the helm release.

> The CRD created by the chart are not removed by default and should be manually cleaned up (if required)

Manually delete the CRD:
```
kubectl delete crd multicasskop.dfy.orange.com
```

> **!!!!!!!!WARNING!!!!!!!!**
>
> If you delete the CRD then **!!!!!!WAAAARRRRNNIIIIINNG!!!!!!**
>
> It will delete **ALL** Clusters that has been created using this CRD!!!
>
> Please never delete a CRD without very very good care


Helm always keeps records of what releases happened. Need to see the deleted releases? `helm list --deleted`
shows those, and `helm list --all` shows all of the releases (deleted and currently deployed, as well as releases that
failed):

Because Helm keeps records of deleted releases, a release name cannot be re-used. (If you really need to re-use a
release name, you can use the `--replace` flag, but it will simply re-use the existing release and replace its
resources.)

Note that because releases are preserved in this way, you can rollback a deleted resource, and have it re-activate.



To purge a release
```console
$ helm delete --purge multi-casskop
```


## Troubleshooting

### Install of the CRD

By default, the chart will install via a helm hook the Casskop CRD, but this installation is global for the whole
cluster, and you may deploy a chart with an existing CRD already deployed.

In that case you can get an error like :


```
$ helm install --name multi-casskop orange-incubator/multi-casskop
Error: customresourcedefinitions.apiextensions.k8s.io "multicasskop.db.orange.com" already exists
```

In this case there si a parameter to say to not uses the hook to install the CRD :

```
$ helm install --name multi-casskop orange-incubator/multi-casskop --no-hooks
```
