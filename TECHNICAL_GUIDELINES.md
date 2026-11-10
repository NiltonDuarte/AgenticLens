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

### Portable Ledger Contract

AgenticLens is multi-developer, multi-machine and multi-harness by design. Versioned ledgers in the configured workspace are trusted, durable project records intended to be committed, shared and reviewed through Git. A supported, valid ledger remains usable when the developer, machine, harness, checkout path or availability of its original sources changes.

**Source of truth, per session: local harness database/session store -> existing ledger.** During generation, update or backfill, successfully extracted facts from a session available on the current machine are authoritative for that session's recorded values. When its source session is unavailable, preserve the existing ledger's record. The local harness source means OpenCode's database or the equivalent Codex rollout store; a capture registry, cache or receipt is not a higher authority over shared history.

**Preserve every session outside a successful complete local refresh. Replace refreshed sessions from their authoritative local snapshots. Never duplicate records or discard unavailable history.**

**Explicit migration exceptions:** The single schema 5-to-6 transition drops ownerless legacy skill/task rows with separate audited counts, projects owned lean skills/delegations and drops retired event, provenance, outcome and opaque child-reference payloads with loss accounting. Preserve proven child IDs and recorded summaries. These explicit transformations do not permit discarding known owners or their unavailable lean history, fabricating ownership, or migrating/deleting facts during ordinary reads. See the owning [session-owned occurrence schema contract](docs/OWNED_OCCURRENCES.md) for the migration, backup and current-schema read boundary.

- **Persisted records and extraction have different responsibilities.** Original session logs are needed to extract new facts or recover missing details. Preserving, aggregating and presenting already recorded facts does not require re-extracting or independently re-verifying their original sources. Keep recorded source/owner identity and uncertainty; do not claim a fresh verification occurred merely because a ledger was loaded.
- **Local state is operational state.** Capture databases, caches, receipts, installation identities and absolute source paths must not become prerequisites for accepting otherwise valid historical ledger facts. A receiving developer does not need to copy the producing machine's private database. Git transports the recorded facts; it does not recreate missing traces or prove a new source observation.
- **Session ID is the stable identity.** Match a stored session to its exact native session ID in the selected local harness source. A locally present, successfully read session is trusted; completeness determines whether its snapshot can replace the stored snapshot. Do not require an additional machine, checkout, developer, receipt or fingerprint identity to establish the match. Preserve recorded source labels without inventing a new identity scheme.
- **Refresh has an owner boundary.** Refresh only the owners successfully extracted from their selected sources. Preserve unavailable owners, other developers' sessions, other harnesses, legacy history outside the explicit migration exception above, and legitimate multiplicity, including when a local extraction succeeds only partially. Missing sources are not measured zero or permission to delete history.
- **A complete local snapshot replaces its session.** Successful complete extraction replaces that exact session's stored snapshot, including when it contains fewer or zero invocations. Do not retain stale stored-only invocations by unioning them into the refreshed snapshot. Preserve every session outside that complete refresh set. Missing, failed or partial source reads do not establish a complete snapshot or an authoritative empty result.
- **Local authority can reduce recorded totals.** An older restored local database/session store can still yield a complete authoritative snapshot. Replacing that session may reduce invocation counts and derived totals. Report that replacement and its resulting counts; do not silently override local authority with a stored snapshot merely because its totals are higher or its recorded timestamp is newer.
- **Updates preserve data and are idempotent.** Outside the authoritative session replacement above, commands must not discard sessions/invocations as a side effect of generation, backfill or migration. Never duplicate records: reconcile updates by established identity rather than appending another copy of a cumulative snapshot. Preserve distinct genuine occurrences; unavailable or partial source data is not a deletion instruction.
- **Aggregation covers the resulting ledger.** During generation/update, the platform runtime recomputes derived totals from the full resulting ledger: refreshed local sessions together with all preserved sessions/invocations. Do not aggregate only the current machine's extraction, add a new cumulative total to an old one, or require an agent to maintain aggregate values manually.
- **Cross-ticket session overlap is a user decision.** The same session ID recorded in multiple tickets must not trigger automatic deletion, reassignment or cross-ticket deduplication. Expose the overlap and affected tickets so the user can decide how to correct their records. Per-session snapshot replacement and within-ledger idempotence do not authorize silently changing ticket attribution or aggregate allocation.
- **Migration is best-effort and resumable.** Apply supported schema transformations without losing recorded data, subject only to the explicit schema migration exceptions above. A hard blocker leaves an explicit, durable mid-migration state with the remaining work/diagnostic so another developer or machine can continue. Preserve facts requiring unavailable source enrichment and do not mark incomplete work complete. A failure must not leave a truncated ledger or erase the prior record.
- **Representation is not authority.** Whitespace formatting, supported schema migration and moving a supported ledger between checkouts must not invalidate recorded facts merely by changing file bytes or paths. Checksums may protect downloads, detect corruption, coordinate writes or select recovery artifacts; an exact file fingerprint must not become a machine-local provenance gate for portable persisted facts.
- **Validation still applies.** Validate supported schema, required fields, identities, ownership consistency and safe file access. Report malformed/unsupported data and unavailable enrichment explicitly. Preserve supported recorded fields when an enrichment step is pending; do not invent missing values, attribute ambiguous history to an owner, bypass required schema transitions or advertise unsupported runtime ingestion.

