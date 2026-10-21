# Code Review Report: Ticket REV-1 Iteration 01

Audit against `alspaces/REV-1/CONTRACT.md`, `alspaces/REV-1/iterations/01/PLAN.md`, and `TECHNICAL_GUIDELINES.md`.

---

## Axis 1: Spec & Contract Alignment

### 1. Completion Date Fallback Elimination
- **Go Domain (`backend/internal/domain/aggregations.go`)**:
  - `GetLedgerCompletionDate`: Fallback cascade to `updated_at` and `started_at` is completely removed. It evaluates `parseTimestamp(ledger.CompletedAt)` exclusively, returning zero `time.Time` and `ok = false` if nil, empty, or unparseable.
  - `GetTicketEffectiveDate`: Stripped of `startedAt` and `updatedAt` fallbacks. Returns `parseISODate(completedAt)`.
  - Temporal aggregators (`FilterLedgersByTimeWindow`, `FilterLedgersByDateRange`, `FilterLedgersByTimeRange`, `SortLedgers`) filter strictly on valid completion dates, ensuring uncompleted ledgers do not skew delivery metrics.
  - Tests in `aggregations_test.go` exhaustively assert nil, empty, whitespace-only, and malformed `completed_at` inputs return `ok = false` even when `started_at` and `updated_at` are populated.
- **TypeScript Domain (`web/src/domain/`)**:
  - `formatters.ts`: `getEffectiveCompletionDateMs` inspects only `item.completed_at`. Zero fallback to `updated_at` or `started_at`.
  - `progression.ts`: `aggregateMetricsOverTime` scopes delivered code strictly to `ledger.completed_at` and invocation metrics to `inv.completed_at ?? ledger.completed_at`. Tickets missing completion timestamps are excluded from completed bucket totals.
  - Tests in `formatters.test.ts` and `progression.test.ts` assert strict null returns and zero bucket allocation when `completed_at` is absent.

### 2. Mandatory `completed_at` in Telemetry Generation
- **Go CLI (`backend/cmd/agentic-lens/main.go`)**:
  - In `runGenerate`, `activeCompletedAt` resolves from explicit CLI flag, cached facts, or invocation timestamps.
  - If nil or whitespace-only, it assigns `updatedAtISO` (current UTC ISO-8601 timestamp).
  - Explicit assertion guards against nil/empty values before constructing `factLedger`, ensuring `factLedger.CompletedAt` is always populated.
  - Verified via CLI tests in `main_test.go` covering explicit flag parsing and whitespace sanitization.

### 3. Session Resolver Tier 4 Deletion & `ErrSessionNotFound` Fast Failure
- **Go Adapters (`backend/internal/adapters/opencode/resolver.go`)**:
  - Exported sentinel `ErrSessionNotFound = errors.New("opencode root session not found")`.
  - Retains Priority 1 (nonce), Priority 2 (explicit session ID), and Priority 3 (session title / workspace path).
  - Priority Tier 4 database query (`SELECT ... FROM message m JOIN session s ... ORDER BY time_created DESC LIMIT 1`) is completely deleted.
  - Unmatched sessions fail fast, returning `nil` and an error wrapping `ErrSessionNotFound` with ticket context.
  - Tests in `extractor_test.go` confirm Tier 4 deletion and assert `errors.Is(err, opencode.ErrSessionNotFound)`.

### 4. Workspaces Directory Resolution Precedence
- **Go Adapters (`backend/internal/adapters/repository.go`)**:
  - `ResolveWorkspacesDir` checks: (1) explicit CLI override, (2) configured project path via `config.Load("")` (`cfg.Workspaces.Dir`), (3) current working directory, (4) parent directory traversal, (5) container volume mount `/app/workspaces`, (6) default string `"workspaces"`.
  - Host filesystem sniffing no longer overrides repository configuration.
  - Tested in `repository_test.go` (`TestResolveWorkspacesDir`).

### 5. Frontend UI Unclassified Badge
- **TypeScript Presentation (`web/src/views/TicketDeepDive.tsx`, `web/src/index.css`)**:
  - Replaced `{ledger.ticket_type ?? 'feature'}` with `{ledger.ticket_type ?? 'unclassified'}`.
  - Sets pill class `ticket-type-pill ${ledger.ticket_type ? '' : 'unclassified'}`.
  - CSS rule `.ticket-type-pill.unclassified` applies neutral background with amber border warning styling.
  - Tested in `TicketDeepDive.test.tsx` verifying absence of silent fallback to `feature`.

### 6. Default Build Version Harmonization
- **Backend & Frontend Sync**:
  - `backend/internal/domain/version.go`: `DefaultVersion = "v0.0.0-dev"`.
  - `web/src/components/Navbar.tsx`: `DEFAULT_BUILD_INFO.version = "v0.0.0-dev"`.
  - Tested in `version_test.go` and `Navbar.test.tsx`.

---

## Axis 2: Engineering Standards & Architecture

### 1. Functional Core, Imperative Shell (FCIS)
- `backend/internal/domain/aggregations.go` imports only standard packages (`fmt`, `math`, `regexp`, `sort`, `strconv`, `strings`, `time`). Zero outer layer imports (`adapters`, `cmd`), zero I/O (`os`, `net/http`, `database/sql`).
- Domain functions are pure and deterministic. Timestamps are parsed, value structs returned. Zero calls to `time.Now()` inside domain packages.

### 2. Typing & Error Wrapping
- Zero `any` or empty `interface{}` across Go backend and TypeScript frontend code.
- Errors wrapped with domain context using `%w`.
- TypeScript strict mode respected; read-only interfaces used across `TicketDateComparable` and component props.

### 3. UI & Hygiene
- Zero emojis across code, tests, CSS, log strings, and UI components.
- Graceful empty states handled across UI and domain filters.
- Scope discipline maintained: no edits outside files targeted by `PLAN.md`.

---

## Verdict

`APPROVED`
