# AL-70 iteration 03: Explicit missing-model state

A12 is explicitly approved by the user: replace the invented missing-model label default with explicit unknown; retain observed names, aggregate rows/counts/counters and existing ledgers without a schema change. Approval is already established and does not require another question. This material acceptance delta supersedes the pending A12 decision only. [Confirmed A12 evidence](../../LEGACY_AND_FALLBACKS.md#confirmed-verdict-a12--missing-model-representation) and the [contract](../../CONTRACT.md) govern this iteration.

Baseline is 54d58be40c74864d645c5168ba23820a21389c01, with PR 137 open per parent handoff. Earlier accepted AL-70 behavior and valid unchanged verification remain in force; earlier producer-label expectations do not establish completion for A12.

## Representation and ownership

Use empty string for missing model identity in existing domain/API model fields, rendered Unknown in human-facing labels. Existing aggregate tests preserve empty model groups, so no nullable field, new sentinel name, schema version or migration is needed. Native literal default and unknown remain observed names. Preserve OpenCode v2 required assistant-model validation and existing nonempty metadata parsing.

The parent owns integration, consolidated review, artifacts, telemetry and PR delivery. The assigned source coder owns OpenCode/Codex producer changes; the live coder owns live resolution; the frontend owner owns display-only model labels. Shared-checkout owners preserve others' edits and receive exact writable paths. Tester adds meaningful adversarial coverage; planner changes only acceptance/plan artifacts.

## Ordered work and verification

1. Change invented labels in OpenCode v1 assistant/metadata-only accumulators, v2 metadata-only accumulators and live resolveModelName. Preserve metadata-backed rows, counters, native names and v2 input rejection.
2. Correct Codex initial/blank model state and provisional-turn reassignment. Audit map-presence versus empty-key checks, including turnNumber/finalization; observed default/unknown must never be treated as missing-model sentinels. Preserve attribution, duration, tokens and counters.
3. Display empty model as Unknown in live session tree/inspector, ticket model table, thread details and relevant human summaries. Preserve raw model keys and existing filtering/sorting/aggregation/count behavior.
4. Extend existing real SQLite/rollout and API/UI tests for absent/blank/default/unknown/known names, mixed groups, metadata-only zero-token rows and late Codex attribution. Compare row counts, turn/iteration/compaction counters, tokens, cost and duration; preserve multiple-turn attribution. Verify existing ledger reads do not relabel or rewrite files. Do not assert metadata-only means zero invocation rows.
5. Run affected adapter/domain/API/frontend checks with canonical embedding prerequisites. Freeze the candidate for risk-directed review of behavior, data integrity, architecture and test effectiveness under existing four-focus/two-model policy. Parent carries forward unchanged evidence with a dependency rationale, integrates main, runs make and completes delivery/late telemetry only after current evidence passes.

Regressions selected: source fact fidelity, optional hydration, portable history, repeated grouping/attribution and frontend embedding. No schema, migration, metric formula, aggregation redesign or new live capability is authorized. Source inspection supports this plan; no new runtime verification has occurred in this planning pass.

Integration correction within approved A12: the first full make at f6a2f543 exposed two backfill preservation cases still expecting the retired invented model label. Update that expectation to the empty identity while preserving the retained-row, zero-turn, unavailable-cost, complete-owner replacement and unrelated-history assertions. Source behavior and acceptance do not change. Tester owns the correction/backfill race check; both test-effectiveness profiles recheck the corrected assertion before the parent repeats the required full gate.

Completion evidence: corrected candidate 07d07a9d61f3bcdae70fb2a72cf1ac7cdef38c96 passed full make (426.197s), host manager/engine execution and same-revision standalone embedded UI/API verification. Current origin/main was already integrated. All eight focus/profile passes and both affected test rechecks approved; final outcome and retained finding attribution are linked from SUMMARY.md.
