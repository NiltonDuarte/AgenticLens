# AL-71: Execution tasks

Use the [design](design.md) for scope approval and the [dashboard query](specs/dashboard-query-reliability/spec.md) and [live reporting](specs/live-source-reporting/spec.md) deltas for acceptance.

## Implementation and focused verification

- [x] **Coder** — Fence Comparisons and Progression settlements and teardown, including catalogue updates. Preserve Fleet's paired requests and reuse only demonstrated shared request and filter invariants.
- [x] **Tester** — Exercise deferred A/B success and rejection in both orders, same-selection retry and teardown. Verify rows, diagnostics, catalogue, errors and loading through the affected views. Preserve Fleet ownership, AL-60 exclusion search and the existing trace and historical enrichment ownership regressions.
- [x] **Coder** — Retrieve all dated and undated candidate ledgers for the type and prefix selection. Move Progression bounds into the domain and show the resolved dates without changing custom range behavior or ticket and invocation attribution.
- [x] **Tester** — Drive the real workspace HTTP handler with historical ledgers. Verify the rendered progression range, buckets and totals with and without an all-matching type or prefix filter. Cover removal of the anchor, membership at UTC boundaries, a custom open side and invocation dates that differ from ticket completion.
- [x] **Coder** — Make live trace decoding strict at the shared REST/SSE boundary. Preserve source diagnostic distinctions, supported optional metadata, and the last valid tree or trace when validation fails.
- [x] **Tester** — Exercise both ingress paths with null, primitive, wrong scalar and nested malformed bodies, plus valid empty and partial data from a Go wire fixture. Replace the assertions that currently expect malformed traces to become empty successes. Verify a malformed payload after a valid one retains the previous observation.
- [x] **Coder** — Serialize fallback polls at the configured cadence and preserve read ordering, failure thresholds and recovery.
- [x] **Tester** — Hold a poll pending through multiple ticks and verify polling resumes after settlement. Verify threshold-to-success recovery, obsolete success and failure against a newer manual or stream observation, and no callbacks after teardown. Retain the ordering behavior from existing overlap tests while removing their reliance on concurrent automatic polls.
- [x] **Coder and tester** — Resolve verified review findings for visible live diagnostics, timestamp precision and validity, and live token counter validation. Add regression assertions for the affected acceptance scenarios, including incomplete-ledger anchor fallback.

## Integration and delivery

- [x] **Orchestrator** — Record implementation approval, coordinate independent behavior, data, architecture and test review, and resolve or document every finding against the tested revision.
- [ ] **Orchestrator** — Integrate with main, run the full `make` gate and affected frontend checks, confirm current embedded assets, and verify the standalone engine. Record actual checks, limits and review attribution in delivery.
- [ ] **Orchestrator** — Publish final source-backed telemetry and the PR with evidence from the integrated candidate.

## Workflow follow-up

Archive after verified delivery and reconcile roadmap status separately.
