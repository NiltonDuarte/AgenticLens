# FIX-11 Contract: Configurable Ceilings, Enriched Anomaly Thresholds, and Source Roots Consolidation

## Problem Statement

1. HealthyCeilingTokens (200,000) and related context health thresholds (specifically the hardcoded 100,000 token boundary for elevated context health) along with live anomaly rule constants (consecutive tool execution errors, token velocity surges, and tool thrashing iterations) are hardcoded in Go domain constants rather than configurable via .agenticlens-config.yaml. As repositories operate models with varying context windows (such as 128k, 200k, or 1M tokens), health labels and anomaly alerts cannot adapt to repository-specific operational targets.
2. workspaces.src_folders conceptually conflicts and overlaps with code_metrics.include, code_metrics.prod_patterns, and code_metrics.test_patterns. The workspaces configuration section is architecturally intended for workspace artifact storage and lifecycle tracking (such as dir: alspaces and artifacts definitions), not source code classification. Source folder root boundaries belong under code_metrics as source_roots, leaving workspaces dedicated solely to workspace artifact management.

## Intended Outcome

1. Add a dedicated MetricsConfig domain model under ProjectConfig to allow configuring healthy_ceiling_tokens via YAML, defaulting to 200,000 when omitted or non-positive.
2. Dynamically calculate the elevated context health threshold in GetBaseHealthStatus and GetHealthLabel relative to the effective healthy ceiling (ceiling / 2) rather than the fixed 100,000 token constant.
3. Enrich AnomalyThresholdsConfig with explicit fields for all live anomaly rules (consecutive tool error streak thresholds, context ceiling approach ratios, token velocity runaway limits and deltas, and tool thrashing counts), falling back safely to domain defaults when omitted or non-positive.
4. Consolidate source code root boundaries under code_metrics.source_roots, migrating workspaces.src_folders while maintaining backward compatibility for existing configurations.
5. Update git classification and diff analysis to respect code_metrics.source_roots while continuing to tolerate legacy workspaces.src_folders definitions during transition.
6. Update .agenticlens-config.yaml to declare the metrics section and consolidated code_metrics configuration.

## Existing System Context

- backend/internal/domain/constants.go defines HealthyCeilingTokens as 200_000 and hardcodes 100_000 for ContextHealthElevated evaluation in GetBaseHealthStatus and GetHealthLabel.
- backend/internal/domain/config.go defines ProjectConfig, WorkspacesConfig with SrcFolders, AnomalyThresholdsConfig (currently limited to stall, error rate, and context exhaustion percentage), and CodeMetricsConfig without source roots.
- backend/internal/domain/live_anomaly.go contains hardcoded rule constants: ConsecutiveToolErrorsStreakWarning (3), ConsecutiveToolErrorsStreakCritical (5), ContextCeilingApproachRatioWarning (0.80), ContextCeilingApproachRatioCritical (1.00), TokenVelocityWarningTokens (30,000), TokenVelocityCriticalTokens (60,000), TokenVelocityWarningDeltaTokens (40,000), TokenVelocityCriticalDeltaTokens (80,000), ToolThrashingMinFailures (3), and ToolThrashingCriticalFailures (5).
- backend/internal/domain/aggregations.go references HealthyCeilingTokens across ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, ComputeThreadSummaries, and RefreshAnomalySummaries.
- backend/internal/adapters/config/loader.go validates workspaces.src_folders as mandatory and lacks parsing for metrics or code_metrics.source_roots.
- backend/internal/adapters/config/writer.go writes workspaces.src_folders and omits metrics.
- backend/internal/adapters/git/classifier.go and analyzer.go check cfg.Workspaces.SrcFolders and ValidateSrcFolders.
- web/src/domain/types.ts defines ProjectConfigResponse for the frontend configuration client.

## Core Architectural Invariants

1. Functional Core, Imperative Shell (FCIS):
   - Pure domain models in backend/internal/domain/config.go define ProjectConfig, MetricsConfig, enriched AnomalyThresholdsConfig, and CodeMetricsConfig.
   - Pure aggregations and live anomaly detection in backend/internal/domain/ accept configured ceilings and thresholds, cleanly defaulting to safe constants when non-positive or unconfigured.
   - GetBaseHealthStatus and GetHealthLabel compute elevated status dynamically as ceiling / 2 rather than using a hardcoded 100,000 token literal.
   - Adapters in backend/internal/adapters/config/ handle serialization, deserialization, validation, and backward-compatible fallbacks without polluting pure domain functions with I/O or YAML decoding mechanics.
   - Git analyzer and classifier in backend/internal/adapters/git/ inspect consolidated source roots and glob patterns cleanly.
2. Typing and Zero Any:
   - Zero interface{} or any permitted across Go backend and TypeScript frontend codebases.
   - Strong typing for all configuration fields, thresholds, and DTOs.
3. Backward Compatibility:
   - Config loading seamlessly accepts legacy manifests that supply workspaces.src_folders without code_metrics.source_roots.
   - Config validation passes if either code_metrics.source_roots or workspaces.src_folders provides valid source directories.
   - If both are provided, code_metrics.source_roots takes precedence.
   - Existing telemetry ledgers and live trace evaluators continue functioning identically under default configuration.
