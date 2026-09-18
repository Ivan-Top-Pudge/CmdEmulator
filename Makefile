APP_NAME := cmd-emulator
MAIN_PACKAGE := ./src/cmd/emulator
BUILD_DIR := build
BINARY := $(BUILD_DIR)/$(APP_NAME)

.PHONY: run build test clean

run:
	go run $(MAIN_PACKAGE) $(ARGS)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BINARY) $(MAIN_PACKAGE)

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)
