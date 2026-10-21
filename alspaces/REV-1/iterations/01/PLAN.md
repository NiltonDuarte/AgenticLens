# Implementation Plan: REV-1 Fallback Architecture Remediations and Deterministic System Directives

## 1. Goal and Iteration Scope

Deliver the vertical slice resolving all 5 locked fallback remediations across the Go backend and TypeScript frontend:
- Enforce mandatory `completed_at` in `cmd/agentic-lens/main.go`.
- Remove date fallback chains in `aggregations.go`, `formatters.ts`, and `progression.ts`.
- Delete Tier 4 database fallback in `resolver.go` and return `ErrSessionNotFound`.
- Prioritize configured `cfg.Workspaces.Dir` over `/app/workspaces` in `repository.go`.
- Eliminate UI fallback to `feature` in `TicketDeepDive.tsx` and style `unclassified`.
- Harmonize default version strings on `v0.0.0-dev` across backend and frontend.

---

## 2. Grounded Reference Architecture & File Anchors

The implementation directly touches the following packages and symbols:

### Go Functional Core (`backend/internal/domain/`)
- `backend/internal/domain/aggregations.go`:
  - `GetLedgerCompletionDate(ledger AgenticMetricsLedger) (time.Time, bool)`
  - `GetTicketEffectiveDate(completedAt, startedAt *string, updatedAt string) (time.Time, bool)`
- `backend/internal/domain/aggregations_test.go`:
  - `TestGetLedgerCompletionDate`
  - `TestFilterLedgersByTimeWindow`
  - `TestFilterLedgersByDateRange`
- `backend/internal/domain/version.go`:
  - `DefaultVersion` constant
- `backend/internal/domain/version_test.go`:
  - Version string formatting assertions

### Go Boundary Adapters & CLI (`backend/internal/adapters/`, `backend/cmd/`)
- `backend/cmd/agentic-lens/main.go`:
  - `generateCmd` execution flow
  - Timestamp resolution and `factLedger` construction
- `backend/internal/adapters/opencode/resolver.go`:
  - `ErrSessionNotFound` sentinel error
  - `ResolveActiveRootSession(db *sql.DB, nonce string, ticketID string, explicitSessionID string) (*string, error)`
- `backend/internal/adapters/opencode/extractor_test.go`:
  - `TestResolveActiveRootSession`
- `backend/internal/adapters/repository.go`:
  - `ResolveWorkspacesDir(overridePath string) string`
- `backend/internal/adapters/repository_test.go`:
  - `TestResolveWorkspacesDir`

### TypeScript Domain Layer (`web/src/domain/`)
- `web/src/domain/formatters.ts`:
  - `getEffectiveCompletionDateMs(item: TicketDateComparable): number | null`
- `web/src/domain/formatters.test.ts`:
  - Completion date extraction test suite
- `web/src/domain/progression.ts`:
  - `aggregateMetricsOverTime`
  - Code metrics and invocation date resolution
- `web/src/domain/progression.test.ts`:
  - Progression temporal aggregation test suite

### TypeScript Presentation Layer (`web/src/views/`, `web/src/components/`, `web/src/`)
- `web/src/views/TicketDeepDive.tsx`:
  - Workspace header ticket type pill rendering
- `web/src/views/TicketDeepDive.test.tsx`:
  - Header pill presentation test suite
- `web/src/index.css`:
  - `.ticket-type-pill.unclassified` CSS rule
- `web/src/components/Navbar.tsx`:
  - `DEFAULT_BUILD_INFO` constant
- `web/src/components/Navbar.test.tsx`:
  - Navbar default version rendering test suite

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Enforce Mandatory `completed_at` in Telemetry Generation (Go CLI)
- Task 1.1: In `backend/cmd/agentic-lens/main.go` inside `runGenerate`:
  - Locate active timestamp resolution for `activeCompletedAt`.
  - After evaluating `completedAtFlag`, `cachedFacts`, and `invocations`, verify whether `activeCompletedAt` is nil or empty.
  - If nil or whitespace-only, assign the current UTC ISO-8601 timestamp string (`updatedAtISO`).
  - Prior to constructing `factLedger`, assert that `activeCompletedAt` is non-nil and non-empty.
  - Ensure `factLedger.CompletedAt` is populated with this validated timestamp.

### Phase 2: Dismantle Date Fallback Chains in Go Functional Core (Go Domain)
- Task 2.1: In `backend/internal/domain/aggregations.go`:
  - Refactor `GetLedgerCompletionDate(ledger AgenticMetricsLedger) (time.Time, bool)`:
    - Attempt to parse `ledger.CompletedAt` using `parseTimestamp`.
    - If valid, return the parsed timestamp and true.
    - If `ledger.CompletedAt` is nil, empty, or unparseable, return zero `time.Time` and false.
    - Remove all fallback evaluations of `ledger.UpdatedAt` and `ledger.StartedAt`.
  - Refactor `GetTicketEffectiveDate(completedAt, startedAt *string, updatedAt string) (time.Time, bool)`:
    - Attempt to parse `completedAt` using `parseISODate`.
    - If valid, return the parsed timestamp and true.
    - If `completedAt` is nil, empty, or unparseable, return zero `time.Time` and false.
    - Remove all fallback evaluations of `startedAt` and `updatedAt`.
  - Verify that callers (`FilterLedgersByDateRange`, `FilterLedgersByTimeWindow`, `SortLedgers`) remain pure functions with zero mutations.
