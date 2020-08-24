#!/usr/bin/env bash
set -e

# Copies any default config files from CassKop bootstrapper image to the /etc/cassandra volume which will replace the one used in cassandra image
cp -rLv /${BOOTSTRAP_CONF}/* /etc/cassandra/

# If User has submited a configMap, we uses them to deplace default ones
# (overwriting the above)
if [[ -d ${CONFIGMAP} && ! -z `ls -A ${CONFIGMAP}` ]] ; then
    echo "We have a ConfigMap, we surcharge default configuration files"
    cp -rLv ${CONFIGMAP}/* /etc/cassandra/
fi

# Copies any extra libraries from this bootstrapper image to the extra-lib empty-dir
if [[ -d /${BOOTSTRAP_LIBS} && ! -z `ls -A /${BOOTSTRAP_LIBS}` ]] ; then
    echo "We have additional libraries, we copy them over"
    cp -v ${BOOTSTRAP_LIBS}/* $CASSANDRA_LIBS/
fi

cp -v /${BOOTSTRAP_TOOLS}/* $CASSANDRA_TOOLS/

if [ -f ${CONFIGMAP}/pre_run.sh ]; then
    echo "We found pre_run.sh script, we execute it"
    ${CONFIGMAP}/pre_run.sh
fi

# Bootstrap Cassandra configuration
echo " == We execute bootstrap script run.sh"
/${BOOTSTRAP_CONF}/run.sh

if [ -f ${CONFIGMAP}/post_run.sh ]; then
    echo " == We found post_run.sh script, we execute it"
    ${CONFIGMAP}/post_run.sh
fi

echo '== bootstrap ended :-)'
