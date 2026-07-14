PROJECT_NAME := "manager"
DOCKER_TAG := "1.0.0"
PKG := "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ./...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)
ROOT_DIR := "../.."

.PHONY: all tidy dep-up dep fmt vet lint test race msan cover coverhtml escape app docker help

all: vet lint test race app docker

tidy: ## go mod tidy
	@go mod tidy

dep-up: ## upgrade all dependencies
	@go get -u ./...
	@go mod tidy

dep: ## upgrade local dependencies
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server
	@go mod tidy

fmt: tidy ${GO_FILES}  ## format all source code
	@go fmt ${PKG_LIST}

vet: fmt  ## vet all source code
	@go vet ${PKG_LIST}

lint: fmt ## lint all source code (requires revive tool)
	@revive -set_exit_status ${PKG_LIST}

test: ## run all unit tests
	@go test -cover ${PKG_LIST}

bench: ## run all benchmarks
	@go test -bench . -benchmem ${PKG_LIST}

race: ## run race detector
	@go test -race ${PKG_LIST}

msan: ## run memory sanitizer
	@go test -msan ${PKG_LIST}

cover: ## run coverage report
	@${ROOT_DIR}/scripts/coverage.sh

coverhtml: ## run coverage report and display in browser (requires xdg-open tool)
	@${ROOT_DIR}/scripts/coverage.sh html
	@xdg-open cover.html

escape: fmt ## run escape analysis
	@go build -gcflags "-m" ${PKG_LIST} 2>&1

app: fmt ## build binaries
	@mkdir -f ${ROOT_DIR}/bin
	@go build -ldflags "-s -w" -o ${ROOT_DIR}/bin/${PROJECT_NAME} cmd/*.go

docker:
	@cd ${ROOT_DIR} && docker build -t ${PROJECT_NAME}:${DOCKER_TAG} -f ./apps/manager/Dockerfile .

help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
