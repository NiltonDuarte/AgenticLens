# FIX-11 Iteration 02 Delta Plan: Root Configuration Harmonization and Fixed Token Ceilings

## Objective

Incorporate user steering feedback on FIX-11 by restructuring configuration hierarchies and standardizing context saturation rules:
1. Promote anomaly_thresholds from live to the root configuration level so that both offline anomaly detection (generate command via domain/anomalies.go) and real-time session tracking (live service via domain/live_anomaly.go) share identical rule thresholds.
2. Promote healthy_ceiling_tokens directly to root configuration level and eliminate the redundant metrics wrapper struct.
3. Replace ratio-based context ceiling approach evaluation with explicit, fixed token thresholds: context_ceiling_warning_tokens (default 160,000) and context_ceiling_critical_tokens (default 200,000). Apply these fixed thresholds consistently across EvaluateContextCeilingApproach in domain/live_anomaly.go and detectContextSaturationSpikes in domain/anomalies.go.

## User Steering Delta Requirements

1. Root Anomaly Thresholds:
Move anomaly_thresholds to root ProjectConfig so both offline fact ledger aggregations and live session evaluations read from the single source of truth.

2. Root Healthy Ceiling Tokens:
Place healthy_ceiling_tokens at root ProjectConfig (defaulting to 200,000) and remove MetricsConfig.

3. Fixed Token Ceilings:
Eliminate context_ceiling_approach_ratio_warning and context_ceiling_approach_ratio_critical. Introduce context_ceiling_warning_tokens (160,000) and context_ceiling_critical_tokens (200,000) in AnomalyThresholdsConfig. Update both live and offline evaluators to trigger warning alerts at warning tokens and critical alerts at critical tokens.

## Implementation Tasks

### Phase 1: Pure Domain Model and Rule Engine (FCIS Core)

1. Configuration Domain Model (backend/internal/domain/config.go):
- Remove MetricsConfig struct entirely.
- Add HealthyCeilingTokens int64 to ProjectConfig with struct tags for json and yaml.
- Relocate AnomalyThresholds AnomalyThresholdsConfig from LiveConfig to ProjectConfig root with struct tags for json and yaml.
- In AnomalyThresholdsConfig, replace ContextCeilingApproachRatioWarning and ContextCeilingApproachRatioCritical float64 fields with ContextCeilingWarningTokens int64 and ContextCeilingCriticalTokens int64.
- In DefaultProjectConfig, set HealthyCeilingTokens to 200,000 tokens, assign root AnomalyThresholds with default constants (including 160,000 warning tokens and 200,000 critical tokens), and keep LiveConfig focused solely on real-time stream parameters (ActiveThresholdSeconds, PollIntervalMs, MaxTraceBufferMessages).

2. Rule Constants and Live Anomaly Evaluators (backend/internal/domain/live_anomaly.go):
- Replace ContextCeilingApproachRatioWarning and ContextCeilingApproachRatioCritical constants with ContextCeilingWarningTokens (160,000) and ContextCeilingCriticalTokens (200,000).
- Update EvaluateContextCeilingApproach signature to accept fixed token thresholds (warningTokens int64, criticalTokens int64) rather than ratios. Fall back safely to ContextCeilingWarningTokens and ContextCeilingCriticalTokens when arguments are non-positive.
- In EvaluateContextCeilingApproach, evaluate peak context tokens directly against warningTokens and criticalTokens. If peak context reaches or exceeds criticalTokens, issue LiveSeverityCritical. If it reaches or exceeds warningTokens, issue LiveSeverityWarning. Format diagnostic messages with token amounts rather than ceiling ratios.
- Update LiveAnomalyOptions to retain Thresholds AnomalyThresholdsConfig and HealthyCeilingTokens int64, and update EvaluateLiveAnomalies to pass Thresholds.ContextCeilingWarningTokens and Thresholds.ContextCeilingCriticalTokens to EvaluateContextCeilingApproach.

3. Offline Anomaly Detection and Fact Aggregations (backend/internal/domain/anomalies.go & backend/internal/domain/events.go):
- Update detectContextSaturationSpikes to accept configurable token thresholds from AnomalyThresholdsConfig instead of hardcoding 160,000 tokens.
- In detectContextSaturationSpikes, trigger SeverityWarning when single-turn jumps exceed TokenVelocityWarningDeltaTokens (40,000) or turn context reaches ContextCeilingWarningTokens (160,000). Trigger SeverityCritical when token jumps exceed TokenVelocityCriticalDeltaTokens (80,000) or context reaches ContextCeilingCriticalTokens (200,000). Update v1 fallback to evaluate peak context against the warning and critical ceiling thresholds.
- Update DetectAnomalies, RefreshAnomalySummaries, DetectAnomaliesWithEvidence, and PersistAnomalySummaries to accept optional AnomalyThresholdsConfig arguments, falling back cleanly to DefaultProjectConfig().AnomalyThresholds when omitted.
- In backend/internal/domain/aggregations.go, replace all cfg.Metrics.HealthyCeilingTokens usages with cfg.HealthyCeilingTokens across ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, ComputeThreadSummaries, and DeriveMetricsLedger. Pass cfg.AnomalyThresholds from DeriveMetricsLedger into RefreshAnomalySummaries and DetectAnomalies so offline telemetry ledger generation uses configured root anomaly thresholds.

