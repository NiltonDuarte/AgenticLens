
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
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 1h 50m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 126.4k tokens (63% of 200k healthy ceiling, 2 compactions) | Subagents peak 168.4k (Coder)
- **Role Effort Distribution (Subagents):** Planner 10.6%, Coder 66.8%, Tester 11.4%, Reviewer 11.2%
- **Subagent Turn Velocity & Thrashing:** 114.3 avg turns/run (peak: 283 turns) | Planner 44.0, Coder 283.0, Tester 62.0, Reviewer 68.0
- **Test-to-Production Code Ratio:** 1.6:1 (+766 test lines / +478 prod lines)
- **Code Delivered:** +1,244 / -168 lines across 12 code files (excluding docs)
- **Reasoning Effort:** 147,165 reasoning tokens (50.8% of generated output)
- **Tool Health & Self-Correction:** 563 tool calls (22 self-corrections, 96.1% success rate)

### Anomalies & Health Flags
- **Status:** 6 execution anomalies detected (4 CRITICAL, 2 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (167.2k tokens) — Context tokens reached 167.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+139.6k tokens) — Context tokens surged by 139.6k to 139.6k tokens.
- [CRITICAL] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: mcp-gateway_mcp-exec (100.0%) — Tool 'mcp-gateway_mcp-exec' failed 5 out of 5 calls (100.0% failure rate, threshold: >=50%).
- [CRITICAL] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: task (66.7%) — Tool 'task' failed 4 out of 6 calls (66.7% failure rate, threshold: >=50%).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+61.1k tokens) — Context tokens surged by 61.1k to 61.1k tokens.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: mcp-gateway_mcp-exec (4 consecutive errors) — Tool 'mcp-gateway_mcp-exec' encountered 4 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 1,241,799 | 24,867 | 7,684,979 | 86.1% | 126.4k | ELEVATED | $1.67 |
| **planner** | 1 | 414,364 | 17,565 | 1,997,902 | 82.8% | 118.0k | ELEVATED | $0.61 |
| **coder** | 1 | 1,743,025 | 78,746 | 25,140,938 | 93.5% | 168.4k | ELEVATED | $3.81 |
| **tester** | 1 | 356,097 | 16,598 | 3,716,678 | 91.3% | 101.7k | ELEVATED | $0.65 |
| **reviewer** | 1 | 402,920 | 4,640 | 3,683,078 | 90.1% | 101.6k | ELEVATED | $0.64 |
| **Total** | **5** | **4,158,205** | **142,416** | **42,223,575** | **91.0%** | **168.4k** | **ELEVATED** | **$7.37** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 5 | 4,158,205 | 142,416 | 42,223,575 | 91.0% | $7.37 |
| **Total** | **5** | **4,158,205** | **142,416** | **42,223,575** | **91.0%** | **$7.37** |
