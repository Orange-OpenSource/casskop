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
SERVICE_NAME := multi-casskop

BUILD_FOLDER = multi-casskop
MOUNTDIR = $(PWD)/..

TELEPRESENCE_REGISTRY ?= datawire
KUBESQUASH_REGISTRY:=

KUBECONFIG ?= ~/.kube/config

MINIKUBE_CONFIG ?= ~/.minikube
MINIKUBE_CONFIG_MOUNT ?= /home/circleci/.minikube

HELM_VERSION    := $(shell cat helm/multi-casskop/Chart.yaml| grep version | awk -F"version: " '{print $$2}')
HELM_TARGET_DIR ?= ../docs/helm

params:
	@echo "CIRCLE_BRANCH = '$(CIRCLE_BRANCH)'"
	@echo "CIRCLE_TAG = '$(CIRCLE_TAG)'"
	@echo "Version = '$(VERSION)'"
	@echo "Image= '$(REPOSITORY):$(VERSION)'"

# CMDs
MOCKS_CMD := go generate ./mocks


#
#
################################################################################

# The default action of this Makefile is to build the development docker image
default: build

clean:
	@rm -rf $(OUT_BIN) || true
	@rm -f apis/db/v2/zz_generated.deepcopy.go || true

helm-package:
	@echo Packaging $(HELM_VERSION)
	helm package helm/multi-casskop
	mv multi-casskop-$(HELM_VERSION).tgz $(HELM_TARGET_DIR)
	helm repo index $(HELM_TARGET_DIR)/

# Build casskop executable file in local go env

export CGO_ENABLED:=0
export PURE:="on"

BUILD_FOLDER = multi-casskop
GENERATE_CMD = CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -o build/_output/bin/health \
				 ../tools/health/main.go
CRD_V1 = deploy/crds/multicluster_v1alpha1_cassandramulticluster_crd.yaml
CRD = deploy/crds/db.orange.com_multicasskops.yaml
FIRST_VERSION = .spec.versions[0]

.PHONY: update-crds
update-crds:
	echo Update CRD - Remove protocol and set config type to object CRD
# 	# We checkout v1alpha1 CRD and add it to v2 CRD as it must be known to do an upgrade
	git show v1.1.5-release:./$(CRD_V1) > /tmp/$$(basename $(CRD))
	sed -e '1,/versions/d' $(CRD) >> /tmp/$$(basename $(CRD))
	cp /tmp/$$(basename $(CRD)) $(CRD)
	yq -i e '$(FIRST_VERSION).storage = false' $(CRD)
	mv $(CRD) $$(echo $(CRD)|sed 's/.yaml/_crd.yaml/')
	cp -v deploy/crds/* helm/*/crds/

include ../shared.mk
include ../kube.mk

docker-generate-files:
	cd ../ && make docker-generate-k8s
	cp -v ../deploy/crds/* helm/*/crds/
	$(MAKE) docker-generate-k8s
	$(MAKE) docker-generate-crds

# Build the Operator and its Docker Image
docker-build: docker-generate-k8s docker-build-operator

# Build the Operator and its Docker Image
docker-build: docker-generate-files docker-build-operator

# Build CassKop executable file in local go env
.PHONY: build
build: generate
	echo "Build Cassandra Operator"
	operator-sdk build $(REPOSITORY):$(VERSION) --image-build-args "--build-arg https_proxy=$$https_proxy --build-arg http_proxy=$$http_proxy"
ifdef PUSHLATEST
	docker tag $(REPOSITORY):$(VERSION) $(REPOSITORY):latest
endif


build-local: generate
	@echo "Build Cassandra Operator for $(GOOS)"
	go build -o build/_output/bin/multi-casskop-$(GOOS) -gcflags all=-trimpath=github.com/Orange-OpenSource -asmflags all=-trimpath=github.com/Orange-OpenSource github.com/Orange-OpenSource/casskop/multi-casskop/cmd/manager

debug-telepresence:
	export TELEPRESENCE_REGISTRY=$(TELEPRESENCE_REGISTRY) ; \
	echo "execute : cat multi-casskop.env" ; \
	sudo mkdir -p /var/run/secrets/kubernetes.io ; \
	sudo mkdir -p /var/run/secrets/admiralty.io ; \
	sudo ln -s /tmp/known/var/run/secrets/kubernetes.io/serviceaccount /var/run/secrets/kubernetes.io/ ; \
	sudo ln -s /tmp/known/var/run/secrets/admiralty.io/serviceaccountimports /var/run/secrets/admiralty.io/ ; \
	tdep=$(shell kubectl get deployment -l app=multi-casskop -o jsonpath='{.items[0].metadata.name}') ; \
	telepresence --swap-deployment $$tdep --mount=/tmp/known --env-file multi-casskop.env 

ifneq (,$(filter $(firstword $(MAKECMDGOALS)),run run-local run-docker))
	RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
	$(eval $(RUN_ARGS):;@:)
endif


NAMESPACE ?= cassandra-demo
# Run the development environment (in local go env) in the background using local ~/.kube/config
.PHONY: run
run:
	export POD_NAME=multi-caaskop; \
	export WATCH_NAMESPACE=$(NAMESPACE); \
	export LOG_LEVEL=Debug; \
	operator-sdk up local --namespace cassandra-demo  --operator-flags "$(RUN_ARGS)"

run-local:
	export POD_NAME=multi-caaskop; \
	export WATCH_NAMESPACE=$(NAMESPACE); \
	export LOG_LEVEL=Debug; \
	./build/_output/bin/multi-casskop-$(GOOS) $(RUN_ARGS)

.PHONY: push
push:
	docker push $(REPOSITORY):$(VERSION)
ifdef PUSHLATEST
	docker push $(REPOSITORY):latest
endif

.PHONY: tag
tag:
	git tag $(VERSION)

.PHONY: publish
publish:
	@COMMIT_VERSION="$$(git rev-list -n 1 $(VERSION))"; \
	docker tag $(REPOSITORY):"$$COMMIT_VERSION" $(REPOSITORY):$(VERSION)
	docker push $(REPOSITORY):$(VERSION)
ifdef PUSHLATEST
	docker push $(REPOSITORY):latest
endif

.PHONY: release
release: tag image publish

unit-test-with-vendor:
	echo TODO
