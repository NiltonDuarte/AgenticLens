# AgenticLens Technical Guidelines

Mandatory technical guidelines for developing AgenticLens using **Go** (backend, domain engine, CLI, and HTTP services) and **TypeScript** (frontend dashboard and UI presentation). All code must strictly adhere to these rules.

---

## 1. Architectural Boundaries and Layer Separation

Dependencies flow strictly inward across layers:

### Backend (Go)
1. **CLI / Service Layer (`backend/cmd/agentic-lens/`)**:
   - Application entrypoint, CLI commands (`scan`, `metrics`, `summary-append`, `server`), HTTP/gRPC server wiring.
2. **Adapter Layer (`backend/internal/adapters/`)**:
   - External boundary serialization (JSON encoding/decoding), filesystem repository scanner (`workspaces/` traversal), HTTP handlers, API clients.
3. **Domain Layer / Functional Core (`backend/internal/domain/`)**:
   - Pure domain models, value objects, aggregation algorithms, health score derivations, constants, and typed errors.
   - **Zero External Dependencies**: Must NOT import outer layers (`internal/adapters`, `cmd/`) or network/disk libraries (`net/http`, `os`, `database/sql`).

### Frontend (TypeScript)
1. **Views & Components (`web/src/components/`, `web/src/views/`)**:
   - Interactive analytics dashboard, KPI widgets, comparative charts, filtering bars.
2. **API & Boundary Layer (`web/src/api/`, `web/src/adapters/`)**:
   - Network clients, JSON boundary validation (via Zod or strict type guards).
3. **Domain Types & Pure Formatters (`web/src/domain/`, `web/src/types/`)**:
   - Strictly typed data contracts mirroring the Go domain model, pure mathematical or date formatting utilities.

---

## 2. Functional Core, Imperative Shell (FCIS)

### Go Functional Core
- All mathematical rollups, context window health evaluations (HEALTHY, ELEVATED, BLOATED, CRITICAL), token formatting, turn velocity, and financial cost derivations must be pure functions in `internal/domain`.
- Pure functions receive immutable structs/slices and return computed models or explicit errors.
- No disk I/O, no network calls, no clocks (`time.Now()`), and no random number generators in the domain core. Inject timestamps or random seeds as function arguments.
- Defensively guard against division by zero (e.g. empty workspaces, zero turns, zero duration, zero tokens).

### Imperative Shell (Go & TypeScript)
- The shell handles filesystem traversal, directory scanning, JSON deserialization, HTTP routing, and UI rendering.
- Errors from the shell must be caught, wrapped with descriptive context (`fmt.Errorf("...: %w", err)`), and handled gracefully.

### Fact Fidelity and Refresh Boundaries

- Required identity and metadata are validated before persistence. Optional or unknown values remain explicit; do not invent completion dates from started/updated dates, assign feature types to unclassified tickets, or represent unavailable measurements as measured zero.
- Resolve stored sessions by exact identity by default. Heuristic recovery must be explicitly enabled and must not silently overwrite unrelated stored facts.
- Refresh each source-owned session/tree only after discovery, parsing, and grouping succeed. Preserve all roots, children, legitimate repeated invocation rows, unavailable sessions, and other-harness facts. Replace cumulative snapshots for refreshed owners; repeated refresh of unchanged input must be idempotent.
- Persisted facts are the baseline for reads. Optional hydration may add information but must not hide accepted baseline data on failure, delay, or partial success. Keep compact persistence transforms separate from read projections; never persist hydrated payloads as a side effect of GET.
- Reuse canonical domain types and source-backed fields. Do not add a persisted metric with no defined producer or infer execution failure solely from error-like text returned by a successful inspection tool.

### Hierarchy, Evidence, and UI Recovery

- Structural parent linkage determines root/child classification; persona names are raw facts and configured presentation mappings must not rewrite source identity. Count SDLC steering from the source records defined by the telemetry contract, never from iteration directories or a forced minimum.
- Keep source diagnostic payloads separate from generated explanatory copy. Preserve verbatim evidence except for explicit approved redaction; text cleanup must not change execution semantics.
- Controls needed to recover from empty filters remain available in empty states. Layout fixes must preserve tooltip visibility, table containment, long-value access, keyboard behavior, and existing collapse defaults together.
- Distinguish a fatal project setup error from a rejected individual ledger. Preserve valid sibling workspace results and surface per-ticket diagnostics; never count invalid ledgers in aggregates.

### Configuration and Live Work

- Trace shared configuration through loader, validation, writer/templates, CLI, live/offline calculations, API, and UI. Test non-default values end to end. Required settings fail explicitly when absent; do not revive removed configuration aliases implicitly.
- Polling designs must state a work bound in terms of session count, history size, subscribers, and cadence. Prefer changed-data reads; avoid rebuilding every full trace on unchanged ticks.
- Failed optional lookups must surface error or partial-data semantics. An unsuccessful lookup is not evidence of an empty result.

