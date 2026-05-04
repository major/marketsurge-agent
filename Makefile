.PHONY: build test smoke lint clean docs

build:
	go build -o marketsurge-agent ./cmd/marketsurge-agent/

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
