# AgenticLens Architecture, Code & Test Review Report

## 1. Executive Summary
A comprehensive re-audit of the AgenticLens repository was conducted following the synchronization with latest main (commit f8ffbc9, incorporating tickets AL-8, AL-13, AL-21, AL-22, and PO-5). The audit evaluated the Go backend packages (`backend/`), TypeScript web frontend (`web/`), and OpenCode agent/skill definitions (`.opencode/`) against the four review tiers defined in `.opencode/skills/agentic-lens-review/SKILL.md`.

The codebase demonstrates exemplary engineering standards: strict domain purity with zero external or I/O imports in `backend/internal/domain`, zero `any` or `interface{}` across production Go and TypeScript code, and strict adherence to the zero-emoji invariant.

However, several high-leverage architectural friction points and performance bottlenecks remain:
1. Leaky Adapter Boundary: Filesystem modification logic (`UpdateSummaryFile`) resides in `backend/internal/adapters/formatters.go` rather than within repository persistence adapters.
2. Unnecessary O(N) Traversal: `WorkspaceDetailHandler` in `backend/internal/adapters/http/handlers.go` executes a full directory scan of all workspaces (`ScanWorkspaces`) to fetch details for a single ticket.
3. Skill and Agent Path Drift: Agent personas (`.opencode/agent/*.md`) and skills (`.opencode/skills/agentic-lens-dev/SKILL.md`) still reference obsolete local developer machine paths (`/Users/guedesduarte/...`) and omit the `backend/` directory prefix.
4. Test Suite Execution Delays: `backend/cmd/agentic-lens/main_test.go` executes CLI subprocesses serially with repeated 100ms sleep delays in polling loops, and Go unit tests lack parallel execution declarations (`t.Parallel()`).
5. Global Frontend JSDOM Overhead: Pure mathematical domain tests in `web/src/domain/` run under heavy JSDOM emulation instead of the native node environment.

---

## 2. Test Suite Performance Diagnostics
- **Total Backend Suite Execution Time**: ~2.5s
- **Total Frontend Suite Execution Time**: ~1.7s
- **Target Execution Time**: < 1.0s (Backend) / < 1.0s (Frontend)

### Slowest Tests & Bottlenecks
| Duration | Package | Test Name | Diagnosed Bottleneck | Recommended Optimization |
|---|---|---|---|---|
| 0.44s | `backend/cmd/agentic-lens` | `TestCLIGenerate_AdversarialEdgeCases` | Subprocess compilation and serial execution | Parallelize independent subtests; reuse binary |
| 0.33s | `backend/cmd/agentic-lens` | `TestCLIGenerateCommand` | Serial subprocess execution | Add `t.Parallel()` and fast fixture reuse |
| 0.20s | `backend/cmd/agentic-lens` | `TestCLIServerPortFlagPrecedence` | Subprocess spawning and serial 100ms readiness sleep | Direct port resolution unit testing + 10ms poll interval |
| 0.19s | `backend/cmd/agentic-lens` | `TestCLIAnomaliesCommand` | CLI subprocess execution with file fixtures | Optimize fixture creation and reduce sleep overhead |
| 0.10s | `backend/cmd/agentic-lens` | `TestCLITicketFilterAndMarkdownOutput` | Subprocess startup with 100ms sleep | Immediate trial GET before backoff sleep |
| 0.07s | `backend/internal/adapters/git` | `TestGitAnalyzer_AdversarialEdgeCases` | Multiple subprocess git invocations in loop | Memory cache stub and parallel subtests |
| 0.05s | `backend/internal/adapters/backfill` | `TestBackfillEngine_DryRunPreservesFiles` | Serial disk workspace scanning | Parallel execution |

---

## 3. Architecture & Code Smells by Severity

### Tier 1: Layering & Boundary Discipline (Go + TypeScript FCIS)
- **Go FCIS Purity**: All packages in `backend/internal/domain` strictly adhere to the functional core pattern. Imports are limited to standard library utilities (`fmt`, `math`, `sort`, `strings`, `time`), with zero external dependencies, zero I/O imports (`os`, `io`, `net/http`), and zero database drivers.
- **Clock Injection**: Time is properly injected at the outer boundary. In HTTP handlers and CLI entrypoints, `time.Now()` is sampled at handler entry and passed as an explicit parameter (`now time.Time`) to domain migration, configuration, and aggregation functions.
- **Leaky Adapter Boundary**: `backend/internal/adapters/formatters.go` contains `UpdateSummaryFile`, which performs filesystem reads and writes (`os.ReadFile`, `os.OpenFile`, `os.WriteFile`). Formatting functions must remain pure string transformations. Persistence logic belongs in `backend/internal/adapters/repository.go`.
- **TypeScript Pure Domain Separation**: `web/src/domain/` encapsulates pure mathematical calculation and formatting functions (`formatters.ts`, `progression.ts`, `events.ts`). UI components in `web/src/components/` and views in `web/src/views/` import pure utilities without leaking UI framework concerns into domain logic.

