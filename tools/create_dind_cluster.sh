#!/bin/bash
echo $https_proxy
#export DIND_PROPAGATE_HTTP_PROXY=true
export DIND_HTTP_PROXY=$http_proxy
export DIND_HTTPS_PROXY=$https_proxy
export DIND_NO_PROXY=$no_proxy

echo "Create K8s cluster"
NUM_NODES=3 dind-cluster up

echo "Configure local-storage"
kubectl create namespace local-provisioner
kubectl config set-context $(kubectl config current-context) --namespace=local-provisioner
tools/configure-dind-local-storage.sh create


echo "label nodes"
kubectl label node kube-node-1 failure-domain.beta.kubernetes.io/region="europe-west2"
kubectl label node kube-node-1 location.physical/rack="1"
kubectl label node kube-node-1 failure-domain.beta.kubernetes.io/region="europe-west2"
kubectl label node kube-node-1 location.physical/rack="2"

echo "create cassandra namespace"
kubectl create namespace cassandra-demo
kubectl config set-context $(kubectl config current-context) --namespace=cassandra-demo

echo "Create CRD"
kubectl apply -f config/crd/bases/db.orange.com_cassandraclusters_crd.yaml

echo "configure helm"
helm init
