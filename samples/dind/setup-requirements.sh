#!/bin/bash

NAMESPACE=${1:-cassandra-e2e}
$(dirname $0)/configure-dind-local-storage.sh create
kubectl create namespace $NAMESPACE
kubens $NAMESPACE
helm init
kubectl apply -f deploy/crds/db_v1alpha1_cassandracluster_crd.yaml
kubectl create secret generic jolokia-auth --from-literal=username=jolokia-user --from-literal=password=M0np455w0rd