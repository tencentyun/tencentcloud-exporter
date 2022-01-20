.PHONY: help build lint

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOLANGCI_LINT_VERSION ?= "v1.35.2"

build:
	env GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o "build/$(version)/qcloud_exporter" ./cmd/qcloud-exporter/

lint:
	if [[ ! -e ./bin/golangci-lint ]]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LINT_VERSION); \
	fi; \
	./bin/golangci-lint run ./...

deploy:
	env GOOS=linux GOARCH=amd64 go build -o "build/qcloud_exporter" ./cmd/qcloud-exporter/

deps:  ## Update vendor.
	go mod verify
	go mod tidy -v
