# AL2-6: Execution tasks

Reference [design.md](design.md) for approved scope and [specs/api-equivalent-cost-estimation/spec.md](specs/api-equivalent-cost-estimation/spec.md) for observable acceptance. Confirm explicit user approval is recorded in `design.md` before starting implementation.

## Implementation and focused verification

- [x] **1. Pure domain pricing catalog, matcher and equivalent-cost rollups** (Owner: Coder for `backend/internal/domain/{config,cost,pricing,models,aggregations}.go`; Tester for `backend/internal/domain/*_test.go`)
  - Add `ModelPricingConfig` and `ModelPricingRate` to `ProjectConfig` with pure validation for nonblank models, finite nonnegative rates, valid glob syntax and duplicate normalized rule rejection.
  - Implement pure canonicalization, deterministic exact-before-glob matching (literal count, then length), fail-closed ambiguity diagnostics and canonical SHA-256 `effective_catalog_id`.
  - Implement `AppliedRateSnapshot`, `CostEstimate`, `CostProvenance` (`empty`, `measured`, `estimated`, `hybrid`, `partial`, `unavailable`), `EquivalentCostSummary` and invocation estimation using `(input * input_per_million + cache_read * cache_read_per_million + output * output_per_million) / 1,000,000`.
  - Enforce guardrails: never overwrite observed spend (including `$0.00`), never estimate partial measured invocations, leave unmatched models or negative/invalid tokens unavailable.
  - Compute `EquivalentCostSummary` and deduplicated `AppliedRates` across ticket totals, roles, models, fleet summaries, progression buckets and comparison rows without mutating existing `CostUSD` / `CostSummary`.
  - Verify with table-driven domain unit tests covering exact vs. glob precedence, ambiguity errors, `$0.00` observed preservation, partial non-double-counting, tiny-amount summation precision, and all six provenance states (`empty`, `measured`, `estimated`, `hybrid`, `partial`, `unavailable`).

- [x] **2. Shared YAML configuration loading and additive schema v6 ledger persistence** (Owner: Coder for `backend/internal/adapters/config/{policy_nodes,loader}.go`, `backend/internal/adapters/{invocation_fact_dto,fact_ledger_dto}.go`; Tester for adapter config and DTO tests)
  - Parse and validate optional shared `model_pricing.rates` in YAML node validation and `ValidateConfig`; ensure absent `model_pricing` and empty `rates: []` stay valid and non-estimating.
  - Add optional `cost_estimate` (`omitempty`) to `InvocationFactDTO` with strict decode and `Validate()` checks requiring `cost_evidence.state == "unavailable"`, finite nonnegative amounts/rates and nonblank `matched_pattern` and `effective_catalog_id`.
  - Verify existing schema v6 ledgers decode unchanged, ledgers without estimates serialize identically to baseline v6, invalid `cost_estimate` payloads (such as attached to observed or partial facts) fail validation, and round-trip persistence preserves exact saved estimates.

- [x] **3. Generation, source-backed refresh and portable cross-machine immutability** (Owner: Coder for `backend/cmd/agentic-lens/` and `backend/internal/adapters/backfill/`; Tester for CLI and backfill tests)
  - Apply the configured rate card during `generate` and source-backed `backfill` refresh only to newly extracted / completely refreshed sessions before saving `agentic_metrics.json`.
  - Preserve unavailable owners' and other-harness sessions' previously persisted `cost_estimate` untouched during partial refresh.
  - Guarantee ordinary reads (`DeriveMetricsLedger`, CLI inspection/export, HTTP GET) read saved estimates directly without repricing from local `model_pricing.rates` or rewriting ledgers.
  - Verify portable cross-machine fixtures: Machine A generates estimates with Rate Card A, Machine B reads identical values with no/different local rate card, and Machine B refreshes only B's session while preserving A's saved estimates.

- [x] **4. CLI Markdown summaries and HTTP/export projections** (Owner: Coder for `backend/internal/adapters/{formatters,metric_projection_dto,metrics_export_dto}.go`, `backend/internal/adapters/http/{workspace,fleet,progression,comparison}_dto.go`; Tester for formatter and HTTP handler tests)
  - Expose `measured_cost_usd`, `estimated_cost_usd`, `total_equivalent_cost_usd`, `cost_provenance`, `equivalent_cost_summary` and per-model `applied_rates` across computed exports and HTTP workspace, fleet, progression and comparison DTOs alongside unchanged `cost_usd` and `cost_summary`.
  - Update `FormatMarkdownMetricsTable` (`delivery.md`) to distinguish measured spend from `EST` estimates, display hybrid breakdowns and partial/unavailable coverage cleanly, and append applied rate details (plus a concise `CacheWrite` exclusion note) when estimates exist while leaving measured-only Markdown unchanged.
  - Verify formatter snapshots and HTTP contract tests across measured, estimated, hybrid, partial and unavailable fixtures.

- [x] **5. Web dashboard Deep Dive presentation, CacheWrite warning and embedded assets** (Owner: Coder for `web/src/**` and `make web-assets`; Tester for `web/src/**/*.test.{ts,tsx}` and standalone-binary check)
  - Extend TypeScript domain types (`web/src/domain/types.ts`, `web/src/domain/cost.ts`, `web/src/domain/formatters.ts`) and strict boundary validators (`web/src/api/costValidation.ts`, `web/src/api/validation.ts`) with zero `any`.
  - Update `TicketDeepDive.tsx` Model Resource Consumption table to display `EST` badges on estimated/hybrid model rows with accessible `InfoTooltip` rate details (`input / cache_read / output` per 1M, matched pattern and catalog ID).
  - Update `CostKPICard.tsx` and Deep Dive cost summaries to display measured vs. estimated vs. total equivalent spend and honest partial/unavailable coverage.
  - Render a visible, accessible UI warning whenever API-equivalent estimates are shown stating that `CacheWrite` is currently not considered if there are cache-write surcharges for a model (such as GPT-5 or GPT-6); omit `EST` badges and estimate warnings on pure measured views.
  - Rebuild embedded frontend assets (`make web-assets`) and verify unit/component tests plus standalone-binary serving of updated UI assets.

- [x] **6. Cost semantics documentation** (Owner: Coder for `docs/COST_SEMANTICS.md`)
  - Document `model_pricing.rates`, matching precedence, `effective_catalog_id`, additive v6 `cost_estimate` compatibility (including older-binary rejection notice), provenance rules, cross-machine read immutability and the `CacheWrite` exclusion caveat.

## Integration and delivery

- [x] **7. Verification and standalone-binary checks** (Owner: Tester / Orchestrator)
  - Execute focused backend unit/adapter/HTTP/CLI tests, frontend vitest/typecheck/lint suites, embedded-asset sync tests and standalone-binary verification against non-default rate-card fixtures.
- [ ] **8. Independent review and disposition of findings** (Owner: Reviewers / Orchestrator)
  - Run focused independent reviews across domain/config, persistence/portability and HTTP/web boundaries; retain compact findings at `evidence/REVIEW_SUMMARY.json` and resolve or explicitly record all findings.
- [ ] **9. Integrated `make` gate, final artifacts, telemetry and PR handoff** (Owner: Orchestrator)
  - Integrate with `main`, run full `make`, record tested revision/checks/limits in `delivery.md`, generate final session telemetry with a literal nonce as the last extraction step, and open/link the PR.

## Workflow follow-up

Archive only after verified delivery and consolidated review gates pass; reconcile roadmap status in `alspaces/plan/` separately upon actual ticket completion.
