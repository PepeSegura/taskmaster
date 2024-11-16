MAKEFLAGS += --silent
# Variables
NAME := taskmaster
PKG := ./...

SRCS = 							\
		srcs/main.go			\

# Default target
all: build

# Build the project
build:
	go build -o $(NAME) $(SRCS)

# Run the project
run: build
	./$(NAME)

# Test the project
test:
	go test $(PKG)

# Run tests with verbose output
test-verbose:
	go test -v $(PKG)

# Run tests with coverage
coverage:
	go test -cover $(PKG)

# Format the code
fmt:
	go fmt $(PKG)

# Clean up binaries and cache
clean:
	go clean
	rm -f $(NAME)

# Install dependencies
deps:
	go mod tidy

# Lint the code (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Vendor dependencies (optional: if you want to keep dependencies locally)
vendor:
	go mod vendor

mod:
	go mod init $(NAME)

init: mod deps

# Targets not associated with file names
.PHONY: all build run test test-verbose coverage fmt clean deps lint vendor