# Feature Specification: Model API Pricing Configuration and Rate-Card Cost Approximation

## Problem Statement

AgenticLens tracks multi-agent engineering sessions originating from diverse runtime harnesses, including OpenCode and OpenAI Codex. While OpenCode records direct billable API spend in its local SQLite database, Codex sessions executed under subscription plans (such as ChatGPT Team or Enterprise) do not log financial expenditure, emitting `$0.00` or null for session costs.

Consequently, workspaces combining OpenCode and Codex sessions (or entirely Codex-driven tickets) present misleading financial data:
1. Fleet and ticket-level metrics under-report actual engineering cost by omitting subscription agent runs.
2. Cross-ticket comparisons and ROI analyses skew in favor of Codex tickets, making them appear "free" despite massive token consumption (e.g. hundreds of millions of cached and uncached tokens).
3. Teams lack a mechanism to project what a subscription-backed agentic workload would cost if executed against standard provider API rate cards.

By establishing configurable model API rate cards in `.agenticlens-config.yaml` and a pure domain approximation engine, AgenticLens can bridge this gap: computing deterministic API-equivalent costs for uncosted sessions while strictly preserving the boundary between measured spend and rate-card estimates.

## Target Personas & Use Cases

- **AI Platform Engineers & Engineering Managers**: Evaluate the true financial footprint and API-equivalent spend of multi-agent development workflows across teams and harnesses.
- **Agent Developers & Orchestrator Architects**: Compare cost-efficiency between models (e.g. Gemini 3.8 Flash vs GPT-6 Luna) on equal financial footing.
- **Operations & Finance Leads**: Run what-if simulations to estimate cloud API budget requirements before transitioning from fixed subscriptions to pay-per-token API tiers.

## User Stories

1. As an engineering manager, I want to configure model pricing rate cards in `.agenticlens-config.yaml`, so that uncosted agent invocations automatically receive deterministic API cost approximations.
2. As a platform engineer, I want AgenticLens to distinguish observed billing from rate-card estimates across CLI reports and web dashboard views, so that financial audits remain auditable and verifiable.
3. As a developer inspecting ticket metrics (such as AL-56), I want to see both measured spend ($23.20) and approximated Codex spend (~$7.11) with clear rate-card badges, so that I understand the complete resource investment.

## Functional Scope & Architectural Breakdown

### 1. Configuration & Schema (`.agenticlens-config.yaml`)
Define a declarative `model_pricing` section in the root configuration:
```yaml
model_pricing:
  default_currency: USD
  rates:
    - model: "gpt-6-luna"
      input_per_million: 0.10
      cache_read_per_million: 0.01
      output_per_million: 0.50
    - model: "gpt-6-astra"
      input_per_million: 0.10
      cache_read_per_million: 0.01
      output_per_million: 0.50
    - model: "gpt-6.1-sol"
      input_per_million: 0.10
      cache_read_per_million: 0.01
      output_per_million: 0.50
    - model: "gemini-3.8-flash"
      input_per_million: 0.15
      cache_read_per_million: 0.0375
      output_per_million: 0.60
```
- Support exact model match and case-insensitive normalization.
- Support wildcard/glob prefix fallback (e.g. `gpt-6*`).

### 2. Functional Core (`internal/domain`)
- Value Objects: `ModelRateCard`, `PricingCatalog`, `CostFact`.
- Pure Derivation:
  `ApproximateTokenCost(tokens InvocationTokens, rateCard ModelRateCard) float64`
  Formula: `(tokens.Input * rate.InputPerMillion + tokens.CacheRead * rate.CacheReadPerMillion + tokens.Output * rate.OutputPerMillion) / 1,000,000.0`.
- Invariant & Provenance Separation:
  - If `CostUSD` is explicitly recorded (> 0) from an authoritative billing source (OpenCode), `CostKind = CostKindMeasured`.
  - If `CostUSD` is zero, nil, or flagged from a subscription source (Codex), compute `EstimatedCostUSD` and set `CostKind = CostKindEstimated`.
  - Do not overwrite measured spend. Provide hybrid aggregates: `MeasuredCostUSD`, `EstimatedCostUSD`, and `TotalEquivalentCostUSD`.

### 3. Imperative Shell & Boundary Adapters (`internal/adapters`)
- Config Loader: Parse and validate `model_pricing` with default rate-card fallbacks.
- CLI Engine: Wire rate cards into `agentic-lens generate`, `scan`, and `backfill`.
- HTTP REST API: Expose pricing catalog and hybrid cost fields (`measured_cost_usd`, `estimated_cost_usd`, `cost_kind`) across `/api/workspaces`, `/api/fleet`, and `/api/workspaces/{ticketId}` endpoints.

