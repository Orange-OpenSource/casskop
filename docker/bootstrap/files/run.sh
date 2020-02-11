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

default_value()
{
    echo $(grep -v '#' ${CASSANDRA_CFG}|grep -i \\b$1|awk '{print $2}')
}

HOSTNAME=$(hostname -f)

# we are doing StatefulSet or just setting our seeds
if [ -n "$CASSANDRA_SEEDS" ]; then
    echo "CASSANDRA_SEEDS=$CASSANDRA_SEEDS"
    # Try to connect to each seed and if no one is found then it's the first node
    echo $IFS
    IFS=',' read -a array <<<$CASSANDRA_SEEDS
    firstNode=true
    for cassandra in ${array[@]}
    do
        echo "Try to connect to $cassandra"
        if nc -z -w5 $cassandra 9042
        then
            echo "Connected!"
            firstNode=false
            break
        fi
    done

    [ "$firstNode" = true ] && CASSANDRA_SEEDS=$HOSTNAME

fi

echo "CASSANDRA_SEEDS=$CASSANDRA_SEEDS"

# The following vars relate to there counter parts in $CASSANDRA_CFG for instance rpc_address
CASSANDRA_LISTEN_ADDRESS=${POD_IP:-$HOSTNAME}
CASSANDRA_BROADCAST_ADDRESS=${POD_IP:-$HOSTNAME}
CASSANDRA_BROADCAST_RPC_ADDRESS=${POD_IP:-$HOSTNAME}
CASSANDRA_RPC_ADDRESS="${CASSANDRA_RPC_ADDRESS:-$(default_value rpc_address)}"
CASSANDRA_NUM_TOKENS="${CASSANDRA_NUM_TOKENS:-$(default_value num_tokens)}"
CASSANDRA_DISK_OPTIMIZATION_STRATEGY="${CASSANDRA_DISK_OPTIMIZATION_STRATEGY:-$(default_value disk_optimization_strategy)}"
CASSANDRA_ENDPOINT_SNITCH="${CASSANDRA_ENDPOINT_SNITCH:-$(default_value endpoint_snitch)}"
CASSANDRA_MIGRATION_WAIT="${CASSANDRA_MIGRATION_WAIT:-1}"
CASSANDRA_RING_DELAY="${CASSANDRA_RING_DELAY:-30000}"
CASSANDRA_AUTO_BOOTSTRAP="${CASSANDRA_AUTO_BOOTSTRAP:-true}"
CASSANDRA_SEEDS="${CASSANDRA_SEEDS:false}"
CASSANDRA_SEED_PROVIDER="${CASSANDRA_SEED_PROVIDER:-org.apache.cassandra.locator.SimpleSeedProvider}"
CASSANDRA_DC="${CASSANDRA_DC}"
CASSANDRA_RACK="${CASSANDRA_RACK}"
CASSANDRA_CLUSTER_NAME="${CASSANDRA_CLUSTER_NAME:='Test Cluster'}"
CASSANDRA_AUTHENTICATOR="${CASSANDRA_AUTHENTICATOR:-$(default_value authenticator)}"
CASSANDRA_AUTHORIZER="${CASSANDRA_AUTHORIZER:-$(default_value authorizer)}"

# Enable cassandra exporter
CASSANDRA_EXPORTER_AGENT="${CASSANDRA_EXPORTER_AGENT:-true}"
# Enable JMX
CASSANDRA_ENABLE_JMX="${CASSANDRA_ENABLE_JMX:-true}"
# Enable Jolokia
CASSANDRA_ENABLE_JOLOKIA="${CASSANDRA_ENABLE_JOLOKIA:-true}"
# Activate basic authentication. Expects JMX_USER and JMX_PASSWORD to be set
CASSANDRA_AUTH_JOLOKIA="${CASSANDRA_AUTH_JOLOKIA:-false}"

# send GC to STDOUT
CASSANDRA_GC_STDOUT="${CASSANDRA_GC_STDOUT:-false}"

# verbose GC logging
CASSANDRA_GC_VERBOSE="${CASSANDRA_GC_VERBOSE:-false}"

echo Starting Cassandra on ${CASSANDRA_LISTEN_ADDRESS}
echo Configuration used :
set|grep CASSANDRA

echo "configuring DC/Racks"
# if DC and RACK are set, use GossipingPropertyFileSnitch
if [[ $CASSANDRA_DC && $CASSANDRA_RACK ]]
then
  echo "dc=$CASSANDRA_DC" > $CASSANDRA_CONF/cassandra-rackdc.properties
  echo "rack=$CASSANDRA_RACK" >> $CASSANDRA_CONF/cassandra-rackdc.properties
  CASSANDRA_ENDPOINT_SNITCH="GossipingPropertyFileSnitch"
fi

if [ -n "$CASSANDRA_MAX_HEAP" ]
then
    sed -ri -e "s/^(#)?-Xmx[0-9]+.*/-Xmx$CASSANDRA_MAX_HEAP/"  \
        -e "s/^(#)?-Xms[0-9]+.*/-Xms$CASSANDRA_MAX_HEAP/" "$CASSANDRA_CONF/jvm.options"
fi

if [ -n "$CASSANDRA_REPLACE_NODE" ]
then
   echo "-Dcassandra.replace_address=$CASSANDRA_REPLACE_NODE/" >> "$CASSANDRA_CONF/jvm.options"
