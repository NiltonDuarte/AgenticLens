# FIX-11 Iteration 01 Plan: Configurable Ceilings, Enriched Anomaly Thresholds, and Source Roots Consolidation

## Goal

Decouple context health and live anomaly detection from hardcoded domain constants by introducing configurable metrics and enriched anomaly thresholds in .agenticlens-config.yaml, and eliminate architectural overlap between workspace artifact storage and source code classification by consolidating source directories into code_metrics.source_roots while preserving full backward compatibility.

## Grounded Findings

- backend/internal/domain/constants.go defines HealthyCeilingTokens as constant 200_000, and hardcodes 100_000 in GetBaseHealthStatus and GetHealthLabel to determine ContextHealthElevated. When repositories operate smaller or larger model context windows, this threshold cannot scale.
- backend/internal/domain/live_anomaly.go contains hardcoded constant thresholds for all four live anomaly rules: ConsecutiveToolErrors (3 warning, 5 critical), ContextCeilingApproach (0.80 warning, 1.00 critical), TokenVelocityRunaway (30k/60k token limits, 40k/80k delta limits), and ToolThrashing (3 warning, 5 critical). AnomalyThresholdsConfig in config.go currently exposes only stall, error rate, and context exhaustion percentage.
- backend/internal/domain/config.go houses ProjectConfig, WorkspacesConfig, AnomalyThresholdsConfig, and CodeMetricsConfig. WorkspacesConfig currently contains SrcFolders, which conceptually belongs to code change classification under CodeMetricsConfig rather than workspace artifact storage.
- backend/internal/adapters/config/loader.go strictly enforces workspaces.src_folders as a mandatory configuration field during ValidateConfig.
- backend/internal/adapters/config/writer.go serializes workspaces.src_folders and omits metrics or code_metrics.source_roots.
- backend/internal/adapters/git/classifier.go and analyzer.go rely on cfg.Workspaces.SrcFolders and ValidateSrcFolders to determine monitored source paths for commit diff analysis.
- .agenticlens-config.yaml contains workspaces.src_folders with web and backend, but lacks a metrics section and code_metrics.source_roots.

## Proposed Implementation Tasks

### Phase 1: Pure Domain Layer (FCIS Core)

1. backend/internal/domain/config.go:
   - Define MetricsConfig struct with HealthyCeilingTokens int64, using json and yaml struct tags healthy_ceiling_tokens.
   - Add Metrics MetricsConfig to ProjectConfig with tags metrics.
   - Enrich AnomalyThresholdsConfig with explicit fields: ConsecutiveErrorsStreakWarning int, ConsecutiveErrorsStreakCritical int, ContextCeilingApproachRatioWarning float64, ContextCeilingApproachRatioCritical float64, TokenVelocityWarningTokens int64, TokenVelocityCriticalTokens int64, TokenVelocityWarningDeltaTokens int64, TokenVelocityCriticalDeltaTokens int64, ToolThrashingMinFailures int, ToolThrashingCriticalFailures int.
   - Add SourceRoots []string to CodeMetricsConfig with tags source_roots.
   - Update DefaultProjectConfig to initialize Metrics.HealthyCeilingTokens to 200_000, AnomalyThresholds with default constants (3, 5, 0.80, 1.00, 30_000, 60_000, 40_000, 80_000, 3, 5), and CodeMetrics.SourceRoots to an empty string slice.
   - Update backend/internal/domain/config_test.go to verify new structs, default values, serialization tags, and slice mutation isolation.

2. backend/internal/domain/constants.go:
   - Refactor GetBaseHealthStatus(toks int64, ceiling int64) ContextHealthStatus: if ceiling <= 0, fallback to HealthyCeilingTokens. Evaluate ContextHealthElevated as toks >= ceiling / 2.
   - Refactor GetHealthLabel(toks int64, ceiling int64) string: if ceiling <= 0, fallback to HealthyCeilingTokens. Evaluate ContextHealthElevated as toks >= ceiling / 2.
   - Update backend/internal/domain/constants_test.go with tests covering custom ceiling values (e.g. 100k, 300k, 1M) verifying that elevated status dynamically triggers at ceiling / 2, while preserving exact behavior for default 200k ceiling and non-positive inputs.

