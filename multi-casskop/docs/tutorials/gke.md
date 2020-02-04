# Setting up Multi-Casskop on Google Kubernetes Engine

## Pre-requisites

User should need :

* [terraform](https://learn.hashicorp.com/terraform/getting-started/install.html) version v0.12.7+
* [kubectl](https://kubernetes.io/fr/docs/tasks/tools/install-kubectl) version v1.13.3+
* [kubectx](https://github.com/ahmetb/kubectx) & kubens
* [Helm](https://helm.sh/docs/intro/using_helm/) version v2.15.1+
* [gcloud sdk](https://cloud.google.com/sdk/install?hl=fr) version 272.0.0+
* A service account with enough rights (for this example : `editor`)
* Having a DNS zone in google cloud dns.

## Setup GCP environment

To setup the GCP environment we will use terraform provisionning, to instantiate the following infrastructure :

* 2 GKE clusters :
  *  First on europe-west1-b which will be the `master`
  *  Second on europe-west1-c which will be the `slave`
* Firewall rules to allow clusters to communicate
* External DNS on each cluster to expose cassandra nodes
* Casskop operator on each cluster to focus on multi-casskop usage


### Environment setup

Start to set variables needed for the instantiation : 

```sh
$ export CASSKOP_WORKSPACE=<path to cassandra-k8s-operateur project>
$ export PROJECT=<gcp project>
$ export SERVICE_ACCOUNT_KEY_PATH=<path to service account key>
$ export NAMESPACE=cassandra-demo
$ export DNS_ZONE_NAME=external-dns-test-gcp-trycatchlearn-fr     # -> change with your own one
$ export DNS_NAME=external-dns-test.gcp.trycatchlearn.fr          # -> change with your own one
$ export MANAGED_ZONE=tracking-pdb                                # -> change with your own one
```

### Setup base infrastructure 

```sh
$ cd ${CASSKOP_WORKSPACE}/multi-casskop/samples/gke/terraform
$ terraform init
```

#### Master provisionning

With the master provisionning, we will deploy firewall and Cloud dns configuration :

```sh
$ terraform workspace new master
$ terraform workspace select master
$ terraform apply \
    -var-file="env/master.tfvars" \
    -var="service_account_json_file=${SERVICE_ACCOUNT_KEY_PATH}" \
    -var="namespace=${NAMESPACE}" \
    -var="project=${PROJECT}" \
    -var="dns_zone_name=${DNS_ZONE_NAME}" \
    -var="dns_name=${DNS_NAME}" \
    -var="managed_zone=${MANAGED_ZONE}"
```

#### Slave provisionning

```sh
$ terraform workspace new slave
$ terraform workspace select slave
$ terraform apply \
    -var-file="env/slave.tfvars" \
    -var="service_account_json_file=${SERVICE_ACCOUNT_KEY_PATH}" \
    -var="namespace=${NAMESPACE}" \
    -var="project=${PROJECT}" \
    -var="dns_zone_name=${DNS_ZONE_NAME}" \
    -var="dns_name=${DNS_NAME}" \
    -var="managed_zone=${MANAGED_ZONE}"
```

### Check installation 

#### Check master configuration

Now we will check that everything is well deployed in the GKE master cluster : 

```sh
$ gcloud container clusters get-credentials cassandra-europe-west1-b-master --zone europe-west1-b --project ${PROJECT}
$ kubectl get pods -n ${NAMESPACE}
NAME                                          READY   STATUS    RESTARTS   AGE
casskop-cassandra-operator-54c4cfcbcb-b4qxq   1/1     Running   0          4h9m
external-dns-6dd96c985-h76gh                  1/1     Running   0          4h16m
```

#### Check slave configuration

Now we will check that everything is well deployed in the GKE slave cluster : 

```sh
$ gcloud container clusters get-credentials cassandra-europe-west1-c-slave --zone europe-west1-c --project ${PROJECT}
$ kubectl get pods -n ${NAMESPACE}
NAME                                          READY   STATUS    RESTARTS   AGE
casskop-cassandra-operator-54c4cfcbcb-sxjz7   1/1     Running   0          4m56s
external-dns-7f947c5b5b-mq7kg                 1/1     Running   0          5m46s
```

#### Check DNS zone configuration

Make a note of the nameservers that were assigned to your new zone : 

```sh
$ gcloud dns record-sets list \
    --zone "${DNS_ZONE_NAME}" \
    --name "${DNS_NAME}." \
    --type NS
NAME                                     TYPE  TTL    DATA
external-dns-test.gcp.trycatchlearn.fr.  NS    21600  ns-cloud-e1.googledomains.com.,ns-cloud-e2.googledomains.com.,ns-cloud-e3.googledomains.com.,ns-cloud-e4.googledomains.com.
```

#### Check Firewall configuration

@TODO : rework firewall source

```sh
$ gcloud compute firewall-rules describe gke-cassandra-cluster
allowed:
- IPProtocol: udp
- IPProtocol: tcp
creationTimestamp: '2019-12-05T13:31:01.233-08:00'
description: ''
direction: INGRESS
disabled: false
id: '8270840333953452538'
kind: compute#firewall
logConfig:
  enable: false
name: gke-cassandra-cluster
network: https://www.googleapis.com/compute/v1/projects/poc-rtc/global/networks/default
priority: 1000
selfLink: https://www.googleapis.com/compute/v1/projects/poc-rtc/global/firewalls/gke-cassandra-cluster
sourceRanges:
- 0.0.0.0/0
targetTags:
- cassandra-cluster
```

#### Check Storage Class 

```sh
$ kubectl get storageclasses.storage.k8s.io 
NAME                 PROVISIONER            AGE
standard (default)   kubernetes.io/gce-pd   28m
standard-wait        kubernetes.io/gce-pd   24m
```

## Multi casskop deployment

### Bootstrap API access to Slave from Master

Multi-Casskop will be deployed in `master cluster`, change your `kubectl` context to point this cluster.

In order to allow Multi-CassKop controller to have access to `slave` from `master`, we are going to use [kubemcsa](https://github.com/admiraltyio/multicluster-service-account/releases/tag/v0.6.1) from [admiralty](https://admiralty.io/) to be able to export secret from `slave` to `master`.

Install `kubemcsa` : 

```sh
$ export RELEASE_VERSION=v0.6.1
$ wget https://github.com/admiraltyio/multicluster-service-account/releases/download/${RELEASE_VERSION}/kubemcsa-linux-amd64
$ mkdir -p ~/tools/kubemcsa/${RELEASE_VERSION} && mv kubemcsa-linux-amd64 tools/kubemcsa/${RELEASE_VERSION}/kubemcsa
$ chmod +x ~/tools/kubemcsa/${RELEASE_VERSION}/kubemcsa
$ sudo ln -sfn  ~/tools/kubemcsa/${RELEASE_VERSION}/kubemcsa /usr/local/bin/kubemcsa
```

Generate secret for `master` : 

```sh
$ kubectx # Switch context on master cluster
Switched to context "gke_<Project name>_europe-west1-b_cassandra-europe-west1-b-master".
$ kubens # Switch context on correct namespace
Context "gke_<Project name>_europe-west1-b_cassandra-europe-west1-b-master" modified.
Active namespace is "<Namespace>".
$ kubemcsa export --context=gke_poc-rtc_europe-west1-c_cassandra-europe-west1-c-slave --namespace ${NAMESPACE} cassandra-operator --as gke-slave-west1-c | kubectl apply -f -
secret/gke-slave-west1-c created
```

Check that the secret is correctly created

```sh
$ kubectl get secrets -n ${NAMESPACE}
...
gke-slave-west1-c                Opaque                                5      28s
```

### Install Multi-CassKop

@TODO : To correct once the watch object will be fixed

Add MultiCasskop crd on the `slave` cluster : 

```sh
$ kubectx # Switch context on slave cluster
Switched to context "gke_<Project name>_europe-west1-c_cassandra-europe-west1-c-slave".
$ kubectl apply -f https://raw.githubusercontent.com/Orange-OpenSource/casskop/master/multi-casskop/deploy/crds/multicluster_v1alpha1_cassandramulticluster_crd.yaml
```

Deployment with Helm : 

```sh
$ kubectx # Switch context on master cluster
Switched to context "gke_<Project name>_europe-west1-b_cassandra-europe-west1-b-master".
$ helm init --client-only
$ helm repo add orange-incubator https://orange-kubernetes-charts-incubator.storage.googleapis.com
$ helm repo update
$ cd ${CASSKOP_WORKSPACE}
$ helm install --name multi-casskop orange-incubator/multi-casskop --set k8s.local=gke-master-west1-b --set k8s.remote={gke-slave-west1-c} --set image.tag=0.5.0-multi-cluster #--no-hooks if crd already install
```

### Create the MultiCasskop CRD

Now we are ready to deploy a MultiCassKop CRD instance.
We will use the example in `multi-casskop/samples/gke/multi-casskop-gke.yaml` :

```sh
$ kubectl apply -f multi-casskop/samples/gke/multi-casskop-gke.yaml
```

### Check multi cluster installation

We can see that each cluster has the required pods : 

```sh
$ kubectx # Switch context on master cluster
Switched to context "gke_<Project name>_europe-west1-b_cassandra-europe-west1-b-master".
$ kubectl get pods -n ${NAMESPACE}
NAME                                          READY   STATUS    RESTARTS   AGE
cassandra-demo-dc1-rack1-0                    1/1     Running   0          8m30s
casskop-cassandra-operator-54c4cfcbcb-8qncr   1/1     Running   0          34m
external-dns-6dd96c985-7jf6w                  1/1     Running   0          35m
multi-casskop-67dc74dff7-z4642                1/1     Running   0          11m
$ kubectx # Switch context on slave cluster
Switched to context "gke_<Project name>_europe-west1-c_cassandra-europe-west1-c-slave".
$ kubectl get pods -n ${NAMESPACE}
NAME                                          READY   STATUS    RESTARTS   AGE
cassandra-demo-dc3-rack3-0                    1/1     Running   0          6m55s
cassandra-demo-dc4-rack4-0                    1/1     Running   0          4m59s
cassandra-demo-dc4-rack4-1                    1/1     Running   0          3m20s
casskop-cassandra-operator-54c4cfcbcb-sxjz7   1/1     Running   0          71m
external-dns-7f947c5b5b-mq7kg                 1/1     Running   0          72m
```

If we go in one of the created pods, we can see that nodetool see pods of both clusters : 

```sh
$ kubectx # Switch context on master cluster
Switched to context "gke_<Project name>_europe-west1-b_cassandra-europe-west1-b-master".
$ kubectl exec -ti cassandra-demo-dc1-rack1-0 nodetool status
Datacenter: dc1
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address    Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.52.2.3  108.62 KiB  256          49.2%             a0958905-e1fa-4410-baca-fc86f4457f1a  rack1
Datacenter: dc3
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address    Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.8.3.2   74.95 KiB  256          51.5%             03f8eede-4b69-43be-a0c1-73f73470398b  rack3
Datacenter: dc4
===============
Status=Up/Down
|/ State=Normal/Leaving/Joining/Moving
--  Address    Load       Tokens       Owns (effective)  Host ID                               Rack
UN  10.8.4.3   107.87 KiB  256          47.8%             1a7432e2-4ca8-4767-acdb-3b40e6ff4a57  rack4
UN  10.8.2.5   107.85 KiB  256          51.6%             272037ce-4146-42c1-9079-ef4561249254  rack4
```

## Clean up everything

If you have set the `deleteCassandraCluster` to true, then when deleting the `MultiCassKop` object, it will cascade the deletion of the CassandraCluster object in the targeted k8s clusters. Then each local CassKop will delete their Cassandra clusters (else skip this step)

```sh
$ kubectl delete multicasskops.db.orange.com multi-casskop-demo
$ helm del --purge multi-casskop
```

### Cleaning slave cluster

```sh
$ cd ${CASSKOP_WORKSPACE}/multi-casskop/samples/gke/terraform
$ terraform workspace select slave
$ terraform destroy \
    -var-file="env/slave.tfvars" \
    -var="service_account_json_file=${SERVICE_ACCOUNT_KEY_PATH}" \
    -var="namespace=${NAMESPACE}" \
    -var="project=${PROJECT}" \
    -var="dns_zone_name=${DNS_ZONE_NAME}" \
    -var="dns_name=${DNS_NAME}" \
    -var="managed_zone=${MANAGED_ZONE}"
```

### Cleaning master cluster

Before running the following command, you need to clean dns records set.

```sh
$ terraform workspace select master
$ terraform destroy \
    -var-file="env/master.tfvars" \
    -var="service_account_json_file=${SERVICE_ACCOUNT_KEY_PATH}" \
    -var="namespace=${NAMESPACE}" \
    -var="project=${PROJECT}" \
    -var="dns_zone_name=${DNS_ZONE_NAME}" \
    -var="dns_name=${DNS_NAME}" \
    -var="managed_zone=${MANAGED_ZONE}"
```
