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

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 17
- **Active Agent Duration:** 4h 26m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 169.2k tokens (85% of 200k healthy ceiling, 8 compactions) | Subagents peak 167.9k (Coder)
- **Role Effort Distribution (Subagents):** Coder 100.0%
- **Subagent Turn Velocity & Thrashing:** 115.0 avg turns/run (peak: 352 turns) | Coder 115.0
- **Test-to-Production Code Ratio:** 2.7:1 (+1,546 test lines / +576 prod lines)
- **Code Delivered:** +2,208 / -907 lines across 41 code files (excluding docs)
- **Reasoning Effort:** 327,540 reasoning tokens (65.6% of generated output)
- **Tool Health & Self-Correction:** 1,158 tool calls (11 self-corrections, 99.1% success rate)

### Anomalies & Health Flags
- **Status:** 14 execution anomalies detected (6 CRITICAL, 8 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.7k tokens) — Context tokens reached 160.7k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+114.2k tokens) — Context tokens surged by 114.2k to 114.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+102.5k tokens) — Context tokens surged by 102.5k to 102.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+49.3k tokens) — Context tokens surged by 49.3k to 49.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+57.8k tokens) — Context tokens surged by 57.8k to 57.8k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+50.6k tokens) — Context tokens surged by 50.6k to 50.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+52.7k tokens) — Context tokens surged by 52.7k to 52.7k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+47.3k tokens) — Context tokens surged by 47.3k to 47.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+74.7k tokens) — Context tokens surged by 74.7k to 74.7k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+49.5k tokens) — Context tokens surged by 49.5k to 49.5k tokens.
- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: task (28.6%) — Tool 'task' failed 2 out of 7 calls (28.6% failure rate, threshold: >20%).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 4,300,899 | 86,432 | 47,819,822 | 91.7% | 169.2k | ELEVATED | $7.66 |
| **coder** | 5 | 3,377,121 | 85,180 | 45,514,230 | 93.1% | 167.9k | ELEVATED | $6.97 |
| **Total** | **6** | **7,678,020** | **171,612** | **93,334,052** | **92.4%** | **169.2k** | **ELEVATED** | **$14.63** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 6 | 7,678,020 | 171,612 | 93,334,052 | 92.4% | $14.63 |
| **Total** | **6** | **7,678,020** | **171,612** | **93,334,052** | **92.4%** | **$14.63** |
