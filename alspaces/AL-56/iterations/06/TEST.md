# AL-56 Iteration 06 Test Report

## Verification

- `make` — passed (exit 0). This ran `go vet ./...`, frontend typecheck and ESLint, `go test -race ./...`, Vitest (31 files, 722 tests), frontend production build, and Linux/macOS amd64/arm64 binary builds.
- `(cd backend && go test ./internal/domain -run 'TestPersistSparseAnomalyEvidenceSelectsExactRetryRunWhenSameTurnRepeats|TestPersistSparseAnomalyEvidenceKeepsSessionRetryStreakAcrossInterleavedSessions|TestPruneEventsForPersistencePropagatesMissingEvidenceError' -v -count=1)` — passed all three targeted retry/error-propagation regressions.
- `(cd backend && go test ./internal/domain ./internal/adapters/backfill ./internal/adapters/codex ./internal/adapters -run 'Sparse|Subagent|Anomaly|MigrationRunner_NormalizesPreIterationV4' -count=1)` — passed for all four packages before the source-ID contract was expanded; the final full `make` covers the expanded tests.
- `(cd backend && go test ./internal/domain -run TestSparseEvidenceMeetsLedgerSizeTargets -v -count=1)` — passed. Synthetic fixture: 619 baseline event rows, 132 retained rows (21.3%), 135,551 baseline JSON bytes, 38,130 sparse JSON bytes (71.9% reduction). Retained rows are failed tool calls referenced by the aggregate anomaly.
- `git diff --check` — passed.

## Focused Behavior Checks

- Retry-streak summaries are derived from the full event stream, nested under the owning session without a serialized `session_id`, and retain the error sequence. The metrics assertion compares persisted summaries with detection from the full stream.
- Cross-session high tool-failure summaries are ticket-scoped and include tool name, error count, total calls, failure rate, and threshold. They reference every failed tool-call event by `(session_id,event_id)`; DTO validation rejects duplicate, missing, wrong-tool, wrong-status, or wrong-event-type evidence. Source-less legacy aggregates fail closed with a record-level missing-evidence diagnostic.
- Retry-streak summaries retain the ordered error sequence; context-spike coverage verifies before/after samples. The subagent rework test retains the six role-transition events for three cycles; an interleaved independent-parent test prevents false cycles. Repeated sparse persistence preserves stored summaries. Totals tests compare per-session tool calls/errors and derived invocation totals before and after persistence.
- Retry detection partitions by owning session and sorts each session chronologically, so interleaved tool calls from other sessions cannot split a streak. Detected retry summaries carry exact source refs that the selector resolves by `(session_id,event_id)`; a regression interleaves session A's failures at t1/t3/t5 with successful session B calls at t2/t4 and verifies only A's three calls are retained. A second regression covers a five-error streak followed by two errors on the same turn and verifies only the first run is referenced. `PruneEventsForPersistence` now returns selection errors instead of silently replacing them with an empty event list.
- DTO round-trip tests cover session and ticket anomaly scopes. Migration tests cover pre-iteration-v4 normalization, payload sanitization, and failure when selected legacy evidence lacks a stable event ID.
- UI tests cover exact `(session_id,event_id)` correlation, reject unrelated same-tool failures, committed evidence labels, offline state, and local SQLite hydrated state. Existing OpenCode hydration tests ran as part of `make`.
- The Codex/OpenCode adapter parity fixture confirms equivalent high-failure summary counts and evidence cardinality, that each source-specific reference resolves to a retained failed event, and that both retain the same number of direct evidence events. Source event IDs are intentionally harness-specific.

## Ledger and Backfill Scope

- Final scoped `generate --source codex` updated only `alspaces/AL-56/agentic_metrics.json` and its metrics table in `SUMMARY.md`. No workspace-wide backfill or migration command was run, and no other ticket ledger changed.
- The regenerated AL-56 ledger contains 24 sessions and 231 retained events. All 231 events have distinct source IDs; all 231 unique evidence references resolve, with no unreferenced retained events.
- Against the prior 20 sessions, retained events fell from 619 to 124 (80.0%); event JSON fell from 114,993 to 31,610 bytes (72.5%); serialized session facts fell from 129,048 to 56,865 bytes (55.9%).
- The complete ledger is 5,166 lines / 176,299 bytes versus the prior 7,094 lines / 209,603 bytes (15.9% smaller) while including four additional sessions. The added sessions and event IDs/evidence references account for part of the full-ledger size; the common-session comparison isolates the retention change.