3. backend/internal/domain/aggregations.go:
   - In DeriveMetricsLedger, determine effectiveCeiling from cfg.Metrics.HealthyCeilingTokens, falling back to HealthyCeilingTokens if <= 0.
   - Propagate effectiveCeiling into ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, ComputeThreadSummaries, and RefreshAnomalySummaries.
   - Support effectiveCeiling in ComputeWorkflowEfficiency by adding an optional ceiling parameter or helper function so that rootSaturationPct, cwhStatus, and ContextWindowHealth.HealthyCeilingTokens reflect the configured ceiling.
   - Support effectiveCeiling across role and model metrics calculation so ContextHealth is assigned based on the configured ceiling.
   - Update backend/internal/domain/aggregations_test.go to verify aggregation output when configured with custom ceiling thresholds.

4. backend/internal/domain/live_anomaly.go:
   - Update rule evaluators to accept configured thresholds with fallback to constants when values are <= 0:
     - EvaluateConsecutiveToolErrors accepts optional streak thresholds.
     - EvaluateContextCeilingApproach accepts optional ratio thresholds.
     - EvaluateTokenVelocityRunaway accepts optional warning/critical and delta thresholds.
     - EvaluateToolThrashing accepts optional failure counts.
     - EvaluateLiveAnomalies accepts an optional AnomalyThresholdsConfig or ProjectConfig parameter, routing thresholds down to each evaluator while maintaining safe defaults on nil or omitted inputs.
   - Update backend/internal/domain/live_anomaly_test.go to verify rule triggers under both default and customized threshold parameters.

### Phase 2: Imperative Shell and Adapters

1. backend/internal/adapters/config/loader.go:
   - Add rawMetricsConfig DTO with HealthyCeilingTokens *int64.
   - Enrich rawAnomalyThresholdsConfig DTO with the new optional threshold fields.
   - Add SourceRoots []string to rawCodeMetricsConfig DTO.
   - Add Metrics *rawMetricsConfig to rawProjectConfigDTO.
   - Update ValidateConfig: check for non-empty source directories across either cfg.CodeMetrics.SourceRoots or cfg.Workspaces.SrcFolders. If neither has non-empty entries, report code_metrics.source_roots as missing.
   - Update mergeConfig:
     - Map dto.Metrics into cfg.Metrics.
     - Map all enriched anomaly threshold fields when non-nil.
     - Map dto.CodeMetrics.SourceRoots.
     - Implement backward compatibility bridge: if CodeMetrics.SourceRoots is empty and Workspaces.SrcFolders is populated, populate CodeMetrics.SourceRoots from Workspaces.SrcFolders. If Workspaces.SrcFolders is empty and CodeMetrics.SourceRoots is populated, populate Workspaces.SrcFolders from CodeMetrics.SourceRoots.
   - Update backend/internal/adapters/config/loader_test.go:
     - Test parsing metrics and enriched anomaly thresholds.
     - Test parsing code_metrics.source_roots.
     - Test legacy YAML without code_metrics.source_roots (only workspaces.src_folders) smoothly populating SourceRoots.
     - Test modern YAML without workspaces.src_folders passing validation and populating both.
     - Test validation error when neither source root field is defined.

2. backend/internal/adapters/config/writer.go:
   - In GenerateConfigYAML:
     - Serialize metrics section with healthy_ceiling_tokens and explanatory comments.
     - Serialize enriched anomaly_thresholds fields under live section.
     - Serialize source_roots under code_metrics section.
     - Maintain workspaces.src_folders as optional legacy output or annotate as superseded by code_metrics.source_roots.
   - Update GenerateDefaultYAML to populate Metrics and CodeMetrics.SourceRoots.
   - Update backend/internal/adapters/config/writer_test.go to verify generated YAML syntax, round-trip loading, and comment annotations.

