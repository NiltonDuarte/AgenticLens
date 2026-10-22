.DEFAULT_GOAL := all

BIN_DIR := bin
BINARY := $(BIN_DIR)/agentic-lens
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

.PHONY: all install install-backend install-frontend check test test-backend test-frontend lint vet build clean help server serve release-build release-package release-verify release-clean

all: install check test build

help:
	@echo "Available targets:"
	@echo "  all                 - Install dependencies, run checks, tests, and build"
	@echo "  install             - Install all dependencies (backend and frontend)"
	@echo "  install-backend     - Download Go backend module dependencies"
	@echo "  install-frontend    - Install npm packages for frontend dashboard"
	@echo "  check               - Run static checks (vet, lint, git diff check)"
	@echo "  test                - Run all tests (backend and frontend if present)"
	@echo "  test-backend        - Run Go test suite with race detector"
	@echo "  test-frontend       - Run frontend test suite (if web directory exists)"
	@echo "  vet                 - Run go vet on all Go packages"
	@echo "  lint                - Run Go vet and frontend linting"
	@echo "  build               - Build backend executable (and web bundle if present)"
	@echo "  release-build       - Cross-compile binaries for all supported platforms"
	@echo "  release-package     - Package release archives and generate SHA256SUMS"
	@echo "  release-verify      - Verify release checksums and host binary execution"
	@echo "  release-clean       - Remove release distribution artifacts"
	@echo "  server              - Build and run HTTP server (PORT=$(PORT))"
	@echo "  serve               - Alias for server target"
	@echo "  clean               - Remove build and release artifacts"

install: install-backend install-frontend

install-backend:
	@echo "Installing backend dependencies..."
	(cd $(BACKEND_DIR) && go mod download)

install-frontend:
	@if [ -d "$(WEB_DIR)" ]; then \
		echo "Installing frontend dependencies..."; \
		(cd $(WEB_DIR) && npm install); \
	fi

check: vet lint
	@git diff --check

vet:
	(cd $(BACKEND_DIR) && go vet ./...)

lint: vet
	@if [ -d "$(WEB_DIR)/node_modules" ]; then \
		echo "Linting frontend..."; \
		(cd $(WEB_DIR) && npm run typecheck && npm run lint); \
	fi

test: test-backend test-frontend

test-backend:
	(cd $(BACKEND_DIR) && go test -race ./...)

test-frontend:
	@if [ -d "$(WEB_DIR)/node_modules" ]; then \
		echo "Testing frontend..."; \
		(cd $(WEB_DIR) && npm test); \
	fi

build:
	@if [ -d "$(WEB_DIR)/node_modules" ]; then \
		echo "Building frontend bundle..."; \
		(cd $(WEB_DIR) && npm run build); \
	fi
	@mkdir -p $(BIN_DIR)
	@echo "Building multi-platform binaries..."
	(cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../$(BIN_DIR)/agentic-lens-linux-amd64 ./cmd/agentic-lens)
	(cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../$(BIN_DIR)/agentic-lens-linux-arm64 ./cmd/agentic-lens)
	(cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../$(BIN_DIR)/agentic-lens-darwin-arm64 ./cmd/agentic-lens)
	(cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../$(BIN_DIR)/agentic-lens-darwin-amd64 ./cmd/agentic-lens)
	@cp scripts/agentic-lens-wrapper.sh $(BINARY)
	@chmod +x $(BINARY)

server: build
	./$(BINARY) server --port $(PORT)

serve: server

release-build:
	@if [ -d "$(WEB_DIR)/node_modules" ]; then \
		echo "Building frontend bundle..."; \
		(cd $(WEB_DIR) && npm run build); \
	fi
	@mkdir -p $(RELEASE_DIR)/bin
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		output="$(RELEASE_DIR)/bin/agentic-lens-$$os-$$arch$$ext"; \
		echo "Compiling $$output for $$os/$$arch (CGO_ENABLED=0)..."; \
		(cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o ../$$output ./cmd/agentic-lens) || exit 1; \
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
