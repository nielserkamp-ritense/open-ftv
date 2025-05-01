PROJECT_NAME := "cerbos-api"
PKG := "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ./...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.PHONY: all tidy dep-up dep fmt vet lint test race msan cover coverhtml escape examples help

all: vet lint test race examples

tidy: ## go mod tidy
	@go mod tidy

dep-up: ## upgrade all dependencies
	@go get -u ./...
	@go mod tidy

dep: ## upgrade local dependencies
	@go mod tidy

fmt: tidy ${GO_FILES}  ## format all source code
	@go fmt ${PKG_LIST}

vet: fmt  ## vet all source code
	@go vet ${PKG_LIST}

lint: fmt ## lint all source code (requires revive tool)
	@revive -set_exit_status ${PKG_LIST}

test: fmt ## run all unit tests
	@./start_api.sh
	@go test -cover ${PKG_LIST}
	@./stop_api.sh

bench: fmt ## run all benchmarks
	@./start_api.sh
	@go test -bench . -benchmem ${PKG_LIST}
	@./stop_api.sh

race: fmt ## run race detector
	@./start_api.sh
	@go test -race ${PKG_LIST}
	@./stop_api.sh

msan: fmt ## run memory sanitizer
	@./start_api.sh
	@go test -msan ${PKG_LIST}
	@./stop_api.sh

cover: fmt ## run coverage report
	@./start_api.sh
	@../../scripts/coverage.sh
	@./stop_api.sh

coverhtml: fmt ## run coverage report and display in browser (requires xdg-open tool)
	@./start_api.sh
	@../../scripts/coverage.sh html
	@./stop_api.sh
	@xdg-open cover.html

escape: fmt ## run escape analysis
	@go build -gcflags "-m" ${PKG_LIST} 2>&1

examples: fmt ## build example binaries

help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
