VERSION?=$(shell git describe --tags)
GITCOMMIT=$(shell git rev-parse --verify --short HEAD)
BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

all: test build
.PHONY: all

hamdeck:
	go build -v -o hamdeck -ldflags "-X github.com/ftl/hamdeck/cmd.version=$(VERSION) -X github.com/ftl/hamdeck/cmd.gitCommit=$(GITCOMMIT) -X github.com/ftl/hamdeck/cmd.buildTime=$(BUILDTIME)" .

build: hamdeck
.PHONY: build

test:
	go test ./... -v
.PHONY: test

clean:
	-@rm ./hamdeck
.PHONY: clean

install: hamdeck
	cp ./hamdeck /usr/bin/hamdeck
	mkdir -p /usr/share/hamdeck
	cp ./example_conf.json /usr/share/hamdeck/example_conf.json
	cp ./.debpkg/lib/systemd/system/hamdeck.service /lib/systemd/system/hamdeck.service
	cp ./.debpkg/lib/udev/rules.d/99-hamdeck.rules /lib/udev/rules.d/99-hamdeck.rules
.PHONY: install

uninstall:
	-rm /usr/bin/hamdeck
	-rm /usr/share/hamdeck/example_conf.json
	-rmdir /usr/share/hamdeck
	-rm /lib/systemd/system/hamdeck.service
	-rm /lib/udev/rules.d/99-hamdeck.rules
.PHONY: uninstall
