# ====================================================================================
# Setup Project

PLATFORMS := linux_amd64 linux_arm64
COMMA := ,
include ../../../build/makelib/common.mk

# ====================================================================================
# Options
IMAGE ?= $(BUILD_REGISTRY)/provider-namecheap:$(VERSION)
TARGETOS := $(word 1, $(subst _, ,$(PLATFORM)))
include ../../../build/makelib/imagelight.mk

# ====================================================================================
# Image Build
img.build:
	@$(INFO) docker build $(IMAGE)
	@$(MAKE) -C ../../.. go.build PLATFORM=$(PLATFORM)
	@$(MAKE) img.build.shared
	@$(OK) docker build $(IMAGE)

img.build.shared:
	@cp Dockerfile $(IMAGE_TEMP_DIR)
	@cp -r $(OUTPUT_DIR)/bin/ $(IMAGE_TEMP_DIR)/
	@docker build $(BUILD_ARGS) --build-arg TARGETOS=$(TARGETOS) --build-arg TARGETARCH=$(TARGETARCH) -t $(IMAGE) $(IMAGE_TEMP_DIR)

# ====================================================================================
# Image Publish
img.publish:
ifeq ($(findstring $(COMMA),$(IMAGE_PLATFORMS)),$(COMMA))
	@$(INFO) building and pushing multiarch image $(IMAGE) for platforms $(IMAGE_PLATFORMS)
	@$(MAKE) -C ../../.. go.build PLATFORM=linux_amd64
	@$(MAKE) -C ../../.. go.build PLATFORM=linux_arm64
	@cp Dockerfile $(IMAGE_TEMP_DIR)
	@cp -r $(OUTPUT_DIR)/bin/ $(IMAGE_TEMP_DIR)/
	@docker buildx create --use --name multiarch-builder 2>/dev/null || docker buildx use multiarch-builder
	@docker buildx build $(BUILD_ARGS) \
		--platform $(IMAGE_PLATFORMS) \
		--tag $(IMAGE) \
		--push \
		$(IMAGE_TEMP_DIR)
	@$(OK) multiarch image $(IMAGE) pushed for $(IMAGE_PLATFORMS)
else
	@$(MAKE) img.build
	@$(INFO) docker push $(IMAGE)
	@docker push $(IMAGE)
	@$(OK) docker push $(IMAGE)
endif