- Task 2.2: In `backend/internal/domain/aggregations_test.go`:
  - Update `TestGetLedgerCompletionDate`:
    - Remove assertions expecting fallback to `updated_at` or `started_at` when `completed_at` is nil.
    - Add tests asserting that a ledger with nil `completed_at` returns `ok = false` and zero time, even if `updated_at` or `started_at` is set.
    - Retain tests validating valid RFC3339 timestamps, date-only strings, nano precision, and timezone offsets in `completed_at`.
  - Update test fixtures in `TestFilterLedgersByTimeWindow` and `TestFilterLedgersByDateRange`:
    - Add explicit `CompletedAt` timestamps to test ledgers representing completed tickets so that time-window filtering logic tests valid date filtering without relying on deprecated fallbacks.

### Phase 3: Drop Session Resolver Tier 4 and Fail Fast (Go Adapters)
- Task 3.1: In `backend/internal/adapters/opencode/resolver.go`:
  - Define and export `var ErrSessionNotFound = errors.New("opencode root session not found")`.
  - In `ResolveActiveRootSession(db *sql.DB, nonce string, ticketID string, explicitSessionID string) (*string, error)`:
    - Retain Priority 1 (exact nonce matching against `part` table).
    - Retain Priority 2 (explicit session ID parameter).
    - Retain Priority 3 (ticket ID matching session title, followed by `ResolveSessionByWorkspacePath`).
    - Completely delete the Priority 4 query (`SELECT s.id, s.parent_id FROM message m JOIN session s ON m.session_id = s.id ORDER BY m.time_created DESC LIMIT 1`).
    - When Tiers 1-3 fail to find a session, return `nil, fmt.Errorf("%w: unable to resolve OpenCode session for ticket %q; specify --session-id or use --nonce", ErrSessionNotFound, cleanTicket)`.
- Task 3.2: In `backend/internal/adapters/opencode/extractor_test.go`:
  - Update `TestResolveActiveRootSession`:
    - Delete the "Priority 4 Latest Message Fallback" subtest or rewrite it to assert that when neither nonce nor ticket ID is supplied, `ResolveActiveRootSession` returns an error wrapping `opencode.ErrSessionNotFound`.
    - Update the "Unresolved Ticket Returns Nil" subtest to assert that `errors.Is(err, opencode.ErrSessionNotFound)` evaluates to true.

### Phase 4: Workspaces Directory Resolution Precedence (Go Adapters)
- Task 4.1: In `backend/internal/adapters/repository.go` inside `ResolveWorkspacesDir(overridePath string) string`:
  - Reorder directory resolution strategies:
    1. Check explicit `overridePath`: if non-empty, return it immediately.
    2. Check configured workspace directory: call `config.Load("")` and `config.FindConfigFile("")`. If `cfg.Workspaces.Dir` is populated, resolve candidate path relative to config directory or working directory. If `os.Stat` confirms the candidate exists and is a directory, return it.
    3. Check current working directory: candidate `./<wsDir>`. If it exists and is a directory, return it.
    4. Check parent directory traversal for `<wsDir>`. If found, return candidate.
    5. Check container volume mount: probe `/app/workspaces`. If it exists and is a directory, return `"/app/workspaces"`.
    6. Default fallback: return `wsDir` string.
- Task 4.2: In `backend/internal/adapters/repository_test.go`:
  - Add a unit test verifying that when `cfg.Workspaces.Dir` points to an existing directory, `ResolveWorkspacesDir` returns that configured directory even if `/app/workspaces` were present.

### Phase 5: Standardize Default Build Version in Go (Go Domain)
- Task 5.1: In `backend/internal/domain/version.go`:
  - Update constant `DefaultVersion = "v0.0.0-dev"`.
- Task 5.2: In `backend/internal/domain/version_test.go`:
  - Update the expected string in line 109 from `v0.1.0-dev` to `v0.0.0-dev`.

### Phase 6: Dismantle Date Fallback Cascades in Frontend (TypeScript Domain)
- Task 6.1: In `web/src/domain/formatters.ts`:
  - Refactor `getEffectiveCompletionDateMs(item: TicketDateComparable): number | null`:
    - Remove the candidate array loop (`[item.completed_at, item.updated_at, item.started_at]`).
    - Check only `item.completed_at`. If string is present, trim and parse via `Date.parse`.
    - If valid number, return milliseconds.
    - Otherwise, return `null`.
