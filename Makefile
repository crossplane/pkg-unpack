# Project Setup
PROJECT_NAME := pkg-unpack
PROJECT_REPO := github.com/crossplane/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64
include build/makelib/common.mk

S3_BUCKET ?= crossplane.releases/pkg-unpack
include build/makelib/output.mk

# Setup Go
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/pkg-unpack
GO_SUBDIRS += cmd
GO_LDFLAGS += -X $(GO_PROJECT)/pkg/version.Version=$(VERSION)
GO111MODULE = on
include build/makelib/golang.mk

# Docker images
DOCKER_REGISTRY = crossplane
IMAGES = pkg-unpack
include build/makelib/image.mk

# We want submodules to be set up the first time `make` is run.
# We manage the build/ folder and its Makefiles as a submodule.
# The first time `make` is run, the includes of build/*.mk files will
# all fail, and this target will be run. The next time, the default as defined
# by the includes will be run instead.
fallthrough: submodules
	@echo Initial setup complete. Running make again . . .
	@make
