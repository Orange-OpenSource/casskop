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
kubectl apply -f deploy/crds/db_v1alpha1_cassandracluster_crd.yaml

echo "configure helm"
helm init
#sleep 60


#echo "install Operator"
#helm install --name cassandra-demo ./helm/cassandra-k8s-operator
#sleep 5

#echo "Deploy cassandra cluster"
#kubectl apply -f samples/cassandra-jvm-configmap-full.yaml
#kubectl apply -f samples/cc-local.yaml

#Delete cassandra cluster
#k delete CassandraCluster cassandra

#Delete PVCs
#k get pvc|awk '!/NAME/ {print $1}'|xargs -I XX kubectl delete pvc XX


#echo "check proxy settings in dind"
#docker exec -ti kube-node-1 cat /etc/systemd/system/docker.service.d/30-proxy.conf


#in docker
#systemctl show --property=Environment docker

#local proxy_env="[Service]"$'\n'"Environment="
#if [[ $http_proxy ]] ;    then
#    proxy_env+="\"HTTP_PROXY=$http_proxy\" "
#fi
#
#if [[ $https_proxy ]] ;    then
#    proxy_env+="\"HTTPS_PROXY=$https_proxy\" "
#fi
#
#if [[ $no_proxy ]] ;    then
#    proxy_env+="\"NO_PROXY=$no_proxy\" "
#fi
#
#container_id=kube-node-1
#docker exec -i ${container_id} /bin/sh -c "cat > /etc/systemd/system/docker.service.d/30-proxy.conf" <<< "${proxy_env}"
#docker exec ${container_id} systemctl daemon-reload
#docker exec ${container_id} systemctl restart docker
#