4. Zero Emojis:
   - No emojis permitted in code, comments, docstrings, commit messages, PR drafts, UI strings, logs, or workspace artifacts.

## Scope Boundaries

- Go Backend:
  - backend/internal/domain/config.go:
    - Add MetricsConfig struct with HealthyCeilingTokens int64.
    - Add Metrics MetricsConfig to ProjectConfig.
    - Enrich AnomalyThresholdsConfig with fields: ConsecutiveErrorsStreakWarning int, ConsecutiveErrorsStreakCritical int, ContextCeilingApproachRatioWarning float64, ContextCeilingApproachRatioCritical float64, TokenVelocityWarningTokens int64, TokenVelocityCriticalTokens int64, TokenVelocityWarningDeltaTokens int64, TokenVelocityCriticalDeltaTokens int64, ToolThrashingMinFailures int, ToolThrashingCriticalFailures int.
    - Add SourceRoots []string to CodeMetricsConfig.
    - Update DefaultProjectConfig to populate safe defaults matching domain constants.
  - backend/internal/domain/constants.go:
    - Update GetBaseHealthStatus and GetHealthLabel to evaluate ContextHealthElevated against ceiling / 2 when ceiling > 0, falling back to HealthyCeilingTokens / 2 when ceiling <= 0.
  - backend/internal/domain/aggregations.go:
    - Propagate configured ceiling to WorkflowEfficiency, ContextWindowHealth, RoleMetrics, ModelMetrics, ThreadSummaries, and Anomaly summaries.
    - Maintain default constant fallback when ceiling <= 0.
  - backend/internal/domain/live_anomaly.go:
    - Update live anomaly evaluators (EvaluateConsecutiveToolErrors, EvaluateContextCeilingApproach, EvaluateTokenVelocityRunaway, EvaluateToolThrashing, EvaluateLiveAnomalies) to accept configured thresholds or fall back to established constants when <= 0.
  - backend/internal/adapters/config/loader.go:
    - Add DTO fields for metrics, enriched anomaly thresholds, and code_metrics.source_roots.
    - Update ValidateConfig to require source roots from either code_metrics.source_roots or workspaces.src_folders.
    - Implement fallback in mergeConfig: if code_metrics.source_roots is empty, copy from workspaces.src_folders; if workspaces.src_folders is empty, copy from code_metrics.source_roots.
  - backend/internal/adapters/config/writer.go:
    - Generate YAML with metrics section and code_metrics.source_roots.
    - Deprecate workspaces.src_folders in comments and omit or maintain as optional legacy compatibility.
  - backend/internal/adapters/git/classifier.go & analyzer.go:
    - Support SourceRoots on ClassificationConfig.
    - Update path monitoring and validation to inspect SourceRoots, falling back to SrcFolders.
  - backend/internal/adapters/live/service.go:
    - Pass configured live anomaly thresholds and metrics ceiling to domain evaluators.
  - .agenticlens-config.yaml:
    - Declare metrics section with healthy_ceiling_tokens: 200000.
    - Declare source_roots under code_metrics: [web, backend].
    - Remove or deprecate src_folders under workspaces.
  - Comprehensive unit and integration test coverage across all modified packages.
- TypeScript Frontend:
  - web/src/domain/types.ts and web/src/test/fixtures/mockData.ts:
    - Update ProjectConfigResponse and mock fixtures if backend /api/config exposes the new fields.
- Out of Scope:
  - Database schema changes to SQLite event stores.
  - Removing workspaces.dir or workspaces.artifacts (these remain core workspace settings).
  - Telemetry ledger format v4 changes (ledger schema version remains 4).

## Acceptance Criteria

1. Metrics Configuration: ProjectConfig supports a metrics section containing healthy_ceiling_tokens int64. DefaultProjectConfig initializes it to 200,000 tokens.
2. Dynamic Health Status: When ceiling is configured (e.g. 100,000 or 500,000), GetBaseHealthStatus and GetHealthLabel identify ContextHealthElevated starting at ceiling / 2, ContextHealthBloated above ceiling, and ContextHealthCritical above 2 * ceiling. Non-positive ceiling defaults to 200,000.
3. Live Anomaly Rule Configuration: Live anomaly detection rules for consecutive tool errors, ceiling approach, token velocity, and tool thrashing accept configurable thresholds from AnomalyThresholdsConfig, defaulting safely to domain constants when values are zero or negative.
4. Source Roots Migration: CodeMetricsConfig includes source_roots []string. Configuration loading accepts code_metrics.source_roots and seamlessly falls back to workspaces.src_folders for legacy configurations.
5. Config Validation: ValidateConfig succeeds when source roots are specified in either code_metrics.source_roots or workspaces.src_folders, and fails only when neither contains non-empty entries.
6. Git Diff Classification: Git change analysis in backend/internal/adapters/git correctly uses configured source_roots to monitor modified paths without regressions in line counts or file categories.
7. YAML Serialization Parity: GenerateConfigYAML outputs valid YAML matching the updated schema with metrics and code_metrics.source_roots. Round-trip serialization and loading preserves all fields.
8. Zero Any and Zero Emojis: All Go code and TypeScript types maintain strict typing with zero any or interface{}, and zero emojis exist across the codebase, tests, and artifacts.
9. Verification: Full test suite (make, go test ./..., npm test) passes with zero regressions.
