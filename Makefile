OUT := hamdeck
VERSION?=$(shell echo `git rev-parse --abbrev-ref HEAD`-`git rev-parse --verify HEAD | head -c10`)

all: test build

build:
	go build -v -o $(OUT) -ldflags "-X main.version=$(VERSION)" .

test:
	@go test ./... -v

clean:
	-@rm $(OUT)

.PHONY: test build
