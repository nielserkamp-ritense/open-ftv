PROJECT_NAME := "data"
PKG := "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasource/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ./...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.PHONY: all tidy dep-up dep fmt vet lint test race cover coverhtml escape app docker help

all: test

tidy: ## go mod tidy
	@go mod tidy

lint: ## lint all source code (requires 'revive' to be installed!)
	@revive -set_exit_status ${PKG_LIST}

dep-up: ## upgrade all dependencies
	@go get -u ./...
	@go mod tidy

dep: ## upgrade local dependencies
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities
	@go mod tidy

fmt: tidy dep ${GO_FILES}  ## format source code
	@go fmt ${PKG_LIST}

test: fmt ## run all unit tests
	@go test -cover ${PKG_LIST}

bench: fmt ## run all benchmarks
	@go test -bench . -benchmem ${PKG_LIST}

race: fmt ## run race detector
	@go test -race ${PKG_LIST}

cover: fmt ## run coverage report
	@../../../scripts/coverage.sh

coverhtml: fmt ## run coverage report and display in browser (requires 'xdg-open' to be installed!)
	@../../../scripts/coverage.sh html
	@xdg-open cover.html > /dev/null 2>&1

escape: fmt ## run escape analysis
	@go build -gcflags "-m" ${PKG_LIST} 2>&1

help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
