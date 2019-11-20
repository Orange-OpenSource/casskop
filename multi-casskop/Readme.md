
# Multi-CassKop Project


## Pré-réquisite

### Install CassKop

CassKop must be deployed on each targeted Kubernetes clusters.

Add the Help repository for CassKop

```console
$ helm repo add casskop https://Orange-OpenSource.github.io/cassandra-k8s-operator/helm
```

Connect to each kubernetes you wants to deploy your Cassandra clusters to and install CassKop:

```console
$ helm install --name casskop casskop/cassandra-operator
$ helm install --name casskop casskop/cassandra-operator --set image.repository=ext-dockerio.artifactory.si.francetelecom.fr/orangeopensource/cassandra-k8s-operator --set image.tag=0.4.0-master
```

### Install External-DNS

Configure your external DNS and deploy it in the namespace 

```console
helm install -f /private/externaldns-values.yaml --name casskop-dns rickaastley/external-dns 
```


### Set-up multi service account

CONTEXT1=dex-kaas-prod-priv-sph # change me
CLUSTER1=dex-kaas-prod-priv-sph # change me

CONTEXT2=dex-kaas-prod-priv-bgl # change me
CLUSTER2=dex-kaas-prod-priv-bgl # change me

RELEASE_URL=https://github.com/admiraltyio/multicluster-service-account/releases/download/v0.5.1

Install multicluster-service-account in cluster1:

k apply -f ./config/install/deployment.yaml

Cluster1 is now able to import service accounts, but it hasn't been given permission to import them from cluster2 yet.
This is a chicken-and-egg problem: cluster1 needs a token from cluster2, before it can import service accounts from it.
To solve this problem, download the kubemcsa binary and run the bootstrap command:

```
OS=darwin # or linux / darwin (i.e., OS X) or windows
ARCH=amd64 # if you're on a different platform, you must know how to build from source
BINARY_URL="$RELEASE_URL/kubemcsa-$OS-$ARCH"
curl -Lo kubemcsa $BINARY_URL
chmod +x kubemcsa
sudo mv kubemcsa /usr/local/bin
```

bootstrap the cluster
```
kubemcsa bootstrap --target-context $CONTEXT1 --source-context $CONTEXT2
```
