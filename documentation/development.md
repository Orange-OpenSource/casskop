

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
- [go](go_tool) version v1.13+.
- [docker](docker_tool) version 18.09+.
- [kubectl](kubectl_tool) version v1.13.3+.
- [Helm](https://helm.sh/) version v2.12.2.
- Fork from [Operator sdk](https://github.com/operator-framework/operator-sdk) version v0.15.0 : [Operator sdk - forked](https://github.com/erdrix/operator-sdk/tree/v0.15.0-pr137)


### Install the Operator SDK CLI

First, checkout and install the operator-sdk CLI:

```sh
$ mkdir -p $GOPATH/src/github.com/operator-framework/
$ cd $GOPATH/src/github.com/operator-framework/
$ git clone git@github.com:erdrix/operator-sdk.git
$ git checkout v0.15.0-pr137
$ make tidy
$ make install
```

**Note :** Use fork from operator-sdk, waiting for [PR #317](https://github.com/kubernetes-sigs/controller-tools/pull/317/files#diff-8963ad7c8dcbe3931bc61c3e2954ccf2R300-R301) will be merged.

### Initial setup

Checkout the project.

```sh
$ git clone https://github.com/Orange-OpenSource/casskop.git
$ cd casskop
```

### Local kubernetes setup

We use [kind](https://kind.sigs.k8s.io) in order to run a local kubernetes cluster with the version we chose. We think it deserves some words as it's pretty useful and simple to use to test one version or another

#### Install

The following requires kubens to be present. On MacOs it can be installed using brew :

```
brew install kubectx
```

The installation of kind is then done with (outside of the cassandra operator folder if you want it to run fast) :

```sh
$ GO111MODULE="on" go get sigs.k8s.io/kind@v0.5.1
```

#### Setup

The following actions should be run only to create a new kubernetes cluster.

```sh
$ samples/kind/create-kind-cluster.sh
```

or if you want to enable network policies

```sh
$ samples/kind/create-kind-cluster-network-policies.sh
```

It creates namespace cassandra-e2e by default. If a different namespace is needed it can be specified on the setup-requirements call

```sh
$ samples/kind/setup-requirements.sh other-namespace
```

To interact with the cluster you then need to use the generated kubeconfig file :

```sh
$ export KUBECONFIG=$(kind get kubeconfig-path --name=kind)
```

Before using that newly created cluster, it's better to wait for all pods to be running by continously checking their status :

```sh
$ kubectl get pod --all-namespaces -w
```

#### Pause/Unpause the cluster

In order to kinda freeze the cluster because you need to do something else on your laptop, you can use those two aliases. Just put them in your ~/.bashrc or ~/.zshrc :

```sh
alias kpause='kind get nodes|xargs docker pause'
alias kunpause='kind get nodes|xargs docker unpause'
```

#### Delete cluster

The simple command `kind delete cluster` takes care of it.

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

#### Run the Operator Locally with the Go Binary
This method can be used to run the operator locally outside of the cluster. This method may be preferred during development as it facilitates faster deployment and testing.

Set the name of the operator in an environment variable

```
$ export OPERATOR_NAME=cassandra-operator
```

Deploy the CRD

```
$ kubectl apply -f deploy/crds/db.orange.com_cassandraclusters_crd.yaml
```

```
$ make run
```

This will run the operator in the `default` namespace using the default Kubernetes config file at `$HOME/.kube/config`.

**Note:** JMX operations cannot be executed on Cassandra nodes when running the operator locally. This is because the operator makes JMX calls over HTTP using jolokia and when running locally the operator is on a different network than the Cassandra cluster.

#### Deploy using the Helm Charts
This section provides an instructions for running the operator Helm charts with an image that is built from the local branch.

Build the image from the current branch.

```
$ export DOCKER_REPO_BASE=<your-docker-repo>
$ make docker-build
```
Push the image to docker hub (or to whichever repo you want to use)

```
$ make push
```
**Note:** In this example we are pushing to docker hub.

**Note:** The image tag is a combination of the version as defined in `verion/version.go` and the branch name.

Install the Helm chart.

```
$ helm install ./helm/cassandra-operator \
    --set-string image.repository=orangeopensource/casskop,image.tag=0.4.0-local-dev-helm \
    --name local-dev-helm
```

**Note:** The `image.repository` and `image.tag` template variables have to match the names from the image that we pushed in the previous step.

**Note:** We set the chart name to the branch, but it can be anything.

Lastly, verify that the operator is running.

```
$ kubectl get pods
NAME                                                READY   STATUS    RESTARTS   AGE
local-dev-helm-cassandra-operator-8946b89dc-4cfs9   1/1     Running   0          7m45s
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
 
>You need to install it before see : https://www.telepresence.io/


If your cluster don't have Internet access, you can change the telepresence image to use to one your cluster have access
exemple:
```
TELEPRESENCE_REGISTRY=you-private-registry/datawire  make debug-telepresence-with-alias
```

#### Configure the IDE

Now we just need to configure the IDE :

![](images/ide_debug_configutation.png)

and let's the magic happened

![](images/ide_debug_action.png)


# Release the Project

There are several things to do when you want to make a release of the project:
Todo: things should be automatize ;)

For ease, we have same version for casskop and multi-casskop
- [ ] Update Changelog.md with informations for the new release
- [ ] update version/version.go with the new release version
- [ ] update multi-casskop/version/version.go with the new release version
- [ ] update helm/cassandra-operator/Chart.yaml and values.yaml
- [ ] update multi-casskop/helm/multi-casskop/Chart.yaml and values.yaml
- [ ] generate casskop helm with `make helm-package`
- [ ] add to git docs/helm, commit & push
- [ ] once the PR is merged to master, create the release with content of changelog for this version
    - https://github.com/Orange-OpenSource/casskop/releases

## With Helm

The CassKop operator is released in the helm/charts/incubator see : https://github.com/helm/charts/pull/14414

We also have a helm repository hosted on GitHub pages.

### Release helm charts on GitHub

In order to release the Helm charts on GitHub, we need to generate the package locally
```
make helm-package
```

then add to git the package and make a PR on the repo.


## With OLM (Operator Lifecycle Manager)

OLM is used to manage lifecycle of the Operator, and is also used to puclish on https://operatorhub.io

### Create new OLM release

You can create new version of CassKop OLM bundle using:

Exemple for generating version 0.0.4
```
operator-sdk olm-catalog gen-csv --csv-version 0.4.0 --update-crds
```

> You may need to manually update some fileds (such as description..), you can refere to previous versions for that

### Instruction to tests locally with OLM

Before submitting the operator to operatorhub.io you need to install and test OLM on a local Kubernetes.

These tests and all pre-requisite can also be executed automatically in a single step using a
[Makefile](https://github.com/operator-framework/community-operators/blob/master/docs/using-scripts.md).

Go to github/operator-framework/community-operators to interract with the OLM makefile

Install OLM
```
make operator.olm.install
```

Launch lint
```
make operator.verify OP_PATH=community-operators/casskop VERBOSE=true
```

Launch tests
```
make operator.test OP_PATH=community-operators/casskop VERBOSE=true
```


# How this repository was initially build

## Boilerplate CassKop

We used the SDK to create the repository layout. This command is for memory ;) (or for applying sdk upgrades)

> You need to have first install the SDK.

```
#old version
 operator-sdk new casskop --api-version=db.orange.com/v1alpha1 --kind=CassandraCluster
#new version
operator-sdk new casskop --dep-manager=modules --repo=github.com.Orange-OpenSource/casskop --type=go
```
Then you want to add managers:

```
# Add a new API for the custom resource CassandraCluster
$ operator-sdk add api --api-version=db.orange.com/v1alpha1 --kind=CassandraCluster

# Add a new controller that watches for CassandraCluster
$ operator-sdk add controller --api-version=db.orange.com/v1alpha1 --kind=CassandraCluster
```

## Useful Infos for developers

### Parsing Yaml from String

For parsing Yaml from string to Go Object we uses this library : `github.com/ghodss/yaml` because with the official one
not all fields of the yaml where correctly populated. I don't know why..
