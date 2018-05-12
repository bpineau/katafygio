FROM golang:1.10.1 as builder
WORKDIR /go/src/github.com/bpineau/katafygio
COPY . .
RUN go get -u github.com/Masterminds/glide
RUN make deps
RUN make build

FROM alpine:3.7
RUN apk upgrade --no-cache && apk --no-cache add ca-certificates git
COPY --from=builder /go/src/github.com/bpineau/katafygio/katafygio /usr/bin/
USER nobody
ENTRYPOINT ["/usr/bin/katafygio"]
