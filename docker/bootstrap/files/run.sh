#!/bin/bash
#
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

set -e

CASSANDRA_CFG=$CASSANDRA_CONF/cassandra.yaml

# The following vars relate to there counter parts in $CASSANDRA_CFG for instance rpc_address
CASSANDRA_SEED_PROVIDER="${CASSANDRA_SEED_PROVIDER:-org.apache.cassandra.locator.SimpleSeedProvider}"

# Enable cassandra exporter
CASSANDRA_EXPORTER_AGENT="${CASSANDRA_EXPORTER_AGENT:-true}"

# Activate basic authentication. Expects JMX_USER and JMX_PASSWORD to be set
CASSANDRA_AUTH_JOLOKIA="${CASSANDRA_AUTH_JOLOKIA:-false}"

echo Starting Cassandra on ${CASSANDRA_LISTEN_ADDRESS}
echo Configuration used :
set|grep CASSANDRA

if [ -n "$CASSANDRA_REPLACE_NODE" ]
then
   echo "-Dcassandra.replace_address=$CASSANDRA_REPLACE_NODE/" >> "$CASSANDRA_CONF/jvm.options"
fi

sed -ri 's/- class_name: .*/- class_name: '"$CASSANDRA_SEED_PROVIDER"'/' $CASSANDRA_CFG

JAVA_AGENT="-javaagent:/extra-lib/jolokia-agent.jar=host=0.0.0.0,executor=fixed"

if [[ $JOLOKIA_USER == 'true' ]]
then
    JAVA_AGENT="${JAVA_AGENT},authMode=basic,user=$JOLOKIA_USER,password=$JOLOKIA_PASSWORD"
fi

cat  <<EOF >>$CASSANDRA_CONF/cassandra-env.sh

# Enable Jolokia
JVM_OPTS="\$JVM_OPTS $JAVA_AGENT"
EOF

if [[ $CASSANDRA_EXPORTER_AGENT == 'true' ]]
then
    cat  <<EOF >>$CASSANDRA_CONF/cassandra-env.sh

# Prometheus exporter from Instaclustr
JVM_OPTS="\$JVM_OPTS -javaagent:/extra-lib/cassandra-exporter-agent.jar=@/etc/cassandra/exporter.conf"
EOF

fi
