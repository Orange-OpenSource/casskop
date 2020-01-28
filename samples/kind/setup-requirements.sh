#!/bin/bash

NAMESPACE=${1:-cassandra-e2e}
$(dirname $0)/configure-local-storage.sh create
kubectl create namespace $NAMESPACE
kubens $NAMESPACE
kubectl apply -f deploy/crds/db_v1alpha1_cassandracluster_crd.yaml
