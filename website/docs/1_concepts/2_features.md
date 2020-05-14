---
id: 2_features
title: Features
sidebar_label: Features
---

To highligt some of the features we needed and were not possible with the operators available, please keep reading 

CassKop deals with Cassandra clusters on one datacenter. For multi-dacenters deployments, please use [Multi-Casskop](https://github.com/Orange-OpenSource/casskop/tree/master/multi-casskop) in addition to CassKop. This second operator is part of this same repository.

The following features are supported by CassKop:
- [x] Deployment of a C* cluster (rack or AZ aware)
- [x] Scaling up the cluster (with cleanup)
- [x] Scaling down the cluster (with decommission prior to Kubernetes scale down)
- [x] Pods operations (removenode, upgradesstable, cleanup, rebuild..)
- [x] Adding a Cassandra DC
- [x] Removing a Cassandra DC
- [x] Setting and modifying configuration files
- [x] Setting and modifying configuration parameters
- [x] Update of the Cassandra docker image
- [x] Rolling update of a Cassandra cluster
    - [x] Update of Cassandra version (including upgradesstable in case of major upgrade)
    - [x] Update of JVM
    - [x] Update of configuration
- [x] Rolling restart of a Cassandra rack
- [x] Stopping a Kubernetes node for maintenance
    - [x] Process a remove node (and create new Cassandra node on another Kubernetes node)
    - [x] Process a replace address (of the old Cassandra node on another Kubernetes node)
- [x] Manage operations on pods through CassKop plugin (cleanup, rebuild, upgradesstable, removenode..)
- [x] Monitoring (using Instaclustr Prometheus exporter to Prometheus/Grafana)
- [x] Use official Cassandra Image (configuration for Casskop is done through a bootstrap init-container)
- [x] Performing live Cassandra repairs through the use of [Cassandra reaper](http://cassandra-reaper.io/)
- [x] Pause/Restart & rolling restart operations through CassKoP plugin.