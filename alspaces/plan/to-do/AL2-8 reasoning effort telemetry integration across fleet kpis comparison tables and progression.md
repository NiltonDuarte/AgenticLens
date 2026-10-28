h3. Title: AL2-8: Reasoning Effort Telemetry Integration across Fleet KPI Cards, Comparison Tables, and Progression Metrics

*TARGET*: AgenticLens (domain | adapters | http | presentation | web | tests)

*CONTEXT*

Modern reasoning models (such as Gemini Flash with thinking enabled, OpenAI o-series, and GPT-6) spend substantial output volume on internal deliberation tokens prior to emitting visible responses. While AgenticLens extracts `tokens.reasoning` from OpenCode SQLite databases and Codex transcripts into `WorkflowEfficiency.ReasoningTokens` and displays it in CLI text summaries and Ticket Deep Dive profile cards, reasoning effort remains absent from primary fleet-level observability surfaces:
1. `FleetOverview.tsx` lacks a KPI card for fleet reasoning volume and deliberation ratio, despite `FleetSummary` already providing `total_reasoning_tokens` and `overall_reasoning_ratio_pct`.
2. `TicketComparisonRow` in domain and web contracts omits reasoning metrics, preventing the Ticket Directory table (`TicketTable.tsx`) from displaying or sorting by reasoning effort.
3. Cross-ticket comparisons (`TicketComparisons.tsx`) and metric progression charts lack reasoning series, preventing teams from analyzing how reasoning intensity correlates with code churn, test ratios, or tool reliability.

*GOAL*

Expose reasoning effort as a first-class metric across the entire AgenticLens platform, adding reasoning volume and ratio to Fleet Overview KPI cards, Ticket Comparison rows and sortable table columns, comparison view metrics, and progression chart series.

*VERTICAL SLICE SCOPE*

* Domain Core: Extend `TicketComparisonRow` in `internal/domain/models.go` and `internal/domain/aggregations.go` with `ReasoningTokens int64` and `ReasoningRatioPct float64`. Extend `MetricProgressionPoint` with `ReasoningTokens int64`. Ensure pure rollup functions compute reasoning metrics deterministically with zero I/O.
* Boundary Adapter: Update HTTP DTOs in `internal/adapters/schemas.go` and `internal/adapters/http/dto.go` to serialize `reasoning_tokens` and `reasoning_ratio_pct` across workspace list and comparison endpoints.
* Web Presentation & Views:
  - `FleetOverview.tsx`: Add a "Reasoning Effort" KPI card displaying total reasoning volume (`formatTokensK`) and overall reasoning ratio percentage with an explanatory tooltip.
  - `TicketTable.tsx`: Add a sortable "Reasoning Effort" column showing formatted reasoning tokens and percentage of generated output.
  - `TicketComparisons.tsx`: Expose reasoning tokens and ratio in comparison metric cards and scatter plot metric pickers.
  - Model tables: Surface reasoning tokens in model breakdown tables alongside input, output, and cache read columns.
* Quality: Go unit tests for domain comparison row derivation and progression point mapping with non-zero and zero reasoning token cases. Vitest tests for `FleetOverview` KPI card rendering, `TicketTable` column sorting, and comparison metric display. Pass `make`.

*ACCEPTANCE CRITERIA*

* OpenCode SQLite extraction continues to populate reasoning tokens accurately from `message.data.tokens.reasoning` and `part.data.tokens.reasoning`.
* `FleetOverview.tsx` renders a prominent "Reasoning Effort" KPI card showing formatted tokens and ratio (e.g. `16.4k tokens (63.4%)`).
* `TicketTable.tsx` includes a sortable "Reasoning Effort" column with ascending and descending natural sort order.
* `TicketComparisonRow` API response carries non-null `reasoning_tokens` and `reasoning_ratio_pct`.
* Zero regressions in existing Deep Dive reasoning display or CLI summary formatters.
* Zero `any` in TypeScript and Go; all tests pass with `make`.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make`).
* Go and TypeScript type checking clean with zero any.
* Verified end-to-end in embedded dashboard.
