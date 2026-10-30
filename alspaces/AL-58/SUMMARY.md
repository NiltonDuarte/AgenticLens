# AL-58 Work Summary

## Domain

- No domain model or aggregation changes were required. Both OpenCode storage families normalize into the existing version-agnostic fact contracts.

## Adapters

- Added read-only SQLite schema-family detection that requires exactly one complete supported table/column family. Partial, mixed, malformed, and unknown schemas produce actionable errors.
- Added v2 `session_v2` / `session_message` extraction and session resolution while preserving v1 `session` / `message` / `part` behavior.
- Normalized v1 `filePath` and v2 `path` tool inputs, tool results/status, message text, event order, tokens, costs, timing, and model data.
- Refactored adapter ownership into shared dispatch/detection, v1-specific extraction/resolution, and v2-specific extraction/resolution files. Version-specific filenames and symbols identify both schema family and responsibility; shared resolver comments and family diagnostics describe both layouts accurately.
- Removed redundant v1 `time_updated` metadata queries because schema-family detection already requires those columns. Synthetic required-column rejection tests now use isolated in-memory SQLite databases and parallel subtests.
- Preserved `--db-path` and the existing default database path. No normal version override flag was added.

## Presentation and documentation

- No frontend changes were needed.
- Updated `ARCHITECTURE.md` and `docs/SETUP_MANUAL.md` with schema-family detection, supported evidence versions, paths, workflows, and evidence limitations.

## Tests and workflows

- Added runtime-derived v1.19.5 and v2.2.0 SQLite fixtures and deterministic adapter/CLI coverage for extraction, schema rejection, empty selections, ordering, read-only behavior, tool outcomes, and semantic parity.
- Added isolated `make opencode-compat-v1`, `make opencode-compat-v2`, and `make opencode-compat` workflows using separate temporary HOME/XDG/project/database roots and a local deterministic provider.
- Retained sanitized runtime provenance and provider setup under `evidence/`.

## Verification

- After merging `origin/main` and the iteration 03 cleanup, `make` passed: Go vet and race tests, frontend typecheck/lint, 820 frontend tests, frontend bundle, and configured multi-platform builds.
- `OPENCODE_V1_BIN=/home/niltonduarte/.opencode/bin/opencode OPENCODE_V2_BIN=/home/niltonduarte/.local/bin/opencode2 make opencode-compat` passed both isolated runtime sessions and focused adapter/CLI tests.
- Independent tester confirmed the renamed extraction/resolution/detection tests retain the existing cases; `go test -count=1 ./internal/adapters/opencode` passed.
- The uncached OpenCode adapter profile improved from 1.023s to 0.077–0.096s; the required-column omission matrix decreased from 0.910s to below the timer's 1ms reporting precision. No individual adapter test exceeded 100ms.
- Iteration 02 reviewer verdict: APPROVED. `git diff --check` passed.
- Iteration 03 reviewer verdict: APPROVED. `git diff --check` passed.

## Evidence limits

The runtime evidence targets are v1.19.5 and v2.2.0. Nonzero native-provider pricing, imported history, partial migration states, duplicate handling, and releases outside those pinned targets remain unverified. npm reported six dependency advisories while installing frontend dependencies during `make`; no dependency updates were made for this ticket.

## Delivery

Iteration 02 naming refactor is committed as `ab374c8`; iteration 03 cleanup is committed as `238f3e0` on the AL-58 feature branch. The branch already includes the merge from `origin/main`; the existing pull request and AgenticLens telemetry are updated in the delivery step.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 26
- **Active Agent Duration:** 1h 19m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 233.2k tokens (117% of 200k healthy ceiling, 1 compactions) | Subagents peak 229.4k (Coder)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0, Coder 1.0, Tester 1.0, Reviewer 1.0
- **Test-to-Production Code Ratio:** 1.4:1 (+1,969 test lines / +1,449 prod lines)
- **Code Delivered:** +3,418 / -2,374 lines across 14 code files (excluding docs)
- **Reasoning Effort:** 70,978 reasoning tokens (34.9% of generated output)
- **Tool Health & Self-Correction:** 254 tool calls (16 self-corrections, 93.7% success rate)

### Anomalies & Health Flags
- **Status:** 8 execution anomalies detected (7 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (214.0k tokens) — Context tokens reached 214.0k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (218.8k tokens) — Context tokens reached 218.8k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+221.6k tokens) — Context tokens surged by 221.6k to 221.6k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (221.6k tokens) — Context tokens reached 221.6k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+219.9k tokens) — Context tokens surged by 219.9k to 219.9k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (219.9k tokens) — Context tokens reached 219.9k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (206.0k tokens) — Context tokens reached 206.0k, breaching the 200k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.2k tokens) — Context tokens reached 162.2k, breaching the 160k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 52,822 | 2,566 | 1,044,352 | 95.2% | 222.9k | BLOATED (111%) | $0.00 |
| **coder** | 1 | 19,833 | 4,561 | 3,812,736 | 99.5% | 229.4k | BLOATED (115%) | $0.00 |
| **tester** | 1 | 9,083 | 1,445 | 1,327,616 | 99.3% | 225.3k | BLOATED (113%) | $0.00 |
| **reviewer** | 1 | 10,435 | 1,959 | 1,109,120 | 99.1% | 226.7k | BLOATED (113%) | $0.00 |
| **rootagent** | 1 | 1,258,045 | 122,120 | 41,470,080 | 97.1% | 233.2k | BLOATED (117%) | $0.00 |
| **Total** | **5** | **1,350,218** | **132,651** | **48,763,904** | **97.3%** | **233.2k** | **BLOATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 5 | 1,350,218 | 132,651 | 48,763,904 | 97.3% | $0.00 |
| **Total** | **5** | **1,350,218** | **132,651** | **48,763,904** | **97.3%** | **$0.00** |
