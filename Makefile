.DEFAULT_GOAL := help

CACHE_DIR = $(shell pwd)/bin/.cache
PROJECT = github.com/darkowlzz/ignite-cntr
GOARCH ?= amd64
GO_VERSION = 1.14.2

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "To run any of the above in docker, suffix the command with '-docker':"
	@echo ""
	@echo "  make build-docker"
	@echo ""

.PHONY: build
build: ## Build ignite-cntr binary.
	GOARCH=$(GOARCH) go build -mod=vendor -v -o bin/ignite-cntr

.PHONY: install
install: ## Install ignite-cntr.
	go install -mod=vendor -v

.PHONY: clean
clean: ## Clear all the generated files and assets.
	rm -rf bin/

.PHONY: test
test: ## Run all the tests.
	go test -mod=vendor -v ./... -count=1

# This target matches any target ending in '-docker' eg. 'build-docker'. This
# allows running makefile targets inside a container by appending '-docker' to
# it.
%-docker:
	mkdir -p $(CACHE_DIR)/go $(CACHE_DIR)/cache
	docker run -it --rm \
		-v $(CACHE_DIR)/go:/go \
		-v $(CACHE_DIR)/cache:/.cache/go-build \
		-v $(shell pwd):/go/src/${PROJECT} \
		-w /go/src/${PROJECT} \
		-u $(shell id -u):$(shell id -g) \
		-e GOARCH=$(GOARCH) \
		--entrypoint "make" \
		golang:$(GO_VERSION) \
		"$(patsubst %-docker,%,$@)"

tidy: ## Prune, add and vendor go dependencies.
	go mod tidy -v
	go mod vendor -v
