---
id: 4_multicasskop
title: MultiCasskop
sidebar_label: MultiCasskop
---

`MultiCasskop` describes the desired state of the Cassandra cluster in a multi-site mode we want to setup through the operator.
 
 ```yaml
apiVersion: db.orange.com/v1alpha1
kind: MultiCasskop
metadata:
  name: multi-casskop-demo
spec:
  deleteCassandraCluster: true
  base: #<-- Specify the base of our CassandraCluster
    apiVersion: "db.orange.com/v1alpha1"
    kind: "CassandraCluster"
    metadata:
      name: cassandra-demo
      namespace: cassandra-demo
      labels:
        cluster: casskop
    spec:
      cassandraImage: orangeopensource/cassandra-image:3.11
      bootstrapImage: orangeopensource/cassandra-bootstrap:0.1.4
      configMapName: cassandra-configmap-v1
      service:
        annotations:
          external-dns.alpha.kubernetes.io/hostname: casskop.external-dns-test.gcp.trycatchlearn.fr.
      rollingPartition: 0
      dataCapacity: "20Gi"
      dataStorageClass: "standard-wait"
      imagepullpolicy: IfNotPresent
#      imagepullpolicy: Always
      hardAntiAffinity: false
      deletePVC: true
      autoPilot: false
      gcStdout: false
      autoUpdateSeedList: false
      debug: false
      maxPodUnavailable: 1
      nodesPerRacks: 1
      runAsUser: 999
      resources:
        requests: &requests
          cpu: '1'
          memory: 2Gi
        limits: *requests
    status:
      seedlist:   #<-- at this time the seedlist must be fullfilled manually with known predictive name of pods
        - cassandra-demo-dc1-rack1-0.casskop.external-dns-test.gcp.trycatchlearn.fr
        - cassandra-demo-dc3-rack3-0.casskop.external-dns-test.gcp.trycatchlearn.fr
        - cassandra-demo-dc4-rack4-0.casskop.external-dns-test.gcp.trycatchlearn.fr
        - cassandra-demo-dc4-rack4-1.casskop.external-dns-test.gcp.trycatchlearn.fr
  override: #<-- Specify overrides of the CassandraCluster depending on the target kubernetes cluster
    gke-master-west1-b:
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
    gke-slave-west1-d:
      spec:
        imagepullpolicy: IfNotPresent
        topology:
          dc:
            - name: dc3
              nodesPerRacks: 1
              numTokens: 256
              labels:
                failure-domain.beta.kubernetes.io/region: europe-west1
              rack:
                - name: rack3
                  labels:
                    failure-domain.beta.kubernetes.io/zone: europe-west1-d
            - name: dc4
              nodesPerRacks: 2
              numTokens: 256
              labels:
                failure-domain.beta.kubernetes.io/region: europe-west1
              rack:
                - name: rack4
                  labels:
                    failure-domain.beta.kubernetes.io/zone: europe-west1-d
 ```

## MultiCasskop

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|metadata|[ObjectMetadata](https://godoc.org/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta)|is metadata that all persisted resources must have, which includes all objects users must create.|No|nil|
|spec|[MultiCasskopSpec](#multicasskopspec)|defines the desired state of MultiCasskop.|No|nil|
|status|[MultiCasskopStatus](#multicasskopstatus)|defines the observed state of MultiCasskop.|No|nil|

## MultiCasskopSpec

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|
|deleteCassandraCluster|bool|If you have set to true, then when deleting the `MultiCassKop` object, it will cascade the deletion of the `CassandraCluster` object in the targeted k8s clusters. Then each local CassKop will delete their Cassandra clusters.|Yes|true|
|base|[CassandraCluster](/casskop/docs/6_references/1_cassandra_cluster#cassandracluster)|Define for all `CassandraCluster` the default configuration|Yes| - |
|override|map\[string\][CassandraCluster](/casskop/docs/6_references/1_cassandra_cluster#cassandracluster)|Define for each `CassandraCluster` a specific configuration not shared across all of them| Yes | -  |

## MultiCasskopStatus

|Field|Type|Description|Required|Default|
|-----|----|-----------|--------|--------|