.PHONY: test build run lint cover cover-check cover-by-pkg ci dev fmt clean

# Coverage threshold
COVERAGE_THRESHOLD := 70

# Test
test:
	go test -v -race ./...

# Coverage measurement
cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Coverage threshold check (for CI)
cover-check:
	@go test -race -coverprofile=coverage.out ./... > /dev/null 2>&1
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $${COVERAGE}%"; \
	if [ $$(echo "$${COVERAGE} < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "ERROR: Coverage $${COVERAGE}% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi; \
	echo "OK: Coverage meets threshold"

# Coverage by package
cover-by-pkg:
	@echo "=== Coverage by package ==="
	@go test -race -coverprofile=coverage.out ./... > /dev/null 2>&1
	@go tool cover -func=coverage.out | grep -E '^(github.com|total:)'

# Build
build:
	go build -o bin/lazyactions ./cmd/lazyactions

# Run
run:
	go run ./cmd/lazyactions

# Lint
lint:
	golangci-lint run

# CI (lint + coverage check + build)
ci: lint cover-check build

# Development (format + lint + test)
dev: fmt lint test

# Format
fmt:
	go fmt ./...
	goimports -w .

# Clean
clean:
	rm -rf bin/ coverage.out coverage.html
