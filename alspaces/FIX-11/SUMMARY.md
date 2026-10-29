# FIX-11 Summary: Critical Context Ceilings, Enriched Anomaly Thresholds, and Source Roots Consolidation

## Overview

This change enforces strict, configurable context ceiling thresholds under `anomaly_thresholds`, excising redundant `healthy_ceiling_tokens` across the codebase, establishes `ContextCeilingCriticalTokens` as the single authoritative source of truth for saturation calculations and health grading, enriches live anomaly detection thresholds, and requires canonical source roots under `code_metrics.source_roots` with no `workspaces.src_folders` fallback.

## Domain Changes

- Excised `healthy_ceiling_tokens` from `ProjectConfig`, `LiveAnomalyOptions`, and manifest templates.
- Enforced `AnomalyThresholds.ContextCeilingCriticalTokens` as the sole source of truth for context saturation percentage and health grading across all domain aggregations and live stream evaluations.
- Retained fixed token context ceiling thresholds: `ContextCeilingWarningTokens` (default 160,000) and `ContextCeilingCriticalTokens` (default 200,000).
- Dynamically calculated elevated context health threshold in `GetBaseHealthStatus` and `GetHealthLabel` as `ceiling / 2`.
- Propagated configured context ceilings and anomaly thresholds through aggregations (`ComputeWorkflowEfficiencyWithCeiling`, `ComputeRoleMetricsWithCeiling`, `ComputeModelMetrics`, `ComputeThreadSummariesWithCeiling`, `RefreshAnomalySummaries`, and `DeriveMetricsLedger`).
- Added `LiveAnomalyOptions` and configurable thresholds for all live anomaly rules (consecutive tool errors, token velocity tokens/deltas, tool thrashing counts) in `internal/domain/live_anomaly.go`.
- Added `SourceRoots []string` to `CodeMetricsConfig`.

## Adapter Changes

- `internal/adapters/config/loader.go`:
  - Excised `healthy_ceiling_tokens` and `rawMetricsConfig` with zero fallback.
  - Implemented fail-fast validation in `ValidateConfig`: immediately rejects configurations missing or with non-positive `context_ceiling_critical_tokens` or `context_ceiling_warning_tokens`, or where warning exceeds critical tokens.
  - Added root DTO fields for `anomaly_thresholds` and `code_metrics.source_roots`.
  - Requires non-empty `CodeMetrics.SourceRoots`; legacy `workspaces.src_folders` input does not satisfy validation.
- `internal/adapters/config/writer.go`:
  - Serializes root `anomaly_thresholds` and `code_metrics.source_roots`, cleanly omitting `healthy_ceiling_tokens`.
  - Deprecated and removed legacy `metrics` and `workspaces.src_folders` sections.
- `internal/adapters/git`:
  - Enriched `ClassificationConfig` with `SourceRoots`.
  - Implemented `ValidateSourceRoots`.
  - Updated `IsMonitoredSourcePath` and git diff analysis to respect source roots.
- `internal/adapters/live/service.go`:
  - Integrated configured root ceiling and anomaly thresholds into live trace analysis.
- `.agenticlens-config.yaml`:
  - Dropped `healthy_ceiling_tokens`.
  - Declared `context_ceiling_warning_tokens: 160000` and `context_ceiling_critical_tokens: 200000` under root `anomaly_thresholds`.
  - Consolidated source paths under `code_metrics.source_roots: [web, backend]`.

## Verification

- Backend tests with race detector (`go test -race -count=1 ./...`): PASS (12 packages)
- Frontend tests (`vitest run`): 32 test files, 734 tests passed: PASS
- Static analysis and lint (`go vet ./...`, `tsc --noEmit`, `eslint`): PASS
- Production build & multi-platform compilation (`make`): PASS

## Iteration 04: Source Roots Validation Guidance

