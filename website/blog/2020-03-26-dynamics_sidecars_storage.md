---
id: dynamics_sidecars_storage
title: Casskop 0.5.1 - Dynamic sidecars and storage configuration feature
author: Alexandre Guitton
author_title: Alexandre Guitton
author_url: https://github.com/erdrix
author_image_url: https://avatars0.githubusercontent.com/u/10503351?s=460&u=ea08d802388c79c17655c314296be58814391572&v=4
tags: [casskop, cassandra, 0.5.2, sidecars, storage]
---


In a previous post, I was talking about how [Setting up Cassandra Multi-Site on Google Kubernetes Engine with Casskop](/casskop/blog/2020/01/15/multicasskop_gke).
Since then, two new versions [0.5.1](https://github.com/Orange-OpenSource/casskop/releases/tag/v0.5.1-release) and [0.5.2](https://github.com/Orange-OpenSource/casskop/releases/tag/v0.5.2-release) had been released.
In another post, Cyril Scetbon focused on the [New Probes feature](https://medium.com/@cscetbon/new-probes-in-casskop-0-5-1-bfd1d6547967) which was added with the [PR #184Ø(https://github.com/Orange-OpenSource/casskop/pull/184), in this post I will focus on the dynamic sidecars and storage configurations added to the operator, which give more flexibility to users to configure their Cassandra cluster deployments.

## Purposes

During our production migration from bare metal Cassandra Cluster to Kubernetes, the main challenge was to perform the smoothest transition for our OPS teams, allowing them to reuse their homemade tools, to facilitate the cluster operationalization. 
However, the operator in this previous form did not leave much room for tuning statefulset and therefore the Cassandra Cluster deployed. 
You could use the bootstrap image to customize your cassandra node configuration, but not for the tools revolving around. 
That is why we added to the **CassandraCluster** the possibility to define containers into the pod in addition to the cassandra ones, these are the **sidecars**, and to configure extract storage for the pods (ie **VolumeClaimTemplates** to the Statefulset configuration).

## Dynamics sidecars configurations

To keep the [container’s best practices](https://cloud.google.com/blog/products/gcp/7-best-practices-for-building-containers) and address our OPS needs, we added the ability to define a dynamic list of containers into a **CassandraCluster.Spec** resource definition: [cassandracluster_types.go#L803](https://github.com/Orange-OpenSource/casskop/blob/master/pkg/apis/db/v1alpha1/cassandracluster_types.go#L803).

```yaml 
spec:
  ...
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
  ...
```

These sidecars are classic [kubernetes container resources](https://godoc.org/k8s.io/api/core/v1#Container), leaving you the full power on what you want to do. 
With this example, we add two simple sidecars allowing to distinguish cassandra and GC logs in two different stdout.

:::note
with this feature you can do everything you want, and obviously some bad things. This feature is not here to make a Cassandra Cluster works, the operator has everything for this, but to allow you to simplify some add-ons usage around Cassandra.
:::

## Sidecars : environment variables

All sidecars added with this configuration will have, at the container init, some of the environment variables from **cassandra container** merged with those defined into the sidecar container

- CASSANDRA_CLUSTER_NAME
- CASSANDRA_SEEDS
- CASSANDRA_DC
- CASSANDRA_RACK

## Storage configuration

In the previous version, the only option about storage was the [data volume configuration](/casskop/docs/3_configuration_deployment/3_storage) allowing you to define :

- `dataCapacity`: Defines the size of the persistent volume claim, for example, "1000Gi".
- `dataStorageClass`: Defines the type of storage to use (or use default one). We recommend to use local-storage for better performances but it can be any storage with high ssd throughput.

The dynamic sidecar doesn’t really suit, unless you put everything in one folder.

:::warning Spoiler alert
It’s not a good idea
:::

That is why we add the `CassandraCluster.Spec.StorageConfig` field, to the `CassandraCluster` resource definition :

```yaml
spec:
 ...
 storageConfigs:
   - mountPath: "/var/lib/cassandra/log"
     name: "gc-logs"
     pvcSpec:
       accessModes:
         - ReadWriteOnce
       storageClassName: local-storage
       resources:
         requests:
           storage: 5Gi
   - mountPath: "/var/log/cassandra"
     name: "cassandra-logs"
     pvcSpec:
       accessModes:
         - ReadWriteOnce
       storageClassName: local-storage
         resources:
           requests:
             storage: 10Gi
 ...
```

`storageConfigs` : Defines the list of storage config object, which will instantiate `Persitence Volume Claim` and associate volume to pod of cassandra node.

- `mountPath`: Defines the path into cassandra container where the volume will be mounted.
- `name`: Used to define the PVC and VolumeMount names.
- `pvcSpec`: pvcSpec describes the PVC used for the mountPath described above, it requires a kubernetes PVC spec.

In this example, we add the two volumes required by our sidecars previously configured, to be able via the sidecars to access to the logs that we want to expose on the stdout.

## Volume Claim Template and statefulset

Keep in mind that Casskop operator works on Statefulset, but have some constraints such as :

```log
updates to statefulset spec for fields other than 'replicas', 'template', and 'updateStrategy' are forbidden.
```

So if you want to add or remove some storages configurations, today you have to perform manually it, by removing the Statefulset, which will be recreated by the operator.

:::note
It’s not a sake operation, and should be performed carefully, because you will loose a rack. Maybe in some releases we will manage it, but today we assume that this operation is an exceptional one.
:::

[CassKop](https://github.com/Orange-OpenSource/casskop) is open source so don’t hesitate to try it out, contribute by first trying to fix a discovered issue and let’s enhance it together!

In a next post, I will speak about the IP management into Casskop, and the [cross IPs issue](https://github.com/Orange-OpenSource/casskop/issues/170), so stay connected !