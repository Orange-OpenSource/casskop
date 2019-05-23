

# CassKop Cassandra Kubernetes Operator Changelog

## 0.3.1

- GitHub open source version

## 0.2.1

- Add `spec.gcStdout` (default: true): to send gc logs to docker stdout
- Add `spec.topology.dc[].numTokens` (default: 256): to specify different number of vnodes for each DC
- Move RollingPartition from `spec.RollingPartition` to `spec.topology.dc[].rack[].RollingPartition 
- Add Cassandra psp config files in deploy

## 0.2.0

- add `spec.maxPodUnavailable` (default: 1): If there is pod unavailable in the ring casskop will refuse to make
change on statefulses. we can bypass this by increasing the maxPodUnavailable value. 

## 0.1.6

- Upgrade to Operator-sdk 0.2.0

## 0.1.5

- Upgrade to Operator-sdk 0.1.1

## 0.1.4

- Add and Remove DC
- Decommission now using JMX call instead of exec nodetool


## 0.1.3

### Features

- Configurable operator resyncPeriod via environment RESYNC_PERIOD
- No more uses of the Kubernetes subdomain in Cassandra Seeds --> Need Cassandra Docker Image > cassandra-3.11-v1.1.0
    - This also fixes the First node SeedList. we know via dns request if the first node exists or not, and if not it is the first creation of the cluster.
      So next times we can properly remove node1 from it's seedlist.
- Add new parameter `imagePullPolicy: "IfNotPresent"` to the CRD (default is "Always")
- Add `securityContext: runAsUser: 1000` to allow pod operator to launch with higher cluster security

### Fixes

- Fix [Issue 60](https://github.com/Orange-OpenSource/cassandra-k8s-operator/issues/60): Error when RollingUpdate on
  UpdateResource
- Fix [Issue 59](https://github.com/Orange-OpenSource/cassandra-k8s-operator/issues/59): Error on UpdateConfigMap
  vs UpdateStatefulset

## 0.1.2

### Features

- [x] **SeedList Management**
  - new param `AutoUpdateSeedList` which defines if operator need to automatically compute and apply best seedList

- [x] **CRD Improvement** : 
  - CRD protection against forbidden changed in the CRD. the operator now **refuses to change**:
    - the `dataCapacity`
    - the `dataStorageClass`
  - We can now specify/surcharge the global `nodesPerRack` in each DC section
- [x] **Better Status Management**
  - [x] Add Cluster level status to have a global view of whole cluster (composed of several statefulsets)
    - lastClusterAction
    - lastClusterActionStatus
    >Thoses status are used to know there is an ongoing action at cluster level, and that enables for instance to
    >completely finish an ScaleUp on all Racks, before executing PodLevel actions such as NodeCleanup.

  - [x] Add new status :
    - `UpdateResources` - if we change requested Pod resources in the CRD
    - `UpdateSeedList`- when the operator need to make rolling Update to apply new seedlist
      > We won't update the Seedist if not All Racks are staged to this modification (no other actions ongoing on the cluster)

- [x] Add `ImagePullSecret` parameter in the CRD to allow provide docker credentials to pull images 

- [x] **SeedList Management**
  - [x] SeedList Initialisation before Startup: We try (if available) to take 1 seed in each rack for each DC
  - [x] the Operator will try to apply the best SeedList in case of cluster topology evolution (Scaling, Add DC/Racks..)
  - [x] The Operator will make a Rolling Update (see nes status above)
  - [x] The DFY Cassandra Image in couple with the Operator will make that a Pod in the SeedList will be removed from it's own seedList.
  >Limitation: The first Pod of the cluster will be in it's own SeedList

>	- We can manually update the SeedList on the CRD Object, this will RollingUpdate each statefulset sequentially starting with the First

- [x] **Operator Debug**
  - [x] Allow specific `docker-build-debug` target in the Makefile and in the Pipeline to build debug version of the operator
    - debug version of go application
    - debug version of Image docker
    - debug version of helm chart (see below)

-  [x] **Helm Chart Improvment**
  - Add Possibility to use images behind authentication (imagePullSecret)
```
  imagePullSecrets:
    enabled: true
    name: <name of your docker registry secret>
