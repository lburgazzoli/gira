.PHONY: build test deps clean

# Build the binary
build:
	go build -o bin/gira .

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