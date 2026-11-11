# A12 data integrity review — gpt-6-luna/medium

**Outcome: APPROVED (scoped).** First-pass data review of AL-70 iteration 03, against baseline `54d58be40c74864d645c5168ba23820a21389c01` and the frozen candidate identified by parent fingerprint `974d1449fc440115b563067461d43b9b6381baf60e178ff45e6854f884b10e6a`. Requested reviewer identity: Codex, gpt-6-luna/medium. Scope: missing-model identity production, model aggregates and retained-ledger behavior. This does not certify other review focuses or the whole change.

## Findings

No blocking data-integrity findings in the inspected scope.

## Hypotheses assessed

- **Missing identity could collide with an actual `default` or `unknown` model.** Source changes use empty string only at missing-value producers: OpenCode v1 assistant and metadata-only paths, OpenCode v2 metadata-only fallback, live resolver, and Codex's initial/blank usage bucket. Observed nonempty values are trimmed/preserved; Codex late context accepts `default` and `unknown` as names, while reassignment requires the prior turn-map value to be empty. This keeps empty, `default`, and `unknown` separate. `ComputeModelMetrics` groups by the raw string key and accumulates each group without filtering empty names (`backend/internal/domain/aggregations.go:468-490`).
- **Codex's empty map value could be mistaken for a missing turn.** `assignTurnModel` checks map presence and requires an empty prior value before reassignment; `turnNumber` likewise checks presence before looking up the associated model bucket (`rollout.go:535-560, 895-900`). The added unknown-turn lookup test covers an empty key and a completed earlier turn. The late-context test also preserves earlier unknown-bucket tokens/duration while attributing the turn and later usage to the newly observed model; this matches existing attribution timing and preserves total usage.
- **Metadata-only rows or counters could disappear.** OpenCode only changes the fallback key when no model accumulator exists; its existing accumulator creation and invocation emission remain intact. Codex finalization still creates a model aggregate for the current model when necessary and emits invocation rows per map entry (`rollout.go:920-995`). The tester's supplied evidence reports real SQLite/rollout coverage for absent/blank/default/unknown names, metadata-backed zero-token rows, counters and mixed groups, with all targeted backend checks passing.
- **Existing ledger facts might be rewritten or schema behavior changed.** The candidate diff contains no schema/migration change or read-side rewrite. Tester evidence includes repeated historical API reads with byte preservation and unchanged recorded `default`/`unknown` model names and counters. A12 contract explicitly prohibits historical relabeling.

## Verification and limits

Reviewed the frozen diff/source paths, `ComputeModelMetrics`, the relevant regression boundaries, and tester evidence at `/tmp/al70-a12-tester-evidence.json`. That evidence records passing focused Go source/domain/API tests, live SQLite and rollout cases, frontend consumers, and `git diff --check`, with no missing oracles. No suite was rerun in this review. No live developer stores were read. Remaining delivery checks (main integration, canonical `make`, standalone verification and publication) are outside this scoped review and remain with the parent.

## Frozen-input verification

At this review's exit, recomputed SHA-256 for all 29 paths in `/tmp/al70-a12-review-nlyoe7ub/manifest.json`; every path matched its recorded digest, and the manifest identifies fingerprint `974d1449fc440115b563067461d43b9b6381baf60e178ff45e6854f884b10e6a`. I inspected the manifest at review entry but did not perform a path-by-path hash check then; this is current verification only.
