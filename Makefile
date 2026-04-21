.PHONY: build test lint clean

build:
	/usr/local/go/bin/go build -o marketsurge-agent ./cmd/marketsurge-agent/

test:
	/usr/local/go/bin/go test -v ./...

lint:
	golangci-lint run ./...

clean:
	/usr/local/go/bin/go clean
	rm -f marketsurge-agent
	rm -rf dist/
