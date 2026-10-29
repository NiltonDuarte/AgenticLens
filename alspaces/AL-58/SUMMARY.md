# AL-58 Work Summary

## Domain

- No domain model or aggregation changes were required. Both OpenCode storage families normalize into the existing version-agnostic fact contracts.

## Adapters

- Added read-only SQLite schema-family detection that requires exactly one complete supported table/column family. Partial, mixed, malformed, and unknown schemas produce actionable errors.
- Added v2 `session_v2` / `session_message` extraction and session resolution while preserving v1 `session` / `message` / `part` behavior.
- Normalized v1 `filePath` and v2 `path` tool inputs, tool results/status, message text, event order, tokens, costs, timing, and model data.
- Preserved `--db-path` and the existing default database path. No normal version override flag was added.

## Presentation and documentation

- No frontend changes were needed.
- Updated `ARCHITECTURE.md` and `docs/SETUP_MANUAL.md` with schema-family detection, supported evidence versions, paths, workflows, and evidence limitations.

## Tests and workflows

- Added runtime-derived v1.19.5 and v2.2.0 SQLite fixtures and deterministic adapter/CLI coverage for extraction, schema rejection, empty selections, ordering, read-only behavior, tool outcomes, and semantic parity.
- Added isolated `make opencode-compat-v1`, `make opencode-compat-v2`, and `make opencode-compat` workflows using separate temporary HOME/XDG/project/database roots and a local deterministic provider.
- Retained sanitized runtime provenance and provider setup under `evidence/`.

## Verification

- After merging `origin/main`, `make` passed: Go vet and race tests, frontend typecheck/lint, 734 frontend tests, frontend bundle, and configured multi-platform builds.
- `OPENCODE_V1_BIN=/home/niltonduarte/.opencode/bin/opencode OPENCODE_V2_BIN=/home/niltonduarte/.local/bin/opencode2 make opencode-compat` passed both isolated runtime sessions and focused adapter/CLI tests.
- Independent tester passed focused adapter/CLI suites and the review regression checks.
- `git diff --check` passed; reviewer verdict: APPROVED.

## Evidence limits

The runtime evidence targets are v1.19.5 and v2.2.0. Nonzero native-provider pricing, imported history, partial migration states, duplicate handling, and releases outside those pinned targets remain unverified. npm reported six dependency advisories while installing frontend dependencies during `make`; no dependency updates were made for this ticket.

## Delivery

Implementation is committed on the AL-58 feature branch. Pull request and AgenticLens telemetry are finalized in the delivery step.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 10
- **Active Agent Duration:** 19m 5s (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 230.5k tokens (115% of 200k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0.6:1 (+364 test lines / +610 prod lines)
- **Code Delivered:** +974 / -7 lines across 7 code files (excluding docs)
- **Reasoning Effort:** 34,073 reasoning tokens (34.0% of generated output)
- **Tool Health & Self-Correction:** 131 tool calls (11 self-corrections, 91.6% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (1 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (206.0k tokens) — Context tokens reached 206.0k, breaching the 200k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.2k tokens) — Context tokens reached 162.2k, breaching the 160k warning ceiling.

### Agentic Execution Metrics

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **rootagent** | 1 | 403,174 | 65,995 | 23,337,344 | 98.3% | 230.5k | BLOATED (115%) | $0.00 |
| **Total** | **1** | **403,174** | **65,995** | **23,337,344** | **98.3%** | **230.5k** | **BLOATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 1 | 403,174 | 65,995 | 23,337,344 | 98.3% | $0.00 |
| **Total** | **1** | **403,174** | **65,995** | **23,337,344** | **98.3%** | **$0.00** |
