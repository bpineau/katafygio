GOLANGCI_VERSION=1.21.0
export GO111MODULE := on

all: build

tools:
	which golangci-lint || ( \
	  curl -sfL "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_VERSION}/golangci-lint-${GOLANGCI_VERSION}-linux-amd64.tar.gz" | \
	  tar xz --strip-components 1 --wildcards '*/golangci-lint' && \
	  chmod 755 golangci-lint && \
	  mv golangci-lint ${GOPATH}/bin/ \
	)
	which goveralls || go get github.com/mattn/goveralls

lint:
	@# govet, errcheck etc are already on by default. this -E enable extra linters:
	golangci-lint run -E gofmt,golint,unconvert,dupl,goimports,misspell,maligned,stylecheck

man:
	go run assets/manpage.go

fmt:
	go fmt ./...

build:
	env CGO_ENABLED=0 go build

install:
	env CGO_ENABLED=0 go install

clean:
	rm -rf dist/
	go clean

coverall:
	goveralls -coverprofile=profile.cov -service=travis-ci -package github.com/bpineau/katafygio/pkg/...

e2e:
	kubectl get ns >/dev/null || exit 1
	go test -count=1 -v assets/e2e_test.go

test:
	go test github.com/bpineau/katafygio/...
	go test -race -cover -coverprofile=profile.cov github.com/bpineau/katafygio/...

.PHONY: tools lint fmt install clean coverall test all
