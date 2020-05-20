---
id: 3_multi_casskop
title: Multi-CassKop
sidebar_label: Multi-CassKop
---

Here is describes how perform some operations based on the MultiCasskop Operator.

The MultiCasskop operator manage nothing else than CassandraCluster ressources. Today, this is not through it that you will manage cassandra operations (this is the duty of Cassskop operator).
So the only things we will be able to work with are the CassandraCluster's informations and Kubernetes Cluster's client used.

## Remove a Kubernetes site used in a Cassandra Ring

Performing a scale down at the MultiCasskop operator level, is by designed scaledown the number of CassandraCluster resource deployed, and so the number of Kubernetes Cluster clients used in the cassandra Ring.

To achieve this scaledown the following steps are required :

1.  First of all you need to remove all cassandra DC associated to the Kubernetes Cluster client that you want to remove of the `MultiCasskop` resource : 

    i - Ensure that there is no more data replicated on it. For example you can check and perform it in following this instructions : 
    
    ```sh
    cqlsh> DESCRIBE keyspace <keyspace name>
    
    CREATE KEYSPACE system_distributed WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': '1', 'dc2': '1', dc3': '1'}  AND durable_writes = true;
    cqls> ALTER KEYSPACE system_distributed WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': '1', 'dc3': '1'}  AND durable_writes = true;
    ```
    
    ii - Change decrease the number of node per rack for the DC to 0 : 
    
    ```yaml 
    ...
    site_a:
      spec:
        topology:
          dc:
            - name: dc1
              nodesPerRacks: 1
              numTokens: 256
              labels:
                failure-domain.beta.kubernetes.io/region: europe-west1
              rack:
                - name: rack1
                  rollingPartition: 0
                  labels:
                    failure-domain.beta.kubernetes.io/zone: europe-west1-b
    site_b:
      spec:
        topology:
          dc:
            - name: dc2
              nodesPerRacks: 0       <---- Downsize to 0
              numTokens: 256
              labels:
                failure-domain.beta.kubernetes.io/region: europe-west1
              rack:
                - name: rack1
                  rollingPartition: 0
                  labels:
                    failure-domain.beta.kubernetes.io/zone: europe-west1-c
    site_c:
      spec:
        topology:
          dc:
            - name: dc3
              nodesPerRacks: 1
              numTokens: 256
              labels:
                failure-domain.beta.kubernetes.io/region: europe-west1
              rack:
                - name: rack1
                  rollingPartition: 0
                  labels:
                    failure-domain.beta.kubernetes.io/zone: europe-west1-c
    ...
    ```
    
    iii -  This will perform the downscale of the nodes into the DC.
    
2 - Remove all DC from site list :

```yaml 
...
site_b:
  spec:
    topology:
      dc:
...
```

3 - Remove the site of `MultiCasskop` resource.
4 - Remove manually the `CassandraCluster` resource on the remote site.