### 4. Presentation & Visualization (`web/` & CLI)
- CLI Summary & Tables: Add `Estimated Cost` column and footnotes when rate-card approximations are active.
- Web Metric Cards: Display total equivalent cost with breakdown indicator (e.g. `$23.20 observed + $7.11 est`).
- Ticket Deep Dive & Model Breakdown Table: Annotate model rows with an `EST` pill badge and contextual tooltip detailing the rate card applied ($0.10 / $0.01 / $0.50 per 1M).
- Comparison Matrix & Scatter Plot: Provide toggle between "Observed Only", "Estimated Only", and "Hybrid Combined" cost views.

## Non-Functional & Architectural Invariants

- **FCIS (Functional Core, Imperative Shell)**: Domain pricing calculations remain pure functions with zero clock or filesystem dependencies. Rate cards are passed as immutable configuration values.
- **Zero Hallucinated Spend**: Rate-card estimates must never be masqueraded as verified billing. Provenance tags (`measured` vs `estimated`) must accompany every financial figure.
- **Zero Emojis**: Strictly forbid emojis in code, configuration keys, logs, UI tooltips, badges, tables, or documentation.
- **Strict Typing**: Zero `any` in TypeScript or `interface{}` in Go. Strict discriminated unions for cost provenance (`type CostKind = 'measured' | 'estimated' | 'hybrid'`).

## Tracer-Bullet Decomposition

1. **AL2-6**: Model API Pricing Configuration, Domain Rate-Card Cost Approximation Engine, and Hybrid Ledger Ingestion
   - Status: Ready (Tier 1)
   - Vertical Slice: Config schema parsing, pure domain cost calculator, DTO additions (`estimated_cost_usd`, `cost_kind`), CLI summary output, and core UI model badges.
   - Blocks: AL2-7
2. **AL2-7**: Interactive Cost Estimation Dashboards, Hybrid Fleet Analytics, and Rate-Card Simulation
   - Status: Pending (Tier 2)
   - Vertical Slice: Multi-mode cost toggles (Measured vs Estimated vs Hybrid) in Fleet Overview and Ticket Comparisons, dynamic scatter plot re-projection, and rate-card simulation drawer.
   - Blocked by: AL2-6
3. **AL2-8**: Reasoning Effort Telemetry Integration across Fleet KPI Cards, Comparison Tables, and Progression Metrics
   - Status: Ready (Tier 1)
   - Vertical Slice: Exposing reasoning tokens and ratio in `TicketComparisonRow` across domain and DTOs, adding a dedicated Fleet Overview KPI card, sortable Ticket Directory column, cross-ticket comparison cards, and model breakdown telemetry.
   - Blocked by: None

---

## Technical Audit: Reasoning Effort Data in OpenCode & Telemetry Pipelines

### 1. Data Availability in OpenCode
Audit of the OpenCode SQLite schema and ingestion adapters confirms reasoning tokens are already tracked at runtime:
- **`message` table**: The JSON payload in the `data` column contains `tokens.reasoning` (e.g. `{"input": 160, "output": 51, "reasoning": 588, "cache": {"read": 22598}}`).
- **`part` table**: `step-finish` objects record `tokens.reasoning` per step execution.
- **OpenCode Adapter** (`backend/internal/adapters/opencode/extractor.go`): Deserializes `rawTokensInfo.Reasoning` and aggregates it into session models and `domain.InvocationTokens.Reasoning`.
- **Codex Adapter** (`backend/internal/adapters/codex/rollout.go`): Directly maps `Usage.Reasoning` to `domain.InvocationTokens.Reasoning`.

### 2. Current vs Missing Presentation Surfaces
- **Currently Surfaced**:
  - CLI Markdown Summary (`formatters.go`): Emits `Reasoning Effort: X tokens (Y% of generated output)`.
  - Ticket Deep Dive Profile Card (`TicketDeepDive.tsx`): Displays `Reasoning Tokens` and `Reasoning Ratio %`.
  - Live Trace Inspector (`LiveTraceInspector.tsx`): Renders reasoning chips and deliberation blocks.
- **Missing / Gaps Addressed by AL2-8**:
  - Fleet Overview KPI Row: Missing reasoning volume card despite `FleetSummary.total_reasoning_tokens` and `overall_reasoning_ratio_pct` already computed in domain aggregations.
  - Ticket Directory Table (`TicketTable.tsx`): Missing reasoning token column and natural sorting.
  - Ticket Comparison Matrix (`TicketComparisons.tsx`): Missing reasoning metrics in comparison cards and scatter plot pickers.
  - Progression Analytics: Missing reasoning series in historical progression charts.
