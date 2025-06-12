.PHONY: build test deps clean

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build flags
LDFLAGS = -X 'github.com/lburgazzoli/gira/internal/version.Version=$(VERSION)' \
          -X 'github.com/lburgazzoli/gira/internal/version.Commit=$(COMMIT)' \
          -X 'github.com/lburgazzoli/gira/internal/version.Date=$(DATE)'

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o bin/gira .

# Run tests
test:
	go test -v ./...

# Install/update dependencies
deps:
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	rm -rf bin/