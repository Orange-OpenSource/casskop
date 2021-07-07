---
id: 5_upgrade_v1_to_v2
title: Upgrade v1 to v2
sidebar_label: Upgrade v1 to v2
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

Version 2 makes it possible to use Cassandra 3 and 4 using the same bootstrap image. That's why it's recommended to
upgrade to version 2.

:::warning
It's highly recommended that you test this procedure on a testing environment first.
:::

In order to upgrade to version 2 without impacting your users you need to follow those steps.

## Collect your configmap parameters
If you use a ConfigMap, you can't use parameters other than pre_run.sh and post_run.sh
Collect all the non default parameters that you use and also the number of tokens. you'll
need those when it's time to set the configuration in your CassandraCluster objects.

## Uninstall your operator
You need to uninstall it which won't have any effect on your running cluster other than
not allow you to trigger operations, scale it etc...
```shell
helm delete casskop
```

## Update the CRDs
helm does not version CRDs, so you'll need to manually update them (You can get the
new CRDs from our git repo).
```shell
kubectl apply -f deploy/crds
```

## Edit your CassandraCluster object
Now it's time to edit your object and add the cassandra/java configuration from your configmap in there.
You also have to update the bootstrap image version to *0.1.9*.
```shell
kubectl edit cassandraclusters.db.orange.com your-object
```

Here is an example of what you could have after you've edited it:
```yaml
apiVersion: db.orange.com/v1alpha1
kind: CassandraCluster
metadata:
  name: your-object
spec:
  nodesPerRacks: 2
  cassandraImage: cassandra:3.11.9
  bootstrapImage: orangeopensource/cassandra-bootstrap:0.1.9
  config:
    cassandra-yaml:
      num_tokens: 256
    jvm-options:
      initial_heap_size: 32M
      max_heap_size: 256M
  dataCapacity: "1Gi"
  deletePVC: true
  autoPilot: true
  resources:
    limits:
      cpu: 100m
      memory: 512Mi
  topology:
    dc:
      - name: dc1
        rack:
          - name: rack1
```

If you use a version like `cassandra:latest`, you have to add at the
same level a parameter called _serverVersion_ and set it to the version of
the configuration you wanna use cause it's used when generating it.

Also you don't have to set the heap and in that case some automatic values
will be picked for you.


If you have a doubt on what name to use for a parameter in your cassandra.yaml,
you can take a look at https://github.com/datastax/cass-config-definitions/tree/1b7eaf4e50447fc8168c4a6c16d0ed986941edf8/resources/cassandra-yaml/cassandra

## Install the latest version of the operator
Now you can install version 2 of the operator by running the usual install
command:
```shell
helm install casskop orange-incubator/cassandra-operator
```
