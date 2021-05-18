---
id: 2_cassandra_cluster
title: Cassandra Cluster
sidebar_label: Cassandra Cluster
---

The full schema of the `CassandraCluster` resource is described in the [Cassandra Cluster CRD Definition](#cassandra-cluster-crd-definition-version-020).

All labels that are applied to the desired `CassandraCluster` resource will also be applied to the Kubernetes resources
making up the Cassandra cluster. This provides a convenient mechanism for those resources to be labelled in whatever way
the user requires.

For every deployed container, CassKop allows you to specify the resources which should be reserved for it
and the maximum resources that can be consumed by it. We support two types of resources:

- Memory
- CPU

CassKop is using the Kubernetes syntax for specifying CPU and memory resources.

## Resource limits and requests

Resource limits and requests can be configured using the `resources` property in `CassandraCluster.spec.resources`.

### Resource requests

Requests specify the resources

:::important
If the resource request is for more than the available free resources on the scheduled kubernetes node,
the pod will remain stuck in "pending" state until the required resources become available.
:::

```yaml
# ...
resources:
  requests:
    cpu: 12
    memory: 64Gi
# ...
```

### Resource limits

Limits specify the maximum resources that can be consumed by a given container. The limit is not reserved and might not
be always available. The container can use the resources up to the limit only when they are available. The resource
limits should be always higher than the resource requests. If you only set limits, k8s uses the same value to set
requests.

```yaml
# ...
resources:
  limits:
    cpu: 12
    memory: 64Gi
# ...
```

### Supported CPU formats

CPU requests and limits are supported in the following formats:

- Number of CPU cores as integer (`5` CPU core) or decimal (`2.5`CPU core).
- Number of millicpus / millicores (`100m`) where 1000 millicores is the same as `1` CPU core.

For more details about CPU specification, refer to
[kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu)

### Supported memory formats

Memory requests and limits are specified in megabytes, gigabytes, mebibytes, gibibytes.

- to specify memory in megabytes, use the `M` suffix. For example `1000M`.
- to specify memory in gigabytes, use the `G` suffix. For example `1G`.
- to specify memory in mebibytes, use the `Mi` suffix. For example `1000Mi`.
- to specify memory in gibibytes, use the `Gi` suffix. For example `1Gi`.

For more details about CPU specification, refer to
[kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-memory)

## Configuring resource requests and limits

the resources requests and limits for CPU and memory will be applied to all Cassandra Pods deployed in the Cluster.

It is configured directly in the `CassandraCluster.spec.resources`:

```yaml
  resources:
    requests:
      cpu: '1'
      memory: 1Gi
    limits:
      cpu: '2'
      memory: 2Gi
```

Depending on the values specified, Kubernetes will define 3 levels for QoS : (BestEffort < Burstable < Guaranteed).

- BestEffort: if no resources are specified
- Burstable: if limits > requests. if a system needs more resources, thoses pods can be terminated if they use more than
  requested and if there is no more BestEffort Pods to terminated
- Guaranteed: request=limits. It is the recommended configuration for cassandra pods.

When updating the crd resources, this will trigger an [UpdateResources](/casskop/docs/5_operations/1_cluster_operations#updateresources) action.
