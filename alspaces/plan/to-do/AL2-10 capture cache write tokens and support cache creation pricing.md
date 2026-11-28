# AL2-10: Capture cache write tokens and support cache creation pricing

## Goal

Operators can track cache creation and write tokens across supported harnesses and apply 4-tier API-equivalent rate cards with provider-specific cache creation rates without underestimating prompt caching setup costs.

## Context

[AL2-6](AL2-6%20estimate%20api%20equivalent%20costs%20alongside%20measured%20spend.md) establishes 3-tier API-equivalent rate-card estimation (`uncached_input`, `cache_read`, `output`). Prompt caching pricing varies across providers and individual models:

- Anthropic Claude charges an explicit 25% surcharge (1.25x base input rate) for `cache_creation_input_tokens`, while cache reads receive a 90% discount.
- OpenAI models exhibit mixed pricing: while many standard text models (e.g. base GPT-4o / o1) have billed cache creation at regular base input rates ($0 write surcharge) with discounted cache reads, certain OpenAI models, modalities, or configurations charge dedicated cache write fees. Other providers (such as DeepSeek) feature their own caching structures.

Today `domain.InvocationTokens` stores only `Input`, `Output`, `CacheRead`, `Reasoning`, and `PeakContext`. Adapters drop cache-write telemetry at the common invocation boundary: Codex parses `cache_write_input_tokens` into `usage.CacheWrite`, Claude Code decodes `cache_creation_input_tokens` into `rawUsage.CacheCreation` only for context sizing, and OpenCode schema v2 records `tokens.cache.write` (schema v1 captures only `tokens.cache.read`). Because ledger decoders enforce `DisallowUnknownFields()` via `newStrictFactDecoder`, adding `cache_write` to `agentic_metrics.json` requires a coordinated schema transition rather than an unversioned field addition.

## Acceptance criteria

- **Configuration**: `model_pricing.rates` accepts optional `cache_write_per_million` as a finite nonnegative number, allowing operators to specify model-specific cache creation rates (e.g. Anthropic's 1.25x premium or OpenAI models with write surcharges). When omitted for a model, it defaults to `input_per_million` (matching baseline input billing for models without write surcharges).
- **Cross-harness extraction**: Populate `Tokens.CacheWrite` from native sources: Codex maps `usage.CacheWrite`, Claude Code maps `rawUsage.CacheCreation` (`cache_creation_input_tokens`), and OpenCode maps schema v2 `tokens.cache.write` (defaulting to `0` for schema v1). Uncached input calculation must not double-count cache write tokens.
- **Domain, rollups, and schema versioning**: Add `CacheWrite int64` to `InvocationTokens` and `InvocationTokensDTO`, and aggregate `CacheWriteTokens` across `GroupedMetric`, `ModelMetric`, `RoleMetric`, `TimeBucketMetric`, and `FleetSummary`. Address strict JSON decoding (`newStrictFactDecoder` / `DisallowUnknownFields`) through a formal Schema v7 transition or approved additive schema boundary that preserves portable cross-machine decoding of historical ledgers.
- **4-tier estimation formula**: Estimate API-equivalent cost for eligible unavailable measured-cost facts as `(uncached_input * rate_in + cache_write * rate_cache_write + cache_read * rate_cache_read + output * rate_out) / 1,000,000`, preserving AL2-6 observed-cost precedence, provenance, coverage, and unknown-value semantics.
- **Persistence and presentation**: Persist `cache_write` in fact ledgers and expose cache write tokens and applied cache creation rates in CLI table/Markdown formatters, HTTP projections, and web Deep Dive / model breakdown tooltips.

## Implementation notes

Extends the 3-tier rate-card contract in [AL2-6](AL2-6%20estimate%20api%20equivalent%20costs%20alongside%20measured%20spend.md). Downstream [AL2-7](AL2-7%20compare%20cost%20modes%20and%20simulate%20rate%20cards.md) rate-card simulation should incorporate `cache_write_per_million` once this contract lands. Developer contract should confirm whether Codex `input_tokens` includes `cache_write_input_tokens` in native rollouts and preserve the [portable ledger contract](../../../TECHNICAL_GUIDELINES.md#portable-ledger-contract).

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md).