3. backend/internal/adapters/git/classifier.go & analyzer.go:
   - In ClassificationConfig, add SourceRoots []string.
   - In LoadConfig, populate SourceRoots from cfg.CodeMetrics.SourceRoots, falling back to cfg.Workspaces.SrcFolders. Keep SrcFolders populated for backward compatibility.
   - In analyzer.go, update ValidateSourceRoots (aliasing ValidateSrcFolders) and update IsMonitoredSourcePath to inspect cfg.SourceRoots in addition to cfg.SrcFolders.
   - Update ComputeCodeFactsAtRef to inspect cfg.SourceRoots or cfg.SrcFolders without spurious validation errors.
   - Update classifier_test.go and analyzer_test.go to verify source root classification.

4. backend/internal/adapters/live/service.go:
   - In GetSessionTrace, pass s.cfg.Live.AnomalyThresholds and s.cfg.Metrics.HealthyCeilingTokens to EvaluateLiveAnomalies.
   - Update live service tests to assert configured thresholds are used during live session analysis.

### Phase 3: Manifest and Integration

1. .agenticlens-config.yaml:
   - Add metrics section specifying healthy_ceiling_tokens: 200000.
   - Move web and backend directories into code_metrics.source_roots.
   - Retain or comment workspaces.src_folders to verify backward-compatible coexistence.
   - Add default live anomaly threshold values under live.anomaly_thresholds.

2. TypeScript Frontend and HTTP DTO:
   - Verify if ProjectConfigDTO in backend/internal/adapters/http/dto.go needs to expose metrics or code_metrics.
   - If exposed, update web/src/domain/types.ts (ProjectConfigResponse) and web/src/test/fixtures/mockData.ts accordingly.
   - Ensure frontend test suite (npm test) and build (npm run build) compile with zero errors.

### Phase 4: Verification Suite

1. Execute domain unit tests: go test -v ./backend/internal/domain/...
2. Execute adapter unit tests: go test -v ./backend/internal/adapters/...
3. Execute complete repository tests: make test
4. Verify zero code formatting issues: go fmt ./...

## Verification Approach

- Unit Tests:
  - Verify GetBaseHealthStatus and GetHealthLabel calculate elevated status at ceiling / 2 across 100k, 200k, 300k, and 1M ceilings.
  - Verify live anomaly rules trigger at user-configured thresholds and fallback safely to constants on zero or negative values.
  - Verify config loader parses new YAML sections and preserves backward compatibility for legacy configurations.
  - Verify git classifier and analyzer accurately detect modified files within code_metrics.source_roots.
- Boundary Tests:
  - Negative and zero values for healthy_ceiling_tokens fall back to 200_000.
  - Negative and zero values for anomaly thresholds fall back to domain constants.
  - Manifests with only workspaces.src_folders load cleanly into CodeMetrics.SourceRoots.
  - Manifests with only code_metrics.source_roots pass validation cleanly.
- Integration Verification:
  - Run make to verify single-binary embedded build and frontend compilation.
  - Confirm zero any or interface{} types in Go and TypeScript.
  - Confirm zero emojis across all code, tests, and documentation.

## Risk Analysis and Mitigations

- Risk: Existing repositories or fixtures only have workspaces.src_folders, causing validation failures if code_metrics.source_roots is strictly required.
  Mitigation: ValidateConfig checks both fields and mergeConfig performs bidirectional synchronization between CodeMetrics.SourceRoots and Workspaces.SrcFolders.
- Risk: Breaking existing callers of GetBaseHealthStatus or ComputeWorkflowEfficiency that assume 200k ceiling or hardcoded 100k elevated boundary.
  Mitigation: Keep default fallback to HealthyCeilingTokens when ceiling is non-positive, ensuring existing test assertions with default ceiling remain identical (ceiling / 2 of 200k is 100k).
- Risk: Live anomaly evaluator signature changes could break external callers or existing tests.
  Mitigation: Implement variadic or config-accepting function signatures that maintain full backward compatibility with zero-argument calls using domain defaults.