Apply the [portable-ledger acceptance scenarios](docs/REGRESSION_PREVENTION.md#portable-ledgers-across-developers-and-machines) to changes affecting persistence or refresh. Before introducing an additional rejection or authenticity rule, identify its approved product requirement and assess those scenarios. An existing implementation or test assertion alone does not establish such a requirement. A proposed change to this contract requires an explicit product decision; routine hardening or a review finding cannot silently supersede it.

### Fact Fidelity and Refresh Boundaries

- Required identity and metadata are validated before persistence. Optional or unknown values remain explicit; do not invent completion dates from started/updated dates, assign feature types to unclassified tickets, or represent unavailable measurements as measured zero.
- Resolve stored sessions by exact identity by default. Heuristic recovery must be explicitly enabled and must not silently overwrite unrelated stored facts.
- Refresh each source-owned session/tree only after discovery, parsing, and grouping succeed. Preserve all roots, children, legitimate repeated invocation rows, unavailable sessions, and other-harness facts. Reconcile refreshed cumulative snapshots by established identity rather than adding their totals to prior snapshots; repeated refresh of unchanged input must be idempotent.
- Source events are transient calculation inputs and trace responses, separate from canonical SessionFact. Current ledgers store no event stream or partial-event marker. Hydration availability distinguishes complete (including successfully empty), partial and unavailable; lean occurrences never synthesize trace rows. See [owned occurrences](docs/OWNED_OCCURRENCES.md).
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

## 5. Native Manager, Engine & Static Asset Bundling

AgenticLens distributes two native executables: a small `agentic-lens` manager for installation, updates and project engine selection, and `agentic-lens-engine` for telemetry, capture, analysis and the dashboard. The engine embeds the TypeScript web application. The manager never imports the engine CLI or embedded dashboard, and the engine never dispatches to another engine.

### Rules
1. **Embedding Mandate (`embed.FS`)**:
   - The compiled static frontend output (`web/dist/`) must be bundled into Go via `embed.FS` (e.g. `//go:embed all:dist` in `internal/adapters/web/` or `cmd/agentic-lens/`).
   - The Go HTTP server (`cmd/agentic-lens server`) serves the SPA static files alongside the JSON/SSE API endpoints.
   - Zero Node.js runtime dependency in production or container deployments.
2. **Dual-Mode Serving**:
   - Production / Standalone: Serves embedded static assets from `embed.FS` with client-side fallback routing for SPA navigation.
   - Local Development: Vite dev server runs with HMR and proxies API requests (`/api/*`) to the Go backend.
3. **Build Sequence**:
   - Build the manager independently with `(cd backend && go build -o ../bin/agentic-lens ./cmd/agentic-lens-manager)`. Build the engine after `(cd web && npm run build)` with `(cd backend && go build -o ../bin/agentic-lens-engine ./cmd/agentic-lens)`. `make build` emits platform-stamped names for both artifacts.
   - The frontend build copies assets to `backend/internal/adapters/web/dist/`, the Go embed input. Verify that tree after frontend changes. Respect its Git tracking policy; do not add ignored build output.
   - Engine CI/container builds compile the frontend before the engine; the manager builds without frontend assets.

---

## 6. Quality Checks & Verification

Before submitting any iteration or PR, all quality gates must pass:

### Automated Verification (Makefile)

Run `make test-unit` for pure-core feedback and `make` for the full gate. Repository integration targets require successful Go and frontend units before setup/execution, including parallel Make and hosted CI dependencies. Standalone npm integration requires the complete frontend unit lane first. See the canonical [contributor build targets](CONTRIBUTING.md#4-makefile-targets-reference) and [verification prerequisites](CONTRIBUTING.md#5-quality-standards-and-invariants) for disjoint lane classification, race coverage and command applicability. Embed-dependent Go probes require `make web-assets`; pure domain units do not. Raw Go/Vitest/Python commands are diagnostic probes without the public entrypoints' gating guarantee. Frontend checks also include `npm run typecheck`, `npm run lint`, and `npm run build`.

### Test Purpose and Proportionality

Prefer table-driven Go tests and Vitest/Jest for TypeScript.

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
- Verify manager/engine separation and self-contained engine execution without a Node runtime dependency.

