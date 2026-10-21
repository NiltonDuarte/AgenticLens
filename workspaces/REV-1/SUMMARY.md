# Work Summary: REV-1 Fallback Architecture Remediations

## Overview
Dismantled brittle and permissive fallback cascades across AgenticLens, replacing them with deterministic invariants, fail-fast boundary error handling, and strict typing.

---

## Changes by Architectural Layer

### 1. Domain Layer (`backend/internal/domain/`, `web/src/domain/`)
- **Completion Date Invariant (`aggregations.go`)**:
  - Refactored `GetLedgerCompletionDate`: completely eliminated fallback to `updated_at` or `started_at`. Evaluates only `CompletedAt`, returning zero `time.Time` and `ok = false` if missing, empty, or unparseable.
  - Refactored `GetTicketEffectiveDate`: removed all fallbacks to `startedAt` and `updatedAt`.
  - Temporal aggregators (`FilterLedgersByTimeWindow`, `FilterLedgersByDateRange`) filter strictly on valid completion dates, ensuring uncompleted tickets never skew delivery metrics.
  - Expanded table-driven unit tests in `aggregations_test.go` covering corrupted strings, leap years, timezone offsets, and whitespace.
- **Frontend Date Resolution (`web/src/domain/formatters.ts`, `web/src/domain/progression.ts`)**:
  - `getEffectiveCompletionDateMs`: removed candidate array loop (`[completed_at, updated_at, started_at]`); now inspects only `completed_at`.
  - `aggregateMetricsOverTime`: eliminated fallback to `started_at`/`updated_at` in `codeDate` and `invDate` resolution.
  - Updated unit test suites in `formatters.test.ts` and `progression.test.ts`.
- **Version Harmonization (`domain/version.go`)**:
  - Synchronized `DefaultVersion` constant to `"v0.0.0-dev"`.

### 2. Adapters & CLI Layer (`backend/internal/adapters/`, `backend/cmd/`)
- **CLI Telemetry Completion Timestamp Enforcement (`cmd/agentic-lens/main.go`)**:
  - In `runGenerate`, enforced that `activeCompletedAt` resolves to a valid non-empty ISO-8601 string before constructing `domain.AgenticFactLedger`. If omitted or empty, deterministically assigns the finalization UTC ISO-8601 timestamp.
  - Added CLI test in `main_test.go` verifying fact ledgers always contain non-empty `completed_at`.
- **Session Resolver Tier 4 Deletion (`adapters/opencode/resolver.go`)**:
  - Exported typed sentinel error `ErrSessionNotFound = errors.New("opencode root session not found")`.
  - Deleted Tier 4 database query (`SELECT ... FROM message m JOIN session s ... ORDER BY time_created DESC LIMIT 1`).
  - Unmatched sessions immediately return `nil` and an error wrapping `ErrSessionNotFound` with ticket context.
  - Updated adapter tests in `extractor_test.go` verifying that unresolved sessions fail fast.
- **Workspaces Directory Precedence (`adapters/repository.go`)**:
  - Restructured `ResolveWorkspacesDir`: prioritizes explicit override, followed by configured project path (`cfg.Workspaces.Dir`), then current working directory and parent traversals, before checking `/app/workspaces`.
  - Added precedence test in `repository_test.go`.

### 3. Presentation Layer (`web/src/views/`, `web/src/components/`, `web/src/`)
- **UI Unclassified Ticket Badge (`views/TicketDeepDive.tsx`, `index.css`)**:
  - Replaced `{ledger.ticket_type ?? 'feature'}` with `{ledger.ticket_type ?? 'unclassified'}`.
  - Applied `.ticket-type-pill.unclassified` class with neutral background and amber warning border.
  - Added unit test in `TicketDeepDive.test.tsx` verifying absence of silent fallback to `feature`.
- **Default Build Version Synchronization (`components/Navbar.tsx`)**:
  - Standardized `DEFAULT_BUILD_INFO.version` fallback on `'v0.0.0-dev'`.
  - Updated `Navbar.test.tsx`.

---

## Verification Status
- **Go Test Suite (`go test -count=1 -race ./...`)**: PASS (172 top-level tests, 0 data races).
- **Go Vet (`go vet ./...`)**: PASS.
- **Vitest Suite (`npm test`)**: PASS (626 tests across 28 test suites, 0 skipped).
- **TypeScript Typecheck (`npm run typecheck`)**: PASS (0 errors).
- **TypeScript Lint (`npm run lint`)**: PASS (0 warnings).
- **Single Binary Pipeline (`make`)**: PASS (clean multi-platform compilation with embedded production web bundle).

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 6
- **Active Agent Duration:** 2h 18m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 168.3k tokens (84% of 200k healthy ceiling, 2 compactions) | Subagents peak 168.2k (Coder)
- **Role Effort Distribution (Subagents):** Planner 17.0%, Coder 51.1%, Tester 20.3%, Reviewer 11.6%
- **Subagent Turn Velocity & Thrashing:** 90.6 avg turns/run (peak: 187 turns) | Planner 83.0, Coder 118.5, Tester 82.0, Reviewer 51.0
- **Test-to-Production Code Ratio:** 15.1:1 (+801 test lines / +53 prod lines)
- **Code Delivered:** +868 / -148 lines across 24 code files (excluding docs)
- **Reasoning Effort:** 108,371 reasoning tokens (52.2% of generated output)
- **Tool Health & Self-Correction:** 598 tool calls (7 self-corrections, 98.8% success rate)

### Anomalies & Health Flags
- **Status:** 5 execution anomalies detected (3 CRITICAL, 2 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k in turn 79 (session: ses_ea1a4b28affeowyXAPpI4V07XM), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k in turn 158 (session: ses_ea1699b13ffeeKDYzCAellSCs3), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+97.6k tokens) — Context tokens surged by 97.6k tokens in turn 171 (session: ses_ea1699b13ffeeKDYzCAellSCs3, total: 97.6k tokens, threshold: 80000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+55.5k tokens) — Context tokens surged by 55.5k tokens in turn 89 (session: ses_ea1a4b28affeowyXAPpI4V07XM, total: 55.5k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+64.3k tokens) — Context tokens surged by 64.3k tokens in turn 109 (session: ses_ea1a4b28affeowyXAPpI4V07XM, total: 64.3k tokens, threshold: 40000)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 1,855,376 | 31,860 | 10,753,970 | 85.3% | 168.3k | ELEVATED | $2.39 |
| **planner** | 1 | 422,695 | 15,576 | 4,796,211 | 91.9% | 119.8k | ELEVATED | $0.83 |
| **coder** | 2 | 1,247,918 | 31,398 | 17,646,230 | 93.4% | 168.2k | ELEVATED | $2.50 |
| **tester** | 1 | 434,993 | 16,034 | 7,132,878 | 94.3% | 127.9k | ELEVATED | $1.00 |
| **reviewer** | 1 | 463,899 | 4,540 | 2,155,059 | 82.3% | 77.8k | HEALTHY | $0.57 |
| **Total** | **6** | **4,424,881** | **99,408** | **42,484,348** | **90.6%** | **168.3k** | **ELEVATED** | **$7.28** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 6 | 4,424,881 | 99,408 | 42,484,348 | 90.6% | $7.28 |
| **Total** | **6** | **4,424,881** | **99,408** | **42,484,348** | **90.6%** | **$7.28** |
