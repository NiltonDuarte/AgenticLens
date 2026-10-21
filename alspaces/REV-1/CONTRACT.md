# Contract: REV-1 Fallback Architecture Remediations and Deterministic System Directives

## 1. Problem Statement

A system-wide audit cataloged in `alspaces/REV-1/REVIEW_REPORT.md` identified architectural risks arising from permissive fallback cascades across the Go backend and TypeScript web frontend. These cascades compromise data determinism, falsify engineering metrics, and risk cross-ticket data contamination:

1. **Temporal Metrics Falsification (Date Fallback Cascades)**:
   In `backend/internal/domain/aggregations.go` (`GetLedgerCompletionDate`, `GetTicketEffectiveDate`), `web/src/domain/formatters.ts` (`getEffectiveCompletionDateMs`), and `web/src/domain/progression.ts`, missing completion dates cascade into `updated_at` or `started_at`. In-progress tickets are consequently treated as completed, distorting lead times, time-window aggregations, and progression trends. Furthermore, telemetry generation in `backend/cmd/agentic-lens/main.go` permits missing or zero `completed_at` timestamps.

2. **Cross-Ticket Telemetry Contamination (Session Resolver Tier 4)**:
   In `backend/internal/adapters/opencode/resolver.go` (`ResolveActiveRootSession`), Priority Tier 4 falls back to the most recent message in SQLite when explicit identifiers are absent. In concurrent multi-agent executions, this captures an arbitrary foreign session and attaches unrelated telemetry, tokens, tool failures, and git diffs to the target ticket.

3. **Silent Frontend Classification Masking**:
   In `web/src/views/TicketDeepDive.tsx`, unclassified tickets (`ledger.ticket_type == null`) are masked with a fallback to `feature`. While backend inference surfaces unclassified tickets cleanly, the UI conceals missing classifications from the user.

4. **Container Path Hijacking**:
   In `backend/internal/adapters/repository.go` (`ResolveWorkspacesDir`), container volume probing for `/app/workspaces` precedes configured workspace paths in `.agenticlens-config.yaml` (`cfg.Workspaces.Dir`), allowing spurious host directories to hijack project configuration.

5. **Platform Build Version Divergence**:
   Backend defaults in `backend/internal/domain/version.go` (`domain.DefaultVersion = "v0.0.0-dev"`) diverge from frontend fallback constants in `web/src/components/Navbar.tsx` (`DEFAULT_BUILD_INFO.version = "v0.1.0-dev"`).

Ticket REV-1 executes the locked user directives to dismantle these fragile fallbacks, enforce fail-fast invariants, and establish deterministic boundaries.

---

## 2. Functional Core, Imperative Shell (FCIS) Architectural Invariants

All deliverables must strictly comply with `TECHNICAL_GUIDELINES.md`:

### Inward Dependency Flow & Pure Functional Core
1. **Go Functional Core (`backend/internal/domain/`)**:
   - `aggregations.go`: `GetLedgerCompletionDate` and `GetTicketEffectiveDate` must remain pure mathematical functions. They receive immutable structs and return parsed UTC timestamps or explicit boolean false indicators.
   - Zero Fallback: Neither function may fall back to `updated_at` or `started_at`. If `completed_at` is nil, empty, or unparseable, the ticket has no completion date.
   - Zero external imports: No imports of `net/http`, `os`, `database/sql`, or filesystem packages.
   - Value semantics: Functions receive value copies, returning newly allocated slices or primitive records.
   - `version.go`: `DefaultVersion` is standardized as the single source of truth for fallback versioning (`v0.0.0-dev`).

2. **TypeScript Domain & Pure Formatters (`web/src/domain/`)**:
   - `formatters.ts`: `getEffectiveCompletionDateMs` must operate as a pure date extraction function inspecting only `completed_at`. If missing, null, or invalid, return null. Zero fallback to `updated_at` or `started_at`.
   - `progression.ts`: Pure temporal aggregators must rely strictly on `completed_at` for delivered code metrics and invocation dates. Ledgers lacking `completed_at` must not be bucketed as completed work.

### Imperative Boundary Shell (Go Backend & TypeScript Web)
1. **Go CLI & Server (`backend/cmd/agentic-lens/`)**:
   - `main.go` (`generateCmd`): Enforces mandatory, non-zero `completed_at` during telemetry generation. If unresolved from flags or invocations, the generator assigns the current finalization timestamp before serializing to disk.
2. **Go Boundary Adapters (`backend/internal/adapters/`)**:
   - `opencode/resolver.go`: Completely deletes Tier 4 message querying. When Tiers 1-3 fail to match a session by nonce, explicit ID, or ticket title/path, it immediately returns typed `ErrSessionNotFound`.
   - `repository.go`: `ResolveWorkspacesDir` prioritizes explicit CLI overrides, followed by configured project paths (`cfg.Workspaces.Dir`), then local working directory paths, before probing `/app/workspaces`.
3. **TypeScript Presentation Shell (`web/src/views/`, `web/src/components/`)**:
   - `web/src/views/TicketDeepDive.tsx`: Renders explicit `unclassified` badge text and neutral warning styling when `ticket_type` is nil or missing.
   - `web/src/components/Navbar.tsx`: Synchronizes default version fallback to `v0.0.0-dev`.

### Strong Typing & Zero Any Invariant
1. Zero `any` or empty `interface{}` across Go backend and TypeScript frontend.
2. Go errors use typed sentinel errors (`ErrSessionNotFound`) wrapped with context.
3. TypeScript props and state use strict read-only types and discriminated unions.

