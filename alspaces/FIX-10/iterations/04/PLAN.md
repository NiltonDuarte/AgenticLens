# Iteration 04 Delta Plan: Remove Synthetic Invocation Identity

## Goal

Delete `sessions[].invocations[].id` throughout schema v4, implementation, frontend, persisted ledgers, and tests. An invocation has no upstream identity; ownership belongs to its enclosing `SessionFact.ID`. Where multiple invocation rows need to be distinguished inside one session, use their model, role, and timestamps (and sequence/index only as transient ordering when necessary), never persist a generated ID.

## Planned Changes

1. **Go domain, adapters, validation, and migration**
   - Remove `ID` from `domain.InvocationFact` and `InvocationFactDTO` and remove ID copying from DTO/domain conversion in `backend/internal/domain/models.go` and `backend/internal/adapters/schemas.go`.
   - Remove invocation-ID required checks from `AgenticFactLedgerDTO.Validate`, canonical v4 validation, and `domain/migration/engine.go`. For schema v4, reject any invocation object containing `id`; do not silently accept or retain the synthetic key.
   - Remove ID generation in OpenCode and Codex extractors, and update backfill assembly/merge paths to preserve invocation rows by their containing session. Audit formatters, live hydration, session constructors, and test helpers that infer session identity from invocation IDs. In live hydration and anomaly attribution, use the enclosing session ID directly.
   - Replace invocation-ID dedup/count sets in `domain/aggregations.go` with the intended row semantics: each invocation row contributes one invocation; retain model and role grouping, and use model/role/timestamps only for any deduplication or attribution that still has a demonstrated need. Update `ComputeThreadSummaries` callers/contract where its flat input cannot recover ownership; session-based summaries must use `SessionFact.ID` and `ParentID`.
   - Update anomaly association and backfill merge keys. Invocation rows are scoped by session; when a merge must distinguish rows within one session, key by available model, role, and start/completion timestamps, with deterministic occurrence ordering for otherwise indistinguishable rows. Do not reconstruct identity from naming conventions.

2. **TypeScript and consumers**
   - Remove `id` from `InvocationFact` in `web/src/domain/types.ts`; remove it from invocation fixtures and any UI/domain consumers that rely on invocation identity. Keep `SessionFact.id` as the sole session ownership key.
   - Audit progression, migration, and view logic for row counting/deduplication that previously used invocation IDs; count invocation records or group by model/role as the metric requires.

3. **Persisted data and migration compatibility**
   - Remove `id` from every `sessions[].invocations[]` object in historical production ledgers under `alspaces/*/agentic_metrics.json`, including FIX-10's ledger, and test/parity ledgers under `testdata/workspaces/` and `testdata/parity_workspaces/`.
   - Schema compatibility decision: keep schema version 4 and make its canonical persisted shape require invocations without `id`. Existing v4 files containing the former synthetic ID are explicitly normalized by removing that field (preserving all other invocation data); runtime ingestion thereafter rejects `id`. No v5 bump or compatibility alias is needed because the removed value carries no source identity or independent facts. Ensure every saved/serialized ledger omits the field, including ledgers loaded and re-saved by CLI/backfill paths.

4. **Tests and acceptance**
   - Update backend domain, DTO/validator, migration, extractor, backfill/merge, live hydration/anomaly, CLI, and serialization tests; update frontend type fixtures and aggregation/migration tests. Add boundary coverage that v4 rejects invocation `id`, serialization never emits it, multiple model rows in one session remain distinct, and attribution consistently uses the enclosing session ID.
   - Acceptance: no Go or TypeScript invocation type/consumer requires an invocation ID; no synthetic ID generation or `-inv-` parsing remains; schema v4 validation rejects the removed key; all persisted `sessions[].invocations[]` in `alspaces/` and `testdata/` omit `id`; session ownership, model/role metrics, merges, and anomaly attribution remain correct for multiple rows per session.
