.PHONY: build test smoke lint clean docs

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null)
LDFLAGS := -X github.com/major/marketsurge-agent/cmd.version=$(VERSION)
ifneq ($(COMMIT),)
LDFLAGS += -X github.com/major/marketsurge-agent/cmd.commit=$(COMMIT)
endif

build:
	go build -ldflags "$(LDFLAGS)" -o marketsurge-agent ./cmd/marketsurge-agent/

test:
	go test -v -race -coverprofile=coverage.out ./...

smoke:
	go test -v -tags=smoke -run TestSmoke ./cmd

docs:
	go run ./cmd/generate-docs/

lint:
	golangci-lint run ./...

clean:
	go clean
	rm -f marketsurge-agent coverage.out
	rm -rf dist/
