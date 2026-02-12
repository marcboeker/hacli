.PHONY: build test clean run fmt vet

BINARY_NAME=hacli
CMD_PATH=./cmd/hacli
BUILD_DIR=./bin

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

run:
	go run $(CMD_PATH)

fmt:
	go fmt ./...

vet:
	go vet ./...

all: fmt vet test build
