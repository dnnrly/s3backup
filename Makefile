GO111MODULE=on

CURL_BIN ?= curl
GO_BIN ?= go
GORELEASER_BIN ?= goreleaser

PUBLISH_PARAM?=
GO_MOD_PARAM?=-mod vendor
TMP_DIR?=./tmp

BASE_DIR=$(shell pwd)

NAME=abbreviate

export PATH := ./bin:$(PATH)

install: deps

build:
	$(GO_BIN) build -v $(GO_MOD_PARAM)

clean:
	rm -f $(NAME)
	rm -rf dist

clean-deps:
	rm -rf ./bin
	rm -rf ./tmp
	rm -rf ./libexec
	rm -rf ./share

./bin/bats:
	git clone https://github.com/sstephenson/bats.git ./tmp/bats
	./tmp/bats/install.sh .

./bin/golangci-lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.17.1

test-deps: ./bin/bats ./bin/golangci-lint
	$(GO_BIN) get -v ./...
	$(GO_BIN) mod tidy

./bin:
	mkdir ./bin

build-deps: ./bin

deps: build-deps test-deps

test:
	$(GO_BIN) test $(GO_MOD_PARAM) ./...

acceptance-test:
	bats --tap acceptance.bats

ci-test:
	$(GO_BIN) test $(GO_MOD_PARAM) -race -coverprofile=coverage.txt -covermode=atomic ./...

lint:
	golangci-lint run

release: clean
	$(GORELEASER_BIN) $(PUBLISH_PARAM)

update:
	$(GO_BIN) get -u
	$(GO_BIN) mod tidy
	make test
	make install
	$(GO_BIN) mod tidy
