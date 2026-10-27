# Implementation Summary: Ticket FIX-4

## Overview
Remediated backend test failures caused by renaming the repository workspace directory from `workspaces` to `alspaces` and updating `.agenticlens-config.yaml`. Decoupled all backend test suites from outer project directories by introducing isolated test fixtures under `testdata/parity_workspaces/`. Hardened the `server` CLI command to strictly load project configuration on startup, derive the active workspaces directory when `--workspaces` is omitted, and fail immediately if no configuration file is found or if `workspaces.dir` is unconfigured. Completely purged hardcoded fallbacks to `"workspaces"` and all obsolete references to the decommissioned `.agenticlens` directory across adapters, loaders, commands, and tests.

## Artifacts and Pointers
- Technical Contract: [CONTRACT.md](CONTRACT.md)
- Iteration 01 Plan: [iterations/01/PLAN.md](iterations/01/PLAN.md)
- Iteration 01 Review: [iterations/01/REVIEW.md](iterations/01/REVIEW.md)
- Iteration 02 Plan: [iterations/02/PLAN.md](iterations/02/PLAN.md)
- Iteration 02 Review: [iterations/02/REVIEW.md](iterations/02/REVIEW.md)

## Changes Grouped by Layer

### Domain Layer (`backend/internal/domain/`)
- Invariant preservation: Zero modifications to pure functional domain models, aggregations, or anomaly detectors. FCIS boundaries strictly maintained.

### Boundary Adapters Layer (`backend/internal/adapters/`)
- `backend/internal/adapters/config/loader.go` & `writer.go`:
  - Set `Workspaces.Dir` default to `""` in config merging, ensuring unconfigured workspaces directories are explicitly detectable rather than silently masked with `"workspaces"`.
  - Restricted `FindConfigFile` candidate lookup strictly to root configuration candidates (`.agenticlens-config.yaml`, `.agenticlens-config.yml`, `.agenticlens-config.json`).
  - Completely purged legacy `.agenticlens/` subfolder evaluation.
- `backend/internal/adapters/repository.go`:
  - In `ResolveWorkspacesDir`, eliminated hardcoded default and fallback to `"workspaces"`. Returns `""` immediately when unconfigured.
  - Removed container mount `/app/workspaces` probe and obsolete `.agenticlens` directory traversal check.
  - In `ScanWorkspaces`, added guard recording an explicit error when the target workspaces directory is unconfigured or empty.
- `backend/internal/adapters/parity_test.go`:
  - Decoupled from live outer repository workspaces.
  - Implemented `findParityWorkspacesDir()` to resolve `testdata/parity_workspaces` relative to the test file via `runtime.Caller(0)`.
  - Updated all parity tests (`TestCrossEngineParityOnWorkspaceLedgers`, `TestFleetSummaryAndComparisonParity`, `TestStructuralClassificationParityOnWorkspaces`) to run self-contained against the 12 reference fixtures.
- `backend/internal/adapters/http/server_test.go`:
  - Decoupled `TestServerWithProductionWorkspaces_DoesNotTouchTestdataWorkspaces` from outer `../../../../workspaces`, pointing to `testdata/parity_workspaces`.
  - Added `TestServerWorkspaceDirConfigSynchronization` to verify bidirectional synchronization between constructor workspace parameters and `ProjectConfig.Workspaces.Dir`.
- `backend/internal/adapters/http/server.go`:
  - Ensured `NewServer` and `NewServerWithConfig` keep `cfg.Workspaces.Dir` synchronized with the active workspaces directory string across `/api/config` and `/api/workspaces`.

### Presentation, CLI, & Application Layer (`backend/cmd/agentic-lens/`, `testdata/`)
- `backend/cmd/agentic-lens/main.go`:
  - In `runServer`, removed `if configuredDir == "" { configuredDir = "workspaces" }`. If `projCfg.Workspaces.Dir` is unconfigured and `--workspaces` is not provided, terminates with exit code 1 and outputs `Error: workspaces.dir is empty or unconfigured: configure 'workspaces.dir' in .agenticlens-config.yaml or provide --workspaces flag`.
  - Removed `if filepath.Base(parentDir) == ".agenticlens"` when resolving relative workspace paths.
  - In root CLI execution (`main()`), added guard failing with the same actionable error when `workspaces.dir` is empty and `--workspaces` is omitted.
- `backend/cmd/agentic-lens/main_test.go`:
  - Added CLI test cases asserting server failure on missing configuration and on empty `workspaces.dir` without flag.
  - Added tests asserting server startup succeeds when `--workspaces` flag is supplied even if unconfigured in `.agenticlens-config.yaml`.
  - Migrated legacy test helpers from `.agenticlens/` to root configuration files.
