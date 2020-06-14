---
id: 5_sidecars
title: Sidecars
sidebar_label: Sidecars
---

For extra needs not covered by the defaults container managed through the CassandraCluster CRD, we are allowing you to define your own sidecars which will be deployed into the cassandra node pods.
To do this, you will configure the `SidecarConfigs` property in `CassandraCluster.Spec`.

CassandraCluster fragment for dynamic sidecars definition :

```yaml
# ...
  sidecarConfigs:
    - args: ["tail", "-F", "/var/log/cassandra/system.log"]
      image: ez123/alpine-tini
      imagePullPolicy: Always
      name: cassandra-log
      resources:
        limits:
          cpu: 50m
          memory: 50Mi
        requests:
          cpu: 10m
          memory: 10Mi
      volumeMounts:
        - mountPath: /var/log/cassandra
          name: cassandra-logs
    - args: ["tail", "-F", "/var/log/cassandra/gc.log.0.current"]
      image: ez123/alpine-tini
      imagePullPolicy: Always
      name: gc-log
      resources:
        limits:
          cpu: 50m
          memory: 50Mi
        requests:
          cpu: 10m
          memory: 10Mi
      volumeMounts:
        - mountPath: /var/log/cassandra
          name: gc-logs
# ...
```

- `sidecarConfigs` *(required)* : Defines the list of container config object, which will be added into each pod of cassandra node, it requires a list of kubernetes Container spec.

With the above configuration, the following configuration will be added to the `rack statefulset` definition :

```yaml
# ...
  # ...
  containers:
    - args: ["tail", "-F", "/var/log/cassandra/system.log"]
      image: ez123/alpine-tini
      imagePullPolicy: Always
      name: cassandra-log
      resources:
        limits:
          cpu: 50m
          memory: 50Mi
        requests:
          cpu: 10m
          memory: 10Mi
      volumeMounts:
        - mountPath: /var/log/cassandra
          name: cassandra-logs
    - args: ["tail", "-F", "/var/log/cassandra/gc.log.0.current"]
      image: ez123/alpine-tini
      imagePullPolicy: Always
      name: gc-log
      resources:
        limits:
          cpu: 50m
          memory: 50Mi
        requests:
          cpu: 10m
          memory: 10Mi
      volumeMounts:
        - mountPath: /var/log/cassandra
          name: gc-logs
  # ...
# ...
```

:::info
Note that all sidecars added with this configuration will have some of the environment variables from cassandra container merged with those defined into the sidecar container
for example :

- CASSANDRA_MAX_HEAP
- CASSANDRA_SEEDS
- CASSANDRA_CLUSTER_NAME
- CASSANDRA_GC_STDOUT
- CASSANDRA_NUM_TOKENS
- CASSANDRA_DC
- CASSANDRA_RACK
:::
