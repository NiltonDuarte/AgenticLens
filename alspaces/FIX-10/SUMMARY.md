# Summary: FIX-10 Multi-Root Sessions and Explicit Fallbacks

## Overview
Fixed two critical telemetry backfill data corruption bugs in AgenticLens:
1. Multi-root session truncation: `engine.go` previously stopped at the first root session encountered, omitting subsequent root sessions and child invocations.
2. Unsafe implicit heuristic fallbacks: missing local SQLite sessions previously fell back to title substring matching and workspace path matching, which could latch onto unrelated review or linting sessions and overwrite valid historical ledger data.

## Key Changes
- `backend/internal/adapters/backfill/options.go`: Added `AllowFallbacks bool` option to `BackfillOptions`.
- `backend/cmd/agentic-lens/main.go`: Added `--allow-fallbacks` CLI flag and documentation.
- `backend/internal/adapters/backfill/engine.go`:
  - Implemented multi-root session discovery resolving all unique candidate roots present across ledger invocations.
  - Extracted facts across all resolved root session IDs without truncating subsequent trees.
  - Restricted title substring and commit-range workspace path heuristic searches to run only when `AllowFallbacks` is true.
  - Preserved existing invocations and events when sessions are missing from local SQLite, leaving status as `partial` pending host migration.
- `backend/internal/adapters/backfill/engine_test.go`: Added tests for multi-root extraction, fallback rejection, and explicit fallback execution.

## Verification
- Unit test suite: `make test` passed with 100% success across all backend packages and frontend tests.
- Target verification: Verified on 10 workspaces in `/Users/guedesduarte/workspace/git/SinkLayerControlPlane.Code`. Multi-root ticket `DPS-11170` successfully retained both root sessions (12 invocations total, $23.55 cost, 25 SDLC iterations), while missing host tickets `DPS-11173`, `DPS-11187`, `DPS-11197`, and `DPS-11199` preserved all historical invocations intact.

## Iteration 02: Normalize Session-Invocation Hierarchy
- Hoisted `User` and `ParentID` to `SessionFact` as container-level metadata, removing them from `InvocationFact`.
- Assigned distinct invocation IDs (`<session_id>-inv-<index>`) across OpenCode extraction, Codex rollouts, and schema v4 migration.
- Updated schema v4 migration step (`step_v3_v4.go`), invariant validator (`validator.go`), and runtime fact validator (`schemas.go`).
- Updated TypeScript types (`web/src/domain/types.ts`) with `user?: string` on `SessionFact` and optional fields on `InvocationFact` for boundary compatibility.
- Verified all backend and frontend tests pass via `make`.

## Iteration 03: Remove Schema Fallbacks and Normalize All Existing alspaces Ledgers
- Removed legacy `ParentID` and `User` fields from `InvocationFactDTO` in `backend/internal/adapters/schemas.go`.
- Removed backward-compatibility hoisting loop from `AgenticFactLedgerDTO.ToDomain()`, enforcing pure canonical v4 ingestion without unknown invocation fields.
- Removed `parent_id` and `user` from TypeScript `InvocationFact` interface in `web/src/domain/types.ts`.
- Upgraded and normalized all 71 workspace ledgers in `alspaces/*/agentic_metrics.json` and all `testdata/` ledgers to canonical schema v4 with non-empty session users and distinct invocation IDs.
- Updated backend and frontend test suites and fixtures.
- Verified all 71 workspace ledgers load and validate with zero errors, warnings, or diagnostics.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 20
- **Active Agent Duration:** 5h 15m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 169.2k tokens (85% of 200k healthy ceiling, 10 compactions) | Subagents peak 168.4k (Coder)
- **Role Effort Distribution (Subagents):** Coder 100.0%
- **Subagent Turn Velocity & Thrashing:** 148.2 avg turns/run (peak: 352 turns) | Coder 148.2
- **Test-to-Production Code Ratio:** 2.7:1 (+1,546 test lines / +576 prod lines)
- **Code Delivered:** +2,208 / -907 lines across 41 code files (excluding docs)
- **Reasoning Effort:** 398,844 reasoning tokens (64.3% of generated output)
- **Tool Health & Self-Correction:** 1,521 tool calls (15 self-corrections, 99.0% success rate)

### Anomalies & Health Flags
- **Status:** 17 execution anomalies detected (8 CRITICAL, 9 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.7k tokens) — Context tokens reached 160.7k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+114.2k tokens) — Context tokens surged by 114.2k to 114.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+104.2k tokens) — Context tokens surged by 104.2k to 104.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+102.5k tokens) — Context tokens surged by 102.5k to 102.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+49.3k tokens) — Context tokens surged by 49.3k to 49.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+57.8k tokens) — Context tokens surged by 57.8k to 57.8k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+72.3k tokens) — Context tokens surged by 72.3k to 72.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+50.6k tokens) — Context tokens surged by 50.6k to 50.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+52.7k tokens) — Context tokens surged by 52.7k to 52.7k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+47.3k tokens) — Context tokens surged by 47.3k to 47.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+74.7k tokens) — Context tokens surged by 74.7k to 74.7k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+49.5k tokens) — Context tokens surged by 49.5k to 49.5k tokens.
- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: task (33.3%) — Tool 'task' failed 3 out of 9 calls (33.3% failure rate, threshold: >20%).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 4,995,879 | 92,275 | 51,733,760 | 91.2% | 169.2k | ELEVATED | $8.54 |
| **coder** | 6 | 5,011,507 | 129,016 | 73,069,286 | 93.6% | 168.4k | ELEVATED | $10.66 |
| **Total** | **7** | **10,007,386** | **221,291** | **124,803,046** | **92.6%** | **169.2k** | **ELEVATED** | **$19.19** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 7 | 10,007,386 | 221,291 | 124,803,046 | 92.6% | $19.19 |
| **Total** | **7** | **10,007,386** | **221,291** | **124,803,046** | **92.6%** | **$19.19** |
