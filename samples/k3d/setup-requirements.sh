#!/bin/bash

kubectx k3d-$CLUSTER

NAMESPACE=ns1
kubectl create namespace $NAMESPACE
kubens $NAMESPACE
kubectl apply -f deploy/crds/
kubectl apply -f deploy/service_account.yaml
