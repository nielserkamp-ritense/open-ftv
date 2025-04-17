PROJECT_NAME := "fsc-auth"
DOCKER_TAG := "1.2.3"
PKG := "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/fsc-auth"
PKG_LIST := $(shell go list ./...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)
ROOT_DIR := "../.."

.PHONY: all dep-up dep fmt vet lint test race msan cover coverhtml escape app docker help

all: vet lint test race app docker

dep-up: ## upgrade all dependencies
	@go get -u ./...
	@go mod tidy

dep: ## upgrade local dependencies
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/config
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers
	@go get -u gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server
	@go mod tidy

fmt: ${GO_FILES}  ## format source code
	@go fmt ${PKG_LIST}

vet: fmt  ## vet all source code
	@go vet ${PKG_LIST}

lint: ## lint all source code (requires 'revive' to be installed!)
	@revive -set_exit_status ${PKG_LIST}

test: fmt ## run all unit tests
	@go test -cover ${PKG_LIST}

bench: fmt ## run all benchmarks
	@go test -bench . -benchmem ${PKG_LIST}

race: fmt ## run race detector
	@go test -race ${PKG_LIST}

msan: fmt ## run memory sanitizer
	@go test -msan ${PKG_LIST}

cover: fmt ## run coverage report
	@${ROOT_DIR}/scripts/coverage.sh

coverhtml: fmt ## run coverage report and display in browser (requires 'xdg-open' to be installed!)
	@${ROOT_DIR}/scripts/coverage.sh html
	@xdg-open cover.html > /dev/null 2>&1

escape: fmt ## run escape analysis
	@go build -gcflags "-m" ${PKG_LIST} 2>&1

app: fmt ## build binaries
	@mkdir -f ${ROOT_DIR}/bin
	@go build -ldflags "-s -w" -o ${ROOT_DIR}/bin/${PROJECT_NAME} cmd/*.go

docker:
	@cd ${ROOT_DIR} && docker build -t ${PROJECT_NAME}:${DOCKER_TAG} -f ./apps/fsc/Dockerfile .

help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
