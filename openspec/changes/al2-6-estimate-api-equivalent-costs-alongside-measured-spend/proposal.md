# AL2-6: Estimate API-equivalent costs alongside measured spend

## Roadmap ticket

[AL2-6: Estimate API-equivalent costs alongside measured spend](../../../alspaces/plan/to-do/AL2-6%20estimate%20api%20equivalent%20costs%20alongside%20measured%20spend.md)

## Why

Schema v6 records source-reported billing accurately, distinguishing observed zero, partial subtotals and unavailable cost. Subscription harnesses such as Codex and Claude Code still produce unavailable cost while consuming large token volumes, leaving operators unable to compare their workload against metered API sessions such as OpenCode. This change adds an explicit, user-configured rate card that estimates API-equivalent USD costs for eligible unavailable facts alongside measured spend, without overwriting billing evidence or treating estimates as invoices.

## What changes

- **Shared rate-card configuration (`model_pricing.rates`)**: Accept an optional shared `model_pricing.rates` list in `.agenticlens-config.yaml` / `.agenticlens-config.yml` with `model`, `input_per_million`, `cache_read_per_million` and `output_per_million`. Support exact model identifiers and configured glob patterns with deterministic precedence, finite nonnegative rate validation, fail-closed ambiguity diagnostics and a canonical SHA-256 effective catalog identity. No built-in market prices are assumed.
- **Pure domain estimation and provenance**: Estimate USD for eligible unavailable measured-cost invocation facts with valid nonnegative token categories using `(input * input_per_million + cache_read * cache_read_per_million + output * output_per_million) / 1,000,000` without double-charging cached input. Never overwrite observed spend (including genuine `$0.00`) and never double-count partial measured-cost facts. Compute aggregate provenance strictly as:
  - `measured` when 100% of facts are measured (`total_equivalent_cost_usd = measured_cost_usd`),
  - `estimated` when 100% of facts are estimated (`total_equivalent_cost_usd = estimated_cost_usd`),
  - `hybrid` when all facts are either measured or estimated (`total_equivalent_cost_usd = measured + estimated`, complete rollup),
  - `partial` or `unavailable` when any facts remain unpriced (incomplete coverage; `total_equivalent_cost_usd = null`), and
  - `empty` when there are zero invocation facts.
- **Portable persistence without read-time repricing**: Persist invocation-level applied estimates and rate snapshots (with catalog identity) as additive optional fields on schema v6 ledgers during generation and source-backed refresh/backfill. Ordinary reads and cross-machine projections render saved estimates as recorded facts without needing the producing machine's local catalog and without silently repricing history.
- **CLI, Markdown and HTTP projections**: Expose `measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`, provenance and coverage through generation, computed exports and HTTP projections while preserving existing `cost_usd` and `cost_summary` measured-spend contracts. Render explicit measured vs. `EST` distinctions and applied rate details in CLI tables and `delivery.md` summaries.
- **Web dashboard presentation and CacheWrite caveat**: Show `EST` badges and applied rate details on Ticket Deep Dive model breakdowns and cost summaries, and display a visible UI warning whenever API-equivalent estimates are shown stating that `CacheWrite` tokens are not currently considered when a model carries cache-write surcharges (for example GPT-5 or GPT-6).

**Compatibility**: Additive optional fields keep existing schema v6 ledgers valid with no migration required. Ledgers generated without estimates serialize identically to current v6 ledgers. Engines released before this change reject ledgers that contain saved estimate fields via strict JSON decoding (`DisallowUnknownFields`); release and delivery notes must call this out.

**Exclusions**: Global cost display mode switching (`cost_mode=measured|estimated|hybrid`) across Fleet/Progression/Comparisons and the interactive rate-card simulation drawer belong to AL2-7. Persisting `cache_write` / `cache_creation` token categories, external price feeds, built-in model rate tables and ingesting Claude Code's `cost-state` estimates are out of scope.

## Capabilities

- New: [api-equivalent-cost-estimation](specs/api-equivalent-cost-estimation/spec.md)

## Impact

- **Users:** Operators and teams mixing subscription harnesses (Codex, Claude Code) with metered harnesses (OpenCode) gain explicit, auditable API-equivalent cost estimates alongside measured spend, with clear coverage provenance and a CacheWrite surcharge caveat. Projects without `model_pricing.rates` keep existing measured-only behavior.
- **Components:** Shared configuration schema/loader/validator (`internal/domain`, `internal/adapters/config`), pure domain pricing and aggregation engine (`internal/domain`), fact ledger DTOs and validation (`internal/adapters`), CLI generation/backfill and Markdown formatters (`cmd/agentic-lens`, `internal/adapters/formatters.go`), HTTP projection DTOs (`internal/adapters/http`), web types/validation/Deep Dive and cost UI (`web/src`), embedded web assets, and cost/operator documentation (`docs/COST_SEMANTICS.md`).
- **Preserved guarantees:** Observed `$0.00` remains observed; partial measured subtotals are never overwritten or double-counted; unmatched models remain explicitly unavailable; unavailable-owner and other-machine sessions survive refresh untouched; GET requests never mutate or reprice persisted ledgers.
