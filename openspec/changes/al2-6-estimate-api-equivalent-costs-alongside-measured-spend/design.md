# AL2-6: Technical design

## Scope and boundaries

This change adds configured API-equivalent USD cost estimation alongside source-reported measured spend. Operators declare a shared `model_pricing.rates` catalog in `.agenticlens-config.yaml` / `.agenticlens-config.yml`. During generation and source-backed refresh, AgenticLens estimates USD for eligible unavailable-cost invocations using uncached input, cache-read and output tokens, persists each applied estimate and rate snapshot on the invocation fact, and projects `measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`, provenance, coverage and applied rate details across CLI Markdown, HTTP APIs and the Ticket Deep Dive dashboard (including a visible `CacheWrite` surcharge warning).

Acceptance is owned by [api-equivalent-cost-estimation](specs/api-equivalent-cost-estimation/spec.md). Product intent remains in [AL2-6](../../../alspaces/plan/to-do/AL2-6%20estimate%20api%20equivalent%20costs%20alongside%20measured%20spend.md).

**Exclusions:**
- Global cost display modes (`cost_mode=measured|estimated|hybrid`) across Fleet, Progression and Comparisons, plus the interactive rate-card simulation drawer (owned by AL2-7).
- Persisting or extracting `cache_write` / `cache_creation` token counts from source transcripts.
- Built-in default pricing tables, external price lookups or ingesting Claude Code's `cost-state` estimate.
- Schema v7 migration: estimates use additive optional fields on schema v6.

## Decisions and ownership

### 1. Configuration schema, matching precedence and catalog identity

- **Shared config (`backend/internal/domain/config.go`, `backend/internal/adapters/config/policy_nodes.go`, `loader.go`)**:
  - Add optional `ModelPricing ModelPricingConfig` (`yaml:"model_pricing,omitempty"`) to `domain.ProjectConfig` with `Rates []ModelPricingRate`.
  - Each `ModelPricingRate` declares `Model string`, `InputPerMillion float64`, `CacheReadPerMillion float64`, and `OutputPerMillion float64`.
  - Validation requires `model_pricing` (when present) to be a YAML mapping, `rates` to be a YAML sequence, every `model` to be nonblank, every rate to be finite and `>= 0`, valid glob syntax, and no duplicate normalized `model` entries.
  - Absence of `model_pricing` or `rates: []` is valid and disables estimation cleanly.
- **Canonicalization and deterministic matching (`backend/internal/domain/pricing.go`)**:
  - Normalize both invocation model names and configured rules with `strings.ToLower(strings.TrimSpace(value))`.
  - Support exact identifiers (no `*` or `?`) and pure-domain glob patterns (`*` matches any rune sequence including `/`, `?` matches any single rune, so provider-prefixed names like `litellm/gemini-3.8-flash` match predictably).
  - Precedence:
    1. Exact normalized match.
    2. Matching glob pattern ranked by highest literal (non-wildcard) character count, then longest normalized pattern length.
  - **Fail-closed ambiguity**: If two or more matching glob rules tie on literal character count and normalized length, return a typed ambiguity error naming the colliding patterns rather than picking by config order.
- **Effective catalog identity**:
  - Canonicalize all rules (normalized model, sorted ascending, rates formatted with `strconv.FormatFloat(normalizeZero(v), 'f', -1, 64)`), encode deterministically and hash with SHA-256 into `sha256:<64-hex-chars>`.

### 2. Pure domain estimation engine and token-category alignment

- **Verified adapter token semantics**:
  - Codex (`backend/internal/adapters/codex/rollout.go`): `Tokens.Input` already subtracts cached input (`usage.Input - usage.CacheRead`), so `Input` is uncached prompt tokens and `CacheRead` is cached input tokens.
  - OpenCode and Claude Code (`schema_v1_session_extraction.go`, `schema_v2_session_extraction.go`, `claude/session_facts.go`): `Tokens.Input` is uncached input, `Tokens.CacheRead` is cached input, and `Tokens.Reasoning` is a subset of `Tokens.Output`.
- **Formula**: `(float64(t.Input)*rate.InputPerMillion + float64(t.CacheRead)*rate.CacheReadPerMillion + float64(t.Output)*rate.OutputPerMillion) / 1_000_000.0`, rounded to 6 decimal places when persisted on an invocation fact (`math.Round(amount*1e6)/1e6`).
- **Eligibility guardrails**:
  - Only `cost_evidence.state == "unavailable"` facts with nonnegative `input`, `cache_read` and `output` tokens and nonblank `model` are eligible.
  - `observed` facts (including genuine `$0.00`) are NEVER overwritten or estimated.
  - `partial` facts (coalesced billed and unbilled turns in one token bucket) retain `known_subtotal_usd` without estimating unsegmented tokens, preventing double-counting.
  - Unmatched models or negative/invalid token counts produce no estimate and remain `unavailable`.

