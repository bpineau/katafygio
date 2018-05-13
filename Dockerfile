FROM golang:1.10.1 as builder
WORKDIR /go/src/github.com/bpineau/katafygio
COPY . .
RUN go get -u github.com/Masterminds/glide
RUN make deps
RUN make build

FROM alpine:3.7
RUN apk upgrade --no-cache && apk --no-cache add ca-certificates git
RUN install -d -o nobody -g nobody /var/lib/katafygio/data
COPY --from=builder /go/src/github.com/bpineau/katafygio/katafygio /usr/bin/
VOLUME /var/lib/katafygio
USER nobody
ENTRYPOINT ["/usr/bin/katafygio"]
