FROM golang:1.12 as build

ENV GO111MODULE=on

RUN useradd -u 1000 multi-casskop
RUN mkdir -p /tmp && chown multi-casskop /tmp

FROM gcr.io/distroless/base

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /tmp /tmp

ENV OPERATOR=/usr/local/bin/multi-casskop \
    USER_UID=1000 \
    USER_NAME=multi-casskop

# install operator binary
COPY build/_output/bin/multi-casskop ${OPERATOR}
COPY build/_output/bin/health /health

CMD ["/usr/local/bin/multi-casskop"]

USER ${USER_UID}