### 3. Provenance and equivalent-cost rollup contract

- Introduce typed `CostProvenance` (`empty`, `measured`, `estimated`, `hybrid`, `partial`, `unavailable`) and `EquivalentCostSummary` in `backend/internal/domain/cost.go`:
  - `Provenance`:
    - `empty` when `Total == 0`.
    - `measured` when 100% of facts are `observed` (`Measured == Total`); `TotalEquivalentCostUSD = MeasuredCostUSD`.
    - `estimated` when 100% of facts are `estimated` (`Estimated == Total`); `TotalEquivalentCostUSD = EstimatedCostUSD`.
    - `hybrid` when all facts are either `observed` or `estimated` with at least one of each (`Measured + Estimated == Total`); `TotalEquivalentCostUSD = MeasuredCostUSD + EstimatedCostUSD` (complete rollup).
    - `partial` when any fact remains unpriced (`PartialMeasured > 0 || Unavailable > 0`) and at least one measured subtotal or estimated amount exists; `TotalEquivalentCostUSD = nil` (incomplete coverage).
    - `unavailable` when all facts remain unpriced and no measured subtotal or estimate exists; all amounts `nil`.
  - Component fields:
    - `MeasuredCostUSD *float64`: non-null (including `0`) when at least one `observed` or `partial` measured fact exists; `nil` otherwise.
    - `EstimatedCostUSD *float64`: non-null (including `0`) when at least one `estimated` fact exists; `nil` otherwise.
    - `TotalEquivalentCostUSD *float64`: non-null only for complete provenance (`measured`, `estimated`, `hybrid`); `nil` for `partial`, `unavailable`, `empty`.
    - `KnownSubtotalUSD *float64`: combined known measured subtotal + known estimated sum when at least one priced/partial fact exists; `nil` for `unavailable` and `empty`.
    - Coverage counts: `Measured`, `Estimated`, `PartialMeasured`, `Unavailable`, `Total` (sum to `Total`).
- Preserve existing `cost_usd` and `cost_summary` unchanged so AL-57 measured-only contracts and AL2-7 default measured mode stay intact.

### 4. Additive schema v6 persistence and portable history

- Follow the repository precedent set by `provisional_seconds` and `source: "claude"` (AL-98): keep `domain.CurrentSchemaVersion = 6` and add optional `cost_estimate` (`omitempty`) on `InvocationFact` / `InvocationFactDTO`:
  ```json
  "cost_estimate": {
    "amount_usd": 1.0,
    "rate": {
      "matched_pattern": "gpt-5.4",
      "input_per_million": 2.0,
      "cache_read_per_million": 0.25,
      "output_per_million": 8.0,
      "effective_catalog_id": "sha256:..."
    }
  }
  ```
- **Why invocation-level persistence**: Fleet, role, model, progression and comparison views slice invocations by arbitrary filters; invocation-level estimates support truthful re-aggregation across every projection without repricing on read.
- **Boundary validation (`backend/internal/adapters/invocation_fact_dto.go`, `fact_ledger_dto.go`)**:
  - When `cost_estimate` is present, enforce `cost_evidence.state == "unavailable"`, finite nonnegative `amount_usd` and per-million rates, and nonblank `matched_pattern` and `effective_catalog_id`.
- **Write vs. read authority**:
  - Estimation runs during `generate` (`cmd/agentic-lens`) and source-backed refresh/backfill (`internal/adapters/backfill`) only on newly extracted / completely refreshed owners before saving `agentic_metrics.json`.
  - Preserved sessions (unavailable owners and other-harness sessions) keep their persisted `cost_estimate` untouched.
  - Ordinary reads (`DeriveMetricsLedger`, CLI rendering, HTTP GET) read saved `cost_estimate` directly and NEVER consult local `model_pricing.rates` to reprice or rewrite history.
- **Compatibility notice**: Existing v6 ledgers require no migration, and new ledgers without estimates serialize identically to current v6. Older v6 binaries reject ledgers that contain `cost_estimate` via `DisallowUnknownFields`; delivery and release notes must document this.

