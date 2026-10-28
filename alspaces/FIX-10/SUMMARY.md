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
- **Software Development Life Cycle Iterations:** 7
- **Active Agent Duration:** 42m 55s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 169.2k tokens (85% of 200k healthy ceiling, 2 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 51,468 reasoning tokens (64.4% of generated output)
- **Tool Health & Self-Correction:** 227 tool calls (2 self-corrections, 99.1% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (1 CRITICAL, 2 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.7k tokens) — Context tokens reached 160.7k, breaching the 160k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+50.6k tokens) — Context tokens surged by 50.6k to 50.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+52.7k tokens) — Context tokens surged by 52.7k to 52.7k tokens.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **general** | 1 | 1,317,768 | 28,466 | 16,065,669 | 92.4% | 169.2k | ELEVATED | $2.49 |
| **Total** | **1** | **1,317,768** | **28,466** | **16,065,669** | **92.4%** | **169.2k** | **ELEVATED** | **$2.49** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 1 | 1,317,768 | 28,466 | 16,065,669 | 92.4% | $2.49 |
| **Total** | **1** | **1,317,768** | **28,466** | **16,065,669** | **92.4%** | **$2.49** |
