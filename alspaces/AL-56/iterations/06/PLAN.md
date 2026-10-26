# AL-56 Iteration 06 Plan: Remove Event Rows from Persisted Ledgers

## Problem Evidence

The current AL-56 ledger is 7,094 lines and 209,603 bytes. It contains 619 persisted events taking about 114,350 compact-JSON bytes (about 55% of the file): 608 `tool_call` events and 11 compactions. Of the tool calls, 132 are errors and 487 are marked completed. Those completed calls are retained because any anomaly on a turn currently causes every event from that turn to be retained. The event objects contain no payloads, but the repeated timestamp, turn, role, type, tool, status, duration, and context fields still dominate the ledger.

## Objective

Remove event-level records from the committed fact ledger while preserving derived anomaly summaries, session tool call/error totals, and the ability to hydrate detailed timelines from local source logs when available. Keep event parsing and anomaly detection in memory during extraction. Do not add a compatibility path for old event arrays; this v4 contract is still being finalized.

## Recommendation

Use a compact anomaly-summary fact owned by the session, and stop serializing `events` in each session. During generation or migration, detect anomalies from the complete event stream before discarding it, then store only the anomaly rule, severity, title/description, and relevant context. The owning session supplies session identity, so do not repeat `session_id` in nested anomaly facts. Existing per-session tool summaries continue to preserve calls and error counts.

This keeps the dashboard's anomaly counts and descriptions available offline. The event timeline and individual error/compaction records are available only when the local source database or rollout can hydrate them; when it cannot, the endpoint returns an empty, explicitly unhydrated event list. Detailed event order and failure sequence are intentionally no longer committed.

## Alternatives Considered

### Keep only error and compaction events

This removes the 487 completed anomaly-turn events while preserving 132 error and 11 compaction events (143 rows, roughly 77% fewer events). It is a smaller behavior change and retains an offline failure/compaction timeline. However, the remaining rows still repeat event fields, and retry-streak/context/rework anomaly detection depends on event order that is not reliably reconstructible from aggregate tool counts. This still needs persisted anomaly summaries. Prefer this only if the user requires an offline error/compaction timeline.

### Compact or minify event JSON

This reduces line count without removing event facts, but does not materially reduce the 209 KB payload or information retained. It treats the formatting symptom rather than the event data bloat and is not the recommended fix.

## Implementation Scope

1. Add a session-owned compact anomaly fact to schema v4 and derive it from the full in-memory event stream during Codex/OpenCode generation before persistence pruning.
2. Remove serialized event arrays from the canonical session DTO and fact ledger. Keep events transient for anomaly derivation and source-log hydration.
3. Update anomaly aggregation to consume persisted anomaly summaries when no event stream is present. Prevent duplicate anomaly results when transient events and stored summaries are both available.
4. Update `/api/workspaces/{ticketId}/events` and the ticket detail UI: hydrate a detailed timeline from local OpenCode/Codex sources when available; otherwise return/display an empty unhydrated timeline while keeping anomaly summaries visible.
5. Migrate only the AL-56 workspace ledger in this checkout: derive compact anomaly summaries from its existing event rows, remove event rows, preserve invocations and per-session tool totals, and keep migration notes for any anomaly summary that cannot be derived. Leave every other ticket's workspace ledger untouched; the user will run any other backfill on the machine containing that ticket's source data. Do not run a workspace-wide backfill unless explicitly requested.
6. Update schema docs and tests only for event persistence, anomaly summaries, and timeline fallback. Avoid unrelated migration, UI, or telemetry cleanup.

## Acceptance Tests

- Canonical v4 serialization contains no session `events` arrays or root event collection; session anomaly summaries serialize without repeated session IDs or private event payload fields.
- Codex and OpenCode generation derive the same anomaly summaries from full transient event streams before omitting event records.
- Metrics derived from a generated compact ledger preserve the anomaly summaries, tool call/error totals, and invocation totals from the full pre-persistence facts.
- Full source logs hydrate the event timeline; missing/unavailable sources yield an empty timeline marked unhydrated without hiding persisted anomaly summaries.
- AL-56 migration preserves attributed calls/errors and invocations; repeated migration is idempotent; event collections are absent afterward; the AL-56 file byte and line counts materially decrease. Other ticket ledgers remain byte-for-byte unchanged.
- Tests cover retry-streak, context-saturation, high-tool-failure-rate, and subagent-rework summaries where their current rules are event-dependent.
- Event parsing remains privacy-safe. No prompt, reasoning, tool input/output, or error message is persisted in anomaly summaries.

## Deliverables

- Updated v4 session anomaly-summary contract and generation/aggregation/API consumers.
- Migrated tracked ledgers with event rows removed and anomaly summaries retained.
- Focused Go and TypeScript regression coverage, plus full `make` and ledger-size comparison.
- Updated iteration review, ticket summary, PR description, and final telemetry in the existing workflow's required order.
