# FIX-11 Summary: Critical Context Ceilings, Enriched Anomaly Thresholds, and Source Roots Consolidation

## Overview

This change enforces strict, configurable context ceiling thresholds under `anomaly_thresholds`, excising redundant `healthy_ceiling_tokens` across the codebase, establishes `ContextCeilingCriticalTokens` as the single authoritative source of truth for saturation calculations and health grading, enriches live anomaly detection thresholds, and consolidates source roots under `code_metrics.source_roots` in `.agenticlens-config.yaml` with backward-compatible bridging for `workspaces.src_folders`.

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
  - Bidirectional bridging between `CodeMetrics.SourceRoots` and `Workspaces.SrcFolders`.
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


### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 3h 42m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 106.0k tokens (53% of 200k healthy ceiling, 4 compactions) | Subagents peak 174.1k (Coder)
- **Role Effort Distribution (Subagents):** Coder 99.4%, Advisor 0.6%
- **Subagent Turn Velocity & Thrashing:** 105.8 avg turns/run (peak: 256 turns) | Coder 125.2, Advisor 9.0
- **Test-to-Production Code Ratio:** 0.5:1 (+241 test lines / +445 prod lines)
- **Code Delivered:** +772 / -170 lines across 21 code files (excluding docs)
- **Reasoning Effort:** 202,197 reasoning tokens (60.5% of generated output)
- **Tool Health & Self-Correction:** 690 tool calls (11 self-corrections, 98.4% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (1 CRITICAL, 6 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+104.4k tokens) — Context tokens surged by 104.4k to 104.4k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.2k tokens) — Context tokens reached 163.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+63.9k tokens) — Context tokens surged by 63.9k to 63.9k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (172.4k tokens) — Context tokens reached 172.4k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+78.6k tokens) — Context tokens surged by 78.6k to 78.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+62.3k tokens) — Context tokens surged by 62.3k to 62.3k tokens.
- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: task (33.3%) — Tool 'task' failed 3 out of 9 calls (33.3% failure rate, threshold: >20%).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 1,266,654 | 8,336 | 3,122,514 | 71.1% | 106.0k | ELEVATED | $1.30 |
| **coder** | 5 | 4,007,427 | 122,801 | 45,218,948 | 91.9% | 174.1k | ELEVATED | $7.52 |
| **advisor** | 1 | 30,622 | 749 | 120,812 | 79.8% | 24.5k | HEALTHY | $0.04 |
| **Total** | **7** | **5,304,703** | **131,886** | **48,462,274** | **90.1%** | **174.1k** | **ELEVATED** | **$8.87** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 7 | 5,304,703 | 131,886 | 48,462,274 | 90.1% | $8.87 |
| **Total** | **7** | **5,304,703** | **131,886** | **48,462,274** | **90.1%** | **$8.87** |
