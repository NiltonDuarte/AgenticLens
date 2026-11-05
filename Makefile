.DEFAULT_GOAL := all

BIN_DIR := bin
HOST_OS := $(shell go env GOHOSTOS)
HOST_ARCH := $(shell go env GOHOSTARCH)
HOST_EXT := $(if $(filter windows,$(HOST_OS)),.exe,)
BINARY := $(BIN_DIR)/agentic-lens-$(HOST_OS)-$(HOST_ARCH)$(HOST_EXT)
WEB_DIR := web
BACKEND_DIR := backend
PORT ?= 8085
RELEASE_DIR ?= dist
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")

LDFLAGS := -X 'agenticlens/internal/domain.Version=$(VERSION)' \
           -X 'agenticlens/internal/domain.GitCommit=$(GIT_COMMIT)' \
           -X 'agenticlens/internal/domain.BuildDate=$(BUILD_DATE)'

# Escape spaces for Make's file rules and wildcard input, not shell arguments.
empty :=
space := $(empty) $(empty)
make_path = $(subst $(space),\$(space),$(1))

# Dependency preparation is shared with unit execution, including parallel goals.
FRONTEND_DEPS := $(call make_path,$(WEB_DIR)/node_modules/.package-lock.json)

# Build-only commands stay build-only. When test/check goals coexist (including
# default all), their common asset producer and native builds wait for both units.
BUILD_ONLY_GOALS := build verify-build release-build release-package release-verify server serve web-assets install install-backend install-frontend frontend-deps help clean release-clean
VERIFICATION_GOALS := $(filter-out $(BUILD_ONLY_GOALS),$(if $(MAKECMDGOALS),$(MAKECMDGOALS),all))
ifneq ($(strip $(VERIFICATION_GOALS)),)
web-assets build release-build release-package release-verify: test-unit
endif

.PHONY: all install install-backend install-frontend check backend-check test test-backend test-frontend lint vet web-assets build verify-build clean help server serve release-build release-package release-verify release-clean opencode-compat-v1 opencode-compat-v2 opencode-compat

all: install check test build

.PHONY: frontend-deps test-unit test-unit-backend test-unit-frontend test-unit-backend-race test-integration test-backend-integration test-frontend-integration test-orchestration

help:
	@echo "Available targets:"
	@echo "  all                 - Install dependencies, run checks, tests, and build"
	@echo "  install             - Install all dependencies (backend and frontend)"
	@echo "  install-backend     - Download Go backend module dependencies"
	@echo "  install-frontend    - Install npm packages for frontend dashboard"
	@echo "  check               - Run unit gates, static checks and boundary workflow checks"
	@echo "  test                - Run units, then all integration lanes"
	@echo "  test-unit           - Run pure Go and frontend units without embedded assets"
	@echo "  test-integration    - Require both unit lanes, then run integration coverage"
	@echo "  test-backend        - Require both units, then Go unit-race and integration lanes"
	@echo "  test-frontend       - Require both units, then frontend integration tests"
	@echo "  vet                 - Run go vet on all Go packages"
	@echo "  lint                - Run Go vet and frontend linting"
	@echo "  build               - Build source natives (and web bundle if present)"
	@echo "  verify-build        - Build and execute the current host native directly"
	@echo "  release-build       - Cross-compile binaries for all supported platforms"
	@echo "  release-package     - Package release archives and generate SHA256SUMS"
	@echo "  release-verify      - Verify release checksums and host binary execution"
	@echo "  release-clean       - Remove release distribution artifacts"
	@echo "  server              - Build and run HTTP server (PORT=$(PORT))"
	@echo "  opencode-compat-v1  - Run isolated OpenCode v1 compatibility session (requires local provider)"
	@echo "  opencode-compat-v2  - Run isolated OpenCode v2 compatibility session (requires local provider)"
	@echo "  opencode-compat     - Run both isolated sessions and focused adapter/CLI tests"
	@echo "  serve               - Alias for server target"
	@echo "  clean               - Remove build and release artifacts"

install: install-backend install-frontend

install-backend:
	@echo "Installing backend dependencies..."
	(cd "$(BACKEND_DIR)" && go mod download)

install-frontend: frontend-deps

frontend-deps: $(FRONTEND_DEPS)

$(FRONTEND_DEPS): $(call make_path,$(WEB_DIR)/package.json) $(call make_path,$(WEB_DIR)/package-lock.json)
	(cd "$(WEB_DIR)" && npm ci)

check: test-unit vet lint test-release-workflow test-installation test-capture-distribution test-orchestration
	@git diff --check

.PHONY: test-release-workflow test-installation test-capture-distribution
test-installation: test-unit web-assets
	python3 scripts/tests/test_installation.py

test-release-workflow: test-unit
	python3 scripts/tests/test_release_revision.py

test-capture-distribution: test-unit web-assets
	python3 scripts/tests/test_capture_distribution.py

test-orchestration: test-unit
	python3 scripts/tests/test_test_orchestration.py

backend-check: test-unit vet test-backend-integration

web-assets: frontend-deps
	@echo "Building frontend assets required by Go embed..."
	(cd "$(WEB_DIR)" && npm run build)

vet: test-unit web-assets
	(cd "$(BACKEND_DIR)" && go vet ./...)

lint: test-unit vet frontend-deps
	(cd "$(WEB_DIR)" && npm run typecheck && npm run lint)

test-unit: test-unit-backend test-unit-frontend

test-unit-backend:
	(cd "$(BACKEND_DIR)" && go test -count=1 ./internal/domain/...)

test-unit-frontend: frontend-deps
	(cd "$(WEB_DIR)" && npm run test:unit)

