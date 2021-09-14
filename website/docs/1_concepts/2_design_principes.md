---
id: 2_design_principes
title: Design Principes
sidebar_label: Design Principes
---

## Statefulset level management

Cassandra is a stateful application. The first piece of the puzzle is the Node, which is a simple server capable of creating/forming a cluster with other Nodes. 

All Cassandra on Kubernetes setup use [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) to create a Cassandra Cluster. Just to quickly recap from the K8s docs:

:::note
StatefulSet manages the deployment and scaling of a set of Pods, and provide guarantees about their ordering and uniqueness. Like a Deployment, a StatefulSet manages Pods that are based on an identical container spec. Unlike a Deployment, a StatefulSet maintains sticky identities for each of its Pods. These pods are created from the same spec, but are not interchangeable: each has a persistent identifier that is maintained across any rescheduling.
:::

How does this looks from the perspective of Apache Cassandra ?

With StatefulSet we get:
- unique Node IDs generated during Pod startup
- networking between Nodes with headless services
- unique Persistent Volumes for Nodes

The Orange Casskop Operator uses StatefulSet.

## Communication with Casssandra's nodes

CassKop doesn't use [nodetool](https://docs.datastax.com/en/archived/cassandra/3.0/cassandra/tools/toolsNodetool.html) but invokes operations through authenticated [JMX/Jolokia](https://jolokia.org/) call, in this way, the communication result in only HTTP calls.

:::note
We are currently looking to provide an alternative to use the [Management API](https://github.com/datastax/management-api-for-apache-cassandra) from Datastax as a sidecar, instead of Jolokia.
:::

## Multi site management

Multi-CassKop goal is to bring the ability to deploy a Cassandra cluster within different regions, each of them running an independant Kubernetes cluster. Multi-Casskop insure that the Cassandra nodes deployed by each local CassKop will be part of the same Cassandra ring by managing a coherent creation of CassandraCluster objects from it's own MultiCasskop custom ressource.

![multi casskop design](/img/1_concepts/multi-casskop.png)

MultiCassKop starts by iterrating on every contexts passed in parameters then it register the controller. The controller needs to be able to interract with MultiCasskop and CassandraCluster CRD objetcs. In addition the controller needs to watch for MultiCasskop as it will need to react on any changes that occured on thoses objects for the given namespace.
