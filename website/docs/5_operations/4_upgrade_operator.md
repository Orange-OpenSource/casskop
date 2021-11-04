---
id: 4_upgrade_operator
title: Upgrade Operator
sidebar_label: Upgrade Operator
---

## Case : No changes of the CRD's structure

Upgrading the operator consists in uninstalling the current version and installing the new version :

```
helm uninstall casskop
helm repo update
helm install --name casskop casskop/cassandra-operator
```

## Upgrading from v1 to v2

Please refer to [the specific v1 to v2 section](/casskop/docs/2_setup/5_upgrade_v1_to_v2) for the step by step protocol.
