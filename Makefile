BINARY=jobgo
BUILD_DIR=bin

.PHONY: build install test lint clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/jobgo

install:
	go install ./cmd/jobgo

test:
	go test ./... -v

lint:
	golangci-lint run

clean: 
	rm -rf $(BUILD_DIR)