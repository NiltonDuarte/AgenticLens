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
- **Python 3, GitHub CLI (`gh`) and curl**: For release workflow tests. API tests use local fixtures and require no GitHub login.

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

### 3.2 Compiling the Native Manager and Engine

AgenticLens builds a small native manager and a separate telemetry engine. The engine embeds all web assets; neither executable requires Node.js in production. `make server` runs the current source engine directly:

```bash
# Build frontend bundle and compile manager/engine artifacts
make build

# Launch production server
make server
# Or directly on POSIX: ./bin/agentic-lens-engine-$(go env GOHOSTOS)-$(go env GOHOSTARCH) server --port 8085
```

Open `http://localhost:8085` to view the standalone dashboard.

---

## 4. Makefile Targets Reference

| Target | Description |
|---|---|
| `make all` | Default target: install dependencies, run static checks, run test suites, and build manager/engine artifacts |
| `make install` | Install all backend Go dependencies and frontend npm packages |
| `make install-backend` | Download Go modules (`go mod download` in `backend/`) |
| `make install-frontend` | Prepare web dependencies with `npm ci` when manifests changed or dependencies are absent |
| `make backend-check` | Require both unit lanes, prepare shared embedded assets, then Go vet and both Go race lanes |
| `make web-assets` | Prepare dependencies, build frontend and copy assets into the Go embed tree; shared once per Make invocation |
| `make check` | Require both unit lanes, then static checks and Python workflow/distribution/orchestration integration checks |
| `make test-unit` | Run pure Go domain/migration and frontend Node units, without embedded asset preparation |
| `make test-unit-backend` | Run only the fast non-race Go pure-core unit lane |
| `make test-unit-frontend` | Run only frontend Node units |
| `make test-integration` / `make test` | Require both unit lanes, then complete Go race and frontend integration coverage |
| `make test-backend-integration` / `make test-backend` | Require both units, then domain race verification and the complementary Go integration packages with race |
| `make test-frontend-integration` / `make test-frontend` | Require both units, then the complementary frontend DOM/API/filesystem integration lane |
| `make test-release-workflow` | Require both units, then release revision, metadata/guard and response-filter tests against local fixtures |
| `make test-orchestration` | Require both units, then exercise unit failure and integration sequencing with isolated real tool fixtures |
| `make vet` | Run `go vet` across all Go packages |
| `make lint` | Run Go vet, TypeScript typecheck, and ESLint |
| `make build` | Compile frontend bundle and build multi-platform binaries in `bin/` |
| `make server` | Build and run the loopback HTTP server on port 8085; CLI --port selects another explicit port |
| `make serve` | Alias for `make server` |
| `make release-build` | Cross-compile binaries for all supported platforms into `dist/bin/` |
| `make release-package` | Package platform archives (`.tar.gz`) and generate `SHA256SUMS` |
| `make release-verify` | Validate release archive checksums and verify host manager/engine execution |
| `make release-clean` | Remove release distribution artifacts in `dist/` |
| `make clean` | Clean local build outputs and distribution directories |
| `make help` | Display available targets and descriptions |

---

## 5. Quality Standards and Invariants

