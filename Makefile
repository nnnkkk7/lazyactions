.PHONY: test test-unit test-integration build run lint cover cover-check cover-by-pkg ci dev fmt clean tools generate

# Coverage threshold
COVERAGE_THRESHOLD := 70

# Tool paths
GOBIN ?= $(shell go env GOPATH)/bin
GOLANGCI_LINT := $(GOBIN)/golangci-lint
GOTESTSUM := $(GOBIN)/gotestsum
MOQ := $(GOBIN)/moq

# Install tools
tools: $(GOLANGCI_LINT) $(GOTESTSUM) $(MOQ)

$(GOLANGCI_LINT):
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

$(GOTESTSUM):
	go install gotest.tools/gotestsum@latest

$(MOQ):
	go install github.com/matryer/moq@latest

# Test (all tests)
test: $(GOTESTSUM)
	$(GOTESTSUM) --format testdox -- -race ./...

# Unit tests only (exclude integration tests)
test-unit: $(GOTESTSUM)
	$(GOTESTSUM) --format testdox -- -race $(shell go list ./... | grep -v /test/)

# Integration tests only
test-integration: $(GOTESTSUM)
	$(GOTESTSUM) --format testdox -- -race ./test/integration/...

# Coverage measurement
cover: $(GOTESTSUM)
	$(GOTESTSUM) --format testdox -- -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Coverage threshold check (for CI)
cover-check: $(GOTESTSUM)
	$(GOTESTSUM) --format pkgname -- -race -coverprofile=coverage.out ./...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $${COVERAGE}%"; \
	if [ $$(echo "$${COVERAGE} < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "ERROR: Coverage $${COVERAGE}% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi; \
	echo "OK: Coverage meets threshold"

# Coverage by package
cover-by-pkg: $(GOTESTSUM)
	@echo "=== Coverage by package ==="
	$(GOTESTSUM) --format pkgname -- -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep -E '^(github.com|total:)'

# Build
build:
	go build -o bin/lazyactions ./cmd/lazyactions

# Run
run:
	go run ./cmd/lazyactions

# Lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

# CI (lint + test + build)
ci: lint test build

# Development (format + lint + test)
dev: fmt lint test

# Format
fmt:
	go fmt ./...
	goimports -w .

# Generate (mocks, etc.)
generate: $(MOQ)
	go generate ./...
