.PHONY: build test lint clean

build:
	go build -o marketsurge-agent ./cmd/marketsurge-agent/

test:
	go test -v ./...

lint:
	golangci-lint run ./...

clean:
	go clean
	rm -f marketsurge-agent
	rm -rf dist/
