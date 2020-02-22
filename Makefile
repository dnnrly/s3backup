GO111MODULE=on

GOPATH=$(go env GOPATH)

CURL_BIN ?= curl
GO_BIN ?= go
GORELEASER_BIN ?= goreleaser

PUBLISH_PARAM?=
GO_MOD_PARAM?=-mod vendor
TMP_DIR?=./tmp

BASE_DIR=$(shell pwd)

NAME=s3backup

export GOPROXY=https://proxy.golang.org
export PATH := ./bin:$(GOPATH)/bin:$(PATH)

.PHONY: install
install: deps

.PHONY: build
build:
	$(GO_BIN) build -v ./cmd/s3backup

.PHONY: clean
clean:
	rm -f $(NAME)
	rm -rf dist

.PHONY: clean-deps
clean-deps:
	rm -rf ./bin
	rm -rf ./tmp
	rm -rf ./libexec
	rm -rf ./share

./bin/bats:
	git clone https://github.com/bats-core/bats-core.git ./tmp/bats
	./tmp/bats/install.sh .

./bin/golangci-lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.22.2

test-deps: ./bin/bats ./bin/golangci-lint
	$(GO_BIN) get github.com/mfridman/tparse
	$(GO_BIN) get github.com/mikefarah/yq/v3
	$(GO_BIN) mod tidy
	yq --version

./bin:
	mkdir -p ./bin

./tmp:
	mkdir -p ./tmp

.PHONY: build-deps
build-deps: ./bin

.PHONY: deps
deps: build-deps test-deps

.PHONY: test
test:
	$(GO_BIN) test -json -cover ./... | tparse -all

.PHONY: acceptance-test
acceptance-test:
	docker-compose run --rm -v ${BASE_DIR}:${BASE_DIR} -w ${BASE_DIR} test

.PHONY: ci-test
ci-test:
	$(GO_BIN) test -race -cover -covermode=atomic -json ./... | tparse -all

.PHONY: lint
lint:
	golangci-lint run

.PHONY: release
release: clean
	$(GORELEASER_BIN) $(PUBLISH_PARAM)

.PHONY: update
update:
	$(GO_BIN) get -u
	$(GO_BIN) mod tidy
	make test
	make install
	$(GO_BIN) mod tidy