### Phase 2: Imperative Shell and Configuration Adapters

1. Configuration Loader (backend/internal/adapters/config/loader.go):
- Add HealthyCeilingTokens *int64 and AnomalyThresholds *rawAnomalyThresholdsConfig to rawProjectConfigDTO.
- In rawAnomalyThresholdsConfig, replace ratio fields with ContextCeilingWarningTokens *int64 and ContextCeilingCriticalTokens *int64.
- In rawLiveConfig, retain optional AnomalyThresholds for backward-compatible migration of legacy manifests that nested anomaly_thresholds under live.
- Retain rawMetricsConfig with HealthyCeilingTokens for backward compatibility with manifests specifying metrics.healthy_ceiling_tokens.
- In mergeConfig, populate cfg.HealthyCeilingTokens from root dto.HealthyCeilingTokens, falling back to dto.Metrics.HealthyCeilingTokens if root is not specified.
- In mergeConfig, populate cfg.AnomalyThresholds from root dto.AnomalyThresholds, falling back to dto.Live.AnomalyThresholds if root is omitted. If legacy ratio fields are detected, safely derive token equivalents using the effective healthy ceiling.

2. Configuration Writer (backend/internal/adapters/config/writer.go):
- Update GenerateConfigYAML and GenerateDefaultYAML to output healthy_ceiling_tokens at the root level, removing the metrics section entirely.
- Serialize anomaly_thresholds at the root level with updated fixed token fields (context_ceiling_warning_tokens and context_ceiling_critical_tokens), removing anomaly_thresholds from the live section.
- Include descriptive comments indicating root anomaly thresholds govern both offline generation and live stream monitoring.

3. Live Observability Service (backend/internal/adapters/live/service.go):
- In FetchActiveSessions and GetSessionTrace, update references from s.cfg.Live.AnomalyThresholds to s.cfg.AnomalyThresholds.
- In GetSessionTrace, update HealthyCeilingTokens assignment from s.cfg.Metrics.HealthyCeilingTokens to s.cfg.HealthyCeilingTokens.

4. Root Manifest (.agenticlens-config.yaml):
- Move healthy_ceiling_tokens: 200000 to the root configuration level and delete the metrics block.
- Move anomaly_thresholds to the root configuration level outside of live.
- Replace context_ceiling_approach_ratio_warning and context_ceiling_approach_ratio_critical with context_ceiling_warning_tokens: 160000 and context_ceiling_critical_tokens: 200000.

### Phase 3: Comprehensive Test Suite Updates

1. Domain Tests:
- backend/internal/domain/config_test.go: Update tests to assert root HealthyCeilingTokens and root AnomalyThresholds, verifying default values, JSON and YAML marshalling, and immutability.
- backend/internal/domain/live_anomaly_test.go: Update EvaluateContextCeilingApproach test cases to verify fixed warning and critical token boundaries, customized token thresholds, and non-positive value fallbacks.
- backend/internal/domain/anomalies_test.go: Update TestDetectAnomalies_ContextSaturationSpike to verify warning alerts trigger at context_ceiling_warning_tokens (160k) and critical alerts trigger at context_ceiling_critical_tokens (200k), with tests asserting custom threshold overrides.
- backend/internal/domain/aggregations_test.go: Update tests referencing cfg.Metrics.HealthyCeilingTokens to use cfg.HealthyCeilingTokens directly.

2. Adapter Tests:
- backend/internal/adapters/config/loader_test.go: Add tests verifying loading of root healthy_ceiling_tokens and root anomaly_thresholds. Add test cases verifying backward-compatible fallback from legacy metrics and live.anomaly_thresholds.
- backend/internal/adapters/config/writer_test.go: Verify generated YAML emits root healthy_ceiling_tokens and root anomaly_thresholds without a metrics block or live.anomaly_thresholds.
- backend/internal/adapters/live/service_test.go: Ensure live trace evaluations correctly consume root configuration thresholds.

## Architectural and Quality Invariants

- Functional Core, Imperative Shell (FCIS): All anomaly calculations and threshold evaluations remain purely in domain/ with zero I/O or system clock dependencies. Adapters handle YAML serialization and SQLite streaming.
- Static Typing and Zero Any: All new and updated configuration structs and evaluator signatures use concrete Go types with zero interface{} or any.
- Zero Emojis: Absolutely zero emojis in source code, comments, test strings, log messages, or documentation files.
- Backward Compatibility: Configuration loader transparently accepts legacy manifests featuring metrics.healthy_ceiling_tokens and live.anomaly_thresholds.
