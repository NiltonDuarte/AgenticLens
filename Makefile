.DEFAULT_GOAL := all

BIN_DIR := bin
BINARY := $(BIN_DIR)/agentic-lens
WEB_DIR := web
BACKEND_DIR := backend
PORT ?= 8085
WORKSPACES ?= workspaces

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")

LDFLAGS := -X 'agenticlens/internal/domain.Version=$(VERSION)' \
           -X 'agenticlens/internal/domain.GitCommit=$(GIT_COMMIT)' \
           -X 'agenticlens/internal/domain.BuildDate=$(BUILD_DATE)'

.PHONY: all check test test-backend test-frontend lint vet build clean help server serve

all: check test build

help:
	@echo "Available targets:"
	@echo "  all            - Run checks, tests, and build"
	@echo "  check          - Run static checks (vet, lint, git diff check)"
	@echo "  test           - Run all tests (backend and frontend if present)"
	@echo "  test-backend   - Run Go test suite with race detector"
	@echo "  test-frontend  - Run frontend test suite (if web directory exists)"
	@echo "  vet            - Run go vet on all Go packages"
	@echo "  lint           - Run Go vet and frontend linting"
	@echo "  build          - Build backend executable (and web bundle if present)"
	@echo "  server         - Build and run HTTP server (PORT=$(PORT) WORKSPACES=$(WORKSPACES))"
	@echo "  serve          - Alias for server target"
	@echo "  clean          - Remove build artifacts"

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
	(cd $(BACKEND_DIR) && go build -ldflags "$(LDFLAGS)" -o ../$(BINARY) ./cmd/agentic-lens)

server: build
	./$(BINARY) server --port $(PORT) --workspaces $(WORKSPACES)

serve: server

clean:
	rm -rf $(BIN_DIR)
