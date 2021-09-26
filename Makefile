# Copyright 2019 Orange
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# 	You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# 	See the License for the specific language governing permissions and
# limitations under the License.

################################################################################

# Name of this service/application
SERVICE_NAME := casskop

BUILD_FOLDER = .
MOUNTDIR = $(PWD)

BOOTSTRAP_IMAGE ?= orangeopensource/cassandra-bootstrap:0.1.9
TELEPRESENCE_REGISTRY ?= datawire
KUBESQUASH_REGISTRY:=

KUBECONFIG ?= ~/.kube/config



HELM_VERSION    := $(shell cat helm/cassandra-operator/Chart.yaml| grep version | awk -F"version: " '{print $$2}')
HELM_TARGET_DIR ?= docs/helm

# Compute image to use during tests
ifdef CIRCLE_BRANCH
  ifeq ($(CIRCLE_BRANCH),master)
	  E2EIMAGE := $(DOCKER_REPO_BASE)/$(IMAGE_NAME):$(VERSION)
  else
	  E2EIMAGE := $(DOCKER_REPO_BASE_TEST)/$(IMAGE_NAME):$(VERSION)
  endif
else
  ifdef CIRCLE_TAG
	  E2EIMAGE := $(DOCKER_REPO_BASE)/$(IMAGE_NAME):$(BRANCH)
  else
		ifeq ($(BRANCH),master)
	    E2EIMAGE := $(DOCKER_REPO_BASE)/$(IMAGE_NAME):latest
    else
	    E2EIMAGE := $(DOCKER_REPO_BASE_TEST)/$(IMAGE_NAME):$(VERSION)
    endif
  endif
endif

build-image:
	@echo $(BUILD_IMAGE):$(OPERATOR_SDK_VERSION)

params:
	@echo "CIRCLE_BRANCH = '$(CIRCLE_BRANCH)'"
	@echo "CIRCLE_TAG = '$(CIRCLE_TAG)'"
	@echo "Version = '$(VERSION)'"
	@echo "E2EIMAGE = '$(E2EIMAGE)'"


# The default action of this Makefile is to build the development docker image
default: build

.DEFAULT_GOAL := help
help:	
	@grep -E '(^[a-zA-Z_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

get-baseversion:
	@echo $(BASEVERSION)

get-version:
	@echo $(VERSION)

clean:
	@rm -rf $(OUT_BIN) || true
	@rm -f apis/cassandracluster/v2/zz_generated.deepcopy.go || true

helm-package:
	@echo Packaging $(HELM_VERSION)
	helm package helm/cassandra-operator
	mv cassandra-operator-$(HELM_VERSION).tgz $(HELM_TARGET_DIR)
	helm repo index $(HELM_TARGET_DIR)/

FIRST_VERSION = .spec.versions[0]
SPEC_PROPS = $(FIRST_VERSION).schema.openAPIV3Schema.properties.spec.properties

