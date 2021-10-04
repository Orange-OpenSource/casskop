FROM golang:1.17 as build

ENV GO111MODULE=on

RUN useradd -u 1000 casskop
RUN mkdir -p /tmp && chown casskop /tmp

ADD . /casskop

WORKDIR /casskop

RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o casskop main.go

FROM gcr.io/distroless/base

WORKDIR /
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /tmp /tmp
COPY --from=build /casskop/casskop /usr/local/bin/casskop
USER casskop

LABEL org.opencontainers.image.documentation="https://github.com/Orange-OpenSource/casskop/blob/master/README.md"
LABEL org.opencontainers.image.authors="Sébastien Allamand <sebastien.allamand@orange.com>"
LABEL org.opencontainers.image.source="https://github.com/Orange-OpenSource/casskop"
LABEL org.opencontainers.image.vendor="Orange France - Digital Factory"
LABEL org.opencontainers.image.version="0.1"
LABEL org.opencontainers.image.description="Operateur des Gestion de Clusters Cassandra"
LABEL org.opencontainers.image.url="https://github.com/Orange-OpenSource/casskop"
LABEL org.opencontainers.image.title="Operateur Cassandra"

LABEL org.label-schema.usage="https://github.com/Orange-OpenSource/casskop/blob/master/README.md"
LABEL org.label-schema.docker.cmd.devel="N/A"
LABEL org.label-schema.docker.cmd.test="N/A"
LABEL org.label-schema.docker.cmd.help="N/A"
LABEL org.label-schema.docker.cmd.debug="N/A"
LABEL org.label-schema.docker.params="LOG_LEVEL=define loglevel,RESYNC_PERIOD=period in second to execute resynchronisation,WATCH_NAMESPACE=namespace to watch for cassandraclusters,OPERATOR_NAME=name of the operator instance pod"



ENTRYPOINT ["/usr/local/bin/casskop"]
