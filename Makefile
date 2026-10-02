APP_NAME := cmd-emulator
MAIN_PACKAGE := ./src/cmd/emulator
BUILD_DIR := build
BINARY := $(BUILD_DIR)/$(APP_NAME)

.PHONY: run build test fmt check clean

run:
	go run $(MAIN_PACKAGE) $(ARGS)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BINARY) $(MAIN_PACKAGE)

test:
	go test ./...

fmt:
	gofmt -w src

check:
	@unformatted=$$(gofmt -l src) || exit 1; \
	if [ -n "$$unformatted" ]; then \
		printf 'Запустите make fmt для файлов:\n%s\n' "$$unformatted"; exit 1; \
	fi
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
