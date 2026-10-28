# docs(PO-12): specify model rate cards, cost approximation engine, and reasoning metrics

## Summary

This pull request establishes the Product Owner specification and roadmap breakdown for introducing model API pricing configuration, rate-card cost approximation, and reasoning effort metric expansion across AgenticLens (`AL2-6`, `AL2-7`, and `AL2-8`).

Codex sessions executed under subscription plans emit `$0.00` or null costs while consuming hundreds of millions of tokens, skewing fleet aggregations and comparative analyses against OpenCode sessions that report exact billing. This specification introduces declarative rate cards in `.agenticlens-config.yaml`, a pure domain token cost approximation engine, hybrid cost rollups (`measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`), CLI summary reporting, and interactive dashboard indicators.

## Roadmap Breakdown

- **AL2-6**: Model API Pricing Configuration, Domain Rate-Card Cost Approximation Engine, and Hybrid Ledger Ingestion (Ready)
  - Declarative `model_pricing` YAML schema in root config with model matching and default fallbacks.
  - Pure domain formula: `(input * input_rate + cache_read * cache_rate + output * output_rate) / 1,000,000.0`.
  - Invariant preservation: measured spend is never overwritten; estimated costs and provenance flags are attached to uncosted runs.
  - CLI tables and markdown summaries displaying estimated costs.
  - Web deep dive and model breakdown tables rendering `EST` badges and tooltip details.
- **AL2-7**: Interactive Cost Estimation Dashboards, Hybrid Fleet Analytics, and Rate-Card Simulation (Pending, blocked by AL2-6)
  - Multi-mode cost toggles (`Measured Spend`, `Estimated Equivalent`, `Combined Hybrid`) in Fleet Overview and Ticket Comparisons.
  - Dynamic scatter plot and progression timeline re-projections based on active cost mode.
  - Interactive rate-card simulation drawer for modeling pricing adjustments against historical token traces.
- **AL2-8**: Reasoning Effort Telemetry Integration across Fleet KPI Cards, Comparison Tables, and Progression Metrics (Ready)
  - End-to-end exposure of reasoning tokens and ratio in `TicketComparisonRow` across domain models, JSON DTOs, and web views.
  - Prominent "Reasoning Effort" KPI card in Fleet Overview.
  - Sortable Reasoning Effort column in the Ticket Directory table (`TicketTable.tsx`).
  - Reasoning metrics in ticket comparison cards and scatter plot metric pickers.
  - Verification that OpenCode SQLite database already tracks reasoning tokens in `message` and `part` (`step-finish`) records.

## Empirical Findings: AL-56 Cost Analysis

A deep audit of the committed AL-56 telemetry facts was conducted to validate the rate-card approximation model against real multi-harness data:
- **gemini-3.8-flash** (OpenCode, 13 invocations): 15.93M uncached input, 122.94M cache read, 236.1k output -> **$23.20** (measured spend).
- **gpt-6-luna** (Codex, 24 invocations): 8.40M uncached input, 521.56M cache read, 1.38M output -> **$6.74** estimated at standard rates ($0.10 / $0.01 / $0.50 per 1M).
- **gpt-6-astra** (Codex, 4 invocations): 384.99k uncached input, 23.83M cache read, 53.38k output -> **$0.30** estimated at standard rates.
- **gpt-6.1-sol** (Codex, 1 invocation): 162.86k uncached input, 4.24M cache read, 16.41k output -> **$0.07** estimated at standard rates.
- **Total Combined AL-56 API-Equivalent Cost**: **$30.32** ($23.20 measured + $7.11 estimated).

## Invariants & Compliance

- Strictly adheres to PO Non-Execution Invariant: zero modifications to application code or test suites.
- Strictly adheres to Zero Emojis Mandate across all documentation and ticket definitions.
- FCIS architecture maintained: pricing calculations are pure functions; rate cards are passed via immutable configuration.
