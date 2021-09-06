#!/bin/bash

CLUSTER=local-casskop
k3d cluster delete $CLUSTER
k3d cluster create $CLUSTER
. $(dirname $0)/setup-requirements.sh
