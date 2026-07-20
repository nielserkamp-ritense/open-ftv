PKG_LIST := ./...

.PHONY: all tidy dep-up fmt vet lint test test-external bench race cover coverhtml escape oas e2e vlierdam rdw rvig help

all: oas vet lint test vlierdam rdw rvig

tidy: ## go mod tidy
	@go mod tidy

dep-up: ## upgrade all dependencies
	@go get -u $(PKG_LIST)
	@go mod tidy

fmt: ## format all source code
	@go fmt $(PKG_LIST)

vet: fmt ## vet all source code
	@go vet $(PKG_LIST)

lint: fmt ## lint all source code (requires revive tool)
	@revive -set_exit_status $(PKG_LIST)

test: ## run all unit tests
	@go test -cover $(PKG_LIST)

test-external: ## run tests that need external services (OpenSearch, GitHub credentials)
	@go test -tags external ./utilities-no-ci/...

bench: ## run all benchmarks
	@go test -bench . -benchmem $(PKG_LIST)

race: ## run race detector
	@go test -race $(PKG_LIST)

cover: ## run coverage report
	@go test -covermode=count -coverprofile=cover.out $(PKG_LIST)
	@go tool cover -func=cover.out

coverhtml: cover ## run coverage report and display in browser (requires xdg-open tool)
	@go tool cover -html=cover.out -o cover.html
	@xdg-open cover.html

escape: ## run escape analysis
	@go build -gcflags "-m" $(PKG_LIST) 2>&1

oas:
	+$(MAKE) -C ./oas

e2e: vlierdam rdw rvig

vlierdam:
	@./e2e/gemeente-vlierdam/test.sh

rdw:
	@./e2e/rdw/test.sh

rvig:
	@./e2e/rvig/test.sh

help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
