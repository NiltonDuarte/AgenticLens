# FIX-11 Iteration 03 Delta Plan: Excision of healthy_ceiling_tokens and Critical Ceiling Enforcement

## Objective

Incorporate user steering feedback on ticket FIX-11 by executing a complete architectural cleanup:
1. Completely excise healthy_ceiling_tokens across the entire codebase, including .agenticlens-config.yaml, ProjectConfig in domain/config.go, writer.go, and loader.go. Provide zero backward-compatible fallback for healthy_ceiling_tokens in loader.go.
2. Enforce fail-fast validation in ValidateConfig within loader.go if anomaly_thresholds.context_ceiling_critical_tokens is unconfigured or non-positive, and validate that context_ceiling_warning_tokens is configured, positive, and does not exceed context_ceiling_critical_tokens.
3. Establish AnomalyThresholds.ContextCeilingCriticalTokens as the single, authoritative source of truth for context window saturation percentage and health status grading across all domain aggregations, live stream monitoring, and offline anomaly detection.
4. Update all test suites across domain, adapters, and CLI commands to align with the complete removal of healthy_ceiling_tokens and the mandatory enforcement of context_ceiling_critical_tokens.

## User Steering Delta Requirements

1. Complete Excision of healthy_ceiling_tokens:
Remove healthy_ceiling_tokens from .agenticlens-config.yaml, the configuration writer, the configuration loader, and the ProjectConfig domain struct. Remove rawMetricsConfig and any dto.Metrics references. Ensure loader.go provides zero fallback for healthy_ceiling_tokens.

2. Fail-Fast Configuration Validation:
In loader.go, update ValidateConfig to strictly require anomaly_thresholds.context_ceiling_critical_tokens. If context_ceiling_critical_tokens is not set or is less than or equal to 0, immediately reject configuration loading with a ConfigValidationError. Furthermore, validate anomaly_thresholds.context_ceiling_warning_tokens so that it must be positive and cannot exceed context_ceiling_critical_tokens. In mergeConfig, do not inject default fallback values when context_ceiling_critical_tokens is omitted from configuration input, ensuring unconfigured files fail validation immediately.

3. Single Source of Truth for Context Saturation and Health Grading:
AnomalyThresholds.ContextCeilingCriticalTokens serves as the sole source of truth across the entire system. Context window saturation percentage calculations (in ComputeWorkflowEfficiencyWithCeiling) and health status grading (in GetBaseHealthStatus and GetHealthLabel) must evaluate directly against ContextCeilingCriticalTokens. Remove all secondary ceiling pathways and fallback bridges.

4. Comprehensive Test Suite Alignment:
Update domain unit tests, adapter tests, and CLI end-to-end integration tests to remove assertions expecting healthy_ceiling_tokens, add test coverage for strict ceiling validation, and ensure all test fixtures provide explicit context_ceiling_warning_tokens and context_ceiling_critical_tokens.

## Implementation Tasks

### Phase 1: Pure Domain Model and Aggregation Engine (FCIS Core)

1. Configuration Domain Model (backend/internal/domain/config.go):
- Remove HealthyCeilingTokens int64 from ProjectConfig.
- Update DefaultProjectConfig() to remove HealthyCeilingTokens initialization.
- Refactor EffectiveHealthyCeiling() on ProjectConfig to return c.AnomalyThresholds.ContextCeilingCriticalTokens, falling back to ContextCeilingCriticalTokens domain constant (200,000) only when non-positive.
- Ensure AnomalyThresholdsConfig retains ContextCeilingWarningTokens int64 and ContextCeilingCriticalTokens int64 with explicit JSON and YAML struct tags.

2. Health Grading and Saturation Aggregations (backend/internal/domain/aggregations.go):
- In DeriveMetricsLedger, read effectiveCeiling directly from cfg.AnomalyThresholds.ContextCeilingCriticalTokens, establishing AnomalyThresholds.ContextCeilingCriticalTokens as the single source of truth.
- In ComputeWorkflowEfficiencyWithConfig, pass cfg.AnomalyThresholds.ContextCeilingCriticalTokens into ComputeWorkflowEfficiencyWithCeiling.
- In ComputeWorkflowEfficiencyWithCeiling, compute rootSatPct as the percentage of rootPeakVal relative to ceiling (ContextCeilingCriticalTokens), and grade overall context health via GetBaseHealthStatus(overallPeak, ceiling).
- Set ContextWindowHealth.HealthyCeilingTokens to ceiling (ContextCeilingCriticalTokens) to preserve downstream ledger schema parity.
- Ensure ComputeTotals, ComputeRoleMetricsWithCeiling, ComputeModelMetrics, ComputeThreadSummariesWithCeiling, and CalculateRoleEffortRollup consume ContextCeilingCriticalTokens for all context health classifications.

