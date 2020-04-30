.DEFAULT_GOAL := help

CACHE_DIR := $(shell pwd)/bin/cache
DOCKER = docker
PROJECT = github.com/darkowlzz/ignite-cntr
GOARCH ?= amd64
GO_VERSION = 1.14.2

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

GO_MAKE_TARGET := go-in-docker
go-make:
	$(MAKE) $(GO_MAKE_TARGET) COMMAND="$(MAKE) $(TARGETS)"

go-in-docker:
	mkdir -p $(CACHE_DIR)/go $(CACHE_DIR)/cache
	$(DOCKER) run -ti --rm \
		-v $(CACHE_DIR)/go:/go \
		-v $(CACHE_DIR)/cache:/.cache/go-build \
		-v $(shell pwd):/go/src/${PROJECT} \
		-w /go/src/$(PROJECT) \
		-u $(shell id -u):$(shell id -g) \
		-e GOARCH=$(GOARCH) \
		golang:$(GO_VERSION) \
		$(COMMAND)

tidy:
	go mod tidy -v
	go mod vendor -v

tidy-in-docker:
	$(MAKE) go-make TARGETS="tidy"

cntr:
	$(MAKE) go-make TARGETS="cntr-bin"

cntr-bin:
	go build -mod=vendor -v -o bin/ignite-cntr
