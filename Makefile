GO111MODULE=off
GOARCH=amd64
GOOS=darwin
# GOOS=linux
GO_PACKAGE=github.com/alibaba/open-simulator
CGO_ENABLED=0

COMMITID=$(shell git rev-parse --short HEAD)
VERSION=v0.1.1-dev
LD_FLAGS=-ldflags "-X '${GO_PACKAGE}/cmd/version.VERSION=$(VERSION)' -X '${GO_PACKAGE}/cmd/version.COMMITID=$(COMMITID)'"

OUTPUT_DIR=./bin
BINARY_NAME=simon

all: build run

.PHONY: build 
build:
	GO111MODULE=$(GO111MODULE) GOARCH=$(GOARCH) GOOS=$(GOOS) CGO_ENABLED=0 go build -trimpath $(LD_FLAGS) -v -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd
	# chmod +x $(OUTPUT_DIR)/$(BINARY_NAME)
	# bin/simon apply -i -f ./example/simon-config.yaml

.PHONY: run
run:
	# bin/simon apply --extended-resources "gpu" -f example/simon-gpushare-config.yaml
	bin/simon apply --extended-resources "gpu" -f example/simon-paib-snapshot-add-config.yaml

.PHONY: test 
test:
	go test -v ./...

.PHONY: clean 
clean:
	rm -rf ./bin || true
