GOLANGCI_VERSION=1.33.0
export GO111MODULE := on

all: build

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

${GOPATH}/bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b ${GOPATH}/bin v${GOLANGCI_VERSION}

${GOPATH}/bin/goveralls:
	GO111MODULE=off go get github.com/mattn/goveralls

tools: ${GOPATH}/bin/golangci-lint ${GOPATH}/bin/goveralls

man:
	go run assets/manpage.go

build:
	env CGO_ENABLED=0 go build

install:
	env CGO_ENABLED=0 go install

clean:
	rm -rf dist/ profile.cov katafygio katafygio.8.gz
	go clean

e2e: build
	kubectl version
	go test -count=1 -v assets/e2e_test.go

test:
	go test -covermode atomic -coverprofile=profile.cov ./...

lint: ${GOPATH}/bin/golangci-lint
	${GOPATH}/bin/golangci-lint run --timeout 5m \
	  -E gofmt,golint,unconvert,dupl,goimports,maligned,stylecheck # extra linters

coverall: ${GOPATH}/bin/goveralls
	${GOPATH}/bin/goveralls -coverprofile=profile.cov -service=github

.PHONY: tools man build install clean e2e test lint coverall
