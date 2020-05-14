---
id: 5_upgrade_bootstrap_image
title: Upgrade Bootstrap Image
sidebar_label: Upgrade Bootstrap Image
---

## Casskop 0.1.5+ : Bootstrap image to 0.1.4

:::warning
We made a few breaking changes in that image
:::

- `ready-probe.sh` was renamed to readiness-probe.sh
- `/opt/bin` contains the curl command to query Jolokia

Because of it upgrading casskop from 0.5.0 to 0.5.1+ requires to the following steps:

- Uninstall casskop (`helm delete casskop`)
- Edit all statefulsets to manually make those few changes using kubectl edit statefulsets $name-of-your-statefulset
  * Rename ready-probe.sh to readiness-probe.sh
  * Upgrade the bootstrap image to 0.1.4+
  * Add a new emptyDir in the volumes section
  ```
        volumes:
        - emptyDir: {}
        name: tools
        ....
  ```
  * Mount that new volume in all initcontainers and containers (:bulb: init-config does not really need it, but we need to change it first in the operator #189)
  ```
        name: init-config
        ....
        volumeMounts:
        - mountPath: /opt/bin
          name: tools
        ....
        name: bootstrap
        ....
        volumeMounts:
        - mountPath: /opt/bin
          name: tools
        ....
        name: cassandra
        ...
        volumeMounts:
        - mountPath: /opt/bin
          name: tools
        ....
  ```
- Upgrade the version bootstrap image in your cassandracluster object (kubectl edit cassandraclusters *YOUR_OBJECT*)
- Install new version of casskop