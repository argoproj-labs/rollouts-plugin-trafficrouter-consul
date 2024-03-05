VERSION = $(shell ./build-scripts/version.sh pkg/version/version.go)
GOLANG_VERSION?=$(shell head -n 1 .go-version)
GOARCH?=$(shell go env GOARCH)
GIT_COMMIT?=$(shell git rev-parse --short HEAD)
GIT_DIRTY?=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_DESCRIBE?=$(shell git describe --tags --always)
DEV_IMAGE?=rollouts-plugin-trafficrouter-consul-dev

##@ CI

.PHONY: go_lint
go_lint: ## Run linter.
	golangci-lint run ./...

.PHONY: build
build: ## Build the rollouts-plugin-trafficrouter-consul binary
	@$(SHELL) $(CURDIR)/build-support/scripts/build-local.sh --os linux --arch $(GOARCH)

.PHONY: docker
docker-dev: build ## Build rollouts-plugin-trafficrouter-consul dev Docker image.
	docker build -t '$(DEV_IMAGE)' \
       --target=dev \
       --build-arg 'GOLANG_VERSION=$(GOLANG_VERSION)' \
       --build-arg 'TARGETARCH=$(GOARCH)' \
       --build-arg 'GIT_COMMIT=$(GIT_COMMIT)' \
       --build-arg 'GIT_DIRTY=$(GIT_DIRTY)' \
       --build-arg 'GIT_DESCRIBE=$(GIT_DESCRIBE)' \
       -f $(CURDIR)/Dockerfile ./

.PHONY: check-remote-dev-image-env
check-remote-dev-image-env:
ifndef REMOTE_DEV_IMAGE
	$(error REMOTE_DEV_IMAGE is undefined: set this image to <your_docker_repo>/<your_docker_image>:<image_tag>, e.g. REMOTE_DEV_IMAGE=wilko1989/rollouts-plugin-trafficrouter-consul:latest)
endif

.PHONY: docker-multi-arch
docker-multi-arch: check-remote-dev-image-env ## Build rollouts-plugin-trafficrouter-consul dev multi-arch Docker image.
	@$(SHELL) $(CURDIR)/build-support/scripts/build-local.sh --os linux --arch "arm64 amd64"
	@docker buildx create --use && docker buildx build -t '$(REMOTE_DEV_IMAGE)' \
       --platform linux/amd64,linux/arm64 \
       --target=dev \
       --build-arg 'GOLANG_VERSION=$(GOLANG_VERSION)' \
       --build-arg 'GIT_COMMIT=$(GIT_COMMIT)' \
       --build-arg 'GIT_DIRTY=$(GIT_DIRTY)' \
       --build-arg 'GIT_DESCRIBE=$(GIT_DESCRIBE)' \
       --push \
       -f $(CURDIR)/Dockerfile ./

.PHONY: version
version: ## display version
	@echo $(VERSION)

##@ Release
.PHONY: check-env
check-env: ## check env
	@printenv | grep "ROLLOUTS_PLUGIN"

.PHONY: prepare-release-script
prepare-release-script: ## Sets the versions, updates changelog to prepare this repository to release
ifndef ROLLOUTS_PLUGIN_RELEASE_VERSION
	$(error ROLLOUTS_PLUGIN_RELEASE_VERSION is required)
endif
	@source $(CURDIR)/build-support/scripts/functions.sh; prepare_release $(CURDIR) $(ROLLOUTS_PLUGIN_RELEASE_VERSION) $(ROLLOUTS_PLUGIN_PRERELEASE_VERSION); \

.PHONY: prepare-release
prepare-release: prepare-release-script

.PHONY: prepare-dev
prepare-dev: ## prepare main dev
ifndef ROLLOUTS_PLUGIN_NEXT_RELEASE_VERSION
	$(error ROLLOUTS_PLUGIN_NEXT_RELEASE_VERSION is required)
endif
	source $(CURDIR)/build-support/scripts/functions.sh; prepare_dev $(CURDIR) $(ROLLOUTS_PLUGIN_NEXT_RELEASE_VERSION)

##@ Help

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categorises are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php
.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
