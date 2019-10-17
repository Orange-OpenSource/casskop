# Copyright 2019 Orange
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PROJECT_ID?=orangeopensource
ifeq ($(CI_REGISTRY_IMAGE),)
	PROJECT?=${PROJECT_ID}/cassandra-bootstrap
else
	PROJECT:=$(CI_REGISTRY_IMAGE)
endif

VERSION:=0.1.2
TAG?=${VERSION}
ifeq ($(CIRCLE_BRANCH),master)
	BRANCH:=latest
else ifeq ($(CIRCLE_TAG),)
  BRANCH=$(CIRCLE_BRANCH)
else
  BRANCH=$(CIRCLE_TAG)
endif

.PHONY: all build build-cqlsh build-openjre build-openjre-cqlsh push push-cqlsh

all: build

params:
	@echo ${PROJECT} ${BRANCH}
	@echo $(VERSION) $(JAVA_VERSION) $(CASSANDRA_VERSION)
	@echo $(TAG)

build: params
	docker build --pull \
							 --build-arg https_proxy=$(https_proxy) --build-arg http_proxy=$(http_proxy) \
							 -t ${PROJECT}:${TAG} .

push: params
	docker push ${PROJECT}:${TAG}

