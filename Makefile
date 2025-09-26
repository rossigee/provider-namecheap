# Project Setup
PROJECT_NAME := provider-namecheap
PROJECT_REPO := github.com/rossigee/$(PROJECT_NAME)

# Platform support
PLATFORMS ?= linux_amd64 linux_arm64

# Include build system
-include build/makelib/common.mk


# Setup Output
-include build/makelib/output.mk

# Setup Go
# Override golangci-lint version for modern Go support
GOLANGCILINT_VERSION ?= 2.5.0
NPROCS ?= 1
GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/provider
GO_LDFLAGS += -X $(GO_PROJECT)/internal/version.Version=$(VERSION)
GO_SUBDIRS += apis
GO111MODULE = on
-include build/makelib/golang.mk

# Setup Images
REGISTRY_ORGS ?= ghcr.io/rossigee
IMAGES = $(PROJECT_NAME)
-include build/makelib/imagelight.mk

# Fix architecture mismatch BEFORE k8s tools are loaded
# Tools are downloaded to linux_amd64 but build system looks in linux_x86_64
override TOOLS_HOST_DIR := $(CACHE_DIR)/tools/linux_amd64

# Setup K8s tools (for crossplane CLI)
-include build/makelib/k8s_tools.mk

# Setup XPKG
XPKG_REG_ORGS ?= ghcr.io/rossigee
XPKGS = $(PROJECT_NAME)
-include build/makelib/xpkg.mk