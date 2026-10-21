# Ticket AL-8: Anomaly API Endpoints & Fleet Health Aggregation

## Summary of Completed Work
- **Domain Layer**:
  - Implemented pure functional aggregation `ComputeFleetAnomalies(ledgers []AgenticFactLedger) FleetAnomalySummary` in `backend/internal/domain/anomalies.go`. Aggregates fleet-level anomaly totals, severity counts, rule counts, ticket counts, anomaly rates, and stable-sorted `BrittleTools`.
  - Added immutable value structs `BrittleToolSummary` and `FleetAnomalySummary` to `backend/internal/domain/models.go`. Extended `FleetSummary` with `AnomalySummary FleetAnomalySummary`.
  - Integrated `AnomalySummary` calculation into `CalculateFleetSummary` in `backend/internal/domain/aggregations.go`.
- **Adapters & HTTP Layer**:
  - Added boundary DTOs `BrittleToolSummaryDTO` and `FleetAnomalySummaryDTO` with bidirectional mappings in `backend/internal/adapters/schemas.go`.
  - Added `Anomalies` (always initialized empty array `[]` when empty, never omitted or null) to `WorkspaceDetailResponse` in `backend/internal/adapters/http/dto.go`.
  - Registered `GET /api/workspaces/{ticketId}/events` in `backend/internal/adapters/http/server.go` and implemented `WorkspaceEventsHandler` returning chronologically sorted `[]EventFact` (empty array for legacy v1 workspaces, 404 for missing workspaces, 400 for path traversal).
  - Extended `GET /api/fleet` with `anomaly_summary`.
- **Presentation Layer (TypeScript)**:
  - Extended `FleetSummary` and `WorkspaceDetail` in `web/src/domain/types.ts` with `anomaly_summary` and `anomalies: readonly Anomaly[]`.
  - Added `fetchWorkspaceEvents` in `web/src/api/client.ts` with network fallback and error handling.
  - Updated mock data and test suites across components and views.
- **Verification & Quality**:
  - Pure domain unit tests in `backend/internal/domain/anomalies_test.go` with zero mocks.
  - Boundary schema tests in `backend/internal/adapters/schemas_test.go`.
  - HTTP integration tests in `backend/internal/adapters/http/server_test.go`.
  - Frontend typechecks and vitest suites (163/163 passing, 0 skipped).

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 1h 22m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 103.1k tokens (52% of 200k healthy ceiling, 2 compactions) | Subagents peak 173.6k (Coder)
- **Role Effort Distribution (Subagents):** Planner 12.2%, Coder 53.2%, Tester 17.3%, Reviewer 17.3%
- **Subagent Turn Velocity & Thrashing:** 87.0 avg turns/run (peak: 184 turns) | Planner 46.0, Coder 184.0, Tester 48.0, Reviewer 70.0
- **Test-to-Production Code Ratio:** 2.1:1 (+921 test lines / +442 prod lines)
- **Code Delivered:** +1,363 / -123 lines across 19 code files (excluding docs)
- **Reasoning Effort:** 85,455 reasoning tokens (48.0% of generated output)
- **Tool Health & Self-Correction:** 433 tool calls (4 self-corrections, 99.1% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (3 CRITICAL).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (171.1k tokens) — Context tokens reached 171.1k in turn 22 (session: ses_eb0cb2d04ffe4fpKyMAfsZfFiG), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (164.3k tokens) — Context tokens reached 164.3k in turn 100 (session: ses_eb0ddd134ffeADvio7rR7tnsrq), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+82.9k tokens) — Context tokens surged by 82.9k tokens in turn 104 (session: ses_eb0ddd134ffeADvio7rR7tnsrq, total: 82.9k tokens, threshold: 80000)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 800,716 | 19,526 | 4,619,609 | 85.2% | 103.1k | ELEVATED | $1.06 |
| **planner** | 1 | 261,639 | 11,241 | 1,884,097 | 87.8% | 90.0k | HEALTHY | $0.45 |
| **coder** | 1 | 869,762 | 50,137 | 13,099,070 | 93.8% | 173.6k | ELEVATED | $1.94 |
| **tester** | 1 | 428,761 | 6,849 | 2,978,565 | 87.4% | 171.1k | ELEVATED | $0.63 |
| **reviewer** | 1 | 382,323 | 4,898 | 3,857,992 | 91.0% | 89.0k | HEALTHY | $0.63 |
| **Total** | **5** | **2,743,201** | **92,651** | **26,439,333** | **90.6%** | **173.6k** | **ELEVATED** | **$4.71** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 5 | 2,743,201 | 92,651 | 26,439,333 | 90.6% | $4.71 |
| **Total** | **5** | **2,743,201** | **92,651** | **26,439,333** | **90.6%** | **$4.71** |