### 5. CLI, Markdown, HTTP and Web presentation

- **CLI / Markdown (`backend/internal/adapters/formatters.go`)**:
  - When no estimates exist in a row/ledger, keep existing `FormatCostSummary` output unchanged.
  - When estimates exist, display explicit measured vs. `EST` values (`$6.74 EST`, `$23.20 measured + $7.12 EST ($30.32 equiv)`, or partial coverage with known subtotal and unpriced counts) and append a concise applied-rates note under `#### By Model` listing each applied pattern, `in / cache_read / out` per 1M USD, `effective_catalog_id`, and a note that `CacheWrite` tokens are not included.
- **HTTP & export DTOs (`backend/internal/adapters/metric_projection_dto.go`, `metrics_export_dto.go`, `backend/internal/adapters/http/{workspace,fleet,progression,comparison}_dto.go`)**:
  - Expose `measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`, `cost_provenance` and `equivalent_cost_summary` alongside unchanged `cost_usd` and `cost_summary`.
  - Expose `applied_rates` (deduplicated sorted `[]AppliedRateSnapshot`) on `ModelMetricDTO` so Deep Dive renders exact rate details per model without scanning raw session rows.
- **Web dashboard (`web/src/domain/types.ts`, `web/src/domain/cost.ts`, `web/src/api/costValidation.ts`, `web/src/components/CostKPICard.tsx`, `web/src/views/TicketDeepDive.tsx`)**:
  - Enforce strict TypeScript types (zero `any`) and extend `costValidation.ts` to validate every new nullable amount, provenance enum, coverage invariant and applied rate snapshot.
  - In `TicketDeepDive.tsx` Model Resource Consumption table: render an accessible `EST` badge next to estimated/hybrid model costs with an `InfoTooltip` showing the applied `input / cache_read / output` per 1M rates, matched pattern and catalog identity.
  - In `CostKPICard.tsx` / Deep Dive cost summary: show measured spend alongside estimated / hybrid equivalent breakdown when estimates exist without presenting estimates as bills.
  - **CacheWrite UI warning**: Whenever any displayed cost includes API-equivalent estimates, render a clear, accessible UI warning banner/notice stating that `CacheWrite` tokens are currently not considered in API-equivalent estimates when a model has cache-write surcharges (for example GPT-5 or GPT-6).

### 6. Writable-path ownership

- **Coder**:
  - `backend/internal/domain/` (`config.go`, `cost.go`, `pricing.go`, `models.go`, `aggregations.go`)
  - `backend/internal/adapters/config/` (`policy_nodes.go`, `loader.go`, `writer.go` / templates if updated)
  - `backend/internal/adapters/` (`invocation_fact_dto.go`, `fact_ledger_dto.go`, `metric_projection_dto.go`, `metrics_export_dto.go`, `formatters.go`, `backfill/`)
  - `backend/internal/adapters/http/` (`workspace_dto.go`, `fleet_dto.go`, `progression_dto.go`, `comparison_dto.go`, `handlers.go`)
  - `backend/cmd/agentic-lens/` (`main.go` and generation/backfill wiring)
  - `web/src/` (`domain/types.ts`, `domain/cost.ts`, `domain/formatters.ts`, `api/costValidation.ts`, `api/validation.ts`, `components/CostKPICard.tsx`, `views/TicketDeepDive.tsx`, styles) and rebuilt embedded assets via `make web-assets`
  - `docs/COST_SEMANTICS.md` (document API-equivalent estimate semantics alongside measured spend)
- **Tester**:
  - Unit, adapter, HTTP and CLI tests across `backend/internal/domain/*_test.go`, `backend/internal/adapters/**/*_test.go`, `backend/cmd/agentic-lens/*_test.go`
  - Frontend domain, validation and component/view tests across `web/src/**/*.test.{ts,tsx}`
  - Standalone-binary and embedded-asset verification
- **Reviewers**: Read-only codebase investigation; reports under `openspec/changes/al2-6-estimate-api-equivalent-costs-alongside-measured-spend/evidence/` only when explicitly assigned.
- **Orchestrator**: Integration, final telemetry, roadmap update (only upon verified completion), `delivery.md` and PR delivery.

## Implementation approval

