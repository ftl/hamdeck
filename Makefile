OUT := hamdeck
VERSION?=$(shell echo `git rev-parse --abbrev-ref HEAD`-`git rev-parse --verify --short HEAD`)

all: test build
.PHONY: all

build:
	go build -v -o $(OUT) -ldflags "-X github.com/ftl/hamdeck/cmd.version=$(VERSION)" .
.PHONY: build

test:
	@go test ./... -v
.PHONY: test

clean:
	-@rm $(OUT)
.PHONY: clean
