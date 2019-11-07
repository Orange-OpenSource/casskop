#!/bin/bash

echo "** this is a pre-scrip for run.sh that can be edit with configmap"

grep max_hints_delivery_threads /etc/cassandra/cassandra.yaml
sed -i 's/max_hints_delivery_threads: 2/max_hints_delivery_threads: 8/' /etc/cassandra/cassandra.yaml
grep max_hints_delivery_threads /etc/cassandra/cassandra.yaml

echo "Change default Authenticator & Authorizer"
sed -ri 's/(authenticator:).*/\1 PasswordAuthenticator/' /etc/cassandra/cassandra.yaml
sed -ri 's/(authorizer:).*/\1 CassandraAuthorizer/' /etc/cassandra/cassandra.yaml
