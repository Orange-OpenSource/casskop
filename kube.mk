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
# Some usefuls commands used with k8s

NAMESPACE=$(shell kubectl config view --minify -o jsonpath="{.contexts[0].context.namespace}")

################################################################################
#
# this is to use make with args
#
# ex: make nodetool status
ifeq (nodetool,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  NODETOOL_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(NODETOOL_ARGS):;@:)
endif

ifeq (kube-namespace,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  KUBENAMESPACE_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(KUBENAMESPACE_ARGS):;@:)
endif

ifeq (unit-test-dir,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  UNITTEST_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(UNITTEST_ARGS):;@:)
endif

ifeq (watch,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  WATCH_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(WATCH_ARGS):;@:)
endif

ifeq (status,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  STATUS_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(STATUS_ARGS):;@:)
endif

ifeq (node,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  NODE_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(NODE_ARGS):;@:)
endif

ifeq (get,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  GET_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(GET_ARGS):;@:)
endif

.PHONY: nodetool node get


debug:
	curl -ssLO https://raw.githubusercontent.com/kubernetes/contrib/master/scratch-debugger/debug.sh

#watch cassandracluster object
watch:
	watch 'kubectl get -o yaml cassandracluster $(WATCH_ARGS)'
cc:
	kubectl get -o yaml cassandracluster $(STATUS_ARGS)

.PHONY: debug watch cc watch-pods get-pods
watch-pods:
	watch "kubectl get pods -o=go-template='{{\"NAME NODE IP READY REASON\n\"}}{{range .items}}\
	{{printf \"%.30s\" .metadata.name}} \
	{{.spec.nodeName}} \
	{{.status.podIP}} \
	{{range .status.containerStatuses}}{{.ready}}{{\" \"}}{{if .state.running}}{{\"running\"}}{{else}}{{ .state.waiting.reason}}{{end}}{{end}} \
	{{\"\n\"}}{{end}}' | column -t"


list-app-by-nodes:
	KUBE_NODES=`kubectl get nodes -l node-role.kubernetes.io/node=true -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for node in $$KUBE_NODES; do \
	  echo "node $$node" ; \
	  kubectl describe node $$node | egrep "$(APP)|namespace" ; \
  done

#Force kubectl to use the NAMESPACE by default
kubectl-change-namespace:
	kubectl config set-context $$(kubectl config current-context) --namespace=$(NAMESPACE)

# Supprime le namespace
delete:
	kubectl delete namespace $(NAMESPACE)

# Create the Cassandra cluster (namespace + service + staefulset)
create:
	kubectl create namespace $(NAMESPACE)
	kubectl -n $(NAMESPACE) apply -f cassandra-service.yaml
	kubectl -n $(NAMESPACE) apply -f cassandra-service-metrics.yaml
	kubectl -n $(NAMESPACE) apply -f cassandra-PodDisruptionBudget.yaml
	make apply


delete-pvc:
	kubectl get pvc -o jsonpath='{.items[*].metadata.name}' | xargs k delete pvc

# Apply the cassandra statefulset
apply:
	ktmpl cassandra-statefulset-template.yaml -p NAMESPACE $(NAMESPACE) | kubectl apply -f - -n $(NAMESPACE)

get:
	kubectl -n $(NAMESPACE) get $(GET_ARGS) -o wide -w

# Go into each Cassandra pods and print hostname
hostname:
	KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for pod in $$KUBE_PODS; do kubectl -n $(NAMESPACE) exec -it $$pod -- sh -c 'hostname'; done

# Go into each Cassandra pods, and create a /var/lib/cassandra/name file with the host name
fullfil:
	KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for pod in $$KUBE_PODS; do kubectl -n $(NAMESPACE) exec -it $$pod -- sh -c "echo $$(hostname) > /var/lib/cassandra/$(NAMESPACE)-$$pod"; done



# Go into each Cassandra pods, and cat the value of /var/lib/cassandra/name + some env vars
check:
	KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for pod in $$KUBE_PODS; do kubectl -n $(NAMESPACE) exec -it $$pod -- sh -c 'cat /var/lib/cassandra/name ; env | grep POD'; done

#check seeds from config.yaml
check-seeds:
	@KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -l app=cassandracluster -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for pod in $$KUBE_PODS; do echo $$pod; \
    arr=$$(kubectl -n $(NAMESPACE) exec -it $$pod -- sh -c 'grep "seeds:" /etc/cassandra/cassandra.yaml' | cut -d":" -f2- | tr "," "\n"); \
    for x in $$arr; do echo $$x ; done; \
    echo "" ; \
  done
#check seeds from ENV
check-seeds-env:
	@KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -l app=cassandracluster -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for pod in $$KUBE_PODS; do echo $$pod; \
    arr=$$(kubectl -n $(NAMESPACE) exec -it $$pod -- sh -c 'env | grep SEEDS' | cut -d"=" -f2- | tr "," "\n"); \
    for x in $$arr; do echo $$x ; done; \
    echo "" ; \
  done


# for each cassandra pods, extract the docker image name from metadata
check-version:
	kubectl -n $(NAMESPACE) get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.spec.containers[0].image}{"\n"}'

check-annotations:
	kubectl -n $(NAMESPACE) get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.items[*]}{.metadata.annotations}{"\n"}'

check-nodes:
	kubectl get nodes --no-headers | awk '{print $$1}' | xargs -I {} sh -c 'echo {}; kubectl describe node {} | grep Allocated -A 5 | grep -ve Event -ve Allocated -ve percent -ve -- ; echo'

# Execute nodetool in the cassandra-0 pod
nodetool:
	@KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -o jsonpath='{.items[0].metadata.name}{" "}'` ; \
	kubectl -n $(NAMESPACE) exec -it $$KUBE_PODS -- nodetool $(NODETOOL_ARGS);

