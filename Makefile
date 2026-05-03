.PHONY: build test smoke lint clean docs

build:
	go build -o marketsurge-agent ./cmd/marketsurge-agent/

test:
	go test -v -race -coverprofile=coverage.out ./...

smoke:
	go test -v -tags=smoke -run TestSmoke ./cmd

lint:
	golangci-lint run ./...

docs: build
	mkdir -p ./docs
	./marketsurge-agent docs --output ./docs

clean:
	go clean
	rm -f marketsurge-agent coverage.out
	rm -rf dist/