---

## 3. Go Design & Idioms

### Strong Typing & Value Semantics
- Target Go 1.23+.
- Use value semantics and immutable struct patterns for domain data entities.
- Zero `any` or `interface{}` across domain and adapters. Use concrete structs, generics, or type-safe sum types / interfaces.
- Define interfaces where they are consumed, keeping interfaces small (1-2 methods).
- Use typed constants / string enums (`type TimeWindow string`) rather than untyped raw string primitives.

### Error Handling & Validation
- Never discard errors (`_ = ...` is forbidden for I/O, parsing, or domain execution).
- Wrap errors with domain-relevant context: `fmt.Errorf("parsing fact ledger for ticket %s: %w", ticketID, err)`.
- Differentiate between non-fatal scan warnings (e.g. corrupted single JSON file) and fatal execution errors.

---

## 4. TypeScript Design & Idioms

### Strict Typing & No Any
- Target TypeScript 5.x+ in strict mode (`"strict": true`, `"noImplicitAny": true`, `"exactOptionalPropertyTypes": true`).
- The `any` type is strictly forbidden across all TypeScript code. Use `unknown` with runtime type narrowing or Zod schemas at boundaries.
- Discriminated unions for multi-state UI representation (Loading, Success, Empty, Error).

### UI Presentation Rules
- Pure presentation components separated from network fetching.
- Declarative state management and reactive chart updates.
- Zero emojis permitted in UI titles, tab labels, metric cards, tables, charts, tooltips, or console logs.
- Defensive presentation: handle missing or empty datasets gracefully with empty-state messages without uncaught runtime errors.

---

## 5. Single Binary Distribution & Static Asset Bundling

AgenticLens deploys as a single, self-contained executable with the TypeScript web application bundled directly into the Go binary.

### Rules
1. **Embedding Mandate (`embed.FS`)**:
   - The compiled static frontend output (`web/dist/`) must be bundled into Go via `embed.FS` (e.g. `//go:embed all:dist` in `internal/adapters/web/` or `cmd/agentic-lens/`).
   - The Go HTTP server (`cmd/agentic-lens server`) serves the SPA static files alongside the JSON/SSE API endpoints.
   - Zero Node.js runtime dependency in production or container deployments.
2. **Dual-Mode Serving**:
   - Production / Standalone: Serves embedded static assets from `embed.FS` with client-side fallback routing for SPA navigation.
   - Local Development: Vite dev server runs with HMR and proxies API requests (`/api/*`) to the Go backend.
3. **Build Sequence**:
   - Production binary build sequence: `(cd web && npm run build)` followed by `(cd backend && go build -o ../bin/agentic-lens ./cmd/agentic-lens)`.
   - The frontend build copies assets to `backend/internal/adapters/web/dist/`, the Go embed input. Verify that tree after frontend changes. Respect its Git tracking policy; do not add ignored build output.
   - CI and container builds must compile the frontend first before invoking Go build.

---

## 6. Quality Checks & Verification

Before submitting any iteration or PR, all quality gates must pass:

### Automated Verification (Makefile)
```bash
# Bootstrap dependencies (Make prepares embedded assets before Go checks)
make install

# Run all quality checks, tests, and build
make

# Or run specific stages:
make backend-check  # Backend vet and race tests, with embedded-asset prerequisite
make web-assets     # Prepare assets before invoking Go tooling directly
make check          # Static analysis: go vet, frontend lint/typecheck, git diff check
make test           # Unit and race-detected test suites (backend and frontend)
make build          # Single binary build (including frontend bundle if present)
```

### Manual Go Checks

Run `make web-assets` first with frontend dependencies installed. For routine backend verification, prefer `make backend-check`, which carries this prerequisite.

```bash
# Verify compilation and format
(cd backend && go vet ./...)
# Run test suite with race detector
(cd backend && go test -race ./...)
# Build CLI binary
(cd backend && go build -o ../bin/agentic-lens ./cmd/agentic-lens)
```

### TypeScript Checks & Frontend Build
Run these commands from `web/` after installing its dependencies.

```bash
# Typecheck
npm run typecheck  # or npx tsc --noEmit
# Linting
npm run lint
# Unit tests
npm test
# Production frontend build
npm run build      # Produces web/dist/ for Go embedding
```

### Test Purpose and Proportionality

Each test must exercise application behavior or an intentional wire/compiler contract and identify a plausible regression its assertions would detect. Before adding a test, inspect existing cases and helpers. Extend a coherent existing test or table when it can express the missing behavior clearly; add a separate test when it needs distinct setup, ownership, or failure diagnosis. Select boundary cases relevant to the changed contract rather than repeating a generic edge-case checklist for every function.