3. Live Anomaly Options and Evaluators (backend/internal/domain/live_anomaly.go):
- In LiveAnomalyOptions struct, remove HealthyCeilingTokens int64.
- In EvaluateLiveAnomalies, remove logic attempting to fall back to opt.HealthyCeilingTokens. Read ContextCeilingCriticalTokens and ContextCeilingWarningTokens directly from thresholds (AnomalyThresholdsConfig).
- In EvaluateContextCeilingApproach, continue evaluating peak context tokens directly against warningTokens and criticalTokens, returning LiveSeverityCritical when reaching or exceeding criticalTokens and LiveSeverityWarning when reaching or exceeding warningTokens.

4. Offline Anomaly Detection (backend/internal/domain/anomalies.go):
- In detectContextSaturationSpikes, enforce warning alerts at thresholds.ContextCeilingWarningTokens and critical alerts at thresholds.ContextCeilingCriticalTokens without referencing healthy_ceiling_tokens.
- In RefreshAnomalySummaries and DetectAnomaliesWithEvidence, pass cfg.AnomalyThresholds as the single source of truth for saturation spike rules.

5. Domain Constants and Formatters (backend/internal/domain/constants.go & backend/internal/adapters/formatters.go):
- In backend/internal/domain/constants.go, retain ContextCeilingCriticalTokens (or HealthyCeilingTokens alias as an internal fallback constant 200,000) solely to defend pure domain helper functions (GetBaseHealthStatus, GetHealthLabel) against zero or negative function arguments.
- In backend/internal/adapters/formatters.go, format context health indicators using domain.ContextCeilingCriticalTokens rather than legacy healthy ceiling constants.

### Phase 2: Imperative Shell, Configuration Adapters, and Manifest

1. Configuration Loader (backend/internal/adapters/config/loader.go):
- Remove HealthyCeilingTokens *int64 from rawProjectConfigDTO.
- Remove rawMetricsConfig struct and Metrics *rawMetricsConfig from rawProjectConfigDTO.
- In mergeConfig, delete all code mapping dto.HealthyCeilingTokens or dto.Metrics.HealthyCeilingTokens to cfg. Zero fallback for healthy_ceiling_tokens.
- In mergeConfig, do not inject default fallback values if rawAT is nil or if rawAT.ContextCeilingCriticalTokens is nil or non-positive; assign 0 so ValidateConfig can detect omission. Similarly, do not inject default fallback for ContextCeilingWarningTokens if omitted.
- In ValidateConfig, add strict validation checks:
  - If cfg.AnomalyThresholds.ContextCeilingCriticalTokens <= 0, append anomaly_thresholds.context_ceiling_critical_tokens to missing fields list in ConfigValidationError.
  - If cfg.AnomalyThresholds.ContextCeilingWarningTokens <= 0, append anomaly_thresholds.context_ceiling_warning_tokens to missing fields list in ConfigValidationError.
  - If cfg.AnomalyThresholds.ContextCeilingWarningTokens > cfg.AnomalyThresholds.ContextCeilingCriticalTokens, return a validation error stating that context_ceiling_warning_tokens cannot exceed context_ceiling_critical_tokens.

2. Configuration Writer (backend/internal/adapters/config/writer.go):
- In GenerateConfigYAML, completely remove the healthy_ceiling_tokens serialization block and its explanatory comments.
- Ensure anomaly_thresholds block serializes context_ceiling_warning_tokens and context_ceiling_critical_tokens with descriptive comments explaining that context_ceiling_critical_tokens governs context saturation percentage and health status grading across the entire system.
- In GenerateDefaultYAML, remove HealthyCeilingTokens from default struct initialization. Populate AnomalyThresholds with default warning tokens (160,000) and critical tokens (200,000).

3. Live Observability Service (backend/internal/adapters/live/service.go):
- In GetSessionTrace, remove HealthyCeilingTokens assignment when instantiating domain.LiveAnomalyOptions. Pass domain.LiveAnomalyOptions with Thresholds: s.cfg.AnomalyThresholds.

4. Root Project Manifest (.agenticlens-config.yaml):
- Delete healthy_ceiling_tokens: 200000 from the root level of .agenticlens-config.yaml.
- Verify that anomaly_thresholds section explicitly declares context_ceiling_warning_tokens: 160000 and context_ceiling_critical_tokens: 200000.

### Phase 3: Comprehensive Test Suite Updates

1. Domain Unit Tests:
- backend/internal/domain/config_test.go:
  - Remove all assertions checking cfg.HealthyCeilingTokens.
  - Assert that DefaultProjectConfig() sets AnomalyThresholds.ContextCeilingWarningTokens to 160,000 and ContextCeilingCriticalTokens to 200,000.
  - Assert JSON and YAML serialization and deserialization of AnomalyThresholdsConfig.
- backend/internal/domain/live_anomaly_test.go:
  - Remove HealthyCeilingTokens field initializations from LiveAnomalyOptions across all test cases.
  - Verify EvaluateLiveAnomalies evaluates ceiling approach alerts strictly from AnomalyThresholds.ContextCeilingCriticalTokens and ContextCeilingWarningTokens.
- backend/internal/domain/aggregations_test.go:
  - Update tests that previously set cfg.HealthyCeilingTokens to set cfg.AnomalyThresholds.ContextCeilingCriticalTokens instead.
  - Verify that ContextWindowHealth, RootSaturationPct, and ContextHealth grading reflect the configured ContextCeilingCriticalTokens.

