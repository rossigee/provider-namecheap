# Project Setup
PROJECT_NAME := provider-namecheap
PROJECT_REPO := github.com/rossigee/$(PROJECT_NAME)
CROSSPLANE_VERSION = 2.5.0

# Platform support
PLATFORMS ?= linux_amd64 linux_arm64

# Include build system
-include build/makelib/common.mk


# Setup Output
-include build/makelib/output.mk

# Fix architecture mismatch BEFORE golang and k8s tools are loaded
# Tools are downloaded to linux_amd64 but build system looks in linux_x86_64
override TOOLS_HOST_DIR := $(CACHE_DIR)/tools/linux_amd64

# Setup Go
# Override golangci-lint version for modern Go support
GOLANGCILINT_VERSION ?= 2.13.2
NPROCS ?= 1
GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/provider
GO_LDFLAGS += -X $(GO_PROJECT)/internal/version.Version=$(VERSION)
GO_SUBDIRS += apis
GO111MODULE = on
GO_REQUIRED_VERSION ?= 1.27.1
-include build/makelib/golang.mk

# Setup Images
REGISTRY_ORGS ?= ghcr.io/rossigee
IMAGES = $(PROJECT_NAME)
-include build/makelib/imagelight.mk

# Setup K8s tools (for crossplane CLI)
-include build/makelib/k8s_tools.mk

# Setup XPKG
XPKG_REG_ORGS ?= ghcr.io/rossigee
XPKGS = $(PROJECT_NAME)
-include build/makelib/xpkg.mk

# Override xpkg.build target to ensure CROSSPLANE_CLI dependency
xpkg.build.$(PROJECT_NAME): $(CROSSPLANE_CLI)

# Ensure publish only happens on release branches
publish.artifacts: $(CROSSPLANE_CLI)
	@if ! echo "$(BRANCH_NAME)" | grep -qE "$(subst $(SPACE),|,main|master|release-.*)"; then \ 
		$(ERR) Publishing is only allowed on branches matching: main|master|release-.* (current: $(BRANCH_NAME)); \ 
		exit 1; \ 
	fi
	$(foreach r,$(XPKG_REG_ORGS), $(foreach x,$(XPKGS),@$(MAKE) xpkg.release.publish.$(r).$(x)))
	$(foreach r,$(REGISTRY_ORGS), $(foreach i,$(IMAGES),@$(MAKE) img.release.publish.$(r).$(i)))
xpkg.release.publish.ghcr.io/rossigee.provider-namecheap:
	@$(foreach p,$(XPKG_LINUX_PLATFORMS),$(MAKE) xpkg.build.provider-namecheap PLATFORM=$(p) || exit 1;)
	@$(CROSSPLANE_CLI) xpkg push \
		$(foreach p,$(XPKG_LINUX_PLATFORMS),--package-files $(XPKG_OUTPUT_DIR)/$(p)/provider-namecheap-$(VERSION).xpkg ) \
		ghcr.io/rossigee/provider-namecheap:$(VERSION)
	@$(OK) Pushed package ghcr.io/rossigee/provider-namecheap:$(VERSION)



# Neutralize plain image publish for ghcr (xpkg uses same ref)
img.release.publish.ghcr.io/rossigee.provider-namecheap:
	@:
