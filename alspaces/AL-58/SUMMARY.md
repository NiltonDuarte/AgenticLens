# AL-58 Work Summary

## Domain

- No domain model or aggregation changes were required. Both OpenCode storage families normalize into the existing version-agnostic fact contracts.

## Adapters

- Added read-only SQLite schema-family detection that requires exactly one complete supported table/column family. Partial, mixed, malformed, and unknown schemas produce actionable errors.
- Added v2 `session_v2` / `session_message` extraction and session resolution while preserving v1 `session` / `message` / `part` behavior.
- Normalized v1 `filePath` and v2 `path` tool inputs, tool results/status, message text, event order, tokens, costs, timing, and model data.
- Refactored adapter ownership into shared dispatch/detection, v1-specific extraction/resolution, and v2-specific extraction/resolution files. Version-specific filenames and symbols identify both schema family and responsibility; shared resolver comments and family diagnostics describe both layouts accurately.
- Removed redundant v1 `time_updated` metadata queries because schema-family detection already requires those columns. Synthetic required-column rejection tests now use isolated in-memory SQLite databases and parallel subtests.
- Added shared Codex identity resolution: blank or exact case-insensitive `unknown` values fall through from the extracted session to a known same-session ledger identity, then configured Git `user.name`, and finally `unknown`.
- Applied the resolver in generation and Codex backfill so migrated placeholders no longer block a trusted Git identity. Kept the migration placeholder and non-Codex resolution unchanged.
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
- Iteration 04 generation and backfill regressions cover unknown cached identity fallback, same-session and fresh identity precedence, and the unknown final fallback. `make test` and the full `make` passed after the iteration 04 implementation.
- Reprocessed AL-58 through the supported ticket-scoped Codex backfill with `--recompute-db --force`. Session `01a184e0-3822-7631-a3f3-cb33bb3e947f` now has user `Nilton Duarte`, from the configured Git identity. The source run also recovered five child-session records; all existing session IDs were preserved.
- Backend test profile: `go test -json ./...` completed in about 4.8 seconds wall time; the slowest test was `TestCLIGenerate_SessionRefreshMatchesFreshExtraction` at 0.52s. No backend test exceeded 0.52s, and no backend test uses `time.Sleep`.
- The uncached OpenCode adapter profile improved from 1.023s to 0.077–0.096s; the required-column omission matrix decreased from 0.910s to below the timer's 1ms reporting precision. No individual adapter test exceeded 100ms.
- Iteration 02 reviewer verdict: APPROVED. `git diff --check` passed.
- Iteration 03 reviewer verdict: APPROVED. `git diff --check` passed.

## Evidence limits

The runtime evidence targets are v1.19.5 and v2.2.0. Nonzero native-provider pricing, imported history, partial migration states, duplicate handling, and releases outside those pinned targets remain unverified. npm reported six dependency advisories while installing frontend dependencies during `make`; no dependency updates were made for this ticket.

## Delivery

Iteration 02 naming refactor is committed as `ab374c8`; iteration 03 cleanup is committed as `238f3e0` on the AL-58 feature branch. The branch already includes the merge from `origin/main`; the existing pull request and AgenticLens telemetry are updated in the delivery step.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 33
- **Active Agent Duration:** 1h 30m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 244.9k tokens (123% of 200k healthy ceiling, 6 compactions) | Subagents peak 246.1k (Coder)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 1.1 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 1.0, Tester 1.5, Reviewer 1.0
- **Test-to-Production Code Ratio:** 1.4:1 (+2,099 test lines / +1,494 prod lines)
- **Code Delivered:** +3,593 / -2,392 lines across 20 code files (excluding docs)
- **Reasoning Effort:** 93,212 reasoning tokens (33.6% of generated output)
- **Tool Health & Self-Correction:** 355 tool calls (21 self-corrections, 94.1% success rate)

### Anomalies & Health Flags
- **Status:** 14 execution anomalies detected (13 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+244.5k tokens) — Context tokens surged by 244.5k to 244.5k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (244.5k tokens) — Context tokens reached 244.5k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+246.1k tokens) — Context tokens surged by 246.1k to 246.1k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (246.1k tokens) — Context tokens reached 246.1k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (214.0k tokens) — Context tokens reached 214.0k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (218.8k tokens) — Context tokens reached 218.8k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+221.6k tokens) — Context tokens surged by 221.6k to 221.6k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (221.6k tokens) — Context tokens reached 221.6k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+219.9k tokens) — Context tokens surged by 219.9k to 219.9k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (219.9k tokens) — Context tokens reached 219.9k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (238.0k tokens) — Context tokens reached 238.0k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (244.8k tokens) — Context tokens reached 244.8k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (206.0k tokens) — Context tokens reached 206.0k, breaching the 200k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.2k tokens) — Context tokens reached 162.2k, breaching the 160k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 108,220 | 7,662 | 1,837,568 | 94.4% | 244.7k | BLOATED (122%) | $0.00 |
| **coder** | 3 | 89,565 | 16,435 | 5,623,424 | 98.4% | 246.1k | BLOATED (123%) | $0.00 |
| **tester** | 2 | 54,510 | 5,743 | 2,770,304 | 98.1% | 245.0k | BLOATED (123%) | $0.00 |
| **reviewer** | 2 | 21,234 | 3,725 | 1,298,944 | 98.4% | 226.7k | BLOATED (113%) | $0.00 |
| **rootagent** | 1 | 1,383,053 | 150,563 | 48,882,048 | 97.2% | 244.9k | BLOATED (122%) | $0.00 |
| **Total** | **10** | **1,656,582** | **184,128** | **60,412,288** | **97.3%** | **246.1k** | **BLOATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 10 | 1,656,582 | 184,128 | 60,412,288 | 97.3% | $0.00 |
| **Total** | **10** | **1,656,582** | **184,128** | **60,412,288** | **97.3%** | **$0.00** |
