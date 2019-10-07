#!/bin/bash -e

testScriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd ${testScriptDir}

source ../util.sh


createDgossVolumes
# Add test file to user config map
createInitConfigContainer
createCassandraBootstrapContainerNoExtraLib
#createAndCheckCassandraContainer # already tested

#createCassandraContainer

# check using test specific `goss.yaml`
 GOSS_WAIT_OPTS='-r 90s -s 1s > /dev/null' dgoss run \
          -v ${BOOTSTRAP_VOLUME}:/etc/cassandra \
          ${CASSANDRA_IMAGE}
