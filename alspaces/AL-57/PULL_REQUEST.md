# fix(cost): distinguish unavailable telemetry costs (AL-57)

Missing cost currently appears as free usage. Add observed, partial and unavailable evidence across extraction, persisted ledgers, aggregates, CLI/API and the embedded dashboard. Mixed data exposes a labeled known subtotal and invocation coverage while its complete total stays unavailable.

Valid historical numeric costs remain observed, including zero; no separate legacy type, counter or label is added. Schema v5 stores evidence, canonical v4 stays readable without rewriting, and older schemas retain explicit migration. Preserve source ownership, refresh safeguards and established JSON field order in generated and exported ledgers.

Cost rankings, ratios and charts handle incomplete data explicitly. Estimates and rate cards are excluded.

See [contract](CONTRACT.md), [current plan](iterations/02/PLAN.md), [summary](SUMMARY.md), and [cost/API compatibility guide](../../docs/COST_SEMANTICS.md).
