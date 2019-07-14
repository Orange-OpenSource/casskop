

# Instruction for development & contributing on CassKop

## CircleCI build pipeline

We use CircleCI to build and test the operator code on each commit.

### CircleCI config validation hook

To discover errors in the CircleCI earlier, we can uses the [CircleCI cli](https://circleci.com/docs/2.0/local-cli/)
to validate the config file on pre-commit git hook.

Fisrt you must install the cli, then to install the hook, runs:<

```console
cp tools/pre-commit .git/hooks/pre-commit
```

The Pipeline uses some envirenment variables that you need to set-up if you want your fork to build

- DOCKER_REPO_BASE -- name of your docker base reposirory (ex: orangeopensource)
- DOCKERHUB_PASSWORD
- DOCKERHUB_USER
- SONAR_PROJECT
- SONAR_TOKEN

If not set in CircleCI environment, according steps will be ignored.

### CircleCI on PR

When you submit a Pull Request, then CircleCI will trigger build pipeline.
Since this is pushed from a fork, for security reason the pipeline won't have access to the environment secrets, 
and not all steps could be executed.

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


### Initial setup

Checkout the project.

```sh
$ mkdir -p $GOPATH/src/gitlab.si.francetelecom.fr/kubernetes/
$ cd $GOPATH/src/gitlab.si.francetelecom.fr/kubernetes/
$ git clone https://github.com/Orange-OpenSource/cassandra-k8s-operator.git
$ cd cassandra-k8s-operator
```

### Local kubernetes setup

We use [dind](https://github.com/kubernetes-sigs/kubeadm-dind-cluster) in order to run a local kubernetes cluster with the version we chose. We think it deserves some words as it's pretty useful and simple to use to test one version or another

#### Setup

The following requires kubens to be present. On MacOs it can be installed using brew :

```
brew install kubectx
```

The setup is then done with :

```sh
$ wget https://github.com/kubernetes-sigs/kubeadm-dind-cluster/releases/download/v0.2.0/dind-cluster-v1.14.sh -O /usr/local/bin/dind-cluster.sh
$ chmod u+x /usr/local/bin/dind-cluster.sh
$ dind-cluster.sh up
$ samples/dind/setup-requirements.sh
```

It creates namespace cassandra-e2e by default. If a different namespace is needed it can be specified on the setup-requirements call

```sh
$ samples/dind/setup-requirements.sh other-namespace
```

It's better to wait for all pods to be running by continously checking their status

```sh
$ kubectl get pod --all-namespaces -w
```

#### Take a snapshot

To avoid having to rebuild the cluster and do the setup again take a snapshot of the dind cluster, this way everytime it's back up, the snapshot is restored and it restarts from the same point.

```sh
$ dind-cluster.sh snapshot
```

If the cluster is cleaned, snapshots are lost cause they are stored in the containers used for master/nodes. But the cluster can be taken down and when it's restarted the snapshot is restored.

```sh
$ dind-cluster.sh down
# Later on
$ dind-cluster.sh up
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

### Run e2e end to end tests

CassKop also have several end-to-end tests that can be run using makefile.

You need to create the namespace `cassandra-e2e` before running the tests.

```
kubectl create namespace cassandra-e2e
```

to launch different tests in parallel in different temporary namespaces

```
make e2e
```

or sequentially in the namespace `cassandra-e2e`

```
make e2e-test-fix
```

**Note**: `make e2e` executes all tests in different, temporary namespaces in parallel. Your k8s cluster will need a lot of resources to handle the many Cassandra nodes that launch in parallel. 

**Note**: `make e2e-test-fix` executes all tests serially in the `cassandra-e2e` namespace and as such does not require as many k8s resources as `make e2e` does, but overall execution will be slower.

You can choose to run only 1 test using the args ex:

```
make e2e-test-fix-arg ClusterScaleDown
```

**Tip**: When debugging test failures, you can run `kubectl get events --all-namespaces` which produce output like:

`cassandracluster-group-clusterscaledown-1561640024   0s    Warning   FailedScheduling   Pod   0/4 nodes are available: 1 node(s) had taints that the pod didn't tolerate, 3 Insufficient cpu.`

**Tip**: When tests fail, there may be resources that need to be cleaned up. Run `tools/e2e_test_cleanup.sh` to delete resources left over from tests.


### Debug CassKop in remote in a Kubernetes cluster

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


# Release the Project

The CassKop operator is released in the helm/charts/incubator see : https://github.com/helm/charts/pull/14414

We also have a helm repository hosted on GitHub pages.

## Release helm charts on GitHub

In order to release the Helm charts on GitHub, we need to generate the package locally
```
make helm-package
```

then add to git the package and make a PR on the repo.


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
