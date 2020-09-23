# Project Setup
PROJECT_NAME := pkg-unpack
PROJECT_REPO := github.com/crossplane/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64
-include build/makelib/common.mk

S3_BUCKET ?= crossplane.releases/pkg-unpack
-include build/makelib/output.mk

# Setup Go
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/pkg-unpack
GO_SUBDIRS += cmd
GO_LDFLAGS += -X $(GO_PROJECT)/pkg/version.Version=$(VERSION)
GO111MODULE = on
-include build/makelib/golang.mk

# Docker images
DOCKER_REGISTRY = crossplane
IMAGES = pkg-unpack
-include build/makelib/image.mk

# Update the submodules, such as the common build scripts.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# We want submodules to be set up the first time `make` is run.
# We manage the build/ folder and its Makefiles as a submodule.
# The first time `make` is run, the includes of build/*.mk files will
# all fail, and this target will be run. The next time, the default as defined
# by the includes will be run instead.
fallthrough: submodules
	@echo Initial setup complete. Running make again . . .
	@make


.PHONY: submodules fallthrough

# ====================================================================================
# Special Targets

define CROSSPLANE_MAKE_HELP
Crossplane Targets:
    submodules         Update the submodules, such as the common build scripts.

endef
# The reason CROSSPLANE_MAKE_HELP is used instead of CROSSPLANE_HELP is because the crossplane
# binary will try to use CROSSPLANE_HELP if it is set, and this is for something different.
export CROSSPLANE_MAKE_HELP

crossplane.help:
	@echo "$$CROSSPLANE_MAKE_HELP"

help-special: crossplane.help

.PHONY: crossplane.help help-special
