h3. Title: AL2-6: Model API Pricing Configuration, Domain Rate-Card Cost Approximation Engine, and Hybrid Ledger Ingestion

*TARGET*: AgenticLens (config | domain | adapters | cli | http | presentation | tests)

*CONTEXT*

Sessions executed in subscription-based harnesses such as OpenAI Codex do not emit actual billed dollars, populating `cost_usd: 0` or null in telemetry ledgers. Meanwhile, OpenCode sessions record exact billing from local SQLite databases. This discrepancy distorts fleet summaries, model efficiency rankings, and cross-ticket comparisons, treating multi-million-token Codex workflows as zero-cost. Operators need a way to declare standard model API rate cards in configuration so AgenticLens can deterministically approximate token costs for uncosted sessions while strictly distinguishing measured spend from rate-card estimates.

*GOAL*

Implement an end-to-end vertical slice introducing declarative model API rate cards in `.agenticlens-config.yaml`, pure domain token cost approximation, hybrid cost aggregation (`measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`), CLI generation/summary reporting, and UI badges indicating rate-card approximations.

*VERTICAL SLICE SCOPE*

* Domain Core: Define `ModelRateCard`, `PricingCatalog`, `CostFact`, and `CostProvenance` (`measured`, `estimated`, `hybrid`) types in `internal/domain`. Implement pure function `ApproximateTokenCost(tokens InvocationTokens, rateCard ModelRateCard) float64` supporting uncached input, cache read, and output token rates per million. Update `DeriveMetricsLedger` to calculate estimated costs when measured cost is zero or unavailable, preserving observed costs and producing hybrid rollups. Pure functions, immutable value semantics, zero I/O.
* Boundary Adapter: Extend configuration schema and loader in `internal/adapters/config` to parse `model_pricing` (model identifier, rates per million tokens, pattern matching). Update DTOs in `internal/adapters/schemas.go` and `internal/adapters/http` to serialize `estimated_cost_usd` and `cost_provenance`. Update Codex and OpenCode ingestion to retain measured cost where present and flag subscription runs for approximation.
* CLI Presentation: Update `cmd/agentic-lens` summary generators and table formatters (`formatters.go`) to display estimated cost annotations (e.g. `$0.00 ($7.11 est)`) in model breakdown tables and markdown summaries when approximations are active.
* Web Presentation: Update TypeScript telemetry contracts (`web/src/types/telemetry.ts`), client mappers, and formatters (`formatCost`). In `TicketDeepDive.tsx` and `ModelBreakdownTable.tsx`, render an `EST` pill badge with tooltip showing the rate card applied whenever an invocation or model rollup uses approximated pricing.
* Quality: Comprehensive Go table-driven unit tests for config parsing, pattern matching, domain token cost calculations, and ledger aggregation with zero mocks. Vitest tests for TypeScript formatters and model table badges. Full verification with `make`.

*ACCEPTANCE CRITERIA*

* `.agenticlens-config.yaml` supports `model_pricing.rates` with `model`, `input_per_million`, `cache_read_per_million`, and `output_per_million`.
* Domain cost calculation strictly implements `(input * input_rate + cache_read * cache_rate + output * output_rate) / 1,000,000.0`.
* Measured costs from OpenCode are never overwritten or altered by rate-card estimates.
* Invocations without measured cost (or zero-cost Codex runs) receive deterministic estimated cost values with provenance marked as `estimated`.
* CLI summary tables and Markdown exports report estimated costs distinctly from measured spend.
* Web dashboard deep dive surfaces `EST` badges and rate card details on approximated model entries.
* All Go tests pass with race detector enabled; TypeScript typecheck and Vitest suites pass cleanly with zero `any`.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