fi

echo "apply configuration changes"
# TODO what else needs to be modified
for yaml in \
  broadcast_address \
  broadcast_rpc_address \
  cluster_name \
  disk_optimization_strategy \
  endpoint_snitch \
  listen_address \
  num_tokens \
  rpc_address \
  start_rpc \
  key_cache_size_in_mb \
  concurrent_reads \
  concurrent_writes \
  memtable_cleanup_threshold \
  memtable_allocation_type \
  memtable_flush_writers \
  concurrent_compactors \
  compaction_throughput_mb_per_sec \
  counter_cache_size_in_mb \
  internode_compression \
  endpoint_snitch \
  gc_warn_threshold_in_ms \
  listen_interface \
  rpc_interface \
  authenticator \
  authorizer
do
  var="CASSANDRA_${yaml^^}"
  val="${!var}"
  if [ "$val" ]
  then
    sed -ri 's/^(# )?('"$yaml"':).*/\2 '"$val"'/' "$CASSANDRA_CFG"
  fi
done

echo "auto_bootstrap: ${CASSANDRA_AUTO_BOOTSTRAP}" >> $CASSANDRA_CFG

# set the seed to itself.  This is only for the first pod, otherwise
# it will be able to get seeds from the seed provider
if [[ $CASSANDRA_SEEDS == 'false' ]]
then
  sed -ri 's/- seeds:.*/- seeds: "'"$POD_IP"'"/' $CASSANDRA_CFG
else # if we have seeds set them.  Probably StatefulSet
  sed -ri 's/- seeds:.*/- seeds: "'"$CASSANDRA_SEEDS"'"/' $CASSANDRA_CFG
fi

sed -ri 's/- class_name: SEED_PROVIDER/- class_name: '"$CASSANDRA_SEED_PROVIDER"'/' $CASSANDRA_CFG

if [[ $CASSANDRA_GC_STDOUT == 'true' ]]
then
    # send gc to stdout
  sed -ri 's/JVM_OPTS.*-Xloggc:.*//' $CASSANDRA_CONF/cassandra-env.sh
else
    echo "send GC logs to ${CASSANDRA_DATA}/log/gc.log"
  mkdir -p "${CASSANDRA_DATA}/log/"
  echo "-Xloggc:${CASSANDRA_DATA}/log/gc.log" >> $CASSANDRA_CONF/jvm.options
  echo "-XX:+UseGCLogFileRotation" >> $CASSANDRA_CONF/jvm.options
  echo "-XX:NumberOfGCLogFiles=10" >> $CASSANDRA_CONF/jvm.options
  echo "-XX:GCLogFileSize=10M" >> $CASSANDRA_CONF/jvm.options
fi

if [[ $CASSANDRA_GC_VERBOSE == 'true' ]]
then
  echo "-XX:PrintFLSStatistics=1" >> $CASSANDRA_CONF/jvm.options
fi

# getting WARNING messages with Migration Service
echo "-Dcassandra.migration_task_wait_in_seconds=${CASSANDRA_MIGRATION_WAIT}" >> $CASSANDRA_CONF/jvm.options
echo "-Dcassandra.ring_delay_ms=${CASSANDRA_RING_DELAY}" >> $CASSANDRA_CONF/jvm.options

[[ $CASSANDRA_ENABLE_JOLOKIA == 'true' ]] && CASSANDRA_ENABLE_JMX=true

if [[ $CASSANDRA_ENABLE_JMX == 'true' ]]
then
  sed -ri '/^JMX_PORT=/a LOCAL_JMX=no' $CASSANDRA_CONF/cassandra-env.sh
  sed -ri 's@ -Dcom\.sun\.management\.jmxremote\.authenticate=true@ -Dcom\.sun\.management\.jmxremote\.authenticate=false@' $CASSANDRA_CONF/cassandra-env.sh
  sed -ri 's@ -Dcom\.sun\.management\.jmxremote\.password\.file=/etc/cassandra/jmxremote\.password/@@' $CASSANDRA_CONF/cassandra-env.sh

  if [[ $CASSANDRA_ENABLE_JOLOKIA == 'true' ]]
  then
      JAVA_AGENT="-javaagent:/extra-lib/jolokia-agent.jar=host=0.0.0.0,executor=fixed"

      if [[ $CASSANDRA_AUTH_JOLOKIA == 'true' ]]
      then

          [[ -z $JOLOKIA_USER ]] && { echo "Jolokia authentication requires at least JOLOKIA_USER to be set !" >&2; exit 1; }

          JAVA_AGENT="${JAVA_AGENT},authMode=basic,user=$JOLOKIA_USER,password=$JOLOKIA_PASSWORD"

      fi

      cat  <<EOF >>$CASSANDRA_CONF/cassandra-env.sh

# Enable Jolokia
JVM_OPTS="\$JVM_OPTS $JAVA_AGENT"
EOF

  fi
fi

if [[ $CASSANDRA_EXPORTER_AGENT == 'true' ]]
then
    cat  <<EOF >>$CASSANDRA_CONF/cassandra-env.sh

# Prometheus exporter from Instaclustr
JVM_OPTS="\$JVM_OPTS -javaagent:/extra-lib/cassandra-exporter-agent.jar=@/etc/cassandra/exporter.conf"
EOF
fi
