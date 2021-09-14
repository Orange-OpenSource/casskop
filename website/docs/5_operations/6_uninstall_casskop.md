---
id: 5_uninstall_casskop
title: Uninstall Casskop
sidebar_label: Uninstall Casskop
---

## Uninstaling the Charts

If you want to delete the operator from your Kubernetes cluster, the operator deployment 
should be deleted.

```
$ helm uninstall casskop
```
The command removes all the Kubernetes components associated with the chart and deletes the helm release.

> The CRDs created by the chart are not removed by Helm and should be manually cleaned up (if required)

Manually delete the CRDs:
```
kubectl delete crd cassandraclusters.db.orange.com
kubectl delete crd cassandrabackups.db.orange.com
kubectl delete crd cassandrarestores.db.orange.com
```

> :triangular_flag_on_post: If you delete the CRDs then : It will delete **ALL** Clusters that has been created using these CRDs!!!
> Please never delete CRDs without very very good care
