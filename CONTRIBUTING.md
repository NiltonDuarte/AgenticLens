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
| `make backend-check` | Prepare embedded assets, run Go vet and backend race tests (shared local/CI entry point) |
| `make web-assets` | Build frontend and copy assets into the Go embed tree; requires installed web dependencies |
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

Select the applicable cases in the [regression prevention checklist](docs/REGRESSION_PREVENTION.md) before changing extraction, configuration, hydration, live polling, or distribution. Record evidence in the active iteration review.

After `make install`, use `make backend-check` for backend-only work or `make` for the full gate. Both prepare assets before Go loads embed-dependent packages. Before direct `go vet`, `go test`, or `go run`, run `make web-assets`. Build/CI changes require verification with generated assets absent and on a repeated build; frontend source checks alone do not validate the standalone bundle.

Follow [TECHNICAL_GUIDELINES.md](TECHNICAL_GUIDELINES.md) for FCIS boundaries, strict typing, race freedom, and the zero-emoji rule. `git diff --check` must pass.

### 5.1 Before Adding or Removing Tests

Identify the regression and inspect existing tests before writing another case. Follow [test purpose and proportionality](TECHNICAL_GUIDELINES.md#test-purpose-and-proportionality) and the [test review checklist](docs/REGRESSION_PREVENTION.md#adding-reviewing-or-reducing-tests): extend coherent existing coverage, preserve unique assertions when consolidating, and record a specific rationale for deletions. Keep a separate test when it provides independent coverage or clearer failure diagnosis.

Scores and test-to-code ratios are review signals, not quality targets. Include classification/counting assumptions in metric reports and describe what verification actually exercised. A test of workflow text does not prove a release works. Record transferred assertions, actual execution results, and confidence limits in the active iteration review.

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

Follow [AGENTS.md](AGENTS.md#1-ticketed-and-ad-hoc-work) for ticket allocation, configured workspace roots, required artifacts, plan lifecycle, final telemetry, and PR delivery. Ad hoc work still requires an ID and workspace, but no planning specification or index entry.

### 7.1 Recording steering without restarting the ticket

For material scope changes, write a short delta in the next iteration plan:

- **Source and intent:** summarize the current user instruction or linked review finding; label inference separately.
- **Baseline:** link the contract and applicable accepted plan/review, and name the current commit.
- **Acceptance delta:** state the observable behavior added, removed, or replaced. Name the specific decision superseded and reconcile the current contract when its boundaries change.
- **Preserved guarantees:** list adjacent behavior that must continue to work; identify affected paths and their owner.
- **Verification impact:** reuse valid results for unchanged code and select affected checks for the delta. A planned command is not a passing result.
- **Delivery status:** state proposed, approved, implemented, or verified using actual evidence. Approval persists within its accepted scope; a plan file or iteration number is not approval evidence.

Keep status questions, review corrections, and routine synchronization in the current iteration unless acceptance changes. Consolidate plans when explicitly requested, preserving accepted decisions and verification in the surviving artifacts. A missing plan is not evidence of no steering, and iteration-directory counts are not SDLC telemetry.
