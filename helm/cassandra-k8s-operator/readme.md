
# Cassandra Operator

This Helm chart install the Cassandra Operator to create/configure/manage a Cassandra 
cluster in a Kubernetes Namespace.
It will uses a Custom Ressource Definition CRD: `cassandraclusters.db.orange.com`, 
which implements a `CassandraCluster` kubernetes Object.


## Introduction


### Configuration

The following tables lists the configurable parameters of the Cassandra Operator Helm chart and their default values.


| Parameter          | Description                                      | Default                                                             |
|--------------------|--------------------------------------------------|---------------------------------------------------------------------|
| `image.repository` | Image                                            | `orangeopensource/cassandra-k8s-operator` |
| `image.tag`        | Image tag                                        | `0.0.6`                                                             |
| `image.pullPolicy` | Image pull policy                                | `Always`                                                            |
| imagePullSecrets   | Name of the secret to connect to docker registry | -                                                                   |
| `rbacEnable`       | If true, create & use RBAC resources             | `true`                                                              |
| `resources`        | Pod resource requests & limits                   | `{}`                                                                |
| `metricService`    | deploy service for metrics                       | `false`                                                             |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install --name my-release ./helm/cassandra-k8s-operator -f values.yaml
```

### Installing the Chart

You can make a dry run of the chart before deploying :

```console 
helm install --dry-run --debug ./helm/cassandra-k8s-operator --set debug.enabled=true --name cassandra-k8s-operator
```

To install the chart with the release name my-release:

```console
$ helm install --name my-release ./helm/cassandra-k8s-operator
```

We can surcharge default parameters using `--set` flag :

```console
$ helm install --replace --set image.tag=asyncronous --name my-release ./helm/cassandra-k8s-operator
```

> the `-replace` flag allow you to reuses a charts release name


### Listing deployed charts

```
helm list
```

### Get Status for the helm deployment :

```
helm status cassandra-k8s-operator

```

## Cleaning

### Uninstalling the Chart

To uninstall/delete the my-release deployment:

> !! Be Careful: Uninstalling the Cart, will delete all ressources created by the Charts, 
that means it will delete all your cassandra Clusters

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

However, Helm always keeps records of what releases happened. Need to see the deleted releases? `helm list --deleted` shows those, and `helm list --all` shows all of the releases (deleted and currently deployed, as well as releases that failed):

Because Helm keeps records of deleted releases, a release name cannot be re-used. (If you really need to re-use a release name, you can use the `--replace` flag, but it will simply re-use the existing release and replace its resources.)

Note that because releases are preserved in this way, you can rollback a deleted resource, and have it re-activate.



To purge a release
```console
$ helm delete --purge my-release
```


