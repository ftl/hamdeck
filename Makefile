OUT := hamdeck
VERSION?=$(shell git describe --tags)
GITCOMMIT=$(shell git rev-parse --verify --short HEAD)
BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

all: test build
.PHONY: all

build:
	go build -v -o $(OUT) -ldflags "-X github.com/ftl/hamdeck/cmd.version=$(VERSION) -X github.com/ftl/hamdeck/cmd.gitCommit=$(GITCOMMIT) -X github.com/ftl/hamdeck/cmd.buildTime=$(BUILDTIME)" .
.PHONY: build

test:
	@go test ./... -v
.PHONY: test

clean:
	-@rm $(OUT)
.PHONY: clean
