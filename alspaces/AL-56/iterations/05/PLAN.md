# AL-56 Iteration 05 Plan: Finalize Session-Owned Schema v4

## Objective

Finalize the still-unreleased schema-v4 ledger as a lean session-owned document and complete its migration path. Keep session identity in one place: each session owns its invocation records, per-tool call/error summaries, and retained events. Remove the parallel root-level invocation, tool, and event collections. Derive cross-session tool totals and dashboard summaries from session facts when loading or aggregating the ledger.

This iteration is limited to the schema shape, extraction/migration needed to produce it, and consumers that must read it. It does not revisit Codex discovery, model/token extraction, Git comparison, nullable cost, or unrelated cleanup.

## Data Contract

- Add a `SessionFact` keyed by a required, nonempty session ID. It carries parent-session linkage and the facts attributed to that session.
- Keep all model-specific invocations for a session inside that session, preserving multi-model splits and invocation fields such as `user_iterations`.
- Store tool summaries inside the owning session, grouped by tool name with calls and errors. Do not repeat `session_id` on each nested tool row. Do not persist cross-session tool aggregates; calculate those from sessions for domain metrics and presentation.
- Store only the events already eligible for persistence inside the owning session. Nested event rows do not repeat `session_id`. Preserve current event pruning and privacy rules; do not add raw transcript or tool payloads.
- Require each invocation, tool summary, and retained event to resolve to exactly one session. Reject duplicate session IDs, dangling references, and unassignable records with record-level diagnostics. Do not fabricate session IDs.
- Session facts are the merge unit. Fresh extraction replaces a matching session snapshot; sessions absent from local extraction remain. Repeated generation must not duplicate sessions or derived totals.

## Migration and Versioning

- Keep the standalone v3-to-v4 migration implementation in `backend/internal/domain/migration/step_v3_v4.go` and extend it to produce the finalized session-owned v4 shape. This branch has not released schema v4; finalize its contract here without adding a v5 transition.
- For v1-v3 ledgers, group invocation rows by their stable session identity (`InvocationFact.ID`, which the current OpenCode and Codex extractors populate with the source session/thread ID); retain every model-specific invocation row in that group. Attach tools and events using their existing session IDs. Validate that all referenced sessions exist.
- For the in-branch interim v4 shape, normalize its root-level collections to the same final v4 representation. Detect and convert it before the normal current-version fast path; do not treat a same-version interim document as already canonical.
- Preserve every fact that has reliable session attribution. Legacy aggregate tool rows without an identity cannot be safely assigned to a session; omit those obsolete aggregates as part of the intentional v4 cleanup and return an explicit migration diagnostic with their count and re-extraction guidance. Do not silently discard them or block all other ledgers on unavailable source logs. No fabricated IDs or hidden partial-migration state.
- Migration must be atomic: on structural validation failure, leave the input bytes/source ledger untouched. On success, repeated migration is idempotent and output contains no legacy root-level facts.

## Implementation Scope

1. Define the session-owned domain and serialized DTO shape; update schema validation, repository load/save, migration invariants, and TypeScript ledger types.
2. Update Codex and OpenCode extraction assembly to create sessions and place invocation, tool, and pruned event facts under the owning session. Keep extraction semantics and the OpenCode default source unchanged.
3. Update domain aggregation and API/dashboard consumers to traverse sessions, derive cross-session tool totals on demand, and keep existing per-tool display semantics.
4. Extend v3-to-v4 migration and interim-v4 normalization as described above. Update only migration/schema documentation and the AL-56 research artifacts that describe the ledger shape.
5. Add repository-independent fixtures and temporary-directory tests. Do not add ticket-ID-specific branches or source/test assumptions about this repository's directory layout.

## Acceptance Tests

- Serialize and load the finalized v4 shape; assert no root-level `invocations`, `tools`, or `events` fields are emitted.
- Reject missing/blank or duplicate session IDs, dangling invocation parents, tools/events without an owning session, and malformed nested records with actionable errors.
- Migrate representative v1, v2, and v3 ledgers to v4 while preserving attributable sessions, model-split invocations, parent relationships, tool counts/errors, retained events, and exact zero iteration totals.
- Normalize the current in-branch interim-v4 fixture into final v4; assert repeated normalization/migration is byte-stable or structurally idempotent and drops legacy root-level collections.
- Verify unscoped legacy tool aggregates are omitted only with a visible diagnostic and accurate dropped-row count; do not fabricate a session or silently report them as migrated facts.
- Verify session snapshot replacement preserves absent sessions and does not double-count tools or events across repeated generation.
- Verify derived tool totals, anomalies, and per-tool UI rows sum facts across sessions without storing a duplicate aggregate.
- Verify both Codex and OpenCode extraction emit the same session-owned contract, including child sessions; preserve default OpenCode extraction behavior.
- Run focused schema/migration/extractor/consumer tests, then `make`, and a Codex generation smoke check. Inspect the resulting ledger for private payload leakage.

## Completion Criteria

- Schema v4 has one canonical session-owned representation; all current producers, migrations, repositories, aggregators, APIs, and UI consumers use it.
- Existing attributable facts migrate without loss; unattributable legacy aggregate rows are explicitly reported and removed according to the user's decision to accept the migration cost.
- Migration failures are atomic and diagnostic; successful migrations are repeatable.
- Focused tests and required repository checks pass; review records findings before the ticket summary and PR body are refreshed.

## Deliverables

- Updated implementation, schema/migration fixtures, and regression tests for the finalized v4 structure.
- Iteration review and concise updates to AL-56 summary and PR materials.
- Refresh telemetry only as the final workflow step for the iteration.
