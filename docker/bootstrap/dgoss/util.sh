#!/bin/bash

CASSANDRA_IMAGE=cassandra:latest

CASSANDRA_UID=1000

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
    docker run --rm \
           -v ${BOOTSTRAP_VOLUME}:/bootstrap \
           -v ${EXTRA_LIB_VOLUME}:/extra-lib \
           --entrypoint=bash \
           ${CASSANDRA_IMAGE} \
           -c "cp -vr /etc/cassandra/* /bootstrap && chown -R cassandra: /bootstrap /extra-lib"
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
           -e CASSANDRA_MAX_HEAP=1024M \
           -e CASSANDRA_SEEDS=cassandra-demo-dc1-rack1-0.cassandra-demo.ns,cassandra-demo-dc1-rack2-0.cassandra-demo.ns,cassandra-demo-dc1-rack3-0.cassandra-demo.ns \
           -e CASSANDRA_CLUSTER_NAME=cassandra-demo \
           -e CASSANDRA_AUTO_BOOTSTRAP=true \
           -e HOSTNAME=cassandra-seb-dc1-rack1-0 \
           -e POD_NAME=cassandra-demo-dc1-rack1-0 \
           -e POD_NAMESPACE=ns \
           -e CASSANDRA_GC_STDOUT=true \
           -e CASSANDRA_NUM_TOKENS=32 \
           -e CASSANDRA_DC=dc1 \
           -e CASSANDRA_RACK=rack1 \
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
           -e CASSANDRA_MAX_HEAP=1024M \
           -e CASSANDRA_SEEDS=cassandra-demo-dc1-rack1-0.cassandra-demo.ns,cassandra-demo-dc1-rack2-0.cassandra-demo.ns,cassandra-demo-dc1-rack3-0.cassandra-demo.ns \
           -e CASSANDRA_CLUSTER_NAME=cassandra-demo \
           -e CASSANDRA_AUTO_BOOTSTRAP=true \
           -e HOSTNAME=cassandra-seb-dc1-rack1-0 \
           -e POD_NAME=cassandra-demo-dc1-rack1-0 \
           -e POD_NAMESPACE=ns \
           -e CASSANDRA_GC_STDOUT=true \
           -e CASSANDRA_ENABLE_JOLOKIA=false \
           -e CASSANDRA_EXPORTER_AGENT=false \
           -e CASSANDRA_NUM_TOKENS=32 \
           -e CASSANDRA_DC=dc1 \
           -e CASSANDRA_RACK=rack1 \
           -v ${BOOTSTRAP_VOLUME}:/etc/cassandra/ \
           ${IMAGE_TO_TEST} 
}

function createCassandraBootstrapContainerWithConfigMap {
    echo "== createCassandraBootstrapContainer"
    docker run \
          -u 999 \
           --rm -ti \
           -e CASSANDRA_MAX_HEAP=1024M \
           -e CASSANDRA_SEEDS=cassandra-demo-dc1-rack1-0.cassandra-demo.ns,cassandra-demo-dc1-rack2-0.cassandra-demo.ns,cassandra-demo-dc1-rack3-0.cassandra-demo.ns \
           -e CASSANDRA_CLUSTER_NAME=cassandra-demo \
           -e CASSANDRA_AUTO_BOOTSTRAP=true \
           -e HOSTNAME=cassandra-seb-dc1-rack1-0 \
           -e POD_NAME=cassandra-demo-dc1-rack1-0 \
           -e POD_NAMESPACE=ns \
           -e CASSANDRA_GC_STDOUT=true \
           -e CASSANDRA_NUM_TOKENS=32 \
           -e CASSANDRA_DC=dc1 \
           -e CASSANDRA_RACK=rack1 \
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
