FROM golang:1.17 as build

ENV GO111MODULE=on

RUN useradd -u 1000 multi-casskop

COPY . /workspace
WORKDIR /workspace/multi-casskop

RUN go mod edit -replace github.com/Orange-OpenSource/casskop=/workspace
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
RUN go mod vendor

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o multi-casskop main.go

FROM gcr.io/distroless/base

WORKDIR /
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /tmp /tmp
COPY --from=build /workspace/multi-casskop/multi-casskop /usr/local/bin/multi-casskop

ENV OPERATOR=/usr/local/bin/multi-casskop \
    USER_UID=1000 \
    USER_NAME=multi-casskop
    
ENTRYPOINT ["/usr/local/bin/multi-casskop"]

USER ${USER_UID}
