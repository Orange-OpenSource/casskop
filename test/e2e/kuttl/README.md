# Kuttl tests
_Reference : https://kuttl.dev/docs/what-is-kuttl.html_

## Requirement
* kubectl version 1.13.0 or newer

## Install
### Linux / Macos (with brew)
```
brew tap kudobuilder/tap
brew install kuttl-cli
```

### Or with krew
If you have [krew](https://github.com/kubernetes-sigs/krew) installed :
```
kubectl krew install kuttl
```

_refer to the [KUTTL setup page](https://kuttl.dev/docs/cli.html#setup-the-kuttl-kubectl-plugin) for more infos or ways to install_

## Run
First you need to set up a k8s cluster and deploy Casskop in it.  
Then you just have tu run :  
```
kubectl kuttl test path/to/e2e/kuttl/declarative-tests/
```