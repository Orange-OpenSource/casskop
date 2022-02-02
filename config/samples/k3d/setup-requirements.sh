#!/bin/bash

kubectx k3d-$CLUSTER

NAMESPACE=ns1
kubectl create namespace $NAMESPACE
kubens $NAMESPACE
kubectl apply -f config/crd/bases/
kubectl apply -f config/rbac/service_account.yaml
