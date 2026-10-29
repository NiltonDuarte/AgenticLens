# Contributing to AgenticLens

Guidelines, developer workflow, and build instructions for contributors and autonomous agents working on AgenticLens.

---

## 1. Prerequisites

Ensure the following tools are installed on your workstation or development container:

- **Go**: Version 1.23 or newer
- **Node.js**: Version 20 LTS or newer
- **npm**: Version 10 or newer
- **Git**: For version control and diff calculations
- **Make**: Standard GNU Make

---

## 2. Installation

Install all backend Go modules and frontend npm dependencies:

```bash
make install
```

Or install dependencies individually per subsystem:

```bash
make install-backend   # Downloads Go modules in backend/
make install-frontend  # Installs npm dependencies in web/
```

---

## 3. Development Workflow

### 3.1 Development Mode with Hot Reload

For rapid frontend iteration with hot module replacement (HMR), run the backend server and the Vite dev server concurrently:

```bash
# Terminal 1: Run Go API and telemetry engine
make server
# Or directly: cd backend && go run ./cmd/agentic-lens server --port 8085

# Terminal 2: Run Vite dev server with proxy to backend
cd web && npm run dev
```

Open `http://localhost:5173` to interact with the dashboard.

### 3.2 Compiling the Standalone Production Binary

AgenticLens compiles into a single, self-contained binary embedding all web assets. Zero external Node.js runtime is required in production:

```bash
# Build frontend bundle and compile host executable
make build

# Launch production server
make server
# Or directly: ./bin/agentic-lens server --port 8085
```

Open `http://localhost:8085` to view the standalone dashboard.

---

## 4. Makefile Targets Reference

| Target | Description |
|---|---|
| `make all` | Default target: install dependencies, run static checks, run test suites, and build binary |
| `make install` | Install all backend Go dependencies and frontend npm packages |
| `make install-backend` | Download Go modules (`go mod download` in `backend/`) |
| `make install-frontend` | Install web dependencies (`npm install` in `web/`) |
| `make check` | Run static checks (`go vet`, frontend typecheck, frontend linter, `git diff --check`) |
| `make test` | Run full test suite across Go backend (with race detector) and TypeScript frontend |
| `make test-backend` | Run Go test suite with data race detector enabled (`go test -race ./...`) |
| `make test-frontend` | Run frontend test suite via Vitest (`npm test` in `web/`) |
| `make vet` | Run `go vet` across all Go packages |
| `make lint` | Run Go vet, TypeScript typecheck, and ESLint |
| `make build` | Compile frontend bundle and build multi-platform binaries in `bin/` |
| `make server` | Build and run HTTP server on configured `PORT` (default: 8085) |
| `make serve` | Alias for `make server` |
| `make release-build` | Cross-compile binaries for all supported platforms into `dist/bin/` |
| `make release-package` | Package platform archives (`.tar.gz` and `.zip`) and generate `SHA256SUMS` |
| `make release-verify` | Validate release archive checksums and verify host binary execution |
| `make release-clean` | Remove release distribution artifacts in `dist/` |
| `make clean` | Clean local build outputs and distribution directories |
| `make help` | Display available targets and descriptions |

---

## 5. Quality Standards and Invariants

Every contribution must strictly satisfy the architectural invariants defined in [TECHNICAL_GUIDELINES.md](TECHNICAL_GUIDELINES.md):

1. **Functional Core, Imperative Shell (FCIS)**:
   - Pure domain functions live in `backend/internal/domain`. Never import file systems, databases, network libraries, or clocks into domain code.
   - Imperative side effects live in `backend/internal/adapters` and `backend/cmd/agentic-lens`.
2. **Zero `any`**:
   - Strictly prohibit `any` and `interface{}` across Go and TypeScript. Use strong typing, concrete structs, and discriminated unions.
3. **Data Race Free**:
   - All backend code must pass `go test -race ./...` with zero race conditions.
4. **Zero Emojis**:
   - Zero emojis in code, comments, docstrings, commit messages, PR descriptions, UI strings, logs, or workspace artifacts.
5. **Clean Working Tree**:
   - Run `git diff --check` to ensure no whitespace errors or merge marker conflicts.

---

## 6. Releases and Distribution

Artifacts are built as self-contained standalone executables embedding all web dashboard assets.

### 6.1 Release Tiers

- **Release Candidates (Pre-release)**:
  Triggered automatically on every merge to the `main` branch. Artifacts are published as pre-releases on GitHub with the version format `vMAJOR.MINOR.PATCH-rc.RUN_NUMBER+SHORT_SHA` (for example, `v0.1.0-rc.42+a1b2c3d`).
- **Stable Releases**:
  Triggered when a semantic version tag matching `vMAJOR.MINOR.PATCH` (for example, `v1.0.0`) is pushed to the repository.
- **Floating Latest Tag**:
  The git tag `latest` and the GitHub Release marked as `latest` automatically track the most recent stable release.

### 6.2 Target Platform Matrix

| Platform | OS / Architecture | Archive Format | Artifact Name |
|---|---|---|---|
| Linux x86_64 | `linux/amd64` | `.tar.gz` | `agentic-lens-<version>-linux-amd64.tar.gz` |
| Linux ARM64 | `linux/arm64` | `.tar.gz` | `agentic-lens-<version>-linux-arm64.tar.gz` |
| macOS Intel | `darwin/amd64` | `.tar.gz` | `agentic-lens-<version>-darwin-amd64.tar.gz` |
| macOS Apple Silicon | `darwin/arm64` | `.tar.gz` | `agentic-lens-<version>-darwin-arm64.tar.gz` |
| Windows x86_64 | `windows/amd64` | `.zip` | `agentic-lens-<version>-windows-amd64.zip` |

### 6.3 Local Release Validation

```bash
# Cross-compile for all supported architectures
make release-build

# Package archives and produce SHA256SUMS
make release-package

# Verify archive checksums and validate execution of the host platform binary
make release-verify
```

---

## 7. Ticketed and Ad Hoc Work

Work may be ticketed or ad hoc. Use an existing ticket ID and its isolated workspace when the task is associated with a ticket. Do not create a ticket, specification, or workspace solely to satisfy this policy; ad hoc work can proceed without one.

For ticketed work, workspace deliverables must include:
- `alspaces/<TICKET_ID>/SUMMARY.md`: Work log, completed changes, and verification.
- `alspaces/<TICKET_ID>/PULL_REQUEST.md`: Pull request description and context.
- `alspaces/<TICKET_ID>/agentic_metrics.json`: Final telemetry ledger.

When a matching specification exists under `alspaces/plan/to-do/`, move it to `alspaces/plan/done/` when the ticket is completed. Remove its entry from `alspaces/plan/TICKETS.md` if present. Do not create a ticket, specification, or index entry solely to perform this lifecycle step.
