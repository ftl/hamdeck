OUT := hamdeck
VERSION ?= "develop"

all: test build

build:
	go build -v -o $(OUT) -ldflags "-X main.version=$(VERSION)" .

test:
	@go test ./... -v

clean:
	-@rm $(OUT)

.PHONY: test build
