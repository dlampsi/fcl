.PHONY: all clean build linter tests build

PROJECT_NAME=fcl
BINARY_NAME?=fcl
RELEASE?=dev
BUILD_TIME?=$(shell date '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH=$(shell git rev-parse --short HEAD)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: clean build linter test build ## Run linter, tests and build package

clean:
	@go clean

test: ## Run tests
	go test ./...

linter: ## Apply linter
	golangci-lint run -c ./.golangci.yml --timeout 3m ./...

build: clean ## Build package
	go build \
	-ldflags "-s -w -X ${PROJECT_NAME}/info.Version=${RELEASE} \
	-X ${PROJECT_NAME}/info.BuildNumber=${BUILD_NUMBER} \
	-X ${PROJECT_NAME}/info.BuildTime=${BUILD_TIME} \
	-X ${PROJECT_NAME}/info.CommitHash=${COMMIT_HASH}" -o ${BINARY_NAME}
