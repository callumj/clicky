.PHONY: all test lint format lint-install build build-linux build-macos package clean

MAIN_PACKAGE := ./cmd/clicky

all: test lint format build package

test:
	go test ./...

lint: lint-install
	golangci-lint run

lint-install:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

format:
	go fmt ./...

build: build-linux build-macos

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/clicky-linux-amd64 $(MAIN_PACKAGE)

build-macos:
	GOOS=darwin GOARCH=arm64 go build -o bin/clicky-darwin-arm64 $(MAIN_PACKAGE)

package: build
	tar -czvf bin/clicky-linux-amd64.tar.gz -C bin clicky-linux-amd64
	tar -czvf bin/clicky-darwin-arm64.tar.gz -C bin clicky-darwin-arm64

clean: