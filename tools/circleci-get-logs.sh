#!/bin/bash

while true; do
    NAME=$(kubectl -n cassandra-e2e get pods -l app=cassandra-k8s-operator -o jsonpath='{range .items[*]}{.metadata.name}');
    READY=$(kubectl -n cassandra-e2e get pods -l app=cassandra-k8s-operator -o jsonpath='{range .items[*]}{.status.containerStatuses[0].ready}');

    echo "wait for logs: $NAME=$READY";
    sleep 3;

    if [[ "$NAME" != "" ]] && [[ "$READY" == "true" ]]; then
        break;
    fi
done ;

echo "Get Operator logs: $NAME";

kubectl -n cassandra-e2e logs $NAME -f > operator.log
