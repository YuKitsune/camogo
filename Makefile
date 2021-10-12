
.DEFAULT_GOAL := help

.PHONY: help
help: ## Shows this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: format
format: ## Makes the source code nice and clean
	gofmt -s -w .

.PHONY: build
build: ## Builds the package
	go build ./...

.PHONY: test
test: ## Runs all tests
	go test ./...

.PHONY: benchmark
benchmark: ## Runs all benchmarks
	go test -bench=.
