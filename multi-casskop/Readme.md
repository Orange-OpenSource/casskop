
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
