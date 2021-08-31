#!/bin/bash

CASSANDRA_IMAGE=cassandra:3.11.6
CASSANDRA_SEEDS=cassandra-demo-dc1-rack1-0.cassandra-demo.ns,cassandra-demo-dc1-rack2-0.cassandra-demo.ns,cassandra-demo-dc1-rack3-0.cassandra-demo.ns
CASSANDRA_UID=1000
CASSANDRA_DC=dc1
CASSANDRA_RACK=rack1

CONFIG_FILE_DATA=$(cat <<-EOF
    {
      "cassandra-yaml": {
        "max_hints_delivery_threads": 8,
        "authenticator": "PasswordAuthenticator",
        "authorizer": "CassandraAuthorizer"
      },
      "cluster-info": {
        "name": "cassandra-e2e",
        "seeds": "$CASSANDRA_SEEDS"
      },
      "datacenter-info": {
        "name": "$CASSANDRA_DC"
      },
      "jvm-options": {
        "cassandra_ring_delay_ms": 30000,
        "initial_heap_size": "64M",
        "jmx-connection-type": "remote-no-auth",
        "max_heap_size": "256M",
        "print_flss_statistics": true
      },
      "logback-xml": {
        "debuglog-enabled": false
      }
    }
EOF
)

BOOTSTRAP_VOLUME=dgoss-bootstrap-vol
EXTRA_LIB_VOLUME=dgoss-extralib-vol
CONFIGMAP_VOLUME=dgoss-configmap-vol
TOOLS_VOLUME=dgoss-tools-vol

function createDgossVolumes {
    echo "== createDgossVolumes"
    docker volume create ${BOOTSTRAP_VOLUME}
    docker volume create ${EXTRA_LIB_VOLUME}
    docker volume create ${CONFIGMAP_VOLUME}
    docker volume create ${TOOLS_VOLUME}
}

function deleteDgossVolumes {
    echo "== deleteDgossVolumes"
    docker volume rm ${BOOTSTRAP_VOLUME} ${EXTRA_LIB_VOLUME} ${CONFIGMAP_VOLUME} ${TOOLS_VOLUME}
}

function createInitConfigContainer {
    echo "== createInitConfigContainer"
    docker run --rm -u 0 \
           -v ${BOOTSTRAP_VOLUME}:/bootstrap \
           -e PRODUCT_NAME=$(echo $CASSANDRA_IMAGE|cut -d: -f1) \
           -e PRODUCT_VERSION=$(echo $CASSANDRA_IMAGE|cut -d: -f2) \
           -e CONFIG_FILE_DATA="$CONFIG_FILE_DATA" \
           -e CONFIG_OUTPUT_DIRECTORY=/bootstrap \
           -e RACK_NAME=$CASSANDRA_RACK \
           datastax/cass-config-builder:1.0.3

    docker run --rm \
           -v ${BOOTSTRAP_VOLUME}:/bootstrap \
           -v ${EXTRA_LIB_VOLUME}:/extra-lib \
           --entrypoint=bash \
           ${CASSANDRA_IMAGE} \
           -c "chown -R cassandra: /bootstrap /extra-lib"
}

function checkDgossVolumes {
    echo "== checkVolumes (manually debug)"
    docker run \
           --rm -ti \
           -v ${BOOTSTRAP_VOLUME}:/etc/cassandra/ \
           -v ${EXTRA_LIB_VOLUME}:/extra-lib/ \
           --entrypoint=bash \
           ${IMAGE_TO_TEST}
}

function createCassandraBootstrapContainer {
    echo "== createCassandraBootstrapContainer"
    docker run \
           -u 999 \
           --rm -ti \
           -e CASSANDRA_SEEDS=$CASSANDRA_SEEDS \
           -e HOSTNAME=cassandra-seb-dc1-rack1-0 \
           -e POD_NAME=cassandra-demo-dc1-rack1-0 \
           -e CASSANDRA_DC=$CASSANDRA_DC \
           -e CASSANDRA_RACK=$CASSANDRA_RACK \
           -v ${BOOTSTRAP_VOLUME}:/etc/cassandra/ \
           -v ${CONFIGMAP_VOLUME}:/configmap \
           -v ${EXTRA_LIB_VOLUME}:/extra-lib/ \
           -v ${TOOLS_VOLUME}:/opt/bin/ \
           ${IMAGE_TO_TEST} 
}

