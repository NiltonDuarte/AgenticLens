# Reported cost, API-equivalent estimates, and availability

AgenticLens schema v6 retains the cost evidence introduced in v5 and distinguishes reported zero from unavailable USD cost. Reported costs are source-reported amounts, not verified invoice spend. Alongside source-reported measured spend, operators may configure explicit API-equivalent rate cards (`model_pricing.rates`) to estimate USD cost for unavailable-cost invocations (such as subscription-billed Codex or Claude Code sessions). Built-in market price tables, external lookups, and subscription cost allocation are not included.

## Measured cost evidence

Invocation facts contain `cost_usd` and `cost_evidence`. Evidence has `state` (`observed`, `partial`, or `unavailable`) and nullable `known_subtotal_usd`. Observed cost has matching numeric amount and subtotal, including zero. Partial cost has a null complete amount and a numeric known subtotal. Unavailable cost has null amounts. Partial invocation evidence occurs when OpenCode coalesces source records with both reported and missing costs into one model row. New Codex extraction reports unavailable cost because rollout usage does not establish USD cost. Claude Code extraction also reports unavailable cost; Claude Code's own `cost-state` estimate is not observed spend and is never persisted or summed. Claude Code does not write the usage of a compaction request to its transcript, so Claude Code ledgers undercount compaction usage.

Valid historical numeric values are observed, including historical Codex zero. Missing/null legacy values remain unavailable. These explicit migration mappings require no source lookup or additional historical state. A successful explicit source refresh replaces only the recovered owner's snapshot using current source evidence; unavailable trees retain their persisted baseline.

Derived measured cost fields (`cost_usd`, `total_cost_usd`, `role_costs`, `model_costs`, `cumulative_cost_usd`) remain nullable complete measured amounts, paired with `cost_summary`. A measured `cost_summary` contains `state` (`empty`, `complete`, `partial`, or `unavailable`), nullable `known_subtotal_usd`, and `fully_observed`, `partial`, `unavailable` and `total` invocation counts. State counts sum to total. A complete measured amount exists only for a nonempty selection of fully observed facts. Known subtotals include reported contributions of partial facts and are displayed with explicit coverage. Invocation coverage is not financial coverage. The schema 6 lean delegation contract retires auxiliary task costs; they never contributed to invocation totals. Explicit migration drops that retired payload while preserving invocation financial evidence.

## Configured API-equivalent cost estimation

Operators may declare an optional `model_pricing.rates` list in shared `.agenticlens-config.yaml` or `.agenticlens-config.yml`. Each entry declares `model`, `input_per_million`, `cache_read_per_million`, and `output_per_million`. Every rate must be a finite nonnegative number, every `model` pattern must be nonblank with valid glob syntax, and duplicate normalized patterns fail validation. Omitting `model_pricing` or setting `rates: []` is valid and disables estimation cleanly; AgenticLens ships no built-in default rates.

Model matching normalizes both invocation model identifiers and configured patterns by trimming whitespace and lowercasing. Exact identifiers (patterns without `*` or `?`) take precedence over glob patterns (`*` matches any rune sequence including `/`; `?` matches any single rune). Matching globs rank by highest literal (non-wildcard) character count, then longest normalized pattern length. Equal-specificity glob ties fail closed with an explicit ambiguity error naming the colliding patterns rather than choosing by config order. Every non-empty catalog hashes its canonicalized, sorted rules with SHA-256 into a deterministic `effective_catalog_id` (`sha256:<64-hex-chars>`).

Only invocation facts with `cost_evidence.state = "unavailable"`, nonnegative `input`, `cache_read` and `output` tokens, and a matching configured rule receive an estimate. The pure domain formula is:

`(input * input_per_million + cache_read * cache_read_per_million + output * output_per_million) / 1,000,000`

Persisted invocation estimates round to 6 decimal places. Across supported adapters, `input` is uncached prompt tokens, `cache_read` is cached prompt tokens (never double-charged at the uncached input rate), and `reasoning` is a subset of `output` (never billed on top of `output`). Observed measured invocations (including genuine `$0.00`) are never overwritten or estimated. Partial measured invocations retain `known_subtotal_usd` without estimating unsegmented tokens so coalesced turns are never double-counted. Unmatched models or invalid token counts remain unpriced and unavailable.

**CacheWrite caveat:** API-equivalent estimates do not currently consider `CacheWrite` tokens. Models with cache-write surcharges (such as GPT-5 or GPT-6) may have higher actual API costs. CLI/Markdown applied-rate notes and the Ticket Deep Dive dashboard surface this caveat whenever estimates are present.

## Additive schema v6 persistence and portable reads

Applied invocation-level estimates are stored as an additive optional `cost_estimate` object on schema v6 `InvocationFact` records (omitted when absent), containing `amount_usd` and a `rate` snapshot (`matched_pattern`, `input_per_million`, `cache_read_per_million`, `output_per_million`, `effective_catalog_id`). Validation requires `cost_estimate` to appear only when `cost_evidence.state = "unavailable"` and enforces finite nonnegative amounts/rates and nonblank pattern and catalog IDs.

