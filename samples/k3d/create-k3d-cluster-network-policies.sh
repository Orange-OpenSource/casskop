#!/bin/bash

curl -o calico.yaml https://raw.githubusercontent.com/rancher/k3d/main/docs/usage/guides/calico.yaml
CLUSTER=local-casskop
k3d cluster delete $CLUSTER
k3d cluster create $CLUSTER --k3s-server-arg '--flannel-backend=none' \
  --volume $PWD/calico.yaml:/var/lib/rancher/k3s/server/manifests/calico.yaml
. $(dirname $0)/setup-requirements.sh
kubectl apply -f $(dirname $0)/../network-policies.yaml
