FROM golang:1.11 as build

ADD . /go/src/github.com/Orange-OpenSource/cassandra-k8s-operator
WORKDIR /go/src/github.com/Orange-OpenSource/cassandra-k8s-operator
#
RUN useradd -u 1000 cassandra-k8s-operator
RUN go get ./tools/health
RUN mkdir -p /tmp && chown cassandra-k8s-operator /tmp

FROM gcr.io/distroless/base

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /go/bin/health /
COPY --from=build /tmp /tmp

LABEL org.opencontainers.image.documentation="https://github.com/Orange-OpenSource/cassandra-k8s-operator/blob/master/README.md"
LABEL org.opencontainers.image.authors="Sébastien Allamand <sebastien.allamand@orange.com>"
LABEL org.opencontainers.image.source="https://github.com/Orange-OpenSource/cassandra-k8s-operator"
LABEL org.opencontainers.image.vendor="Orange France - Digital Factory"
LABEL org.opencontainers.image.version="0.1"
LABEL org.opencontainers.image.description="Operateur des Gestion de Clusters Cassandra"
LABEL org.opencontainers.image.url="https://github.com/Orange-OpenSource/cassandra-k8s-operator"
LABEL org.opencontainers.image.title="Operateur Cassandra"

LABEL org.label-schema.usage="https://github.com/Orange-OpenSource/cassandra-k8s-operator/blob/master/README.md"
LABEL org.label-schema.docker.cmd="/usr/local/bin/cassandra-k8s-operator"
LABEL org.label-schema.docker.cmd.devel="N/A"
LABEL org.label-schema.docker.cmd.test="N/A"
LABEL org.label-schema.docker.cmd.help="N/A"
LABEL org.label-schema.docker.cmd.debug="N/A"
LABEL org.label-schema.docker.params="LOG_LEVEL=define loglevel,RESYNC_PERIOD=period in second to execute resynchronisation,WATCH_NAMESPACE=namespace to watch for cassandraclusters,OPERATOR_NAME=name of the operator instance pod"

# install operator binary
COPY build/_output/bin/cassandra-k8s-operator /usr/local/bin/

USER cassandra-k8s-operator

CMD ["/usr/local/bin/cassandra-k8s-operator"]