- Avoid assertions that only confirm fixture literals, direct struct assignments, TypeScript casts, standard-library behavior, or an implementation copied into the test. Exercise the production path and use an independent expected result. A local JSON roundtrip plus a TypeScript cast does not validate an API; actual JSON tags, custom codecs, runtime validators, and negative compiler checks can protect real contracts.
- Read helper assertions and compiler directives before calling a test assertion-free. Keep important absence/default/override behavior, failure handling, source provenance, ownership, multiplicity, numerical boundaries, accessible interactions, and asynchronous state transitions where the feature depends on them. Coverage of the same feature at different layers is not automatically redundant.
- Delete a test only with an explicit redundancy or invalid-contract rationale. Before consolidating it, map every meaningful assertion to a named existing recipient exercising the same behavior. Preserve the oracle and setup needed to reach it; retain a separate guard when no coherent recipient exists. Avoid unrelated large test bodies or helper suites created merely to reduce declaration counts.
- Scores, test counts, coverage percentages, runtime, and test-to-production line ratios guide investigation; none is a creation or deletion quota. If a 0-10 score is requested, state the rubric: 0 identifies a candidate with no useful guard; 10 identifies a critical contract guard. A low score requires inspection, and retained tests keep their honest scores. Report whether a review screened source or examined bodies and peers directly; passing the suite alone does not establish equivalent defect detection.

### Match Verification to the Claimed Boundary

Tests of workflow text, help strings, or copied release logic do not establish successful CI execution. Missing-file/checksum guards prove their rejection path; build, packaging, publishing, download, and host execution need evidence at those boundaries when they are in scope. Describe that limit rather than claiming full release coverage. Likewise, source/CSS assertions do not establish browser geometry, and DOM text alone may not establish accessible behavior.

Domain tests remain pure and deterministic with zero mocks. Real filesystem, SQLite, and Git fixtures are appropriate when testing adapter behavior that depends on them. Preserve those boundaries when optimizing; an in-memory substitute must not erase the behavior under test. Parallelize only independent tests; environment, process-global, and linker-metadata mutation require isolation or serial execution. Replace arbitrary sleeps with deterministic synchronization or controlled timers where practical. Measure performance with comparable cache, race-detector, and concurrency settings; rounded test-runner timings do not establish a microsecond SLA.

### Interpreting Test Metrics

When reporting a line ratio, state source roots, classification patterns, exclusions, and treatment of helpers, fixtures, generated files, comments, and blank lines. Distinguish whole-tree physical source counts from executable lines and per-diff added lines. Configured telemetry classification and an adjusted source estimate may differ; label both rather than silently substituting one for the other. There is no universal good ratio: assess risk coverage, maintenance cost, and actual regressions caught.

### Fixture and Verification Boundaries

Select cases from the [regression prevention checklist](docs/REGRESSION_PREVENTION.md) before implementing boundary changes.

- Tests must construct isolated configuration, ledgers, and source fixtures; do not read live `alspaces/` history, developer databases, or depend on actual ticket IDs. A ticket rename must not break tests.
- Exercise configurable roots and classification patterns, unavailable optional enrichment, malformed records, and partial source trees where relevant. Keep provenance explicit for sanitized real-format fixtures versus synthetic examples.
- Run focused checks during rework, then the required full `make` gate on the integrated result. Record commands, results, and skipped stages; a conditional Makefile target that ran no frontend checks is not proof they passed.

### Invariant Checks
- Verify zero `any` in Go and TypeScript.
- Verify zero emojis in code, UI, docstrings, and log messages.
- Verify single binary self-contained execution without Node runtime dependency.

---

## Summary Checklist

| Topic | Decision | Rule |
|---|---|---|
| **Backend Architecture** | Inward Dependencies | `internal/domain` has zero I/O imports. `internal/adapters` handles JSON and disk. `cmd/` handles CLI/server. |
| **Separation** | FCIS | Pure deterministic logic in domain; filesystem I/O and HTTP in shell. |
| **Go Standards** | Go 1.23+, Value Semantics | No `any`/`interface{}`, wrap errors with `%w`, zero unhandled errors. |
| **Frontend Architecture** | TypeScript Strict | `"strict": true`, zero `any`, Zod/type guards at boundaries. |
| **Asset Bundling** | Single Binary via `embed.FS` | `web/dist` embedded into Go executable; zero Node.js in production. |
| **UI & Styling** | Modern Web / Zero Emojis | Pure components, graceful empty states, zero emojis. |
| **Testing** | Deterministic Pure Core | Table-driven tests in Go with zero mocks for domain. Vitest/Jest for TypeScript. |
