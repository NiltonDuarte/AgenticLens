# AL-56 Iteration 06 Test Report

## Verification

- `make` — passed (exit 0). This ran `go vet ./...`, frontend typecheck and ESLint, `go test -race ./...`, Vitest (31 files, 719 tests), frontend production build, and Linux/macOS amd64/arm64 binary builds.
- `(cd backend && go test ./internal/domain ./internal/adapters/backfill ./internal/adapters/codex ./internal/adapters -run 'Sparse|Subagent|Anomaly|MigrationRunner_NormalizesPreIterationV4' -count=1)` — passed for all four packages before the source-ID contract was expanded; the final full `make` covers the expanded tests.
- `(cd backend && go test ./internal/domain -run TestSparseEvidenceMeetsLedgerSizeTargets -v -count=1)` — passed. Synthetic fixture: 619 baseline event rows, 132 retained rows (21.3%), 135,551 baseline JSON bytes, 38,130 sparse JSON bytes (71.9% reduction). Retained rows are failed tool calls referenced by the aggregate anomaly.
- `git diff --check` — passed.

## Focused Behavior Checks

- Retry-streak summaries are derived from the full event stream, nested under the owning session without a serialized `session_id`, and retain the error sequence. The metrics assertion compares persisted summaries with detection from the full stream.
- Cross-session high tool-failure summaries are ticket-scoped and include tool name, error count, total calls, failure rate, and threshold. They reference every failed tool-call event by `(session_id,event_id)`; DTO validation rejects duplicate, missing, wrong-tool, wrong-status, or wrong-event-type evidence. Source-less legacy aggregates fail closed with a record-level missing-evidence diagnostic.
- Retry-streak summaries retain the ordered error sequence; context-spike coverage verifies before/after samples. The subagent rework test retains the six role-transition events for three cycles; an interleaved independent-parent test prevents false cycles. Repeated sparse persistence preserves stored summaries. Totals tests compare per-session tool calls/errors and derived invocation totals before and after persistence.
- DTO round-trip tests cover session and ticket anomaly scopes. Migration tests cover pre-iteration-v4 normalization, payload sanitization, and failure when selected legacy evidence lacks a stable event ID.
- UI tests cover exact `(session_id,event_id)` correlation, reject unrelated same-tool failures, committed evidence labels, offline state, and local SQLite hydrated state. Existing OpenCode hydration tests ran as part of `make`.
- The Codex/OpenCode adapter parity fixture confirms equivalent high-failure summary counts and evidence cardinality, that each source-specific reference resolves to a retained failed event, and that both retain the same number of direct evidence events. Source event IDs are intentionally harness-specific.

## Ledger and Backfill Scope

- `git status --short -- 'alspaces/*/agentic_metrics.json'` — empty; no ledger files changed.
- `git diff --name-only -- 'alspaces/*/agentic_metrics.json'` — empty; no other ticket ledger changed.
- No workspace-wide backfill or migration command was run.
- The current AL-56 ledger remains 7,094 lines / 209,603 bytes. Its final sparse conversion and idempotence against the real ledger were not run because no final-generation/backfill authorization was given. The size result above is synthetic and does not claim a real-ledger before/after measurement.

