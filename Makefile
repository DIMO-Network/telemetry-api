.PHONY: clean run build install dep test lint format docker gqlgen

PATHINSTBIN = $(abspath ./bin)
export PATH := $(PATHINSTBIN):$(PATH)
SHELL := env PATH=$(PATH) /bin/sh

BIN_NAME					?= telemetry-api
DEFAULT_INSTALL_DIR			:= $(go env GOPATH)/bin
DEFAULT_ARCH				:= $(shell go env GOARCH)
DEFAULT_GOOS				:= $(shell go env GOOS)
ARCH						?= $(DEFAULT_ARCH)
GOOS						?= $(DEFAULT_GOOS)
INSTALL_DIR					?= $(DEFAULT_INSTALL_DIR)
.DEFAULT_GOAL 				:= run


VERSION   := $(shell git describe --tags || echo "v0.0.0")
VER_CUT   := $(shell echo $(VERSION) | cut -c2-)

# Dependency versions
GOLANGCI_VERSION   = latest
GQLGEN_VERSION     = $(shell go list -m -f '{{.Version}}' github.com/99designs/gqlgen)
MODEL_GARAGE_VERSION = $(shell go list -m -f '{{.Version}}' github.com/DIMO-Network/model-garage)
MOCKGEN_VERSION    = $(shell go list -m -f '{{.Version}}' go.uber.org/mock)

help:
	@echo "\nSpecify a subcommand:\n"
	@grep -hE '^[0-9a-zA-Z_-]+:.*?## .*$$' ${MAKEFILE_LIST} | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[0;36m%-20s\033[m %s\n", $$1, $$2}'
	@echo ""

build: ## Build the binary
	@CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(ARCH) \
		go build -o $(PATHINSTBIN)/$(BIN_NAME) ./cmd/$(BIN_NAME)

run: build ## Run the binary
	@./$(PATHINSTBIN)/$(BIN_NAME)
all: clean target

clean: ## Remove previous built binaries
	@rm -rf $(PATHINSTBIN)
	
install: build 
	@install -d $(INSTALL_DIR)
	@rm -f $(INSTALL_DIR)/$(BIN_NAME)
	@cp $(PATHINSTBIN)/$(BIN_NAME) $(INSTALL_DIR)/$(BIN_NAME)

tidy: ## tidy go modules
	@go mod tidy

test: ## Run tests
	@go test -v ./...

lint: ## Run linter
	@golangci-lint run

format:
	@golangci-lint run --fix

docker: dep
	@docker build -f ./Dockerfile . -t dimozone/$(BIN_NAME):$(VER_CUT)
	@docker tag dimozone/$(BIN_NAME):$(VER_CUT) dimozone/$(BIN_NAME):latest

gqlgen: ## Generate gqlgen code.
	@gqlgen generate

gql-model: ## Generate gqlgen data model.
	@codegen -generators=custom -custom.output-file=schema/signals_gen.graphqls -custom.template-file=./schema/signals.tmpl
	@codegen -generators=custom -custom.output-file=internal/graph/model/signalSetter_gen.go -custom.template-file=./internal/graph/model/signalSetter.tmpl -custom.format=true
	@codegen -generators=custom -custom.output-file=internal/graph/signals_gen.resolvers.go -custom.template-file=./internal/graph/signals_gen.resolvers.tmpl -custom.format=true

gql: gql-model gqlgen ## Generate all gql code.

generate: gql ## Runs all code generators for the repository.
	@go generate ./...

tools-gqlgen: ## install gqlgen tool
	@mkdir -p $(PATHINSTBIN)
	GOBIN=$(PATHINSTBIN) go install github.com/99designs/gqlgen@${GQLGEN_VERSION}

tools-model-garage: ## install model-garage tool
	@mkdir -p $(PATHINSTBIN)
	GOBIN=$(PATHINSTBIN) go install github.com/DIMO-Network/model-garage/cmd/codegen@${MODEL_GARAGE_VERSION}

tools-golangci-lint: ## install golangci-lint tool
	@mkdir -p $(PATHINSTBIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PATHINSTBIN) $(GOLANGCI_VERSION)

tools-mockgen: ## install mockgen tool
	@mkdir -p $(PATHINSTBIN)
	GOBIN=$(PATHINSTBIN) go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)

tools: tools-golangci-lint tools-gqlgen tools-model-garage tools-mockgen ## Install all tools required for development.