#disable jolokia & exporter
function createCassandraBootstrapContainerNoExtraLib {
    echo "== createCassandraBootstrapContainer"
    docker run \
           -u 999 \
           --rm -ti \
           -e CASSANDRA_SEEDS=$CASSANDRA_SEEDS \
           -e HOSTNAME=cassandra-seb-dc1-rack1-0 \
           -e POD_NAME=cassandra-demo-dc1-rack1-0 \
           -e CASSANDRA_ENABLE_JOLOKIA=false \
           -e CASSANDRA_EXPORTER_AGENT=false \
           -e CASSANDRA_DC=$CASSANDRA_DC \
           -e CASSANDRA_RACK=$CASSANDRA_RACK \
           -v ${BOOTSTRAP_VOLUME}:/etc/cassandra/ \
           ${IMAGE_TO_TEST} 
}

function createCassandraBootstrapContainerWithConfigMap {
    echo "== createCassandraBootstrapContainer"
    docker run \
          -u 999 \
           --rm -ti \
           -e CASSANDRA_SEEDS=$CASSANDRA_SEEDS \
           -e CASSANDRA_CLUSTER_NAME=cassandra-demo \
           -e HOSTNAME=cassandra-seb-dc1-rack1-0 \
           -e POD_NAME=cassandra-demo-dc1-rack1-0 \
           -e CASSANDRA_DC=$CASSANDRA_DC \
           -e CASSANDRA_RACK=$CASSANDRA_RACK \
           -v ${BOOTSTRAP_VOLUME}:/etc/cassandra/ \
           -v ${CONFIGMAP_VOLUME}:/configmap \
           -v ${EXTRA_LIB_VOLUME}:/extra-lib/ \
           ${IMAGE_TO_TEST} 
}

function createSimpleConfigMapFile {
    echo "== createSimpleConfigMapFile"
    docker run \
       --rm \
       -v ${CONFIGMAP_VOLUME}:/configmap \
       --entrypoint=bash \
       ${CASSANDRA_IMAGE} \
       -c "echo 'this is additional file' > /configmap/additional-file.yaml"
}

#the bind-mount does not seams to works in circleci, so I'll do a cp
function createPreRunConfigMapFile {
    echo "== createPreRunConfigMapFile"
    docker container create --name dummy -v ${CONFIGMAP_VOLUME}:/configmap hello-world
    docker cp ${testScriptDir}/pre_run.sh dummy:/configmap/pre_run.sh
    docker rm dummy
}

function createCassandraContainer {
    echo "== createCassandraContainer"
    docker run \
        -v ${BOOTSTRAP_VOLUME}:/etc/cassandra \
        -v ${EXTRA_LIB_VOLUME}:/extra-lib \
        -v ${TOOLS_VOLUME}:/opt/bin/ \
        -p 9500:9500 \
        -p 8778:8778 \
        ${CASSANDRA_IMAGE}

}

function createAndCheckCassandraContainer {
    echo "== createAndCheckCassandraContainer"
    local script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    GOSS_FILES_PATH=${script_dir}/checks GOSS_SLEEP=5 GOSS_WAIT_OPTS='-r 90s -s 1s > /dev/null' dgoss run \
                   -v ${BOOTSTRAP_VOLUME}:/etc/cassandra \
                   -v ${EXTRA_LIB_VOLUME}:/extra-lib \
                   -v ${TOOLS_VOLUME}:/opt/bin/ \
                   -p 9500:9500 \
                   -p 8778:8778 \
                   ${CASSANDRA_IMAGE}

}


trap deleteDgossVolumes EXIT SIGTERM SIGKILL