### Tier 2: Module Depth & Structural Friction
- **Single-Item Workspace Detail Query**: `backend/internal/adapters/http/handlers.go` in `WorkspaceDetailHandler` calls `ScanWorkspaces` to read all workspaces simply to retrieve one ticket, instead of loading the requested ticket directly via a targeted repository function (`LoadTicketWorkspace`). This results in O(N) disk I/O on single-item queries.
- **Formatters Indirection**: In `backend/internal/adapters/formatters.go`, functions such as `FormatLeadTime` and `FormatTokens` merely wrap standard library calls without domain value, adding unnecessary layers of indirection.
- **Duplicated Aggregations**: `backend/internal/domain/aggregations.go` implements `AggregateMetricsOverTime(ledgers, bucketSize)`, which buckets tickets into daily/weekly/monthly slices. Simultaneously, `web/src/domain/progression.ts` implements `aggregateMetricsOverTime(tickets, bucketSize)` with identical date-parsing and aggregation logic in TypeScript.

### Tier 3: Domain Modeling & Fowler Code Smells
- **Strict Typing Compliance**: Zero instances of `any` or `interface{}` exist in Go domain logic or TypeScript production components. All multi-state representations use discriminated unions or typed interfaces.
- **Zero Emojis Mandate**: The entire repository, including Go source, TSX views, CSS, test files, and markdown docs, is 100% free of emojis.
- **Data Clumps & Primitive Obsession**: In `backend/internal/domain/aggregations.go`, `ComputeToolBreakdown` repeatedly unpacks three related metrics: `Calls`, `Errors`, and `ErrorRate`. Grouping these into a dedicated `ToolReliabilityStats` struct eliminates repetitive argument lists.
- **Chart Palette Duplication**: In `web/src/views/MetricProgression.tsx` and `web/src/components/SaturationBar.tsx`, raw hex colors (`#3b82f6`, `#10b981`, `#f59e0b`, `#ef4444`) are hardcoded inline instead of referencing centralized CSS custom properties or design token constants.
- **Agent & Skill Instructions Drift**: Personas in `.opencode/agent/` (`planner.md`, `coder.md`, `tester.md`, `reviewer.md`, `orchestrator.md`) specify layers as `internal/domain/`, `internal/adapters/`, and `cmd/agentic-lens/`, omitting the `backend/` directory prefix required by the actual repository layout. `.opencode/agent/planner.md` and `.opencode/skills/agentic-lens-dev/SKILL.md` contain hardcoded local machine paths (`/Users/guedesduarte/workspace/projects/SinkLayer/AgenticLens`).

### Tier 4: Test Suite Performance, Execution Speed & Flakiness
- **Unconditional Sleep Delays**: In `backend/cmd/agentic-lens/main_test.go`, server readiness polling loops execute `time.Sleep(100 * time.Millisecond)` before issuing HTTP requests. Replacing this with immediate trial requests and low-latency retry backoff (10ms) eliminates over 400ms of idle wait per server test.
- **Absence of Parallel Testing**: Zero tests in `backend/internal/domain/` or `backend/internal/adapters/` call `t.Parallel()`. Pure mathematical rollup tests in `aggregations_test.go`, `anomalies_test.go`, `config_test.go`, and `version_test.go` are completely stateless and can run concurrently.
- **Frontend Test Environment Overhead**: `web/vite.config.ts` sets `environment: 'jsdom'` globally. Pure domain test suites (`src/domain/progression.test.ts`, `src/domain/formatters.test.ts`, `src/domain/events.test.ts`) incur setup penalty initializing browser DOM emulation that is never utilized. Adding per-file environment pragmas (`// @vitest-environment node`) bypasses DOM emulation for pure unit tests.

---

## 4. Top Remediation Recommendations

| Priority | Category | Target Files | Complexity | Architectural Value |
|---|---|---|---|---|
| P1 | FCIS Boundary | `backend/internal/adapters/formatters.go`, `backend/internal/adapters/repository.go` | Low | High: Isolates filesystem I/O to repository adapter; pure formatters become side-effect free. |
| P1 | Performance | `backend/internal/adapters/http/handlers.go`, `backend/internal/adapters/repository.go` | Low | High: Replaces full workspace scans in `WorkspaceDetailHandler` with targeted single-ticket lookups (`LoadTicketWorkspace`). |
| P1 | Performance | `backend/cmd/agentic-lens/main_test.go` | Medium | High: Eliminates unconditional test sleep overhead; optimizes server readiness polling. |
| P1 | Performance | `backend/internal/domain/*_test.go`, `backend/internal/adapters/*_test.go` | Low | High: Enables parallel execution across stateless table tests via `t.Parallel()`. |
| P2 | Configuration | `.opencode/agent/*.md`, `.opencode/skills/**/*.md` | Low | High: Fixes broken path references (`backend/` prefix) and removes hardcoded developer machine roots. |
| P2 | Performance | `web/src/domain/*.test.ts`, `web/vite.config.ts` | Low | Medium: Bypasses heavy JSDOM initialization on pure calculation test suites. |
| P3 | Architecture | `backend/internal/domain/aggregations.go`, `web/src/domain/progression.ts` | Medium | Medium: Establishes a single source of truth for time-series aggregation between backend and frontend. |
| P3 | Design System | `web/src/views/MetricProgression.tsx`, `web/src/components/SaturationBar.tsx` | Low | Medium: Consolidates hardcoded hex colors into design token variables. |