- Initially updated the `generate` CLI validation message to identify the canonical field and legacy option; iteration 05 removes legacy acceptance and its fallback.
- Extended the CLI subprocess regression assertion; iteration 05 updates it to require only `code_metrics.source_roots`.
- Reviewer verdict: APPROVED (`iterations/04/REVIEW.md`).
- Focused CLI race test: PASS (`cd backend && GOCACHE=/tmp/agenticlens-go-cache go test -race ./cmd/agentic-lens`).
- Full backend race suite: PASS through Codex and OpenCode (`GOCACHE=/tmp/agenticlens-go-cache make test`, 12 Go packages).
- Post-merge `GOCACHE=/tmp/agenticlens-go-cache make`: PASS, including Go vet, TypeScript typecheck/lint, 12 Go race-test packages, 32 frontend test files (734 tests), frontend production build, and four platform builds.

## Iteration 05: Require Canonical Source Roots

- Removed `WorkspacesConfig.SrcFolders`, loader and writer bridging, classifier fallback, and legacy analyzer validation. Missing or legacy-only source roots now fail with a `code_metrics.source_roots` validation error.
- Updated YAML/JSON fixtures to prove canonical roots are required, `workspaces.src_folders` cannot supply or override them, and generated configuration omits the removed field while preserving workspace directories and artifacts.
- Updated README and setup instructions to require `code_metrics.source_roots`; amended the ticket contract to remove the backward-compatibility requirements.
- Reviewer verdict: APPROVED (`iterations/05/REVIEW.md`).
- Focused backend race tests and post-merge `GOCACHE=/tmp/agenticlens-go-cache make`: PASS. The full make included 12 Go packages, 32 frontend test files (734 tests), typecheck, lint, production build, and four platform builds.
- `GOCACHE=/tmp/agenticlens-go-cache make test` through OpenCode: PASS (12 Go packages and 734 frontend tests).

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 6
- **Active Agent Duration:** 5h 59m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 126.8k tokens (63% of 200k healthy ceiling, 6 compactions) | Subagents peak 174.1k (Coder)
- **Role Effort Distribution (Subagents):** Coder 99.6%, Advisor 0.4%
- **Subagent Turn Velocity & Thrashing:** 121.1 avg turns/run (peak: 273 turns) | Coder 137.1, Advisor 9.0
- **Test-to-Production Code Ratio:** 1.4:1 (+838 test lines / +610 prod lines)
- **Code Delivered:** +1,534 / -544 lines across 31 code files (excluding docs)
- **Reasoning Effort:** 271,488 reasoning tokens (58.4% of generated output)
- **Tool Health & Self-Correction:** 1,029 tool calls (16 self-corrections, 98.4% success rate)

### Anomalies & Health Flags
- **Status:** 13 execution anomalies detected (3 CRITICAL, 10 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+81.9k tokens) — Context tokens surged by 81.9k to 81.9k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+104.4k tokens) — Context tokens surged by 104.4k to 104.4k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+88.1k tokens) — Context tokens surged by 88.1k to 88.1k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.2k tokens) — Context tokens reached 163.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+63.9k tokens) — Context tokens surged by 63.9k to 63.9k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (172.4k tokens) — Context tokens reached 172.4k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+78.6k tokens) — Context tokens surged by 78.6k to 78.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+62.3k tokens) — Context tokens surged by 62.3k to 62.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (168.5k tokens) — Context tokens reached 168.5k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (168.0k tokens) — Context tokens reached 168.0k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (174.1k tokens) — Context tokens reached 174.1k, breaching the 160k warning ceiling.
- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: task (27.3%) — Tool 'task' failed 3 out of 11 calls (27.3% failure rate, threshold: >20%).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 2 | 1,884,827 | 12,687 | 4,511,658 | 70.5% | 126.8k | ELEVATED | $1.90 |
| **coder** | 7 | 6,415,490 | 179,886 | 72,309,823 | 91.9% | 174.1k | ELEVATED | $11.81 |
| **advisor** | 1 | 30,622 | 749 | 120,812 | 79.8% | 24.5k | HEALTHY | $0.04 |
| **Total** | **10** | **8,330,939** | **193,322** | **76,942,293** | **90.2%** | **174.1k** | **ELEVATED** | **$13.75** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 9 | 8,323,168 | 193,180 | 76,935,125 | 90.2% | $13.75 |
| **gpt-6-luna** | 1 | 7,771 | 142 | 7,168 | 48.0% | $0.00 |
| **Total** | **10** | **8,330,939** | **193,322** | **76,942,293** | **90.2%** | **$13.75** |
