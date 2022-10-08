.PHONY: test
test: ## Run `go vet` and `go test` then outputs test coverage to converage.out
	@go test ./... -v -race -coverprofile=coverage.out

.PHONY: test-report
test-report: test ## Display generated coverage.out file in html on default browser
	@go tool cover -html=coverage.out

.PHONY: build
build: ## Build the project
	@go build

.PHONY: format
format: ## Format project using basic `go fmt`
	@go fmt ./...

.PHONY: help
help: ## Print this help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {printf "\033[36m%-20s\033[0m %s\n", $$1, $$NF}' $(MAKEFILE_LIST)


.DEFAULT_GOAL := build