Approved by the user on 2026-11-26 through the orchestrator's approval question ("Approve as written (Recommended)"). The approval covers this contract and [specs/api-equivalent-cost-estimation/spec.md](specs/api-equivalent-cost-estimation/spec.md) as written: invocation-level additive schema v6 `cost_estimate` persistence (no v7 migration), deterministic exact-before-glob matching with SHA-256 `effective_catalog_id`, exact aggregate provenance rules (`empty`, `measured`, `estimated`, `hybrid`, `partial`, `unavailable`), combined CLI/Markdown Cost column with applied-rates footnote, and Deep Dive `EST` badges + rate tooltips + visible `CacheWrite` surcharge warning. Baseline revision: `133100bb` on branch `feat/AL2-6-estimate-api-equivalent-costs`.

## Implementation approach

1. Build pure domain pricing config, glob matcher, SHA-256 catalog identity, estimation engine and `EquivalentCostSummary` provenance rollups first with exhaustive table-driven unit tests.
2. Wire shared YAML config parsing/validation and additive schema v6 `cost_estimate` DTO decoding/validation.
3. Integrate estimation at generation and source-backed refresh boundaries (preserving unavailable owners and read-time immutability), then extend CLI Markdown formatting and HTTP/export DTOs.
4. Update TypeScript domain types, strict boundary validation, `CostKPICard` and `TicketDeepDive` (`EST` badges, rate tooltips and visible `CacheWrite` surcharge warning), rebuild embedded assets and verify with `make` plus standalone binary checks.

Selected [regression boundaries](../../../docs/REGRESSION_PREVENTION.md):
- **Telemetry extraction and refresh**: Mixed observed + unavailable + partial invocations, observed `$0.00` preservation, ambiguous pattern rejection, repeated refresh idempotence, and preservation of unavailable owners' saved estimates.
- **Portable ledgers across developers and machines**: Machine A generates estimates with Rate Card A; Machine B reads and projects identical saved estimates with absent or different Rate Card B; Machine B refreshes only B's session while preserving A's saved estimates.
- **Configuration and project portability**: Absent `model_pricing`, empty `rates: []`, valid exact + glob rates, invalid negative/NaN/duplicate/ambiguous rates failing closed, and non-default rate verification end to end.
- **Frontend embedding and UI state**: Embedded assets rebuilt via `make web-assets`, standalone binary serves updated Deep Dive `EST` badges, tooltips and `CacheWrite` warning, and measured-only tickets remain free of `EST` badges.

## Steering

- **2026-11-26 (User custom input during `openspec-propose`)**:
  - **Acceptance delta**: Fixed aggregate provenance states strictly to: 100% measured -> `measured` (`equivalent = measured`), 100% estimated -> `estimated` (`equivalent = estimated`), 100% covered by measured + estimated -> `hybrid` (complete rollup), and any unpriced facts -> `partial` or `unavailable` (incomplete coverage, `total_equivalent_cost_usd = null`). Added mandatory UI warning stating that `CacheWrite` is currently not considered if a model has cache-write surcharges (such as GPT-5 or GPT-6).
  - **Superseded decision**: None (initial change proposal).
  - **Preserved guarantees**: Observed `$0.00` remains observed; partial measured invocations are never double-counted; cross-machine reads never silently reprice persisted history; AL-57 `cost_usd` and `cost_summary` contracts stay intact.
  - **Verification impact**: Requires explicit unit, API validation and UI tests covering all five aggregate provenance states (`measured`, `estimated`, `hybrid`, `partial`, `unavailable` plus `empty`) and the conditional rendering of the `CacheWrite` surcharge warning.
- **2026-11-26 (User review-policy override during Step 5)**:
  - **Source**: User answered the orchestrator's review-provider check with `"Review via current Gemini 4 Neon"` because `claudeAgent` (`claude` CLI) is not installed in this environment.
  - **Acceptance delta**: Run the four core independent focused reviews (`reviewer-behavior`, `reviewer-data`, `reviewer-architecture`, `reviewer-tests`) on `opencode` / `litellm/gemini-4-neon` instead of `claudeAgent` (`claude-opus-5-5/medium` + `claude-sonnet-5-5/high`), and record the explicit override in `delivery.md` and `evidence/REVIEW_SUMMARY.json`.
  - **Superseded decision**: Temporary execution override for this environment's unavailable `claudeAgent` provider; does not modify `.agents/skills/agentic-lens-review/references/model-policy.yaml`.
  - **Preserved guarantees**: All four core focuses run independently against the same frozen candidate revision with no cross-reviewer seeding, and every confirmed blocker must be fixed and re-verified before delivery.
  - **Verification impact**: Delivery report will record exact model (`litellm/gemini-4-neon`) and note the unavailable `claudeAgent` profiles explicitly.
