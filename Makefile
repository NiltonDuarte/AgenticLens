.DEFAULT_GOAL := all

BIN_DIR := bin
BINARY := $(BIN_DIR)/agentic-lens
WEB_DIR := web

.PHONY: all check test test-backend test-frontend lint vet build clean help

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
	@echo "  clean          - Remove build artifacts"

check: vet lint
	@git diff --check

vet:
	go vet ./...

lint: vet
	@if [ -d "$(WEB_DIR)/node_modules" ]; then \
		echo "Linting frontend..."; \
		(cd $(WEB_DIR) && npm run typecheck && npm run lint); \
	fi

test: test-backend test-frontend

test-backend:
	go test -race ./...

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
	go build -o $(BINARY) ./cmd/agentic-lens

clean:
	rm -rf $(BIN_DIR)