test: test-integration

test-integration: test-backend-integration test-frontend-integration

test-backend: test-backend-integration

test-frontend: test-frontend-integration

# This second, instrumented core pass retains the race coverage of the former
# full race suite without classifying mixed adapter/CLI packages as pure units.
test-unit-backend-race: test-unit
	(cd "$(BACKEND_DIR)" && go test -race -count=1 ./internal/domain/...)

test-backend-integration: test-unit test-unit-backend-race web-assets
	@packages=$$(cd "$(BACKEND_DIR)" && "$(CURDIR)/scripts/go-test-packages.sh" integration) || exit 1; \
	(cd "$(BACKEND_DIR)" && go test -race -count=1 $$packages)

test-frontend-integration: test-unit frontend-deps
	(cd "$(WEB_DIR)" && npm exec -- vitest run --config vitest.integration.config.ts)

opencode-compat-v1: test-unit
	bash -c 'source ./scripts/opencode-compat.sh; run_compatibility v1'

opencode-compat-v2: test-unit
	bash -c 'source ./scripts/opencode-compat.sh; run_compatibility v2'

opencode-compat: test-unit web-assets
	bash -c 'source ./scripts/opencode-compat.sh; run_compatibility both'

# One asset writer per Make process; the installation build-only fixture uses
# an absent WEB_DIR to reuse an already prepared embed tree without recursion.
ifneq ($(wildcard $(call make_path,$(WEB_DIR)/package.json)),)
build release-build: web-assets
endif

build:
	@mkdir -p "$(BIN_DIR)"
	@echo "Building multi-platform binaries..."
	@build_dir=$$(CDPATH= cd -P -- "$(BIN_DIR)" && pwd -P) || exit 1; \
	for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; ext=""; \
		if [ "$$os" = windows ]; then ext=".exe"; fi; \
		output="$$build_dir/agentic-lens-$$os-$$arch$$ext"; \
		(cd "$(BACKEND_DIR)" && CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o "$$output" ./cmd/agentic-lens) || exit 1; \
	done

# Source development must exercise this build, never a shared-pin installed release.
verify-build: build
	"$(BINARY)" version
	"$(BINARY)" version --short
	"$(BINARY)" version --json
	"$(BINARY)" -v
	"$(BINARY)" --version

server: build
	"$(BINARY)" server --port $(PORT)

serve: server

release-build:
	@mkdir -p $(RELEASE_DIR)/bin
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		output="$(RELEASE_DIR)/bin/agentic-lens-$$os-$$arch$$ext"; \
		echo "Compiling $$output for $$os/$$arch (CGO_ENABLED=0)..."; \
		(cd "$(BACKEND_DIR)" && CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o ../$$output ./cmd/agentic-lens) || exit 1; \
	done

release-package:
	@mkdir -p $(RELEASE_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		bin_name="agentic-lens-$$os-$$arch$$ext"; \
		bin_path="$(RELEASE_DIR)/bin/$$bin_name"; \
		if [ ! -f "$$bin_path" ]; then \
			echo "Binary $$bin_path not found. Run 'make release-build' first." >&2; \
			exit 1; \
		fi; \
		license_file=""; \
		if [ -f "$(CURDIR)/LICENSE" ]; then license_file="LICENSE"; fi; \
		if [ "$$os" = "windows" ]; then \
			archive_name="agentic-lens-$(VERSION)-$$os-$$arch.zip"; \
			echo "Packaging $$archive_name..."; \
			rm -f "$(RELEASE_DIR)/$$archive_name"; \
			zip -q -j "$(RELEASE_DIR)/$$archive_name" "$$bin_path" "$(CURDIR)/README.md" $$( [ -n "$$license_file" ] && echo "$(CURDIR)/$$license_file" ) || exit 1; \
		else \
			archive_name="agentic-lens-$(VERSION)-$$os-$$arch.tar.gz"; \
			echo "Packaging $$archive_name..."; \
			rm -f "$(RELEASE_DIR)/$$archive_name"; \
			tar -czf "$(RELEASE_DIR)/$$archive_name" -C "$(RELEASE_DIR)/bin" "$$bin_name" -C "$(CURDIR)" README.md $$license_file || exit 1; \
		fi; \
	done
	@echo "Generating SHA256SUMS..."
	@(cd $(RELEASE_DIR) && sha256sum agentic-lens-$(VERSION)-*.tar.gz agentic-lens-$(VERSION)-*.zip bin/agentic-lens-* | sort -k2 > SHA256SUMS)

release-verify:
	@if [ ! -f "$(RELEASE_DIR)/SHA256SUMS" ]; then \
		echo "SHA256SUMS not found in $(RELEASE_DIR). Run 'make release-package' first." >&2; \
		exit 1; \
	fi
	@echo "Verifying SHA256SUMS in $(RELEASE_DIR)..."
	@(cd $(RELEASE_DIR) && sha256sum -c SHA256SUMS)
	@host_os=$$(go env GOOS); \
	host_arch=$$(go env GOARCH); \
	ext=""; \
	if [ "$$host_os" = "windows" ]; then ext=".exe"; fi; \
	host_bin="$(RELEASE_DIR)/bin/agentic-lens-$$host_os-$$host_arch$$ext"; \
	if [ -f "$$host_bin" ]; then \
		echo "Verifying host binary: $$host_bin"; \
		./$$host_bin version || exit 1; \
		./$$host_bin version --short || exit 1; \
		./$$host_bin version --json || exit 1; \
	fi

release-clean:
	rm -rf $(RELEASE_DIR)

clean: release-clean
	rm -rf $(BIN_DIR)
