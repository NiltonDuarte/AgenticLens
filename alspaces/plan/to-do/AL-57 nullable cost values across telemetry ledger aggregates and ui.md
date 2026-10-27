h3. Title: AL-57: Represent unavailable telemetry cost as nullable

*TARGET*: AgenticLens (domain / adapters / CLI / presentation / tests)

*CONTEXT*

Some telemetry sources report token usage without establishing a billable USD cost. The Codex adapter currently writes numeric zero into a shared cost field, so dashboards and summaries can present unavailable cost as free usage. Cost is repeated across invocation facts, aggregates, progression data, comparisons, and UI views; changing only one source adapter would leave derived totals misleading.

*GOAL*

Represent cost availability across the full telemetry contract so an actual zero remains distinguishable from an unknown cost. Preserve measured OpenCode costs, emit unavailable cost for sources without billing evidence, and make the CLI, ledger, aggregates, migrations, and dashboard communicate that distinction consistently.

*VERTICAL SLICE SCOPE*

* Domain Core: Define nullable USD cost for invocation facts and all derived totals. Unknown inputs must not be summed as zero or yield a misleading partial aggregate; define and document propagation behavior for mixed known and unknown inputs.
* Boundary Adapter: Map measured source costs to values and sources without billing evidence to null. Preserve legacy numeric ledger values and add a cost-specific compatibility/migration path where source provenance can reliably establish that zero meant unavailable. Coordinate schema versioning with other migration work without depending on it.
* Presentation: Render null as unavailable across ticket details, fleet totals, comparisons, progression charts, and CLI output. Keep genuine zero distinct and define sort/filter/chart behavior for unavailable values.
* Quality: Add Go serialization, aggregation, migration, and source-mapping tests plus TypeScript formatter/component tests; include mixed-source and legacy-ledger cases.

*ACCEPTANCE CRITERIA*

* JSON serialization distinguishes `null` cost from numeric `0` without changing the meaning of measured OpenCode costs.
* Aggregates containing unavailable cost follow the documented rule and never display an incomplete sum as a complete total.
* Every cost display handles unavailable values explicitly; no UI or CLI renders null as `$0.00`.
* Versioned migration preserves known numeric costs and marks unavailable Codex costs null only when source provenance proves the conversion is safe.
* Go vet/tests and TypeScript typecheck/lint/tests pass; generated frontend assets remain embedded in the Go binary.

*DEFINITION OF DONE*

* Cost migration is versioned and compatible with the repository's existing migration framework; it does not require AL-55 to complete first.
* Quality gates pass with `make`.
* Existing ledgers with numeric costs continue to load and display correctly.
