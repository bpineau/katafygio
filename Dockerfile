FROM golang:1.12 as builder
WORKDIR /go/src/github.com/bpineau/katafygio
COPY . .
RUN go get -u github.com/Masterminds/glide
RUN make deps
RUN make build

FROM alpine:3.10
RUN apk upgrade --no-cache && \
    apk --no-cache add ca-certificates git openssh-client tini
RUN install -d -o nobody -g nobody /var/lib/katafygio/data
COPY --from=builder /go/src/github.com/bpineau/katafygio/katafygio /usr/bin/
VOLUME /var/lib/katafygio
USER nobody
ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/katafygio"]
