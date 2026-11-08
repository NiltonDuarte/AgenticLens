h3. Title: AL2-6: Estimate API-equivalent costs alongside measured spend

*GOAL*

Operators can compare an explicitly labeled API-equivalent token estimate alongside measured spend without treating estimates as bills.

*CONTEXT*

AL-57 already distinguishes observed zero, partial and unavailable cost. Subscription harnesses can lack billed-dollar evidence; a declared rate card can supply a separate comparison estimate. Source-reported billing availability must be assessed per record, not assumed from harness name.

*ACCEPTANCE CRITERIA*

* Configuration accepts model_pricing.rates entries with model, input_per_million, cache_read_per_million and output_per_million. Support exact model identifiers and configured model-pattern matching. Rates must be finite nonnegative numbers; invalid or ambiguous matching is diagnosed instead of selecting an arbitrary price. Declare matching precedence and effective catalog identity before implementation.
* For eligible unavailable measured-cost facts with known token categories, estimate USD as (uncached_input * input_rate + cache_read * cache_rate + output * output_rate) / 1,000,000. Do not charge cached input twice. Unknown token categories/rates remain unavailable; no built-in current market price is assumed.
* Observed costs, including genuine numeric zero, are never overwritten or used as missing-cost signals. Preserve partial measured subtotals and coverage; estimate only uncovered eligible facts and never double-count known spend.
* Expose measured_cost_usd, estimated_cost_usd and total_equivalent_cost_usd with measured/estimated/hybrid provenance and coverage through generation, persistence and HTTP projections. Unmatched models retain explicit unavailability.
* CLI tables/Markdown distinguish estimates from measured spend. Deep Dive/model breakdown entries show EST with the applied rate details. Saved estimates remain interpretable on another machine without its local catalog; reads do not silently reprice history.

*Status*: Ready
*Blocked by*: None
*Work type*: feat

*IMPLEMENTATION NOTES*

AL2-7 owns global cost modes and hypothetical repricing. Rate-card matching and token-category mapping need an explicit developer contract grounded in native token semantics; no external pricing lookup or blanket model-pattern guess is required. Preserve portable ledger refresh and unavailable-owner history.

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md). [PO-19](../../PO-19/SPEC.md) preserves the earlier development reconciliation; historical observations require revalidation, not an assumption of current runtime support.
