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
- **Software Development Life Cycle Iterations:** 7
- **Active Agent Duration:** 6h 53m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 168.3k tokens (84% of 200k healthy ceiling, 4 compactions) | Subagents peak 168.2k (Coder)
- **Role Effort Distribution (Subagents):** Planner 12.3%, Coder 62.9%, Tester 15.2%, Reviewer 9.7%
- **Subagent Turn Velocity & Thrashing:** 88.7 avg turns/run (peak: 187 turns) | Planner 61.0, Coder 134.0, Tester 64.0, Reviewer 50.5
- **Test-to-Production Code Ratio:** 10.0:1 (+1,111 test lines / +111 prod lines)
- **Code Delivered:** +2,612 / -272 lines across 38 code files (excluding docs)
- **Reasoning Effort:** 237,777 reasoning tokens (58.4% of generated output)
- **Tool Health & Self-Correction:** 1,062 tool calls (12 self-corrections, 98.9% success rate)

### Anomalies & Health Flags
- **Status:** 9 execution anomalies detected (5 CRITICAL, 4 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.2k tokens) — Context tokens reached 161.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.1k tokens) — Context tokens reached 162.1k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+97.6k tokens) — Context tokens surged by 97.6k to 97.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+40.5k tokens) — Context tokens surged by 40.5k to 40.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+55.5k tokens) — Context tokens surged by 55.5k to 55.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+64.3k tokens) — Context tokens surged by 64.3k to 64.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+57.0k tokens) — Context tokens surged by 57.0k to 57.0k tokens.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 3,080,880 | 42,218 | 14,244,109 | 82.2% | 168.3k | ELEVATED | $3.63 |
| **planner** | 2 | 628,193 | 20,735 | 6,062,294 | 90.6% | 119.8k | ELEVATED | $1.17 |
| **coder** | 4 | 2,758,457 | 77,103 | 42,481,867 | 93.9% | 168.2k | ELEVATED | $5.98 |
| **tester** | 2 | 657,709 | 20,869 | 10,046,635 | 93.9% | 127.9k | ELEVATED | $1.44 |
| **reviewer** | 2 | 668,444 | 8,463 | 4,042,064 | 85.8% | 77.8k | HEALTHY | $0.92 |
| **Total** | **11** | **7,793,683** | **169,388** | **76,876,969** | **90.8%** | **168.3k** | **ELEVATED** | **$13.14** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 11 | 7,793,683 | 169,388 | 76,876,969 | 90.8% | $13.14 |
| **Total** | **11** | **7,793,683** | **169,388** | **76,876,969** | **90.8%** | **$13.14** |