- `testdata/parity_workspaces/`:
  - Established 12 isolated reference ticket fixtures (`AL-1`, `AL-2`, `AL-3`, `AL-4`, `AL-10`, `AL-11`, `AL-12`, `GEN-1`, `GEN-2`, `PO-1`, `PO-2`, `PO-3`) containing valid contracts, summaries, and telemetry ledgers.
  - Preserved `testdata/workspaces/` (5 baseline tickets) intact.

## Verification Status
- Backend test suite (`go test -race ./...`): Passed with zero data races.
- Frontend test suite (`vitest run`): 28 test files passed (622 tests).
- Static analysis & linting (`make check`): Passed (`go vet`, `tsc --noEmit`, `eslint`).
- Binary build (`make build`): All platform binaries compiled cleanly.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 22
- **Active Agent Duration:** 28h 25m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 174.3k tokens (87% of 200k healthy ceiling, 21 compactions) | Subagents peak 171.2k (Coder)
- **Role Effort Distribution (Subagents):** Planner 6.8%, Coder 73.4%, Tester 11.5%, Reviewer 8.3%
- **Subagent Turn Velocity & Thrashing:** 165.5 avg turns/run (peak: 1672 turns) | Planner 56.5, Coder 434.9, Tester 84.0, Reviewer 59.3
- **Test-to-Production Code Ratio:** 6.1:1 (+610 test lines / +100 prod lines)
- **Code Delivered:** +5,452 / -482 lines across 447 code files (excluding docs)
- **Reasoning Effort:** 595,568 reasoning tokens (46.4% of generated output)
- **Tool Health & Self-Correction:** 4,742 tool calls (52 self-corrections, 98.9% success rate)

### Anomalies & Health Flags
- **Status:** 29 execution anomalies detected (21 CRITICAL, 8 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.4k tokens) — Context tokens reached 161.4k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.0k tokens) — Context tokens reached 160.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+110.1k tokens) — Context tokens surged by 110.1k to 110.1k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+105.1k tokens) — Context tokens surged by 105.1k to 105.1k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+93.6k tokens) — Context tokens surged by 93.6k to 93.6k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+108.8k tokens) — Context tokens surged by 108.8k to 108.8k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+114.5k tokens) — Context tokens surged by 114.5k to 114.5k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+117.0k tokens) — Context tokens surged by 117.0k to 117.0k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+84.8k tokens) — Context tokens surged by 84.8k to 84.8k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+107.5k tokens) — Context tokens surged by 107.5k to 107.5k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+120.2k tokens) — Context tokens surged by 120.2k to 120.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+116.1k tokens) — Context tokens surged by 116.1k to 116.1k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+103.7k tokens) — Context tokens surged by 103.7k to 103.7k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+93.6k tokens) — Context tokens surged by 93.6k to 93.6k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+110.3k tokens) — Context tokens surged by 110.3k to 110.3k tokens.
- [CRITICAL] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: mcp-gateway_mcp-exec (100.0%) — Tool 'mcp-gateway_mcp-exec' failed 5 out of 5 calls (100.0% failure rate, threshold: >=50%).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+42.0k tokens) — Context tokens surged by 42.0k to 42.0k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+66.5k tokens) — Context tokens surged by 66.5k to 66.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+79.1k tokens) — Context tokens surged by 79.1k to 79.1k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+58.2k tokens) — Context tokens surged by 58.2k to 58.2k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+58.2k tokens) — Context tokens surged by 58.2k to 58.2k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+72.8k tokens) — Context tokens surged by 72.8k to 72.8k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+66.9k tokens) — Context tokens surged by 66.9k to 66.9k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+72.3k tokens) — Context tokens surged by 72.3k to 72.3k tokens.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 6 | 2,014,718 | 47,071 | 18,142,843 | 90.0% | 139.9k | ELEVATED | $3.54 |
| **coder** | 7 | 20,766,440 | 373,288 | 274,234,612 | 93.0% | 171.2k | ELEVATED | $38.24 |
| **tester** | 6 | 3,415,592 | 101,053 | 35,590,910 | 91.2% | 154.4k | ELEVATED | $6.02 |
| **reviewer** | 7 | 3,002,475 | 34,794 | 21,825,067 | 87.9% | 167.0k | ELEVATED | $4.30 |
| **advisor** | 1 | 10,825,511 | 131,623 | 37,412,063 | 77.6% | 174.3k | ELEVATED | $11.77 |
| **Total** | **27** | **40,024,736** | **687,829** | **387,205,495** | **90.6%** | **174.3k** | **ELEVATED** | **$63.87** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 27 | 40,024,736 | 687,829 | 387,205,495 | 90.6% | $63.87 |
| **Total** | **27** | **40,024,736** | **687,829** | **387,205,495** | **90.6%** | **$63.87** |
