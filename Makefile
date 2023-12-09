VERSION?=$(shell git describe --tags)
GITCOMMIT=$(shell git rev-parse --verify --short HEAD)
BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: all
all: test build

.PHONY: build
build: hamdeck

.PHONY: test
test:
	go test ./... -v

.PHONY: clean
clean:
	-@rm ./hamdeck

.PHONY: install
install: hamdeck
	cp ./hamdeck /usr/bin/hamdeck
	mkdir -p /usr/share/hamdeck
	cp ./example_conf.json /usr/share/hamdeck/example_conf.json
	cp ./.debpkg/lib/systemd/system/hamdeck.service /lib/systemd/system/hamdeck.service
	cp ./.debpkg/lib/udev/rules.d/99-hamdeck.rules /lib/udev/rules.d/99-hamdeck.rules

.PHONY: uninstall
uninstall:
	-rm /usr/bin/hamdeck
	-rm /usr/share/hamdeck/example_conf.json
	-rmdir /usr/share/hamdeck
	-rm /lib/systemd/system/hamdeck.service
	-rm /lib/udev/rules.d/99-hamdeck.rules

hamdeck:
	go build -v -o hamdeck -ldflags "-X github.com/ftl/hamdeck/cmd.version=$(VERSION) -X github.com/ftl/hamdeck/cmd.gitCommit=$(GITCOMMIT) -X github.com/ftl/hamdeck/cmd.buildTime=$(BUILDTIME)" .
