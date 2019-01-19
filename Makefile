all: build

GOLANGCI_VERSION="1.12.5"
tools:
	which golangci-lint || ( \
	  curl -sfL "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_VERSION}/golangci-lint-${GOLANGCI_VERSION}-linux-amd64.tar.gz" | \
	  tar xz --strip-components 1 --wildcards '*/golangci-lint' && \
	  chmod 755 golangci-lint && \
	  mv golangci-lint ${GOPATH}/bin/ \
	)
	which glide || go get -u github.com/Masterminds/glide
	which goveralls || go get github.com/mattn/goveralls

lint:
	@# vet, errcheck, deadcode, etc are already enabled by default; here we just add more checks:
	golangci-lint run -E gofmt,golint,unconvert,dupl,goimports,misspell,maligned

man:
	go run assets/manpage.go

fmt:
	go fmt ./...

deps:
	glide install

build:
	env CGO_ENABLED=0 go build -i

install:
	env CGO_ENABLED=0 go install

clean:
	rm -rf dist/
	go clean -i

coverall:
	goveralls -coverprofile=profile.cov -service=travis-ci -package github.com/bpineau/katafygio/pkg/...

e2e:
	kubectl get ns >/dev/null || exit 1
	go test -count=1 -v assets/e2e_test.go

test:
	go test -i github.com/bpineau/katafygio/...
	go test -race -cover -coverprofile=profile.cov github.com/bpineau/katafygio/...

.PHONY: tools lint fmt install clean coverall test all
