---
id: 2_release_guide
title: Release guide
sidebar_label: Release guide
---

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