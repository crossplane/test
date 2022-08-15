

# ====================================================================================
# Setup Project

PROJECT_NAME := test
PROJECT_REPO := github.com/crossplane/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64
# -include will silently skip missing files, which allows us
# to load those files with a target in the Makefile. If only
# "include" was used, the make command would fail and refuse
# to run a target until the include commands succeeded.
-include build/makelib/common.mk

# Set a sane default so that the nprocs calculation below is less noisy on the initial
# loading of this file
NPROCS ?= 1

# ====================================================================================
# Setup Kubernetes tools

UP_VERSION = v0.12.2
UP_CHANNEL = stable
-include build/makelib/k8s_tools.mk

# ====================================================================================
# Setup XPKG

XPKG_REGISTRY ?= us-west1-docker.pkg.dev
XPKG_ORG ?= crossplane-playground/xp-install-test
XPKG_REPO ?= configuration

# ====================================================================================
# Targets

# run `make help` to see the targets and options

# We want submodules to be set up the first time `make` is run.
# We manage the build/ folder and its Makefiles as a submodule.
# The first time `make` is run, the includes of build/*.mk files will
# all fail, and this target will be run. The next time, the default as defined
# by the includes will be run instead.
fallthrough: submodules
	@echo Initial setup complete. Running make again . . .
	@make

# Update the submodules, such as the common build scripts.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# NOTE(hasheddan): the build submodule currently overrides XDG_CACHE_HOME in
# order to force the Helm 3 to use the .work/helm directory. This causes Go on
# Linux machines to use that directory as the build cache as well. We should
# adjust this behavior in the build submodule because it is also causing Linux
# users to duplicate their build cache, but for now we just make it easier to
# identify its location in CI so that we cache between builds.
go.cachedir:
	@go env GOCACHE

# NOTE(hasheddan): we ensure up is installed prior to running platform-specific
# build steps in parallel to avoid encountering an installation race condition.
build.init: $(UP)

# TODO(hasheddan): make xpkg machinery generic.
xpkg.build: $(UP)
	@$(INFO) Building package xp-install-test-configuration-$(VERSION).xpkg for $(PLATFORM)
	@mkdir -p $(OUTPUT_DIR)/xpkg/$(PLATFORM)
	@$(UP) xpkg build \
		--package-root ./packages/xp-install-test/configuration \
		--output ./_output/xpkg/$(PLATFORM)/xp-install-test-configuration-$(VERSION).xpkg || $(FAIL)
	@$(OK) Built package xp-install-test-configuration-$(VERSION).xpkg for $(PLATFORM)

xpkg.push: $(UP)
	@$(INFO) Pushing package xp-install-test-configuration-$(VERSION).xpkg
	@$(UP) xpkg push \
		--package $(OUTPUT_DIR)/xpkg/linux_amd64/xp-install-test-configuration-$(VERSION).xpkg \
		--package $(OUTPUT_DIR)/xpkg/linux_arm64/xp-install-test-configuration-$(VERSION).xpkg \
		$(XPKG_REGISTRY)/$(XPKG_ORG)/$(XPKG_REPO):$(VERSION) || $(FAIL)
	@$(OK) Pushed package xp-install-test-configuration-$(VERSION).xpkg

build.artifacts.platform: xpkg.build

.PHONY: submodules fallthrough

# ====================================================================================
# Special Targets

define CROSSPLANE_TEST_MAKE_HELP
Crossplane Test Targets:
    submodules         Update the submodules, such as the common build scripts.

endef

export CROSSPLANE_TEST_MAKE_HELP

crossplane.test.help:
	@echo "$$CROSSPLANE_TEST_MAKE_HELP"

help-special: crossplane.test.help

.PHONY: crossplane.test.help help-special