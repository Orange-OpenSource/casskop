#!/bin/bash -e

testScriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [[ -z ${IMAGE_TO_TEST} ]] ; then
    echo Expected IMAGE_TO_TEST to be specified
    exit 1
fi

for testDirectory in `find ${testScriptDir}/test-* -type d` ; do
    echo ===== Running ${testDirectory} =====
    bash ${testDirectory}/run.sh
    echo ===== End of ${testDirectory} =====
    echo
done
