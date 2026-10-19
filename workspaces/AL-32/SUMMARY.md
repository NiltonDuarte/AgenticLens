# AL-32: Telemetry Event Pruning, Database Message Stripping, and Schema V3 Migration

## Overview
Implemented pure telemetry event pruning and stripped error message strings for schema version 3, ensuring persisted fact ledgers in `workspaces/*/agentic_metrics.json` remain compact (<5 KB) and leak-free while retaining full in-memory derivation fidelity for all metrics and health anomalies.

## Work Completed by Layer

### 1. Domain Core (`backend/internal/domain/`)
- Promoted `CurrentSchemaVersion` from `2` to `3` in `backend/internal/domain/constants.go`.
- Defined `MigrationStepSanitizedEvents = "schema_v3_sanitized"`.
- Implemented `PruneEventsForPersistence(ledger AgenticFactLedger) []EventFact` in `backend/internal/domain/events.go`:
  - Retains bare failed tool calls (`Status == "error"`) with `ToolName`, `Status`, `Turn`, `Timestamp`, `DurationMs`, `ContextTokens`, and `ErrorMessage: nil`.
  - Retains compaction events with `ErrorMessage: nil`.
  - Retains anomaly-associated events with `ErrorMessage: nil`.
  - Omits routine completed turn messages and routine successful tool calls.
  - Stably sorts remaining events via `SortEvents`.
- Updated `detectToolHighFailureRates` in `backend/internal/domain/anomalies.go` to evaluate `ledger.Tools` first, preventing pruned event slices from skewing volume or failure rates.
- Implemented `StepV2ToV3Sanitized` in `backend/internal/domain/migration/step_v2_v3.go`.
- Registered `StepV2ToV3Sanitized` in `NewEngine` registry in `backend/internal/domain/migration/registry.go`.
- Updated `InvariantValidator` in `backend/internal/domain/migration/validator.go` with target version 3 assertions (identity preservation, bit-identical derived totals, `len(after.Events) <= len(before.Events)`, and zero non-nil error messages).

### 2. Boundary Adapters (`backend/internal/adapters/`)
- Removed `sanitizeTelemetryText` and `emailPattern` regex from OpenCode SQLite extractor in `backend/internal/adapters/opencode/extractor.go`, recording bare `ErrorMessage: nil` on tool failure events.
- Updated `backend/internal/adapters/backfill/engine.go` to apply `domain.PruneEventsForPersistence` before atomic file writes and set schema version to 3.
- Updated `backend/internal/adapters/backfill/migration_runner.go` to default migration target version to `CurrentSchemaVersion` (3).
- Added batch migration and schema v3 tests in `migration_runner_test.go` and `engine_test.go`.

### 3. Presentation & CLI (`backend/cmd/agentic-lens/`, `web/`)
- Structured `runGenerate` in `backend/cmd/agentic-lens/main.go` so in-memory metric derivation and markdown rendering run against the complete unpruned fact ledger before pruning persisted events and setting schema version 3.
- Updated `runMigrate` and `runBackfill` in `backend/cmd/agentic-lens/main.go` to default to `domain.CurrentSchemaVersion` (3).
- Added CLI integration test `TestCLIGenerate_V3PrunedEventsAndZeroErrorMessage` in `backend/cmd/agentic-lens/main_test.go`.
- Web presentation layer (`web/`): Existing TypeScript components verify clean with optional `error_message` fields; zero frontend regressions.

### 4. Historical Workspaces Migration
- Successfully migrated all 36 historical workspace ledgers under `workspaces/` to schema version 3.
- All `workspaces/*/agentic_metrics.json` verified to contain zero `"error_message"` keys and average <5 KB.

### 5. Quality & Verification
- Unit and integration tests pass with race detector (`go test -race ./...`).
- Static checks pass clean (`make check`, `go vet ./...`).
- Multi-platform binaries compiled successfully (`make build`).
- Strict compliance with `TECHNICAL_GUIDELINES.md`: zero `any` across Go and TypeScript, zero emojis in code, tests, and documentation.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 2h 35m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 150.4k tokens (75% of 200k healthy ceiling, 3 compactions) | Subagents peak 168.5k (Coder)
- **Role Effort Distribution (Subagents):** Planner 9.5%, Coder 66.0%, Tester 18.4%, Reviewer 6.1%
- **Subagent Turn Velocity & Thrashing:** 133.8 avg turns/run (peak: 297 turns) | Planner 62.0, Coder 297.0, Tester 122.0, Reviewer 54.0
- **Test-to-Production Code Ratio:** 5.6:1 (+1,228 test lines / +220 prod lines)
- **Code Delivered:** +1,448 / -158 lines across 16 code files (excluding docs)
- **Reasoning Effort:** 138,820 reasoning tokens (56.6% of generated output)
- **Tool Health & Self-Correction:** 654 tool calls (28 self-corrections, 95.7% success rate)

### Anomalies & Health Flags
- **Status:** 5 execution anomalies detected (3 CRITICAL, 2 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (165.7k tokens) — Context tokens reached 165.7k in turn 50 (session: ses_eaa994517ffewCvJWHeUOib3EW), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k in turn 94 (session: ses_eaac19ffaffev5E6e14HzWvkJY), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+85.2k tokens) — Context tokens surged by 85.2k tokens in turn 201 (session: ses_eaac19ffaffev5E6e14HzWvkJY, total: 85.2k tokens, threshold: 80000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+60.1k tokens) — Context tokens surged by 60.1k tokens in turn 98 (session: ses_eaac19ffaffev5E6e14HzWvkJY, total: 60.1k tokens, threshold: 40000)
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: mcp-gateway_mcp-exec (3 consecutive errors) — Tool 'mcp-gateway_mcp-exec' encountered 3 consecutive execution failures (session: ses_eaad395bbffegFiZyotJbdZzFr, turn: 58)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 1,311,453 | 23,198 | 9,775,971 | 88.2% | 150.4k | ELEVATED | $1.86 |
| **planner** | 1 | 307,007 | 18,026 | 4,760,854 | 93.9% | 152.7k | ELEVATED | $0.75 |
| **coder** | 1 | 4,256,770 | 46,640 | 21,747,639 | 83.6% | 168.5k | ELEVATED | $5.24 |
| **tester** | 1 | 803,462 | 13,573 | 9,444,305 | 92.2% | 167.3k | ELEVATED | $1.46 |
| **reviewer** | 1 | 246,552 | 4,798 | 3,398,500 | 93.2% | 102.6k | ELEVATED | $0.49 |
| **Total** | **5** | **6,925,244** | **106,235** | **49,127,269** | **87.6%** | **168.5k** | **ELEVATED** | **$9.80** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 5 | 6,925,244 | 106,235 | 49,127,269 | 87.6% | $9.80 |
| **Total** | **5** | **6,925,244** | **106,235** | **49,127,269** | **87.6%** | **$9.80** |
