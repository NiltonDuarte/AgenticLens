## Purpose

Provide explicit, portable, rate-card-based API-equivalent USD cost estimates alongside source-reported measured spend for [AL2-6](../../../../alspaces/plan/to-do/AL2-6%20estimate%20api%20equivalent%20costs%20alongside%20measured%20spend.md), preserving measured billing evidence, deterministic matching, honest coverage provenance, cross-machine readability and transparent UI caveats.

## ADDED Requirements

### Requirement: Configured model pricing catalog and validation
The system SHALL accept an optional shared `model_pricing.rates` configuration list whose entries each declare `model`, `input_per_million`, `cache_read_per_million` and `output_per_million`. Every rate MUST be a finite nonnegative number, every `model` MUST be nonblank with valid glob syntax, duplicate normalized rules MUST fail validation, and the system MUST compute a deterministic SHA-256 `effective_catalog_id` without assuming built-in market prices.

#### Scenario: Absent rate card preserves measured-only behavior
- **WHEN** a project configuration omits `model_pricing` or sets `model_pricing.rates: []`
- **THEN** configuration loads without error, no estimates are produced, and unavailable-cost invocations remain explicitly unavailable

#### Scenario: Invalid rate or duplicate normalized model fails closed
- **WHEN** `model_pricing.rates` contains a negative or non-finite rate, invalid glob syntax, or duplicate normalized `model` entries
- **THEN** configuration validation returns an explicit diagnostic error naming the offending rule(s)

### Requirement: Deterministic model matching and ambiguity rejection
Model matching MUST normalize names by trimming whitespace and lowercasing, then apply exact identifier matches before glob pattern matches, ranking matching globs by higher literal character count and then longer normalized pattern length. Any equal-specificity glob tie MUST fail closed with an explicit ambiguity error rather than selecting an arbitrary price.

#### Scenario: Exact model rule beats broader glob pattern
- **WHEN** `model_pricing.rates` declares both an exact rule for `gpt-5.4` and a glob rule `gpt-5*`, and an eligible unavailable-cost invocation uses `GPT-5.4`
- **THEN** the matcher selects the exact `gpt-5.4` rule and records the catalog's canonical `effective_catalog_id`

#### Scenario: Ambiguous equal-specificity glob fails closed
- **WHEN** two patterns of equal specificity (such as `gpt-5*` and `*-mini`, each with 5 literal characters and length 6) both match `gpt-5-mini` with no exact match
- **THEN** matching returns an explicit ambiguity error naming the colliding rules instead of choosing one arbitrarily

### Requirement: Pure domain API-equivalent estimation formula
The system SHALL estimate API-equivalent USD exclusively for invocation facts whose `cost_evidence.state` is `unavailable`, whose token categories (`input`, `cache_read`, `output`) are present and nonnegative, and whose model matches a configured rule. The formula MUST be `(input * input_per_million + cache_read * cache_read_per_million + output * output_per_million) / 1,000,000`, without double-charging `cache_read` or billing `reasoning` on top of `output`.

#### Scenario: Subscription invocation with unavailable cost receives an API-equivalent estimate
- **WHEN** a Codex or Claude Code invocation has `cost_evidence.state = "unavailable"`, `tokens = { input: 200000, cache_read: 800000, output: 50000, reasoning: 10000 }`, and a matching rule `{ input_per_million: 2.0, cache_read_per_million: 0.25, output_per_million: 8.0 }`
- **THEN** the calculated estimate is `$1.00`, `cache_read` is not charged at the uncached input rate, `reasoning` is not billed on top of `output`, and measured `cost_usd` remains `null` with `cost_evidence.state = "unavailable"`

#### Scenario: Unmatched model or invalid token category stays unavailable
- **WHEN** an unavailable-cost invocation uses an unmatched model or carries negative/invalid token counts
- **THEN** no estimate is produced and the invocation remains explicitly unavailable

### Requirement: Observed and partial billing preservation
The system MUST NEVER overwrite observed measured spend (including genuine numeric `$0.00`) and MUST NEVER estimate tokens on `partial` measured-cost invocation facts, preserving their `known_subtotal_usd` without double-counting coalesced turns.

#### Scenario: Observed zero measured spend is never converted into an estimate
- **WHEN** an invocation has `cost_usd = 0` and `cost_evidence = { state: "observed", known_subtotal_usd: 0 }` and its model exists in `model_pricing.rates`
- **THEN** the invocation retains its observed `$0.00` measured cost, no estimate is generated for it, and its provenance remains `measured`

#### Scenario: Partial measured invocation is preserved without double-counting
- **WHEN** an OpenCode invocation has `cost_usd = null` and `cost_evidence = { state: "partial", known_subtotal_usd: 2.00 }` because billed and unbilled turns were coalesced into one token bucket
- **THEN** the invocation retains `known_subtotal_usd = 2.00`, no token estimate is added for its unsegmented token counts, and it counts as an unpriced remainder in aggregate provenance

### Requirement: Aggregate provenance and equivalent-cost rollup contract
Aggregates SHALL expose `measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd` and coverage provenance alongside existing `cost_usd` and `cost_summary`. Provenance MUST be `empty` for zero facts, `measured` when 100% of facts are measured, `estimated` when 100% are estimated, `hybrid` when all facts are either measured or estimated (complete rollup), and `partial` or `unavailable` whenever any fact remains unpriced (`total_equivalent_cost_usd = null`).