- Task 6.2: In `web/src/domain/formatters.test.ts`:
  - Update tests for `getEffectiveCompletionDateMs`:
    - Replace tests expecting fallback to `updated_at` or `started_at` with assertions verifying that when `completed_at` is null or undefined, the return value is strictly `null`.
- Task 6.3: In `web/src/domain/progression.ts`:
  - In `aggregateMetricsOverTime`:
    - Refactor `codeDate` resolution: evaluate `parseIsoDateToUtc(ledger.completed_at)`. Do not fall back to `ledger.started_at` or `ledger.updated_at`.
    - Refactor `invDate` resolution: evaluate `parseIsoDateToUtc(inv.completed_at) ?? parseIsoDateToUtc(ledger.completed_at)`. Do not fall back to `started_at` or `updated_at`.
- Task 6.4: In `web/src/domain/progression.test.ts`:
  - Update test cases verifying fallback behavior to align with strict `completed_at` requirement.

### Phase 7: Eliminate UI Silent Fallback to 'feature' (TypeScript Presentation)
- Task 7.1: In `web/src/views/TicketDeepDive.tsx`:
  - In line 221, replace `{ledger.ticket_type ?? 'feature'}` with `{ledger.ticket_type ?? 'unclassified'}`.
  - Dynamically assign class name: `ticket-type-pill ${ledger.ticket_type ? '' : 'unclassified'}`.
- Task 7.2: In `web/src/index.css`:
  - Add CSS class `.ticket-type-pill.unclassified` styling the badge with neutral background, subtle warning border, and muted text.
- Task 7.3: In `web/src/views/TicketDeepDive.test.tsx`:
  - Add a unit test asserting that when `ledger.ticket_type` is undefined or null, the component renders `unclassified` with class `unclassified` and does not render `feature`.

### Phase 8: Standardize Default Build Version in Frontend (TypeScript Presentation)
- Task 8.1: In `web/src/components/Navbar.tsx`:
  - Update `DEFAULT_BUILD_INFO.version` from `v0.1.0-dev` to `v0.0.0-dev`.
- Task 8.2: In `web/src/components/Navbar.test.tsx`:
  - Update test assertions expecting `v0.1.0-dev` fallback to expect `v0.0.0-dev`.

---

## 4. Verification and Adversarial Testing Tasks for Tester

### Pure Go Domain Testing
- Execute `go test -v -race ./backend/internal/domain/...`
- Adversarial tests for `GetLedgerCompletionDate`:
  - Ledger with missing `completed_at`, populated `started_at` and `updated_at` returns `ok = false`.
  - Ledger with whitespace-only `completed_at` returns `ok = false`.
  - Ledger with corrupt `completed_at` returns `ok = false`.
  - Ledger with valid UTC, non-UTC timezone offsets, and leap years parses accurately.
- Adversarial tests for `FilterLedgersByTimeWindow`:
  - Tickets without `completed_at` are excluded from all bounded windows (`1d`, `7d`, `30d`, `90d`).
- Assert zero mocks used in domain tests.

### Go Adapter Testing
- Execute `go test -v -race ./backend/internal/adapters/...`
- Assert `ResolveActiveRootSession` returns `ErrSessionNotFound` when session cannot be identified.
- Assert `ResolveWorkspacesDir` chooses configured path before checking `/app/workspaces`.

### TypeScript Unit & Component Testing
- Execute `npm test` inside `web/`.
- Test `getEffectiveCompletionDateMs` returns null when `completed_at` is absent, regardless of `started_at`/`updated_at`.
- Test `TicketDeepDive` renders `unclassified` badge when `ticket_type` is missing.
- Test `Navbar` displays `v0.0.0-dev` when version endpoint is unavailable.
- Run `npm run typecheck` and `npm run lint`.

### Full System Build Gate
- Execute root `make` (`make check`, `make test`, `make build`).
- Verify binary runs cleanly and embeds web assets.

---

## 5. Audit Criteria for Reviewer

### Axis 1: Spec & Contract Alignment
- Are completion date fallback chains completely eliminated in Go domain and TypeScript domain?
- Does `agentic-lens generate` enforce non-zero `completed_at`?
- Is Tier 4 completely deleted from `resolver.go` with fast failure on `ErrSessionNotFound`?
- Does `TicketDeepDive.tsx` display `unclassified` with proper badge styling instead of `feature`?
- Does `ResolveWorkspacesDir` prioritize `cfg.Workspaces.Dir` over `/app/workspaces`?
- Are backend and frontend default version strings harmonized on `v0.0.0-dev`?

### Axis 2: Engineering Standards (TECHNICAL_GUIDELINES.md)
- Go FCIS: Are domain aggregations pure functions without I/O, clocks, or external imports?
- Strong Typing: Zero `any` or empty `interface{}` across Go and TypeScript.
- Error Handling: Is `ErrSessionNotFound` a typed sentinel error wrapped with descriptive context?
- Zero Emojis: Are there zero emojis across all code, tests, CSS, and documentation?
- Clean Working Tree: Build artifacts and tests pass without uncommitted artifacts.
