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
- Upgraded and normalized all 71 workspace ledgers in `alspaces/*/agentic_metrics.json` and all `testdata/` ledgers to canonical schema v4 with non-empty session users and no synthetic invocation IDs.
- Updated backend and frontend test suites and fixtures.
- Verified all 71 workspace ledgers load and validate with zero errors, warnings, or diagnostics.

## Iteration 04: Remove Synthetic Invocation IDs

- Removed `sessions[].invocations[].id` from the Go and TypeScript contracts, DTO serialization, canonical schema validation, and all 89 persisted production/test ledgers (604 rows); schema v4 now rejects the removed key.
- Replaced synthetic-ID ownership inference with explicit enclosing session ownership across extraction, migration, anomaly attribution, hydration, aggregation, and thread summaries. Multiple rows within one session remain distinct.
- Changed refresh merges to replace the refreshed session's full cumulative snapshot, preventing double counting as Codex or OpenCode sessions grow while retaining other sessions and duplicate rows in the fresh snapshot.
- Added regression coverage for Codex/OpenCode source growth, repeated row preservation, session ownership, legacy migrations, strict v4 rejection, and frontend progression.
- Verification: `make check`, `make test`, and uncached Go race tests passed; 2,361 Go tests/subtests and 737 frontend tests passed. Independent review approved iteration 04. See `iterations/04/IMPLEMENTATION_TEST_EVIDENCE.txt` and `iterations/04/REVIEW.md`.

## Iteration 05: Ticket and Plan Lifecycle Requirements

- Updated `AGENTS.md`, `CONTRIBUTING.md`, and the standard agent policy block in `docs/SETUP_MANUAL.md` to distinguish ticketed work from ad hoc work.
- The initial wording allowed ad hoc work without a ticket ID or workspace; iteration 06 corrects that overbroad exception.
- Plan lifecycle now moves a matching specification to `plan/done/` only when it exists in `plan/to-do/`, and removes the ticket index entry only when present. No ticket artifacts are created to satisfy lifecycle rules.
- Reviewer verdict: APPROVED (`alspaces/FIX-10/iterations/05/REVIEW.md`).
- Documentation consistency search and `git diff --check`: PASS.

## Iteration 06: Require Ticket IDs and Workspaces

- Clarified that every task, including ad hoc work, requires a ticket ID and isolated workspace. Assign a `GEN-#` ID when no ID is provided.
- Kept planning records optional for ad hoc tickets: move a matching existing specification to `plan/done/` and remove an index entry only when present; do not create either solely for lifecycle compliance.
- Aligned `AGENTS.md`, `CONTRIBUTING.md`, and `docs/SETUP_MANUAL.md`.
- Reviewer verdict: APPROVED (`alspaces/FIX-10/iterations/06/REVIEW.md`).
- Documentation consistency search and `git diff --check`: PASS.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 40
- **Active Agent Duration:** 6h 49m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 240.9k tokens (120% of 200k healthy ceiling, 11 compactions) | Subagents peak 168.4k (Coder)
- **Role Effort Distribution (Subagents):** Coder 100.0%
- **Subagent Turn Velocity & Thrashing:** 153.2 avg turns/run (peak: 352 turns) | Coder 153.2
- **Test-to-Production Code Ratio:** 4.2:1 (+2,778 test lines / +664 prod lines)
- **Code Delivered:** +3,442 / -1,958 lines across 68 code files (excluding docs)
- **Reasoning Effort:** 446,085 reasoning tokens (58.7% of generated output)
- **Tool Health & Self-Correction:** 1,745 tool calls (23 self-corrections, 98.7% success rate)

### Anomalies & Health Flags
- **Status:** 29 execution anomalies detected (14 CRITICAL, 15 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (201.2k tokens) — Context tokens reached 201.2k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.7k tokens) — Context tokens reached 160.7k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+114.2k tokens) — Context tokens surged by 114.2k to 114.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+104.2k tokens) — Context tokens surged by 104.2k to 104.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+102.5k tokens) — Context tokens surged by 102.5k to 102.5k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (168.4k tokens) — Context tokens reached 168.4k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (165.4k tokens) — Context tokens reached 165.4k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (167.9k tokens) — Context tokens reached 167.9k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (169.2k tokens) — Context tokens reached 169.2k, breaching the 160k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (165.8k tokens) — Context tokens reached 165.8k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+49.3k tokens) — Context tokens surged by 49.3k to 49.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+57.8k tokens) — Context tokens surged by 57.8k to 57.8k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+72.3k tokens) — Context tokens surged by 72.3k to 72.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+50.6k tokens) — Context tokens surged by 50.6k to 50.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+52.7k tokens) — Context tokens surged by 52.7k to 52.7k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+47.3k tokens) — Context tokens surged by 47.3k to 47.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+74.7k tokens) — Context tokens surged by 74.7k to 74.7k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+49.5k tokens) — Context tokens surged by 49.5k to 49.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (168.4k tokens) — Context tokens reached 168.4k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (165.4k tokens) — Context tokens reached 165.4k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (167.9k tokens) — Context tokens reached 167.9k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (169.2k tokens) — Context tokens reached 169.2k, breaching the 160k warning ceiling.
- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: task (33.3%) — Tool 'task' failed 3 out of 9 calls (33.3% failure rate, threshold: >20%).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 5,261,607 | 93,224 | 52,654,100 | 90.9% | 169.2k | ELEVATED | $8.81 |
| **coder** | 6 | 5,154,527 | 131,272 | 75,180,300 | 93.6% | 168.4k | ELEVATED | $10.94 |
| **rootagent** | 1 | 615,389 | 89,007 | 39,521,408 | 98.5% | 240.9k | BLOATED (120%) | $0.00 |
| **Total** | **8** | **11,031,523** | **313,503** | **167,355,808** | **93.8%** | **240.9k** | **BLOATED** | **$19.75** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 7 | 10,416,134 | 224,496 | 127,834,400 | 92.5% | $19.75 |
| **gpt-6-luna** | 1 | 615,389 | 89,007 | 39,521,408 | 98.5% | $0.00 |
| **Total** | **8** | **11,031,523** | **313,503** | **167,355,808** | **93.8%** | **$19.75** |
