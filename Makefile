# ──────────────────────────────────────────────────────────────
# githookd Makefile
# ──────────────────────────────────────────────────────────────

BINARY    := ghm
MODULE    := githookd
MAIN      := ./cmd/ghm
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE      := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS   := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

GO        := go
GOTEST    := $(GO) test
GOBUILD   := $(GO) build
GOLINT    := golangci-lint

.PHONY: all build test lint clean install snapshot help

## Build the binary
all: build

## Build the binary with version info
build:
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY) $(MAIN)

## Run all tests
test:
	$(GOTEST) -v -race -count=1 ./...

## Run all tests with coverage
test-cover:
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Run linter
lint:
	$(GOLINT) run ./...

## Remove build artifacts
clean:
	rm -f $(BINARY) coverage.out coverage.html
	rm -rf dist/

## Install to $GOPATH/bin
install:
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(shell $(GO) env GOPATH)/bin/$(BINARY) $(MAIN)

## Build a snapshot release (no publish)
snapshot:
	goreleaser release --snapshot --clean

## Show help
help:
	@echo "githookd Makefile targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Usage: make [target]"
