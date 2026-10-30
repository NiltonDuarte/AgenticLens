# AL-58 Work Summary

## Domain

- No domain model or aggregation changes were required. Both OpenCode storage families normalize into the existing version-agnostic fact contracts.

## Adapters

- Added read-only SQLite schema-family detection that requires exactly one complete supported table/column family. Partial, mixed, malformed, and unknown schemas produce actionable errors.
- Added v2 `session_v2` / `session_message` extraction and session resolution while preserving v1 `session` / `message` / `part` behavior.
- Normalized v1 `filePath` and v2 `path` tool inputs, tool results/status, message text, event order, tokens, costs, timing, and model data.
- Refactored adapter ownership into shared dispatch/detection, v1-specific extraction/resolution, and v2-specific extraction/resolution files. Version-specific filenames and symbols identify both schema family and responsibility; shared resolver comments and family diagnostics describe both layouts accurately.
- Preserved `--db-path` and the existing default database path. No normal version override flag was added.

## Presentation and documentation

- No frontend changes were needed.
- Updated `ARCHITECTURE.md` and `docs/SETUP_MANUAL.md` with schema-family detection, supported evidence versions, paths, workflows, and evidence limitations.

## Tests and workflows

- Added runtime-derived v1.19.5 and v2.2.0 SQLite fixtures and deterministic adapter/CLI coverage for extraction, schema rejection, empty selections, ordering, read-only behavior, tool outcomes, and semantic parity.
- Added isolated `make opencode-compat-v1`, `make opencode-compat-v2`, and `make opencode-compat` workflows using separate temporary HOME/XDG/project/database roots and a local deterministic provider.
- Retained sanitized runtime provenance and provider setup under `evidence/`.

## Verification

- After merging `origin/main` and the iteration 02 refactor, `make` passed: Go vet and race tests, frontend typecheck/lint, 820 frontend tests, frontend bundle, and configured multi-platform builds.
- `OPENCODE_V1_BIN=/home/niltonduarte/.opencode/bin/opencode OPENCODE_V2_BIN=/home/niltonduarte/.local/bin/opencode2 make opencode-compat` passed both isolated runtime sessions and focused adapter/CLI tests.
- Independent tester confirmed the renamed extraction/resolution/detection tests retain the existing cases; `go test -count=1 ./internal/adapters/opencode` passed.
- Iteration 02 reviewer verdict: APPROVED. `git diff --check` passed.

## Evidence limits

The runtime evidence targets are v1.19.5 and v2.2.0. Nonzero native-provider pricing, imported history, partial migration states, duplicate handling, and releases outside those pinned targets remain unverified. npm reported six dependency advisories while installing frontend dependencies during `make`; no dependency updates were made for this ticket.

## Delivery

Iteration 02 naming refactor is committed as `ab374c8` on the AL-58 feature branch. The branch already includes the merge from `origin/main`; the existing pull request and AgenticLens telemetry are updated in the delivery step.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 18
- **Active Agent Duration:** 55m 17s (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 233.2k tokens (117% of 200k healthy ceiling, 1 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 1.3:1 (+1,968 test lines / +1,478 prod lines)
- **Code Delivered:** +3,446 / -2,374 lines across 14 code files (excluding docs)
- **Reasoning Effort:** 61,228 reasoning tokens (35.5% of generated output)
- **Tool Health & Self-Correction:** 211 tool calls (15 self-corrections, 92.9% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (1 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (206.0k tokens) — Context tokens reached 206.0k, breaching the 200k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.2k tokens) — Context tokens reached 162.2k, breaching the 160k warning ceiling.

### Agentic Execution Metrics

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **rootagent** | 1 | 809,684 | 111,360 | 35,875,072 | 97.8% | 233.2k | BLOATED (117%) | $0.00 |
| **Total** | **1** | **809,684** | **111,360** | **35,875,072** | **97.8%** | **233.2k** | **BLOATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 1 | 809,684 | 111,360 | 35,875,072 | 97.8% | $0.00 |
| **Total** | **1** | **809,684** | **111,360** | **35,875,072** | **97.8%** | **$0.00** |
