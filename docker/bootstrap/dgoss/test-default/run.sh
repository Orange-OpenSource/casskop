#!/bin/bash -e

testScriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd ${testScriptDir}

source ../util.sh

createDgossVolumes
createInitConfigContainer
createCassandraBootstrapContainer
createAndCheckCassandraContainer
