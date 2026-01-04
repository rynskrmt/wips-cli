BINARY_NAME=wip
BUILD_DIR=bin
CMD_PATH=./cmd/wip

.PHONY: all build test clean install uninstall dev run help

all: build

build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run tests
	go test -v ./...

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

install: build ## Install binary to /usr/local/bin (requires sudo)
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

uninstall: ## Remove binary from system
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

reset: ## Reset dev environment (remove data dir)
	rm -rf "$(HOME)/Library/Application Support/$(BINARY_NAME)"

dev: ## Build and install to $GOPATH/bin
	go install $(CMD_PATH)

run: build ## Build and run (usage: make run ARGS="tail")
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
