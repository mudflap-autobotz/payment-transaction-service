.PHONY: help start build wire wire-check docs mocks test coverage coverage-html

GO := go
WIRE := $(GO) tool wire
SWAG := $(GO) tool swag
MOCKERY := $(GO) tool mockery
DI_PKG := ./cmd/api
SWAG_MAIN := cmd/api/main.go
SWAG_OUT := docs
COVERAGE_OUT := coverage.txt

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

start: ## Run the API locally
	$(GO) run ./cmd/api

build: wire docs ## Regenerate wire + OpenAPI docs, then build the binary into bin/api
	$(GO) build -o bin/api ./cmd/api

wire: ## Regenerate the wire dependency-injection code
	$(WIRE) $(DI_PKG)

wire-check: ## Verify the wire dependency graph without writing files
	$(WIRE) check $(DI_PKG)

docs: ## Regenerate the OpenAPI spec from handler annotations
	$(SWAG) init -g $(SWAG_MAIN) -o $(SWAG_OUT) --parseDependency --parseInternal --parseDepth 2

mocks: ## Regenerate testify mocks from domain interfaces (.mockery.yml)
	$(MOCKERY)

test: ## Run all tests
	$(GO) test -v ./...

coverage: ## Run tests verbosely and show the coverage summary in the terminal
	$(GO) test -v -coverprofile=$(COVERAGE_OUT) ./...
	@grep -v -e "/mocks/" -e "/docs/" -e "wire_gen.go" $(COVERAGE_OUT) > $(COVERAGE_OUT).tmp && mv $(COVERAGE_OUT).tmp $(COVERAGE_OUT)
	$(GO) tool cover -func=$(COVERAGE_OUT)

coverage-html: ## Run tests verbosely and open the HTML coverage report
	$(GO) test -v -coverprofile=$(COVERAGE_OUT) ./...
	$(GO) tool cover -html=$(COVERAGE_OUT)
