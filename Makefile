GO111MODULE=on

CURL_BIN ?= curl
GO_BIN ?= go
GORELEASER_BIN ?= goreleaser

PUBLISH_PARAM?=
GO_MOD_PARAM?=-mod vendor
TMP_DIR?=./tmp

BASE_DIR=$(shell pwd)

NAME=s3backup

export GOPROXY=https://proxy.golang.org
export PATH := ./bin:$(PATH)

.PHONY: install
install: deps

.PHONY: build
build:
	$(GO_BIN) build -v

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
	$(GO_BIN) mod tidy
	curl -L -o ./bin/yq.v2 https://github.com/mikefarah/yq/releases/download/2.4.0/yq_linux_amd64
	chmod +x ./bin/yq.v2
	yq.v2

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
	$(GO_BIN) test -json ./... | tparse -all

.PHONY: acceptance-test
acceptance-test:
	bats --tap test/*.bats

.PHONY: ci-test
ci-test:
	$(GO_BIN) test -race -coverprofile=coverage.txt -covermode=atomic -json ./... | tparse -all

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
