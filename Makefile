
all: build

tools:
	which gometalinter || ( go get -u github.com/alecthomas/gometalinter && gometalinter --install )
	which glide || go get -u github.com/Masterminds/glide
	which goveralls || go get github.com/mattn/goveralls

lint:
	gometalinter --concurrency=1 --deadline=300s --vendor --disable-all \
		--enable=golint \
		--enable=vet \
		--enable=vetshadow \
		--enable=varcheck \
		--enable=errcheck \
		--enable=structcheck \
		--enable=deadcode \
		--enable=ineffassign \
		--enable=dupl \
		--enable=gotype \
		--enable=varcheck \
		--enable=interfacer \
		--enable=goconst \
		--enable=megacheck \
		--enable=unparam \
		--enable=misspell \
		--enable=gas \
		--enable=goimports \
		--enable=gocyclo \
		./...

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

test:
	go test -i github.com/bpineau/katafygio/...
	go test -race -cover -coverprofile=profile.cov github.com/bpineau/katafygio/...

.PHONY: tools lint fmt install clean coverall test all
