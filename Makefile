.PHONY: build test lint clean docs

build:
	go build -o marketsurge-agent ./cmd/marketsurge-agent/

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

docs: build
	mkdir -p ./docs
	./marketsurge-agent docs --output ./docs

clean:
	go clean
	rm -f marketsurge-agent coverage.out
	rm -rf dist/
