PACKAGES=$(shell go list ./... | grep -v '/simulation')

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=NewApp \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=lendingD \
	-X github.com/cosmos/cosmos-sdk/version.ClientName=lendingCLI

BUILD_FLAGS := -ldflags '$(ldflags)'

.PHONY: all
all: install

.PHONY: install
install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/lendingD
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/lendingCLI

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify

# Uncomment when you have some tests
# test:
# @go test -mod=readonly $(PACKAGES)
.PHONY: lint
# look into .golangci.yml for enabling / disabling linters
lint:
	@echo "--> Running linter"
	@golangci-lint run
	@go mod verify
