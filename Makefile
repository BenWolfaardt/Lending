
# Go related variables.
GOCMD=go
GOBUILD=$(GOCMD) build

GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

APPNAME=lending

VERSION := $(shell echo $(shell git describe --always --tag) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

ldflags = -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		-X github.com/cosmos/cosmos-sdk/version.ServerName=lendingD \
		-X github.com/cosmos/cosmos-sdk/version.ClientName=lendingCLI \
		-X github.com/cosmos/cosmos-sdk/version.Name=lendingnetwork \
		-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \

ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -ldflags '$(ldflags)'

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: help
help: ## Print help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "%-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build bin
	@$(GOBUILD) -o $(GOBIN)/$(APPNAME) $(BUILD_FLAGS) ./app

.PHONY: simplify
simplify:
	@gofmt -s -l -w $(SRC)

.PHONY: run-develop
run-develop: ## Run develop app
	@$(GOBIN)/$(APPNAME) --config develop.json

.PHONY: clean
clean: ## Remove all binaries
	@rm -f $(GOBIN)/*

.PHONY: tidy
tidy: ## Tidy packages
	$(GOCMD) mod tidy

.PHONY: test
test:
	go test ./... -v -coverprofile .coverage.txt
	go tool cover -func .coverage.txt


$(echo $(git describe --tags) | sed 's/^v//')