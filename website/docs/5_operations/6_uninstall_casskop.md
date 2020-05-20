---
id: 6_uninstall_casskop
title: Uninstall Casskop
sidebar_label: Uninstall Casskop
---

## Uninstaling the Charts

If you want to delete the operator from your Kubernetes cluster, the operator deployment 
should be deleted.

```
$ helm delete casskop
```
The command removes all the Kubernetes components associated with the chart and deletes the helm release.

> The CRD created by the chart are not removed by default and should be manually cleaned up (if required)

Manually delete the CRD:
```
kubectl delete crd cassandraclusters.dfy.orange.com
```

> :triangular_flag_on_post: If you delete the CRD then it will delete **ALL** Clusters that has been created using this CRD!!!
> Please never delete a CRD without very very good care