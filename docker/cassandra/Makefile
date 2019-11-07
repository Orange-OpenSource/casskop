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
	PROJECT?=${PROJECT_ID}/cassandra-image
else
	PROJECT:=$(CI_REGISTRY_IMAGE)
endif

VERSION:=3.11
TAG?=${VERSION}

ifdef CIRCLE_BRANCH
	#removing / for fork which lead to docker error
	BRANCH := $(subst /,-,$(CIRCLE_BRANCH))
else
  ifdef CIRCLE_TAG
		BRANCH := $(CIRCLE_TAG)
	else
		BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
	endif
endif

.PHONY: all build build-cqlsh build-openjre build-openjre-cqlsh push push-cqlsh

all: build

params:
	@echo "CIRCLE_BRANCH = '$(CIRCLE_BRANCH)'"
	@echo "CIRCLE_TAG = '$(CIRCLE_TAG)'"
	@echo "Version = '$(VERSION)'"
	@echo "Branch = '$(BRANCH)'"
	@echo "Tag = '$(TAG)'"


build: params
	docker build --pull \
							 --build-arg https_proxy=$(https_proxy) --build-arg http_proxy=$(http_proxy) \
							 -t ${PROJECT}:${TAG} .

push: params
ifeq ($(BRANCH),master)
	docker push ${PROJECT}:${TAG}
endif
	docker tag ${PROJECT}:${TAG} ${PROJECT}:${BRANCH}
	docker push ${PROJECT}:${BRANCH}

