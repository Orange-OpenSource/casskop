

# Instruction for development & contributing on CassKop

## CircleCI build Pipeline

We use CircleCI to build and test the operator code on each commit.

### CircleCI config validation hook

To discover errors in the CircleCI earlier, we can uses the [CircleCI cli](https://circleci.com/docs/2.0/local-cli/)
to validate the config file on pre-commit git hook.

Fisrt you must install the cli, then to install the hook, runs:<

```console
cp tools/pre-commit .git/hooks/pre-commit
```

## Operator SDK

### Prerequisites

Casskop has been validated with :

- [dep](dep_tool) version v0.5.1+.
- [go](go_tool) version v1.12.4+.
- [docker](docker_tool) version 18.09+.
- [kubectl](kubectl_tool) version v1.13.3+.
- [Helm](https://helm.sh/) version v2.12.2.
- [Operator sdk](https://github.com/operator-framework/operator-sdk) version v0.7.0.


### Install the Operator SDK CLI

First, checkout and install the operator-sdk CLI:

```sh
$ cd $GOPATH/src/github.com/operator-framework/operator-sdk
$ git checkout v0.7.0
$ make dep
$ make install
```


### Initial Setup

Checkout the project.

```sh
$ mkdir -p $GOPATH/src/gitlab.si.francetelecom.fr/kubernetes/
$ cd $GOPATH/src/gitlab.si.francetelecom.fr/kubernetes/
$ git clone https://github.com/Orange-OpenSource/cassandra-k8s-operator.git
$ cd cassandra-k8s-operator
```
### Build CassKop


#### Using your local environment

If you prefer working directly with your local go environment you can simply uses :

```
make get-deps
make build
```

> You can check on the Gitlab Pipeline to see how the Project is build and test for each push


#### Or Using the provided cross platform build environment

Build the docker image which will be used to build CassKop docker image

```
make build-ci-image
```

> If you want to change the operator-sdk version change the **OPERATOR_SDK_VERSION** in the Makefile.

Then build CassKop (code & image)

```
make docker-get-deps
make docker-build
```

### Run CassKop

We can quickly run CassKop in development mode (on your local host), then it will use your kubectl configuration file to connect to your kubernetes cluster.

There are several ways to execute your operator :

- Using your IDE directly
- Executing directly the Go binary
- deploying using the Helm charts

If you want to configure your development IDE, you need to give it environment variables so that it will uses to connect to kubernetes.


```
KUBECONFIG=<path/to/your/kubeconfig>
WATCH_NAMESPACE=<namespace_to_watch>
POD_NAME=<name for operator pod>
LOG_LEVEL=Debug
OPERATOR_NAME=ide
```

### Run unit-tests

You can run Unit-test for CassKop

```
make unit-test
```

### Run E2E End to End tests

CassKop also have several end-to-end tests that can be run using makefile.

to launch different tests in parallel in different temporary namespaces

```
make e2e
```

or sequentially in the namespace `cassandra-e2e`

```
make e2e-test-fix
```

you can choose to run only 1 test using the args ex:

```
make e2e-test-fix-arg ClusterScaleDown
```

### Debug CassKop in remote in a Kubernetes Cluster

CassKop makes some specific calls to the Jolokia API of the CassandraNodes it deploys inside the kubernetes
cluster. Because of this, it is not possible to fully debug CassKop when launch outside the kubernetes cluster (in
your local IDE).

It is possible to use external solutions such as KubeSquash or Telepresence.

Telepresence launch a bi-directional tunnel between your dev environment and an existing operator pod in the cluster 
which it will **swap**

To launch the telepresence utility you can launch

```
make telepresence
```
 
>c You need to install it before see : https://www.telepresence.io/


#### Configure the IDE

Now we just need to configure the IDE :

![](images/ide_debug_configutation.png)

and let's the magic happened

![](images/ide_debug_action.png)



# How this repository was initially build

## Boilerplate CassKop

We used the SDK to create the repository layout. This command is for memory ;) (or for applying sdk upgrades)

> You need to have first install the SDK.

```
operator-sdk new cassandra-k8s-operator --api-version=db.orange.com/v1alpha1 --kind=CassandraCluster
```

## Useful Infos for developers

### Parsing Yaml from String

For parsing Yaml from string to Go Object we uses this library : `github.com/ghodss/yaml` because with the official one
not all fields of the yaml where correctly populated. I don't know why..
