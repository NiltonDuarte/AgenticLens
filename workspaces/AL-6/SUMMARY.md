
# AL-6: Chronological Event Extraction and Schema Evolution

## Summary
Completed implementation of chronological event extraction and schema evolution in Go and TypeScript parity for AgenticLens. All acceptance criteria and technical guidelines satisfied without emojis, `any`, or regressions.

## Changes Completed
1. **Domain Models & Enums (`backend/internal/domain/`)**:
   - Added `EventType` string enum and constants (`EventTypeToolCall`, `EventTypeTurnMessage`, `EventTypeCompaction`).
   - Added `EventFact` struct and optional `Events []EventFact` to `AgenticFactLedger` (`json:"events,omitempty"`).
2. **Boundary Schemas & Repository (`backend/internal/adapters/`)**:
   - Added `EventFactDTO` and conversion functions `ToDomain()` and `EventFactDTOFromDomain()`.
   - Updated `AgenticFactLedgerDTO` and `ExistingLedgerFacts` to deserialize and hold chronological events.
3. **OpenCode SQLite Extraction (`backend/internal/adapters/opencode/`)**:
   - Updated `ExtractFacts` signature to return `([]domain.InvocationFact, []domain.ToolFact, []domain.EventFact, *string, *string, error)`.
   - Extracted assistant `turn_message` events with 1-based turn number and context tokens (`input + cache_read`).
   - Extracted `tool_call` events with execution status, truncated error messages, duration, turn, and context tokens.
   - Extracted `compaction` events correlated with session turns.
   - Sorted events chronologically ascending with stable secondary ordering on turn and session ID.
   - Added `tableHasColumn` fallback to gracefully handle SQLite schemas missing `time_updated`.
4. **CLI & Backfill Migration (`backend/cmd/agentic-lens/`, `backend/internal/adapters/backfill/`)**:
   - CLI `generate` writes extracted events to `agentic_metrics.json`.
   - Backfill engine recognizes ledgers needing event backfill under `--recompute-db`, extracting events and persisting updated schema v2 ledgers.
   - Backfilled all 13 existing workspace ledgers in `workspaces/`.
5. **Frontend Contracts (`web/src/domain/types.ts`)**:
   - Added `EventType`, `EventFact`, and `events?: readonly EventFact[]` matching backend Go models.

## Verification
- `make all`: `go vet`, `npm run typecheck`, `npm run lint`, `go test -race ./...`, `npm test` all passed cleanly.
- `make build`: Full production frontend compilation and Go binary embedding succeeded.
- Zero emojis verified across codebase and generated files.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 1h 48m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 111.0k tokens (56% of 200k healthy ceiling, 2 compactions) | Subagents peak 168.4k (Coder)
- **Role Effort Distribution (Subagents):** Planner 10.6%, Coder 66.8%, Tester 11.4%, Reviewer 11.2%
- **Subagent Turn Velocity & Thrashing:** 114.3 avg turns/run (peak: 283 turns) | Planner 44.0, Coder 283.0, Tester 62.0, Reviewer 68.0
- **Test-to-Production Code Ratio:** 1.6:1 (+766 test lines / +478 prod lines)
- **Code Delivered:** +1,244 / -168 lines across 12 code files (excluding docs)
- **Reasoning Effort:** 144,628 reasoning tokens (51.0% of generated output)
- **Tool Health & Self-Correction:** 539 tool calls (47 self-corrections, 91.3% success rate)

### Agentic Execution Metrics

**Contributors:** nilton.guedes_duarte@digits.schwarz

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 918,041 | 21,407 | 4,764,110 | 83.8% | 111.0k | ELEVATED | $1.18 |
| **planner** | 1 | 414,364 | 17,565 | 1,997,902 | 82.8% | 118.0k | ELEVATED | $0.61 |
| **coder** | 1 | 1,743,025 | 78,746 | 25,140,938 | 93.5% | 168.4k | ELEVATED | $3.81 |
| **tester** | 1 | 356,097 | 16,598 | 3,716,678 | 91.3% | 101.7k | ELEVATED | $0.65 |
| **reviewer** | 1 | 402,920 | 4,640 | 3,683,078 | 90.1% | 101.6k | ELEVATED | $0.64 |
| **Total** | **5** | **3,834,447** | **138,956** | **39,302,706** | **91.1%** | **168.4k** | **ELEVATED** | **$6.89** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 5 | 3,834,447 | 138,956 | 39,302,706 | 91.1% | $6.89 |
| **Total** | **5** | **3,834,447** | **138,956** | **39,302,706** | **91.1%** | **$6.89** |
