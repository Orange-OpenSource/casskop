---
id: 1_introduction
title: Introduction
sidebar_label: Introduction
---

The Orange Cassandra operator is a Kubernetes operator to automate provisioning, management, autoscaling and operations of [Apache Cassandra](http://cassandra.apache.org/) clusters deployed to K8s.

## Overview

The CassKop Cassandra Kubernetes operator makes it easy to run Apache Cassandra on Kubernetes. Apache Cassandra is a popular, 
free, open-source, distributed wide column store, NoSQL database management system. 
The operator allows to easily create and manage racks and data centers aware Cassandra clusters.

Some of the high-level capabilities and objectives of Apache Cassandra include, and some of the main features of the **Casskop** are:

- Deployment of a C* cluster (rack or AZ aware)
- Graceful rolling update
- Graceful C* cluster **scaling** (with cleanup and decommission prior to Kubernetes scale down)
- Manage operations on pods through CassKop plugin (cleanup, rebuild, upgradesstable, removenode..)
- Performing live Cassandra repairs through the use of [Cassandra reaper](http://cassandra-reaper.io/)
- Multi-site management through [Multi-Casskop operator](https://github.com/Orange-OpenSource/casskop/tree/master/multi-casskop)
- Live Backup/Restore of Cassandra's datas

The Cassandra operator is based on the CoreOS
[operator-sdk](https://github.com/operator-framework/operator-sdk) tools and APIs.


CassKop creates/configures/manages Cassandra clusters atop Kubernetes and is by default **space-scoped** which means
that :
- CassKop is able to manage X Cassandra clusters in one Kubernetes namespace.
- You need X instances of CassKop to manage Y Cassandra clusters in X different namespaces (1 instance of CassKop
  per namespace).

:::info
This adds security between namespaces with a better isolation, and less work for each operator.
:::

## Presentation

We have some slides for a [CassKop demo](https://orange-opensource.github.io/casskop/slides/index.html?slides=Slides-CassKop-demo.md#1)

You can also play with CassKop on [Katacoda](https://www.katacoda.com/orange)

## Motivation

At [Orange](https://opensource.orange.com/fr/accueil/) we are building some [Kubernetes operator](https://github.com//Orange-OpenSource?utf8=%E2%9C%93&q=operator&type=&language=), that operate NiFi, Galera and Cassandra clusters (among other types) for our business cases.

There are already some approaches to operating C* on Kubernetes, however, we did not find them appropriate for use in a highly dynamic environment, nor capable of meeting our needs.

- [Datastax K8ssandra Cass-Operator](https://github.com/k8ssandra/cass-operator) (see also [K8ssandra project](https://k8ssandra.io)
- [Instaclustr Operator](https://github.com/instaclustr/cassandra-operator)
- [Sky-Uk Operator](https://github.com/sky-uk/cassandra-operator)

Finally, our motivation is to build an open source solution and a community which drives the innovation and features of this operator.
