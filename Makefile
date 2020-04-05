.DEFAULT_GOAL:=help

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build ignite-cntr binary.
	go build -mod=vendor -v -o bin/ignite-cntr

.PHONY: install
install: ## Install ignite-cntr.
	go install -mod=vendor -v

.PHONY: clean
clean: ## Clear all the generated files and assets.
	rm -rf bin/

.PHONY: test
test: ## Run all the tests.
	go test -mod=vendor -v ./...
