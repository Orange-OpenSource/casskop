#!/bin/bash

echo "Deleting test namespaces"
for x in $(kns | egrep "group|main-"); do echo $x ; k delete namespace --grace-period=0 --force $x ; done

echo "Deleting CRD"
kubectl delete -f deploy/crds/db.orange.com_cassandraclusters_crd.yaml