Estimation runs only during `generate` and source-backed refresh/backfill for newly extracted or completely refreshed sessions. Unavailable owners and other-harness sessions keep their previously persisted `cost_estimate` untouched. Ordinary reads (`DeriveMetricsLedger`, CLI formatting, HTTP GET) and cross-machine projections interpret saved `cost_estimate` facts directly without consulting local `model_pricing.rates`, so reading a committed ledger on another machine never silently reprices history.

Existing v6 ledgers require no migration, and ledgers without estimates serialize identically to earlier v6 output. Older v6 binaries reject ledgers that contain `cost_estimate` because strict decoding uses `DisallowUnknownFields`.

## Provenance and equivalent-cost rollups

Derived projections (workspace `totals`, `roles`, `models`, thread models, `FleetSummary`, `TimeBucketMetric`, and `TicketComparisonRow`) retain unchanged `cost_usd` and `cost_summary` while adding `measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`, `cost_provenance`, and `equivalent_cost_summary` (plus `applied_rates` on `ModelMetric`, `role_equivalent_cost_summaries` / `model_equivalent_cost_summaries` on `FleetSummary`, and `cumulative_*` equivalents on `TimeBucketMetric`).

`equivalent_cost_summary` reports `provenance`, nullable `measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`, `known_subtotal_usd`, and exact invocation counts `measured`, `estimated`, `partial_measured`, `unavailable`, and `total` (summing to `total`):

- `empty`: `total = 0`; all four amounts are `null`.
- `measured`: 100% of invocations are observed measured (`measured = total > 0`); `estimated_cost_usd = null`; `total_equivalent_cost_usd = measured_cost_usd`.
- `estimated`: 100% of invocations are rate-card estimated (`estimated = total > 0`); `measured_cost_usd = null`; `total_equivalent_cost_usd = estimated_cost_usd`.
- `hybrid`: every invocation is either observed measured or rate-card estimated with at least one of each (`measured > 0`, `estimated > 0`, `measured + estimated = total`); `total_equivalent_cost_usd = measured_cost_usd + estimated_cost_usd`.
- `partial`: at least one invocation remains unpriced (`partial_measured + unavailable > 0`) and at least one measured subtotal or estimate exists; `total_equivalent_cost_usd = null`; `known_subtotal_usd` sums known measured and estimated contributions.
- `unavailable`: every invocation remains unpriced (`unavailable = total > 0`); all four amounts are `null`.

## Fleet, progression, migration, and UI presentation

Fleet `total_cost_usd` follows the measured-only completeness rule alongside its additive equivalent-cost fields. `role_costs` and `model_costs` contain nullable measured amounts with matching `role_cost_summaries`, `model_cost_summaries`, `role_equivalent_cost_summaries`, and `model_equivalent_cost_summaries`. Time buckets expose `cost_summary`, `cumulative_cost_summary`, and their equivalent-cost counterparts; cumulative complete cost remains unavailable after an unknown contribution. Cost-derived shares require complete nonzero denominators. Otherwise `role_effort_unavailable_reason` explains why a distribution is unavailable. Unrelated non-cost metrics stay usable.

Runtime ingestion and new persistence require schema v6. Older inputs are accepted only by explicit migration through the existing framework. Use the migration command's dry-run, ticket filter and backup options; consult the built CLI help for exact options. v4-to-v5 conversion preserves numeric amounts and does not create a source-recovery requirement solely for missing costs. The single v5-to-v6 transition makes lean skill/delegation occurrences session-owned, deliberately deletes ownerless legacy rows and audits retired auxiliary payloads while preserving invocation cost evidence. See [session-owned occurrences](OWNED_OCCURRENCES.md) for the migration boundary. Pre-v4 shape migrations retain their existing source-recovery requirements.

Old binaries reject v6 files (and older v6 binaries reject files containing `cost_estimate`). Preserve migration backups when rollback is needed. JSON consumers must handle nullable costs and evidence/summary metadata. The bundled frontend updates with the backend. Ordinary GET requests do not migrate or rewrite ledgers.

Typed boundary DTOs declare nullable amounts, evidence and summaries explicitly. Standard JSON encoding retains the established relative order of their fields, including nested objects, with cost evidence beside the related amount. Version-specific decoders handle legacy presence and validation without rewriting an arbitrary JSON tree.

Cost KPI cards separate the short amount from availability, provenance, and coverage details. Partial values are visibly labeled known subtotals; unavailable/empty values use a dash and a clear state. In Ticket Deep Dive, pure measured tickets retain their baseline rendering with no `EST` badges or caveat noise; tickets with estimates show explicit `EST` badges, applied-rate tooltips (`input / cache_read / output` per 1M tokens and catalog ID), measured vs estimated vs equivalent breakdowns, and a visible `CacheWrite` surcharge warning. Comparison Average Spend uses complete-cost tickets within the active filters, includes observed zero, and states the eligible ticket count. Partial/unavailable tickets remain in other comparison metrics and rows; they are excluded only from this average and its denominator. Invocation coverage counts retain their existing meaning.

Financial validation follows explicit API/schema locations; dynamic role/model map keys are opaque identifiers. Persisted fact DTOs, compact computed exports and richer HTTP responses remain distinct contracts. Fact-ledger ingestion keeps its required reverse conversions; computed exports are verified as emitted output rather than treated as a second ingestion format. Internal numeric compatibility and per-ledger scalar rounding remain separate from precise cost summaries.
