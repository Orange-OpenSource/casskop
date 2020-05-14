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

It's also possible to decide to temporarily install a developement release by specifying the image tag to use :

```
helm install --name casskop casskop/cassandra-operator --set debug.enabled=true --no-hooks \
--set image.tag=v0.5.0b-branch1
```