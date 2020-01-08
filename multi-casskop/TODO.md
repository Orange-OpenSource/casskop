

## Features

- [x] **Feature01**: CassKop's special feature using the unlockNextOperation which allows to recover when CassKop fails,
  must be treated with really good care as this parameter is removed when activated by local CassKop. I think We must
  prevent to let this parameter be set at MultiCasskop level, and not to be removed from local CassandraCluster
  if it has been set up locally (remove from the difference detection)

- [ ] **Feature02**: Auto compute and update seedlist at MultiCassKop level

- [x] **Feature03**: Specify the namespace we want to deploy onto for each kubernetes contexts

- [x] **Feature04**: Allow to delete CassandraClusters when deleting MultiCasskop
                 Make uses of a Finalizer to keep track of last MultiCasskop before deleting

- [x] **Feature05**: Managing rollingUpdate changes on CassandraCluster objects (Casskop may remove this flag, and
      multiCassKop re-set-it...)


## Bugs

- [x] **Bug01**: when changing parameter on a deployed cluster (for instance cassandra image), both clusters applied modification
  in the same time, this is not good, we need to only applied on one cluster and when OK apply to the next one
