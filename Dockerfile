FROM golang:1.15 as builder
WORKDIR /go/src/github.com/bpineau/katafygio
COPY . .
RUN make build

FROM alpine:3.12
RUN apk upgrade --no-cache && \
    apk --no-cache add ca-certificates git openssh-client tini
RUN install -d -o nobody -g nobody /var/lib/katafygio/data
COPY --from=builder /go/src/github.com/bpineau/katafygio/katafygio /usr/bin/
VOLUME /var/lib/katafygio
USER 65534
ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/katafygio"]
