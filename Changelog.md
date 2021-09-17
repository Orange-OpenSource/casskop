
# CassKop Cassandra Kubernetes Operator Changelog

## v2.0.2

- PR [363](https://github.com/Orange-OpenSource/casskop/pull/363) - Bump CRDs version from v1aplha1 to v2
- PR [362](https://github.com/Orange-OpenSource/casskop/pull/362) - Added a possibility to configure readinessProbe timeouts for operator deployment
- PR [361](https://github.com/Orange-OpenSource/casskop/pull/361) - Update Documentation
- PR [359](https://github.com/Orange-OpenSource/casskop/pull/359) - Do not mount multiple times the same path
- PR [358](https://github.com/Orange-OpenSource/casskop/pull/358) - Limit ConfigBuilder resources and add ability to specify its image
- PR [356](https://github.com/Orange-OpenSource/casskop/pull/356) - Bump tar from 6.1.4 to 6.1.11 in /website
- PR [354](https://github.com/Orange-OpenSource/casskop/pull/354) - Bump url-parse from 1.5.1 to 1.5.3 in /website 
- PR [353](https://github.com/Orange-OpenSource/casskop/pull/353) - Upgrade to Go 1.17
- PR [352](https://github.com/Orange-OpenSource/casskop/pull/352) - Bump path-parse from 1.0.6 to 1.0.7 in /website 
- PR [350](https://github.com/Orange-OpenSource/casskop/pull/350) - Remove pre_stop.sh and update build image
- PR [349](https://github.com/Orange-OpenSource/casskop/pull/349) - Bump tar from 6.0.5 to 6.1.4 in /website 


## v2.0.1

- PR [#344](https://github.com/Orange-OpenSource/casskop/pull/344) - Revert "Fix #316 by labeling the cluster uid to each PVC (#322)"
- PR [#343](https://github.com/Orange-OpenSource/casskop/pull/343) - Revert uid label
- PR [#341](https://github.com/Orange-OpenSource/casskop/pull/341) - Document how to upgrade from v1 to v2
- PR [#340](https://github.com/Orange-OpenSource/casskop/pull/340) - Fix operator & pods restart when scaling up DC with autoUpdateSeedList set to true
- PR [#337](https://github.com/Orange-OpenSource/casskop/pull/337) - Bump prismjs from 1.23.0 to 1.24.0 in /website
- PR [#336](https://github.com/Orange-OpenSource/casskop/pull/336) - Bump postcss from 7.0.35 to 7.0.36 in /website
- PR [#335](https://github.com/Orange-OpenSource/casskop/pull/335) - Scales up and down a DC in multi-dcs

## v2.0.0

- PR [#331](https://github.com/Orange-OpenSource/casskop/pull/331) - Bump ws from 6.2.1 to 6.2.2 in /website
- PR [#330](https://github.com/Orange-OpenSource/casskop/pull/330) - Do not crash if backup is created with non existing
  Datacenter
- PR [#246](https://github.com/Orange-OpenSource/casskop/pull/246) - Support of Cassandra 4.0
- PR [#329](https://github.com/Orange-OpenSource/casskop/pull/329) - Bump dns-packet from 1.3.1 to 1.3.4 in /website
- PR [#328](https://github.com/Orange-OpenSource/casskop/pull/328) - e2e tests use kuttl only

## v1.1.4

- PR [#325](https://github.com/Orange-OpenSource/casskop/pull/325) - Add datacenter to Restore and take it into accoun
  during backup/restore operations
- PR [#322](https://github.com/Orange-OpenSource/casskop/pull/322) - Label each PVC with the cluster uid
- PR [#319](https://github.com/Orange-OpenSource/casskop/pull/319) - Fix the way FSGroup and RunAsUser are used
- PR [#314](https://github.com/Orange-OpenSource/casskop/pull/314) - Added `fsGroup` to `CassandraCluster.Spec`
- PR [#313](https://github.com/Orange-OpenSource/casskop/pull/313) - Add allow annotations to be passed from
  Multicasskop
- PR [#307](https://github.com/Orange-OpenSource/casskop/pull/307) - Patch kuttl test to work with releases
- PR [#305](https://github.com/Orange-OpenSource/casskop/pull/305) - Bump prismjs from 1.22.0 to 1.23.0 in /website


## v1.1.3

- PR [#302](https://github.com/Orange-OpenSource/casskop/pull/302) - Fix Bootstrap issue
- PR [#303](https://github.com/Orange-OpenSource/casskop/pull/303) - Upgrade Icarus to 1.0.9


## v1.1.2

- PR [#298](https://github.com/Orange-OpenSource/casskop/pull/297) - Fix Helm publishing


## v1.1.1

- PR [#297](https://github.com/Orange-OpenSource/casskop/pull/297) - Fix sonar jdk 
- 
## v1.1.0

### Added

- PR [#285](https://github.com/Orange-OpenSource/casskop/pull/285) - Bump ini from 1.3.5 to 1.3.8 in /website
- PR [#286](https://github.com/Orange-OpenSource/casskop/pull/286) - Add the 2 new pages (Cassandra Backup & Restore) to the sidebar + bump version
- PR [#287](https://github.com/Orange-OpenSource/casskop/pull/287) - Restrict psp role in helm chart to get and list
- PR [#289](https://github.com/Orange-OpenSource/casskop/pull/289) - Do not use data folder to store downloaded sstables
- PR [#290](https://github.com/Orange-OpenSource/casskop/pull/290) - Enforce HARDLINKS restore strategy as it is faster
- PR [#291](https://github.com/Orange-OpenSource/casskop/pull/291) - Add check if cassandraBackup.Annotations map is nil before assignement
- PR [#293](https://github.com/Orange-OpenSource/casskop/pull/293) - Add option to Rename a table when restoring it
- PR [#294](https://github.com/Orange-OpenSource/casskop/pull/294) - Generate CRDs to deploy dir AND both helms CRDs dirs
- PR [#295](https://github.com/Orange-OpenSource/casskop/pull/295) - Upgrade Icarus to 1.0.8 (Fix error catching)

### Changed

- CassandraRestore CRD (restorationStrategyType removed, add rename option)

### Deprecated

### Removed

- v1beta1 CRDs


## v1.0.2

### Added

- PR [#283](https://github.com/Orange-OpenSource/casskop/pull/283) - Add Kuttl (declarative E2E test) implem 


### Changed

- CassandraBackup CRD (fixed cassandraCluster typo)

### Deprecated

### Removed

- v1beta1 CRDs

### Bug Fixed

- PR [#282](https://github.com/Orange-OpenSource/casskop/pull/282) - Fix helm 3 crds installation + add some docs 

## v1.0.1

### Added

- PR [#280](https://github.com/Orange-OpenSource/casskop/pull/280) - Upgrade Icarus to 1.0.5
- PR [#279](https://github.com/Orange-OpenSource/casskop/pull/279) - Add option to specify resources at DCs level
- PR [#278](https://github.com/Orange-OpenSource/casskop/pull/278) - Make separate liveness/readiness probes possible
- PR [#276](https://github.com/Orange-OpenSource/casskop/pull/276) - Upgrade Icarus and add errors to status
- PR [#275](https://github.com/Orange-OpenSource/casskop/pull/275) - Use static name for k3d cluster
- PR [#274](ttps://github.com/Orange-OpenSource/casskop/pull/274) - Fix website doc and add search ability

### Changed

### Deprecated

### Removed

### Bug Fixes

- PR [#273](https://github.com/Orange-OpenSource/casskop/pull/273) - **[CassandraCluster]** Fix Jolokia auth.

## v1.0.0

### Added

- PR [#258](https://github.com/Orange-OpenSource/casskop/pull/268)
- PR [#267](https://github.com/Orange-OpenSource/casskop/pull/267)
- PR [#260](ttps://github.com/Orange-OpenSource/casskop/pull/260)

### Changed
- PR [#265](https://github.com/Orange-OpenSource/casskop/pull/265)

### Deprecated

### Removed

### Bug Fixes

- PR [#263](https://github.com/Orange-OpenSource/casskop/pull/263)

## v0.5.6

### Added

### Changed

### Deprecated

### Removed

### Bug Fixes

- PR [#256](https://github.com/Orange-OpenSource/casskop/pull/256) - **[Chart]** Fix multi-casskop role
 
## v0.5.5

### Added

### Changed

### Deprecated

### Removed

### Bug Fixes

- PR [#252](https://github.com/Orange-OpenSource/casskop/pull/252) - **[Plugin]** Remove metadata.resourceVersion from the applied resource
- PR [#250](https://github.com/Orange-OpenSource/casskop/pull/250) - **[CassandraCluster]** Scale up node at a time

## 0.5.4

### Added

- PR [#233](https://github.com/Orange-OpenSource/casskop/pull/233) - **[CassandraCluster]** Add [ShareProcessNamespace](https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/) option for operator and cassandra nodes
- PR [#245](https://github.com/Orange-OpenSource/casskop/pull/245) - **[Chart]** Explicit roles needed by casskop

### Changed

- PR [#245](https://github.com/Orange-OpenSource/casskop/pull/245) - **[Chart]** Explicit roles
- PR [#240](https://github.com/Orange-OpenSource/casskop/pull/240) - **[Documentation]** Bump lodasg from 4.17.15 to 4.17.19
- PR [#242](https://github.com/Orange-OpenSource/casskop/pull/242) - **[Documentation]** Bump elliptic from 6.5.2 to 6.5.3
- PR [#244](https://github.com/Orange-OpenSource/casskop/pull/244) - **[Documentation]** Bump prismjs from 1.20.0 to 1.21.0

### Deprecated

### Removed

### Bug Fixes

- PR [#234](https://github.com/Orange-OpenSource/casskop/pull/234) - **[CassandraCluster]** Fix having pod to fail during decommissioning / joining, replacing liveness probe.
- PR [#235](https://github.com/Orange-OpenSource/casskop/pull/235) - **[CassandraCluster]** Fix multi decommissioning
- PR [#241](https://github.com/Orange-OpenSource/casskop/pull/241) - **[CassandraCluster]** Do not do more decommissions than needed
- PR [#247](https://github.com/Orange-OpenSource/casskop/pull/247) - **[CassandraCluster]** Update pre-stop bootstrap script

## 0.5.3

### Added

- PR [#203](https://github.com/Orange-OpenSource/casskop/pull/203) - **[CassandraCluster]** Data configuration at DC level
- PR [#215](https://github.com/Orange-OpenSource/casskop/pull/215) - **[CassandraCluster]** Default resources requirements for init containers


### Changed

- PR [#204](https://github.com/Orange-OpenSource/casskop/pull/204) - Fix sonar project
- PR [#217](https://github.com/Orange-OpenSource/casskop/pull/217) - **[Documentation]** Website documentation in replacement of MD folder
- PR [#225](https://github.com/Orange-OpenSource/casskop/pull/217) - **[CI/CD]** Use k3d instead of Minikube


### Deprecated

### Removed

### Bug Fixes

- PR [#205](https://github.com/Orange-OpenSource/casskop/pull/205) - **[MultiCasskop]** Non blocking unused Kubernetes cluster in `MultiCasskop` resources
- PR [#206](https://github.com/Orange-OpenSource/casskop/pull/206) - **[CassandraCluster]** Fix readiness & liveness probe configuration update detection
- PR [#220](https://github.com/Orange-OpenSource/casskop/pull/220) - **[Documentation]** Fix yarn.lock
- PR [#223](https://github.com/Orange-OpenSource/casskop/pull/223) - **[Documentation]** Fix chart name
- PR [#230](https://github.com/Orange-OpenSource/casskop/pull/230) - **[CassandraCluster]** Set boostratp env vars based on Cassandra resources


## 0.5.2

- PR [#201](https://github.com/Orange-OpenSource/casskop/pull/201) - Add liveness and readiness probe configurable in CassandraCluster object
- PR [#200](https://github.com/Orange-OpenSource/casskop/pull/200) - Catch nil pvcSpec error
- PR [#199](https://github.com/Orange-OpenSource/casskop/pull/199) - Fix [Issue #197](https://github.com/Orange-OpenSource/casskop/issues/197) helm release
- PR [#198](https://github.com/Orange-OpenSource/casskop/pull/198) - Add custom metrics to operator
- PR [#196](https://github.com/Orange-OpenSource/casskop/pull/196) - Fix [Issue #170](https://github.com/Orange-OpenSource/casskop/issues/170) cross ip
- PR [#195](https://github.com/Orange-OpenSource/casskop/pull/195) - Ensure generated deepcopy files are always up to date
- PR [#193](https://github.com/Orange-OpenSource/casskop/pull/193) - Fix [Issue #192](https://github.com/Orange-OpenSource/casskop/issues/192) Add check on container length for statefulset comparison

## 0.5.1

**Breaking Change in the bootstrap image**
See [Upgrade section](Readme.md#upgrade-casskop-015bootstrap-image-to-014)

- PR [#190](https://github.com/Orange-OpenSource/casskop/pull/190) - Fix [Issue #189](https://github.com/Orange-OpenSource/casskop/issues/189)  Handle volumemounts per container
- PR [#187](https://github.com/Orange-OpenSource/casskop/pull/187) - Fix helm repo [url](Readme.md#deploy-the-cassandra-operator-and-its-crd-with-helm)
- PR [#185](https://github.com/Orange-OpenSource/casskop/pull/185) - Add the support of [sidecars](documentation/description.md#sidecars-configuration)
- PR [#184](https://github.com/Orange-OpenSource/casskop/pull/184) - Use Jolokia calls instead of nodetool in readiness and liveness probes
- PR [#179](https://github.com/Orange-OpenSource/casskop/pull/179) - Fix [Issue #168](https://github.com/Orange-OpenSource/casskop/issues/168) Do not check toplogy in CassKop (does not work with MultiCassKop) during rebuild but using Cassandra 
- PR [#177](https://github.com/Orange-OpenSource/casskop/pull/177) - Add documentation on how to add [tolerations](documentation/description.md#using-dedicated-nodes)
- PR [#175](https://github.com/Orange-OpenSource/casskop/pull/175) - Fix dgoss tests
- PR [#174](https://github.com/Orange-OpenSource/casskop/pull/174) - Upgrade operator sdk
- PR [#173](https://github.com/Orange-OpenSource/casskop/pull/173) - Fix documentation
- PR [#167](https://github.com/Orange-OpenSource/casskop/pull/167) - Fix plugin remove command
- PR [#165](https://github.com/Orange-OpenSource/casskop/pull/165) - Fix OpenAPI v3.0 schema validation
- PR [#164](https://github.com/Orange-OpenSource/casskop/pull/164) - Rename repository
- PR [#163](https://github.com/Orange-OpenSource/casskop/pull/163) - Add documentation regarding the upgrade of the operator
- PR [#162](https://github.com/Orange-OpenSource/casskop/pull/162) - Adapt CI pipeline for multi-CassKop
- PR [#161](https://github.com/Orange-OpenSource/casskop/pull/161) - Fix helm chart
- PR [#157](https://github.com/Orange-OpenSource/casskop/pull/157) - Add logo for CassKop
- PR [#156](https://github.com/Orange-OpenSource/casskop/pull/156) - Watch only first cluster in MultiCassKop
- PR [#155](https://github.com/Orange-OpenSource/casskop/pull/155) - Refactor PodAffinityTerm
- PR [#153](https://github.com/Orange-OpenSource/casskop/pull/153) - Allow Istio to work with Cassandra and encrypt native connections
- PR [#152](https://github.com/Orange-OpenSource/casskop/pull/152) - Add GKE example


## 0.5.0

Introduce Multi-Casskop, the Operator to manage a single Cassandra Cluster above multiple Kubernetes clusters.

- PR [#145](https://github.com/Orange-OpenSource/casskop/pull/145) - Fix [Issue #142](https://github.com/Orange-OpenSource/casskop/issues/142) PodStatus which rarely fails in
  unit tests
- PR [#146](https://github.com/Orange-OpenSource/casskop/pull/146) - Fix [Issue #143](https://github.com/Orange-OpenSource/casskop/issues/143) External update of SeedList was not possible

- PR [#147](https://github.com/Orange-OpenSource/casskop/pull/147) - Introduce Multi-CassKop operator

- PR [#149](https://github.com/Orange-OpenSource/casskop/pull/149) - Get rid of env var SERVICE_NAME and
  keep current hostname in seedlist

- PR [#151](https://github.com/Orange-OpenSource/casskop/pull/151) - Fix [Issue #150](https://github.com/Orange-OpenSource/casskop/issues/150) Makes JMX port remotely available (again)
    - uses New bootstrap Image 0.1.3 : orangeopensource/cassandra-bootstrap:0.1.3


## 0.4.1

**Breaking Change in API**

The fields `spec.baseImage` and `spec.version` have been removed in favor for `spec.cassandraImage` witch is a merge of
both of thems.

- PR [#128](https://github.com/Orange-OpenSource/casskop/pull/128/files) Fix Issue
  [#96](https://github.com/Orange-OpenSource/casskop/issues/96): cluster stay pending
- PR [#127](https://github.com/Orange-OpenSource/casskop/pull/127) Fix Issue
  [#126](https://github.com/Orange-OpenSource/casskop/issues/126): update racks in parallel
- PR [#124](https://github.com/Orange-OpenSource/casskop/pull/124): Add Support for pod & services
  annotations
- PR [#138](https://github.com/Orange-OpenSource/casskop/pull/138) Add support for Tolerations

Examples of annotation needed in the CassandraCluster Spec:
```
  service:
    annotations:
      external-dns.alpha.kubernetes.io/hostname: my.custom.domain.com.

```

- PR [#119](https://github.com/Orange-OpenSource/casskop/pull/119) Refactoring Makefile
- tests now uses default cassandra docker image


## 0.4.0

- initContainerImage and bootstrapContainerImage used to adapt to official cassandra image.
- ReadOnly Container : `Spec.ReadOnlyRootFilesystem` default true

## 0.3.3

- upgrade to operator-sdk 0.9.0 & go modules (thanks @jsanda)

## 0.3.2

- Released version

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

- Fix [Issue 60](https://github.com/Orange-OpenSource/casskop/issues/60): Error when RollingUpdate on
  UpdateResource
- Fix [Issue 59](https://github.com/Orange-OpenSource/casskop/issues/59): Error on UpdateConfigMap
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
    repository: orangeopensource/casskop
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
    - [x] Pod level get infos for Rack & DC. [PR #33](https://github.com/Orange-OpenSource/casskop/merge_requests/33)
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