### Zero Emojis Mandate
Zero emojis permitted in Go code, TypeScript code, CSS, comments, test assertions, PR drafts, UI elements, or workspace artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. **Go CLI (`backend/cmd/agentic-lens/main.go`)**:
   - Validate and ensure non-zero `completed_at` in `generateCmd`.
2. **Go Domain (`backend/internal/domain/`)**:
   - Refactor `GetLedgerCompletionDate` in `aggregations.go` to remove fallback to `updated_at` and `started_at`.
   - Refactor `GetTicketEffectiveDate` in `aggregations.go` to remove fallback to `started_at` and `updated_at`.
   - Update `domain.DefaultVersion` in `version.go` to `v0.0.0-dev`.
   - Update domain unit tests in `aggregations_test.go` and `version_test.go`.
3. **Go Adapters (`backend/internal/adapters/`)**:
   - Export `ErrSessionNotFound` in `backend/internal/adapters/opencode/resolver.go`.
   - Delete Tier 4 query from `ResolveActiveRootSession` and return `ErrSessionNotFound` on unresolved sessions.
   - Reorder resolution sequence in `ResolveWorkspacesDir` in `backend/internal/adapters/repository.go` to check `cfg.Workspaces.Dir` before `/app/workspaces`.
   - Update adapter tests in `extractor_test.go` and `repository_test.go`.
4. **TypeScript Domain (`web/src/domain/`)**:
   - Update `getEffectiveCompletionDateMs` in `web/src/domain/formatters.ts` to inspect only `completed_at`.
   - Update `web/src/domain/progression.ts` to eliminate `started_at` and `updated_at` fallback chains in temporal bucketing.
   - Update unit tests in `formatters.test.ts` and `progression.test.ts`.
5. **TypeScript Presentation (`web/src/views/`, `web/src/components/`, `web/src/index.css`)**:
   - Replace `{ledger.ticket_type ?? 'feature'}` with `{ledger.ticket_type ?? 'unclassified'}` and add `.unclassified` badge styling in `TicketDeepDive.tsx` and `index.css`.
   - Update `DEFAULT_BUILD_INFO` in `web/src/components/Navbar.tsx` to `v0.0.0-dev`.
   - Update component tests in `TicketDeepDive.test.tsx` and `Navbar.test.tsx`.

### Out of Scope
1. Modifying SQLite schema or table structures in OpenCode database.
2. Modifying Git commit diff extraction or conventional commit inference logic in `backend/internal/adapters/git/`.
3. Altering REST API endpoint schemas or Server-Sent Events protocols.

---

## 4. Concrete Acceptance Criteria

1. **Mandatory Telemetry Completion Timestamp**:
   - When running `agentic-lens generate`, the output fact ledger and derived metrics ledger always contain a valid, non-zero ISO-8601 `completed_at` string.
2. **Zero Date Fallbacks in Domain Aggregations**:
   - Given a ledger with `completed_at = nil`, `GetLedgerCompletionDate` returns a zero `time.Time` and `ok = false`, even if `updated_at` and `started_at` contain valid ISO timestamps.
   - Given a ledger with `completed_at = nil`, `GetTicketEffectiveDate` returns a zero `time.Time` and `ok = false`.
   - `FilterLedgersByTimeWindow` and `FilterLedgersByDateRange` exclude ledgers with nil or unparseable `completed_at`.
3. **Zero Date Fallbacks in Frontend Utilities**:
   - `getEffectiveCompletionDateMs` returns `null` whenever `completed_at` is null, undefined, empty, or unparseable.
   - Progression trend aggregations in `web/src/domain/progression.ts` do not allocate code metrics or invocations to time buckets using `started_at` or `updated_at` when `completed_at` is absent.
4. **Deterministic Session Resolution & Fast Failure**:
   - `ResolveActiveRootSession` resolves exclusively via Priority 1 (nonce), Priority 2 (explicit session ID), or Priority 3 (ticket ID in title or workspace path).
   - If Tiers 1-3 fail, `ResolveActiveRootSession` returns `nil` and an error wrapping `ErrSessionNotFound`. It never inspects the most recent message in the database.
5. **Visible Unclassified Ticket Status in UI**:
   - In `TicketDeepDive`, a ledger with null or missing `ticket_type` displays the text `unclassified` inside a pill with CSS class `ticket-type-pill unclassified`.
   - The UI never silently renders `feature` for unclassified tickets.
6. **Configured Workspaces Precedence**:
   - In `ResolveWorkspacesDir`, when `.agenticlens-config.yaml` defines `workspaces.dir`, this path is resolved and returned if it exists, even if `/app/workspaces` is also present on the filesystem.
7. **Synchronized Platform Version Defaults**:
   - `domain.DefaultVersion` in Go is `v0.0.0-dev`.
   - `DEFAULT_BUILD_INFO.version` in TypeScript `Navbar.tsx` is `v0.0.0-dev`.

---

## 5. Verification Strategy & Quality Gates

1. **Go Test Suite**:
   - Run `go test -race ./...` across domain, adapters, and cmd packages.
   - Assert all pure domain tests run without mocks, checking boundary edge cases (nil dates, missing fields, corrupted formats).
2. **TypeScript Test Suite**:
   - Run `npm test` across domain, views, components, and formatters.
   - Verify `typecheck` and `lint` pass with zero type errors and zero `any`.
3. **Single Binary Build Verification**:
   - Execute `make check`, `make test`, and `make build`.
   - Confirm compiled binary embeds frontend static assets and executes cleanly.