# Execute nodetool cleanup in each cassandra pods
nodetool-cleanup:
	KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for pod in $$KUBE_PODS; do kubectl -n $(NAMESPACE) exec -it $$pod -- sh -c 'nodetool cleanup'; done

nodetool-upgrade-sstable:
	KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for pod in $$KUBE_PODS; do kubectl -n $(NAMESPACE) exec -it $$pod -- sh -c 'nodetool upgradesstables'; done

nodetool-repair-full:
	KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -l app=cassandracluster -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for pod in $$KUBE_PODS; do kubectl -n $(NAMESPACE) exec -it $$pod -- sh -c 'echo "$$HOSTNAME:"; echo nodetool repair -full'; done

nodetool-repair-full2:
	KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -l app=cassandracluster -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
  echo $$KUBE_PODS | xargs -I XX kubectl -n $(NAMESPACE) exec -it XX -- sh -c 'echo "$$HOSTNAME:"; echo nodetool repair -full'


pod-repartition:
	echo to implement

check-pvc:
	PVC=(`kubectl -n $(NAMESPACE) get pvc -o jsonpath='{range .items[*]}{.metadata.name}{" "}'`) ; \
	PV=(`kubectl -n $(NAMESPACE) get pvc -o jsonpath='{range .items[*]}{.spec.volumeName}{" "}'`) ; \
	for i in $${!PVC[*]} ; do \
	  echo "Claim : $${PVC[$$i]}" ; \
	  echo "PV: $${PV[$$i]}" ; \
	  HOST=$$(kubectl -n $(NAMESPACE) get pv $${PV[$$i]} -o jsonpath='{.metadata.annotations.volume\.alpha\.kubernetes\.io/node-affinity}' | jq -r ".requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]") ; \
	  LPATH=$$(kubectl -n $(NAMESPACE) get pv $${PV[$$i]} -o jsonpath='{.spec.local.path}') ; \
		echo $$HOST ; \
		IP=$$(kubectl -n $(NAMESPACE) get node $$HOST -o jsonpath='{.status.addresses[0].address}') ; \
		echo $$LPATH ; \
	  echo $$IP ; \
	  ssh -q cloudwatt-k8s "ssh $$IP 'ls -la $$LPATH ; hostname'" ; \
	  echo "" ; \
	done


#Usage make node 'ls -la'
node:
	KUBE_IPS=`kubectl get nodes -l node-role.kubernetes.io/node=true -o jsonpath='{range .items[*]}{.status.addresses[0].address}{" "}'` ; \
  for x in $$KUBE_IPS; do \
    echo "Serveur $$x :" ; \
    ssh -q cloudwatt-k8s "ssh $$x '$(NODE_ARGS)'" ; \
  done;


annotate-upgradesstables:
	make check-annotations
	KUBE_PODS=`kubectl -n $(NAMESPACE) get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}'` ; \
	for x in $$KUBE_PODS; do kubectl -n $(NAMESPACE) annotate --overwrite pods $$x cc-action=upgradesstables; done
	make check-annotations


os:
	echo $(GOOS)

check-env:
	echo "working on OS type $(GOOS)"
	echo " Working with docker repository $(REPOSITORY)"

#usage: kube-namespace <mynamespace>
#permet de changer le namespace par default de kubectl
kube-namespace:
	kubectl config set-context $$(kubectl config current-context) --namespace=$(KUBENAMESPACE_ARGS)

#list all resources in current namespace
list-all:
	kubectl api-resources --verbs=list --namespaced -o name | grep -v events | xargs -n 1 kubectl get --show-kind --ignore-not-found

#get sorted events
events:
	kubectl get events -w --sort-by .metadata.creationTimestamp

events-all:
	kubectl get events -w --sort-by .metadata.creationTimestamp --all-namespaces
