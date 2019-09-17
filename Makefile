MAKEFLAGS = -j1
SHELL := /bin/bash

all: build 

export GO11MODULE := on 
export CGO_ENABLED := 0

build: 
	@echo "Building..."
	@mkdir -p bin/ 
	@go build  ./...
	@echo "Done..." 


# install golint if not already installed
GOLANGCI_LINT_VERSION=1.16.0
GOLANGCI_LINT_BIN=$(GOLANGCI_LINT_VERSION)-$(shell go env GOOS)-$(shell go env GOARCH)
GOLANGCI_LINT_CMD=$(or $(shell command -v golangci-lint),$(HOME)/bin/golangci-lint)

$(GOLANGCI_LINT_CMD):
	@echo "Downloading golangci-lint..."
	@curl -sLO --fail "https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_BIN).tar.gz"
	@tar -xzf golangci-lint-$(GOLANGCI_LINT_BIN).tar.gz "golangci-lint-$(GOLANGCI_LINT_BIN)/golangci-lint"
	@mkdir -p $(HOME)/bin
	@mv golangci-lint-$(GOLANGCI_LINT_BIN)/golangci-lint $(HOME)/bin
	@rm -rf golangci-lint-$(GOLANGCI_LINT_BIN)*

godeps: ## download go modules to the cache
	@echo "Ensuring Go dependencies..."
	@go mod download

golint: $(GOLANGCI_LINT_CMD) ## lint
	@echo "Running linters..."
	@$(GOLANGCI_LINT_CMD) run --fix