Preserve the [portable ledger contract](TECHNICAL_GUIDELINES.md#portable-ledger-contract). For persistence, capture, refresh or migration changes, select the [cross-developer/machine scenarios](docs/REGRESSION_PREVENTION.md#portable-ledgers-across-developers-and-machines); a passing test that requires private originating-machine proof does not establish a valid product requirement.

Select the applicable cases in the [regression prevention checklist](docs/REGRESSION_PREVENTION.md) before changing extraction, configuration, hydration, live polling, or distribution. Record evidence in the active iteration review.

Use `make test-unit` for fast feedback and `make` for the full gate. `make check`, `make backend-check`, every supported Make integration target and direct compatibility scripts require both Go and frontend units to succeed before boundary work starts. The dependency graph also applies to parallel and mixed goals, such as `make -j8 check test build`; units are not a sibling prerequisite that can race integrations. Frontend dependency preparation is shared with unit execution rather than silently skipping an absent frontend. No persistent unit-success stamp is used.

The Go unit lane is `backend/internal/domain/...`; all other packages form a conservative integration lane, including pure cases in mixed adapter/CLI packages. Race verification covers both lanes after the fast gate. CLI package `TestMain` still builds its native application, and real storage, process and platform oracles remain integration tests. The outer race runner does not implicitly instrument that separate native build.

The strict TypeScript check covers source plus both Vitest configuration roots and their imported shared selectors.

Frontend classification is shared by `web/test-suites.ts`: domain and the named pure utility owners run under Node, while the disjoint complement retains jsdom, setup and explicit per-file environment annotations. The Node-annotated asset-sync test remains integration because it exercises real files and processes. From `web/`, `npm run test:unit` runs frontend units; `npm test` and `npm run test:integration` always run the complete frontend unit lane before integration. Standalone npm commands are frontend-scoped and do not require Go or Make. Repository Make and CI use the stronger two-language gate. Release workflows invoke the selected revision's established `make check` and `make test` entrypoints; revisions containing this Make graph enforce the unit gate internally. Forced rebuilds of historical tags retain that tag's verification commands and semantics, without importing newer source or requiring newer Make targets. CI's Linux distribution job waits for both unit jobs; build verification waits for every required integration branch.

Direct `go test`, `vitest`, Python test files and raw build commands are diagnostic probes, not supported unit-first integration entrypoints. Prepare assets with `make web-assets` before embed-dependent Go probes; pure domain units do not need assets. Standalone `make build`, `make verify-build` and release build/package commands remain build-only. Asset preparation is shared when build and verification goals coexist; an installation fixture may override `WEB_DIR` to reuse already prepared assets without recursively running tests. Build/CI changes require verification with generated assets absent and on a repeated build; frontend source checks alone do not validate the standalone bundle.

Follow [TECHNICAL_GUIDELINES.md](TECHNICAL_GUIDELINES.md) for FCIS boundaries, strict typing, race freedom, and the zero-emoji rule. `git diff --check` must pass.

### 5.1 Before Adding or Removing Tests

Identify the regression and inspect existing tests before writing another case. Follow [test purpose and proportionality](TECHNICAL_GUIDELINES.md#test-purpose-and-proportionality) and the [test review checklist](docs/REGRESSION_PREVENTION.md#adding-reviewing-or-reducing-tests): extend coherent existing coverage, preserve unique assertions when consolidating, and record a specific rationale for deletions. Keep a separate test when it provides independent coverage or clearer failure diagnosis.

Scores and test-to-code ratios are review signals, not quality targets. Include classification/counting assumptions in metric reports and describe what verification actually exercised. A test of workflow text does not prove a release works. Record transferred assertions, actual execution results, and confidence limits in the active iteration review.

---

## 6. Releases and Distribution

Artifacts are built as self-contained standalone executables embedding all web dashboard assets.

### 6.1 Release Tiers

- **Release Candidates (Pre-release)**:
  Triggered after the Release Please workflow succeeds on `main`, or manually through workflow dispatch. The version is captured from the root manifest at the immutable head of the pending Release Please PR. A missing release PR skips the candidate; API errors and invalid versions fail the job. Later changes to `main` or the release PR do not invalidate the captured version. Before publication, an exact `vMAJOR.MINOR.PATCH` tag check skips an already reserved stable target. Artifacts use `vMAJOR.MINOR.PATCH-rc.RUN_NUMBER+SHORT_SHA` (for example, `v2.1.0-rc.42+a1b2c3d`), reusing Release Please's version decision without another calculator.
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

Apply the [shared writing rules](AGENTS.md#artifact-writing-rules). Artifact purposes and length guidance belong to the producing skill: [development](.agents/skills/agentic-lens-dev/SKILL.md#setup--workspace-structure), [PO](.agents/skills/agentic-lens-po/SKILL.md#po-artifact-definitions), [management](.agents/skills/agentic-lens-management/SKILL.md#management-artifact-definitions) and [review](.agents/skills/agentic-lens-review/SKILL.md#evidence-and-outcomes).

### 7.1 Recording steering without restarting the ticket

For material scope changes, write a short delta in the next iteration plan:

- **Source and intent:** summarize the current user instruction or linked review finding; label inference separately.
- **Baseline:** link the contract and applicable accepted plan/review, and name the current commit.
- **Acceptance delta:** state the observable behavior added, removed, or replaced. Name the specific decision superseded and reconcile the current contract when its boundaries change.
- **Preserved guarantees:** list adjacent behavior that must continue to work; identify affected paths and their owner.
- **Verification impact:** reuse valid results for unchanged code and select affected checks for the delta. A planned command is not a passing result.
- **Delivery status:** state proposed, approved, implemented, or verified using actual evidence. Approval persists within its accepted scope; a plan file or iteration number is not approval evidence.

Keep status questions, review corrections, and routine synchronization in the current iteration unless acceptance changes. Consolidate plans when explicitly requested, preserving accepted decisions and verification in the surviving artifacts. A missing plan is not evidence of no steering, and iteration-directory counts are not SDLC telemetry.

### Installed release entrypoints

The contributor build produces `bin/agentic-lens-<os>-<arch>` (manager) and `bin/agentic-lens-engine-<os>-<arch>` (engine). `make server` and the engine checks in `make verify-build` execute the current host engine directly, without selecting the user release cache. The same verification target checks the manager's own identity separately. In a configured adopting repository, the installed `agentic-lens server` command remains supported: the manager selects the project engine and forwards `server` unchanged.

The installed native manager handles offline engine selection and explicit installation/self-update; see the [setup manual](docs/SETUP_MANUAL.md#1-quick-start-and-setup-pathways). It is installed directly on PATH, with no shell launcher or separate dispatcher copy. Project engine pins and manager updates are independent. Setup verifies exact artifact identity/checksum/protocol; tests isolate HOME, cache and project fixtures. Supported platforms are Linux and macOS, each on amd64 and arm64. Windows is unsupported.
