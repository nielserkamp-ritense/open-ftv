.PHONY: all oas

all: oas

oas: ## create models for external OAS specs (via go tool oapi-codegen from go.mod)
	@./gen_oas.sh

lint: ## lint all OpenAPI specs.
	@./lint_oas.sh

test: ## run Specmatic tests for all OpenAPI specs.
	@./test_oas.sh

help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
