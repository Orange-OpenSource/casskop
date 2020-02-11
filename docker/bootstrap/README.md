


# Cassandra bootsrap image for CassKop

This image aims to be used as an init-container to bootstrap Cassandra images to run with CassKop Kubernetes operator. It generates all the configuration necessary for the cassandra container to run. Our bootstrap image does the following :
- Copies any default config files from CassKop bootstrapper image to /etc/cassandra mounted path which will replace those used in Cassandra image
- Copies files from any mounted configmap (CassKop sets CONFIGMAP env var)
- Copies any extra libraries from this bootstrap image to the /extra-lib (the default image provides instaclustr's cassandra-exporter and Jolokia agent)
- Copies tools from /bootstrap/tools to /opt/bin
- Executes script ${CONFIGMAP}/pre_run.sh if provided
- Executes script run.sh
- Executes script ${CONFIGMAP}/post_run.sh if provided

## Custom boostrap image Requirements

In order to use your own bootstrap image, CassKop requires that it provides the binary curl in /bootstrap/tools (used by default readiness/liveness probes) and generates the cassandra configuration files correctly (see [run.sh](files/run.sh)). The easiest is to start with our image and just overwrite what you want to change in order to miss anything like the activation of Jolokia that is used by CassKop. As a recommendation, whenever you think you need a custom bootstrap image, try first to do your changes through a pre-run.sh script. Take a look at [that example](dgoss/test-with-pre-run/pre_run.sh).