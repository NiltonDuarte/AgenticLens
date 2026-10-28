# FIX-11 Summary: Configurable Ceilings, Enriched Anomaly Thresholds, and Source Roots Consolidation

## Overview

This change introduces configurable healthy ceiling tokens, enriches live anomaly detection thresholds, and consolidates source roots under `code_metrics.source_roots` in `.agenticlens-config.yaml` while ensuring backward compatibility with legacy `workspaces.src_folders`.

## Domain Changes

- Added `MetricsConfig` struct under `ProjectConfig` with `HealthyCeilingTokens int64`, defaulting to 200,000 tokens when non-positive.
- Dynamically calculated elevated context health threshold in `GetBaseHealthStatus` and `GetHealthLabel` as `ceiling / 2` instead of hardcoded 100,000.
- Propagated configured context ceilings through aggregations (`ComputeWorkflowEfficiency`, `ComputeRoleMetrics`, `ComputeModelMetrics`, `ComputeThreadSummaries`, `RefreshAnomalySummaries`, and `DeriveMetricsLedger`).
- Added `LiveAnomalyOptions` and configurable thresholds for all live anomaly rules (consecutive tool errors, ceiling approach ratios, token velocity tokens/deltas, tool thrashing counts) in `internal/domain/live_anomaly.go`, falling back safely to domain defaults.
- Added `SourceRoots []string` to `CodeMetricsConfig`.

## Adapter Changes

- `internal/adapters/config/loader.go`:
  - Added DTO fields for `metrics`, `code_metrics.source_roots`, and enriched anomaly thresholds.
  - Bidirectional bridging between `CodeMetrics.SourceRoots` and `Workspaces.SrcFolders`.
  - Updated `ValidateConfig` to accept source definitions from either location.
- `internal/adapters/config/writer.go`:
  - Serializes `metrics.healthy_ceiling_tokens`, enriched anomaly thresholds, and `code_metrics.source_roots`.
  - Deprecated legacy `workspaces.src_folders`.
- `internal/adapters/git`:
  - Enriched `ClassificationConfig` with `SourceRoots`.
  - Implemented `ValidateSourceRoots`.
  - Updated `IsMonitoredSourcePath` and git diff analysis to respect source roots.
- `internal/adapters/live/service.go`:
  - Integrated configured ceiling and anomaly thresholds into live trace analysis.
- `.agenticlens-config.yaml`:
  - Defined `metrics` with `healthy_ceiling_tokens: 200000`.
  - Configured explicit thresholds under `live.anomaly_thresholds`.
  - Consolidated source paths under `code_metrics.source_roots: [web, backend]`.

## Verification

- Backend tests with race detector (`go test -race ./...`): PASS
- Frontend tests (`vitest run`): 32 test files, 734 tests passed: PASS
- Static analysis and lint (`make check`): PASS
- Single binary and multi-platform build (`make build`): PASS

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 2
- **Active Agent Duration:** 1h 8m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 79.6k tokens (40% of 200k healthy ceiling, 2 compactions) | Subagents peak 174.1k (Coder)
- **Role Effort Distribution (Subagents):** Coder 98.7%, Advisor 1.3%
- **Subagent Turn Velocity & Thrashing:** 78.3 avg turns/run (peak: 237 turns) | Coder 101.3, Advisor 9.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 106,371 reasoning tokens (62.9% of generated output)
- **Tool Health & Self-Correction:** 356 tool calls (7 self-corrections, 98.0% success rate)

### Anomalies & Health Flags
- **Status:** 4 execution anomalies detected (1 CRITICAL, 3 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (172.4k tokens) — Context tokens reached 172.4k, breaching the 160k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+78.6k tokens) — Context tokens surged by 78.6k to 78.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+62.3k tokens) — Context tokens surged by 62.3k to 62.3k tokens.
- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: task (42.9%) — Tool 'task' failed 3 out of 7 calls (42.9% failure rate, threshold: >20%).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 517,158 | 4,315 | 2,004,702 | 79.5% | 79.6k | HEALTHY | $0.62 |
| **coder** | 3 | 1,526,681 | 57,706 | 22,595,425 | 93.7% | 174.1k | ELEVATED | $3.38 |
| **advisor** | 1 | 30,622 | 749 | 120,812 | 79.8% | 24.5k | HEALTHY | $0.04 |
| **Total** | **5** | **2,074,461** | **62,770** | **24,720,939** | **92.3%** | **174.1k** | **ELEVATED** | **$4.04** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 5 | 2,074,461 | 62,770 | 24,720,939 | 92.3% | $4.04 |
| **Total** | **5** | **2,074,461** | **62,770** | **24,720,939** | **92.3%** | **$4.04** |