2. Adapter Tests:
- backend/internal/adapters/config/loader_test.go:
  - Update testProjectConfig(), testValidSuffixYAML, and testValidSuffixJSON to declare anomaly_thresholds with context_ceiling_warning_tokens (160000) and context_ceiling_critical_tokens (200000).
  - Update existing YAML string fixtures throughout loader_test.go to include the mandatory anomaly_thresholds ceiling tokens.
  - Add explicit test case: ValidateConfig rejects configuration when anomaly_thresholds.context_ceiling_critical_tokens is missing.
  - Add explicit test case: ValidateConfig rejects configuration when context_ceiling_critical_tokens is zero or negative.
  - Add explicit test case: ValidateConfig rejects configuration when anomaly_thresholds.context_ceiling_warning_tokens is missing or non-positive.
  - Add explicit test case: ValidateConfig rejects configuration when context_ceiling_warning_tokens exceeds context_ceiling_critical_tokens.
  - Add explicit test case: configuration file specifying healthy_ceiling_tokens but omitting context_ceiling_critical_tokens fails validation, proving zero fallback exists.
  - In TestLoadFromFile, remove assertion for cfg.HealthyCeilingTokens == 200000 and replace with assertion for cfg.AnomalyThresholds.ContextCeilingCriticalTokens == 200000.
- backend/internal/adapters/config/writer_test.go:
  - Verify GenerateConfigYAML outputs YAML containing context_ceiling_critical_tokens and context_ceiling_warning_tokens under anomaly_thresholds, and zero occurrences of healthy_ceiling_tokens.
  - Verify round-trip generation and parsing succeeds and strictly preserves context_ceiling_critical_tokens.
- backend/internal/adapters/live/service_test.go:
  - Verify session trace anomaly evaluation operates with LiveAnomalyOptions containing only Thresholds.

3. CLI Integration Tests:
- backend/cmd/agentic-lens/main_test.go:
  - Audit all inline .agenticlens-config.yaml test configurations (cfgContent and variants) and ensure they include anomaly_thresholds with context_ceiling_warning_tokens: 160000 and context_ceiling_critical_tokens: 200000.
  - Verify CLI commands (generate, scan, live, validate) execute cleanly with the updated configuration schema.
  - Add CLI integration test verifying that running agentic-lens generate with a configuration missing context_ceiling_critical_tokens fails with a descriptive validation error message.

## Architectural and Quality Invariants

- Functional Core, Imperative Shell (FCIS):
  - Pure domain rules in backend/internal/domain/ derive context health and saturation percentages from immutable function parameters without I/O or environment state.
  - Adapters in backend/internal/adapters/config/ handle YAML decoding and fail-fast validation boundaries.
- Single Source of Truth:
  - AnomalyThresholds.ContextCeilingCriticalTokens is the sole parameter dictating context window saturation percentage and health status across the entire application lifecycle.
- Zero Fallback:
  - No translation, aliasing, or fallback exists in loader.go for healthy_ceiling_tokens or unconfigured context_ceiling_critical_tokens. Missing configurations fail immediately.
- Strict Typing with Zero Any:
  - Concrete int64 types are enforced across all structs and evaluator signatures. Zero interface{} or any types in Go backend and TypeScript frontend.
- Zero Emojis:
  - Strict absence of emojis across all source code, comments, test strings, log output, and workspace artifacts.

## Verification Approach

- Step 1: Run domain unit tests to verify pure model updates, aggregation math, and live anomaly evaluation:
  go test -v -race ./backend/internal/domain/...
- Step 2: Run adapter unit tests to verify strict loader validation, writer output, and live service:
  go test -v -race ./backend/internal/adapters/...
- Step 3: Run CLI integration tests to verify end-to-end command behavior:
  go test -v -race ./backend/cmd/agentic-lens/...
- Step 4: Run full repository verification suite:
  make test
- Step 5: Verify zero formatting defects and static analysis pass:
  go fmt ./...
  go vet ./...
- Step 6: Verify TypeScript frontend build and tests remain green:
  npm test
  npm run build

## Risk Analysis and Mitigations

- Risk: Existing test fixtures throughout loader_test.go and main_test.go omit anomaly_thresholds and fail ValidateConfig immediately.
  Mitigation: Update testValidSuffixYAML, testValidSuffixJSON, testProjectConfig(), and inline YAML test helpers in a single coordinated pass to include valid ceiling tokens.
- Risk: Downstream telemetry consumers or frontend dashboard components expect healthy_ceiling_tokens in ContextWindowHealth.
  Mitigation: Preserve ContextWindowHealth.HealthyCeilingTokens on the ledger model while assigning its value directly from cfg.AnomalyThresholds.ContextCeilingCriticalTokens, preserving schema compatibility without maintaining the deprecated configuration field.
- Risk: Inadvertent regression where unconfigured files pass validation due to DefaultProjectConfig defaults in loader.go.
  Mitigation: Explicitly zero out or check ContextCeilingCriticalTokens during DTO decoding in loader.go prior to ValidateConfig, backed by negative unit tests asserting validation failure on omitted fields.