.PHONY: update-crds
update-crds:
	echo Update CRD - Remove protocol and set config type to object CRD
	@sed -i '/\- protocol/d' deploy/crds/db.orange.com_cassandraclusters.yaml
	@yq -i e '$(SPEC_PROPS).config.type = "object"' deploy/crds/db.orange.com_cassandraclusters.yaml
	@yq -i e '$(SPEC_PROPS).topology.properties.dc.items.properties.config.type = "object"' deploy/crds/db.orange.com_cassandraclusters.yaml
	@yq -i e '$(SPEC_PROPS).topology.properties.dc.items.properties.rack.items.properties.config.type = "object"' deploy/crds/db.orange.com_cassandraclusters.yaml
	# We checkout v1alpha1 CRD and add it to v2 CRD as it must be known to do an upgrade
	@for crd in deploy/crds/*.yaml; do \
		git show v1.1.5-release:$$(echo $$crd|sed 's/.yaml/_crd.yaml/') $$crd > /tmp/$$(basename $$crd); \
		if [ $$(basename $$crd) == "db.orange.com_cassandraclusters.yaml" ]; then \
			sed -e '1,/versions/d' -e 's/^..//' $$crd >> /tmp/$$(basename $$crd); \
		else \
			sed -e '1,/versions/d' $$crd >> /tmp/$$(basename $$crd); \
		fi; \
		cp /tmp/$$(basename $$crd) $$crd; \
		yq -i e '$(FIRST_VERSION).storage = false' $$crd; \
		mv $$crd $$(echo $$crd|sed 's/.yaml/_crd.yaml/'); \
	done
	cp -v deploy/crds/* helm/*/crds/
	cp -v deploy/crds/* */helm/*/crds/

include share.mk
include kube.mk

docker-generate-files: docker-generate-k8s docker-generate-crds

# Build the Operator and its Docker Image
docker-build: docker-generate-files docker-build-operator

ifdef PUSHLATEST
	docker tag $(REPOSITORY):$(VERSION) $(REPOSITORY):latest
endif

# Build the docker development environment
build-ci-image: deps-development
	docker build --cache-from $(BUILD_IMAGE):latest \
	  --build-arg OPERATOR_SDK_VERSION=$(OPERATOR_SDK_VERSION) \
		-t $(BUILD_IMAGE):latest \
		-t $(BUILD_IMAGE):$(OPERATOR_SDK_VERSION) \
		-f $(DEV_DIR)/Dockerfile \
		.

push-ci-image: deps-development
	docker push $(BUILD_IMAGE):$(OPERATOR_SDK_VERSION)
ifdef PUSHLATEST
	docker push $(BUILD_IMAGE):latest
endif

build-bootstrap-image:
	$(MAKE) -C docker/bootstrap build
push-bootstrap-image:
	$(MAKE) -C docker/bootstrap push

build-cassandra-image:
	$(MAKE) -C docker/cassandra build
push-cassandra-image:
	$(MAKE) -C docker/cassandra push


pipeline:
	docker run -ti --rm --privileged -v $(PWD):/go/src/github.com/Orange-OpenSource/casskop -w /go/src/github.com/Orange-OpenSource/casskop \
  --env https_proxy=$(https_proxy) --env http_proxy=$(http_proxy) \
	$(BUILD_IMAGE):$(OPERATOR_SDK_VERSION) bash

pipeline-e2e:
	docker run -ti --rm --privileged -v $(PWD):/source -w /source \
	$(BUILD_IMAGE):$(OPERATOR_SDK_VERSION) bash

circleci-process:
	circleci config process .circleci/config.yml

circleci-validate:
	circleci config validate

# Run a shell into the development docker image
shell: docker-dev-build
	docker run  --env GO111MODULE=on -ti --rm -v ~/.kube:/.kube:ro -v $(PWD):$(WORKDIR) --name $(SERVICE_NAME) $(BUILD_IMAGE):$(OPERATOR_SDK_VERSION) /bin/bash

debug-port-forward:
	kubectl port-forward `kubectl get pod -l app=casskop -o jsonpath="{.items[0].metadata.name}"` 40000:40000

debug-pod-logs:
	kubectl logs -f `kubectl get pod -l app=casskop -o jsonpath="{.items[0].metadata.name}"`

define debug_telepresence
	export TELEPRESENCE_REGISTRY=$(TELEPRESENCE_REGISTRY) ; \
	echo "execute : cat cassandra-operator.env" ; \
	sudo mkdir -p /var/run/secrets/kubernetes.io ; \
	sudo ln -s /tmp/known/var/run/secrets/kubernetes.io/serviceaccount /var/run/secrets/kubernetes.io/ || true ; \
	tdep=$(shell kubectl get deployment -l app=cassandra-operator -o jsonpath='{.items[0].metadata.name}') ; \
  	echo kubectl get deployment -l app=cassandra-operator -o jsonpath='{.items[0].metadata.name}' ; \
	echo telepresence --swap-deployment $$tdep --mount=/tmp/known --env-file cassandra-operator.env $1 $2 ; \
  	telepresence --swap-deployment $$tdep --mount=/tmp/known --env-file cassandra-operator.env $1 $2
endef

debug-telepresence:
	$(call debug_telepresence)

debug-telepresence-with-alias:
	$(call debug_telepresence,--also-proxy,10.40.0.0/16)
	# $(call debug_telepresence,--also-proxy,172.18.0.0/16)

debug-kubesquash:
	kubesquash --container-repo $(KUBESQUASH_REGISTRY)

# Run the development environment (in local go env) in the background using local ~/.kube/config
run:
	export POD_NAME=casskop; \
	operator-sdk up local

push:
	docker push $(REPOSITORY):$(VERSION)
ifdef PUSHLATEST
	docker push $(REPOSITORY):latest
endif

tag:
	git tag $(VERSION)

publish:
	@COMMIT_VERSION="$$(git rev-list -n 1 $(VERSION))"; \
	docker tag $(REPOSITORY):"$$COMMIT_VERSION" $(REPOSITORY):$(VERSION)
	docker push $(REPOSITORY):$(VERSION)
ifdef PUSHLATEST
	docker push $(REPOSITORY):latest
endif

release: tag image publish

# Test stuff in dev
docker-unit-test:
	docker run --env GO111MODULE=on --rm -v $(PWD):$(WORKDIR) -v $(GOPATH)/pkg/mod:/go/pkg/mod -v $(shell go env GOCACHE):/root/.cache/go-build $(BUILD_IMAGE):$(OPERATOR_SDK_VERSION) /bin/bash -c '$(UNIT_TEST_CMD); cat test-report.out; $(UNIT_TEST_COVERAGE)'

docker-unit-test-with-vendor:
	docker run --env GO111MODULE=on --rm -v $(PWD):$(WORKDIR) -v $(GOPATH)/pkg/mod:/go/pkg/mod -v $(shell go env GOCACHE):/root/.cache/go-build $(BUILD_IMAGE):$(OPERATOR_SDK_VERSION) /bin/bash -c '$(UNIT_TEST_CMD_WITH_VENDOR); cat test-report.out; $(UNIT_TEST_COVERAGE)'

unit-test:
	$(UNIT_TEST_CMD) && echo "success!" || { echo "failure!"; cat test-report.out; exit 1; }
	cat test-report.out
	$(UNIT_TEST_COVERAGE)

unit-test-with-vendor:
	$(UNIT_TEST_CMD_WITH_VENDOR) && echo "success!" || { echo "failure!"; cat test-report.out; exit 1; }
	cat test-report.out
	$(UNIT_TEST_COVERAGE)

define run-operator-cmd
	docker run  --env GO111MODULE=on -ti --rm -v $(PWD):$(WORKDIR) -u $(UID):$(GID) --name $(SERVICE_NAME) $(BUILD_IMAGE):$(OPERATOR_SDK_VERSION) /bin/sh -c $1
endef

docker-go-lint:
	$(call run-operator-cmd,$(GO_LINT_CMD))

# golint is not fully supported by modules yet - https://github.com/golang/lint/issues/409
go-lint:
	$(GO_LINT_CMD)


# Test if the dependencies we need to run this Makefile are installed
deps-development:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif

#Generate dep for graph
UNAME := $(shell uname -s)

dep-graph:
ifeq ($(UNAME), Darwin)
	dep status -dot | dot -T png | open -f -a /Applications/Preview.app
endif
ifeq ($(UNAME), Linux)
	dep status -dot | dot -T png | display
endif

count:
	git ls-files | xargs wc -l

image:
	echo $(REPOSITORY):$(VERSION)

export CGO_ENABLED:=0

ifeq (kuttl-test-fix-arg,$(firstword $(MAKECMDGOALS)))
  KUTTL_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(KUTTL_ARGS):;@:)
endif

kuttl-test-fix-arg:
ifeq ($(KUTTL_ARGS),)
	@echo "A test folder is required" && exit 1
endif
	helm install casskop helm/cassandra-operator --set image.tag=$(BRANCH)
	cd test/kuttl; kuttl test --test $(KUTTL_ARGS) --namespace default

dgoss-bootstrap:
	 IMAGE_TO_TEST=$(BOOTSTRAP_IMAGE) ./docker/bootstrap/dgoss/runChecks.sh

configure-psp:
	kubectl get clusterrole psp:cassie -o yaml
	kubectl -n cassandra get rolebindings.rbac.authorization.k8s.io psp:sa:cassie -o yaml
	kubectl -n cassandra get rolebindings.rbac.authorization.k8s.io psp:sa:cassie -o yaml | grep -vE '(annotations|creationTimestamp|resourceVersion|uid|selfLink|last-applied-configuration)' | sed 's/cassandra/cassandra-e2e/' | kubectl apply -f -

# Usage example:
# REPLICATION_FACTOR=3 make cassandra-stress small
#
ifeq (cassandra-stress,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  STRESS_TYPE := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(STRESS_TYPE):;@:)
endif

REPLICATION_FACTOR ?= 1
DC ?= dc1
USERNAME ?= cassandra
PASSWORD ?= cassandra

.PHONY: cassandra-stress
cassandra-stress:
	kubectl delete configmap cassandra-stress-$(STRESS_TYPE) || true
	cp cassandra-stress/$(STRESS_TYPE)_stress.yaml /tmp/
	echo Using replication factor $(REPLICATION_FACTOR) with DC $(DC) in cassandra-stress profile file
	sed -i -e "s/'dc1': '3'/'$(DC)': '$(REPLICATION_FACTOR)'/" /tmp/$(STRESS_TYPE)_stress.yaml
	kubectl create configmap cassandra-stress-$(STRESS_TYPE) --from-file=/tmp/$(STRESS_TYPE)_stress.yaml
	kubectl delete -f cassandra-stress/cassandra-stress-$(STRESS_TYPE).yaml --wait=false || true
	while kubectl get pod cassandra-stress-$(STRESS_TYPE)>/dev/null; do echo -n "."; sleep 1 ; done
	cp cassandra-stress/cassandra-stress-$(STRESS_TYPE).yaml /tmp/
	sed -i -e 's/user=[a-zA-Z]* password=[a-zA-Z]*/user=$(USERNAME) password=$(PASSWORD)/' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
ifdef CASSANDRA_IMAGE
	echo "using Cassandra image $(CASSANDRA_IMAGE)"
	sed -i -e 's#image:.*#image: $(CASSANDRA_IMAGE)#g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
endif

ifdef CASSANDRA_NODE
	sed -i -e 's/cassandra-demo/$(CASSANDRA_NODE)/g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
else
  ifneq ($(and $(CLUSTER_NAME),$(DC),$(RACK)),)
	sed -i -e 's/cassandra-demo/$(CLUSTER_NAME)-$(DC)-$(RACK)-0.$(CLUSTER_NAME)/g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
  endif

  ifdef CLUSTER_NAME
	sed -i -e 's/cassandra-demo/$(CLUSTER_NAME)/g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
  endif
endif

ifdef CONSISTENCY_LEVEL
	sed -i -e 's/cl=one/cl=$(CONSISTENCY_LEVEL)/g' /tmp/cassandra-stress-$(STRESS_TYPE).yaml
endif

	cat /tmp/cassandra-stress-$(STRESS_TYPE).yaml
	kubectl apply -f /tmp/cassandra-stress-$(STRESS_TYPE).yaml