#### Scenario: Pure measured selection produces measured provenance
- **WHEN** all invocations in a ticket or aggregate have `cost_evidence.state = "observed"` totaling `$4.50`
- **THEN** provenance is `measured`, `measured_cost_usd = 4.50`, `estimated_cost_usd = null`, and `total_equivalent_cost_usd = 4.50`

#### Scenario: Pure estimated selection produces estimated provenance
- **WHEN** all invocations in a ticket or aggregate have unavailable measured cost and valid rate-card estimates totaling `$6.75`
- **THEN** provenance is `estimated`, `measured_cost_usd = null`, `estimated_cost_usd = 6.75`, and `total_equivalent_cost_usd = 6.75`

#### Scenario: Complete mixed measured and estimated selection produces hybrid provenance
- **WHEN** a ticket combines observed OpenCode invocations totaling `$23.20` and rate-card estimated Codex invocations totaling `$7.12` with zero unpriced invocations
- **THEN** provenance is `hybrid`, `measured_cost_usd = 23.20`, `estimated_cost_usd = 7.12`, and `total_equivalent_cost_usd = 30.32`

#### Scenario: Any unpriced invocation yields partial or unavailable provenance
- **WHEN** a ticket has `$2.00` measured, `$3.00` estimated, and one unmatched unavailable invocation (or one `partial` measured invocation)
- **THEN** provenance is `partial`, `total_equivalent_cost_usd = null`, and coverage reports the known equivalent subtotal plus exact counts of measured, estimated, partial-measured and unavailable invocations

#### Scenario: All unpriced invocations yield unavailable provenance
- **WHEN** every invocation in a ticket has unavailable measured cost and no matching rate rule
- **THEN** provenance is `unavailable`, `measured_cost_usd = null`, `estimated_cost_usd = null`, and `total_equivalent_cost_usd = null`

### Requirement: Portable persisted estimates without read-time repricing
The system SHALL persist applied invocation-level estimates and rate snapshots (matched pattern, per-million rates and `effective_catalog_id`) as additive optional fields on schema v6 ledgers during generation and source-backed refresh. Ordinary reads and cross-machine projections MUST interpret saved estimates directly without requiring the producing machine's catalog and MUST NEVER silently reprice history or overwrite unavailable owners during partial refresh.

#### Scenario: Cross-machine read preserves saved estimate without local rate card
- **WHEN** Machine A generates and commits a ledger with saved rate-card estimates, and Machine B opens that ledger in CLI or dashboard with an empty or different local `model_pricing.rates`
- **THEN** Machine B renders Machine A's exact saved `estimated_cost_usd`, applied rate details and `effective_catalog_id` without modifying the file or recalculating from Machine B's rates

#### Scenario: Refreshing one local session preserves unavailable owners' saved estimates
- **WHEN** Developer B refreshes only B's locally available session in a shared ticket ledger while Developer A's session source is unavailable on B's machine
- **THEN** Developer A's invocations retain their previously persisted measured evidence and saved estimates unchanged, while Developer B's refreshed session uses the current active catalog and whole-ledger aggregates combine both

### Requirement: CLI, Markdown and HTTP projections distinguish estimates from measured spend
CLI output, generated Markdown (`delivery.md`), computed exports and HTTP workspace, fleet, progression and comparison APIs SHALL expose `measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`, provenance, coverage and applied rate details alongside unchanged `cost_usd` and `cost_summary`. CLI and Markdown tables MUST clearly separate measured spend from `EST` estimates and list applied per-million rates when estimates exist.

#### Scenario: Markdown summary renders measured, estimated and hybrid rows clearly
- **WHEN** `agentic-lens generate` renders `delivery.md` for a ledger containing both measured OpenCode rows and rate-card estimated Codex rows
- **THEN** measured rows display measured USD without `EST`, estimated rows display `$X.XX EST` with applied rate details (`in/cache/out per 1M`), and totals show the hybrid measured + estimated breakdown alongside complete equivalent cost (or partial/unavailable coverage when unpriced facts exist)

#### Scenario: HTTP workspace and fleet responses pass strict boundary validation
- **WHEN** the web client fetches `/api/workspaces/{id}`, `/api/fleet`, `/api/progression` or `/api/comparisons` containing measured, estimated, hybrid, partial and unavailable cost records
- **THEN** every response includes validated `measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`, provenance, coverage and per-model applied rate details alongside unchanged `cost_usd` and `cost_summary`

### Requirement: Dashboard Deep Dive estimate badges and CacheWrite surcharge warning
The web dashboard SHALL display an explicit `EST` badge with an applied-rate tooltip on estimated Ticket Deep Dive model rows and cost summaries, and MUST render a visible UI warning whenever API-equivalent estimates are displayed stating that `CacheWrite` is currently not considered if a model has cache-write surcharges (such as GPT-5 or GPT-6). Pure measured views MUST NOT display `EST` badges.

#### Scenario: Deep Dive model breakdown shows EST badge, rate tooltip and CacheWrite warning
- **WHEN** an operator opens Ticket Deep Dive for a ticket containing at least one estimated model row
- **THEN** estimated model rows show an `EST` badge with a tooltip detailing the applied input, cache-read and output rates per 1M tokens, measured rows show no `EST` badge, cost summaries distinguish measured vs estimated vs equivalent spend, and a visible UI warning states that `CacheWrite` is currently not considered if there are cache-write surcharges for a model (such as GPT-5 or GPT-6)

#### Scenario: Pure measured ticket displays no EST badge or estimate caveat noise
- **WHEN** an operator opens Ticket Deep Dive for a ticket whose invocations are 100% measured observed spend
- **THEN** no `EST` badges appear on model rows and the measured cost display retains its baseline presentation