```    
  - New way to define Debug Image and delve version API to uses
```
debug:
  enabled: false
  image:
    repository: orangeopensource/cassandra-k8s-operator
    tag: 0.1.2-debug
  version: 2
```

### Fixes

- [x] When NodeCleanup encounters some errors, we can see the status in the CassandraCluster
- [x] Fix Bug #53: Error which prevent PVC to be Deleted when CRD is delete and `deletePVC` flag is true
- [x] Fix Bug #52: The cluster was not deploying if Topology was empty


## 0.1.0

- [x] Rack Aware Deployment
    - [x] Add Topology section to declare Cassandra DC and Racks on their deployment ysing kubernetes nodes labels
    - [x] **Note:** Rename of `nodes` to `nodesPerRacks` in the CRD yaml file
- [x] add `hardAntiAffinity` flag in CRD to manage if we allow only 1 Cassandra Node per Kubernetes Nodes.
>**Limitation:** This parameter check only for Pods in the same kubernetes Namespace!!
- [x] add `deletePVC` flag in CRD to allow to delete all PersistentVolumesClaims in case we delete the cluster

- [x] Uses Jolokia for nodetool cleanup operation
- [x] Add `autoPilot` flag in CRD to enable to automatically execute Pod Operation **cleanup** after a ScaleUp, 
  or to allow to do the Operation manually by editing Pods Labels status to **Manual** to **ToDo**

## 0.0.7

### Features

- [x] Rack Aware Deployment
    - [x] Pod level get infos for Rack & DC. [MR #33](https://github.com/Orange-OpenSource/cassandra-k8s-operator/merge_requests/33)
        - Exposes **CASSANDRA_RACK** env var in the Pod from `cassandraclusters.db.orange.com.rack` Pod Labels
        - Exposes **CASSANDRA_DC** env var in the Pod from `cassandraclusters.db.orange.com.dc` Pod Labels
    
- [x] Make Uses of OLM (Operator Lifecycle Management) to manage the Operator

### Fixes

- [x] #25: change declaration of local-storage in PersistentVolumeClaim

## 0.0.6

### Features
- [x] Upgrade Operator SDK version to latest master (revision=a719b04752a51e5fe723467c7e66bc35830eb179)
- [x] Add start time and end time labels on Pods during Pod Actions
- [x] Add a Test on Operation Name for detecting an end in Cleanup Action
- in ensureDecommission
    - Re-Order Status in ensureDecommission
    - Add test on CassandraNode status to know if decommissioned is ongoing or not
    - Add asynchronous for nodetool decommission operation
- [x] Add Helm charts to deploy the operator
- [x] Add a Pod Disruption Budget which allows to have only 2 cassandra node down at a same time while working on the kubernetes cluster
- [x] Add a Jolokia client to interract with Cassandra

### Fixes
- [x] Remove old unused code
- [x] Add a test on the Pod Readiness before say ScaleUp is Done
- [x] Increase HealthCheck Periods and Timeouts
- [x] Add output messages in health checks requests for debug
- [x] Fix GetLastPod is number of pods > 10
- [x] Better management of decommission status (check with nodetool netstats to get node status), and adapt behaviour
- [x] On scale down, test Date on pod label to not execute several time nodetool decommission until status change from NORMAL to LEAVING

## 0.0.5

- Add test on field readyReplicas of the Statefulset to know operation is Done
- add sample directory for demo manifests.
- Add plantuml algorithm documentation
- If no dataCapacity is specified in the CRD, then No PersistentVolumeClaim is created
    - **WARNING** this is useful for dev but unsafe for production meaning that no datas will be persistent..
- Increase Timeout for HealthCheck Status from 5 to 40 and add PeriodSeconds to 50 between each healthcheck
- remove `nodetool drain` from the PreStop instruction
- Add PodDisruptionBudget with MaxUnavailable=2

## 0.0.4

- Initial version port from cassandra-kooper-operator propject
