# AL-56 Iteration 06 Plan: Persist Sparse Anomaly Evidence

## Problem Evidence

The current AL-56 ledger is 7,094 lines and 209,603 bytes. It contains 619 persisted events taking about 114,350 compact-JSON bytes (about 55% of the file): 608 `tool_call` events and 11 compactions. Of the tool calls, 132 are errors and 487 are marked completed. Those completed calls are retained because any anomaly on a turn currently causes every event from that turn to be retained. The event objects contain no payloads, but repeated timestamp, turn, role, type, tool, status, duration, and context fields still dominate the ledger.

## Objective

Keep a small `events` list containing only direct evidence for detected anomalies. Persist concise anomaly summaries so the complete anomaly results do not need to be recomputed from this partial evidence stream. Preserve per-session tool call/error totals. Keep full event extraction in memory and allow local source logs to hydrate a complete timeline when available.

This finalizes the unreleased schema-v4 contract; no version bump or legacy-event compatibility path is needed.

## Recommended Evidence Contract

- Compute anomaly summaries from the complete in-memory event stream before selecting evidence. Store session-specific anomalies under their owning session with rule ID, severity, title/description, and relevant context; do not repeat the session ID in nested summaries.
- Store aggregate anomalies that combine facts across sessions, such as high tool failure rate, in a ticket-level `anomalies` collection on the fact ledger. The summary remains ticket-scoped, but it must reference the exact direct evidence events that contributed to the anomaly. For a tool failure rate, retain each failed tool-call event contributing to the numerator and identify it by its owning session and stable source event ID. The denominator comes from session tool totals. Never persist an aggregate anomaly that cannot be traced to its contributing records.

  ```json
  {
    "rule_id": "tool_high_failure_rate",
    "severity": "warning",
    "title": "High tool failure rate: bash (30.0%)",
    "description": "Tool 'bash' failed 3 out of 10 calls",
    "context": {
      "tool_name": "bash",
      "count": 3,
      "total_calls": 10,
      "failure_rate_pct": 30.0,
      "threshold": ">20%"
    },
    "evidence": [
      { "session_id": "session-a", "event_id": "call-17" },
      { "session_id": "session-b", "event_id": "call-4" },
      { "session_id": "session-b", "event_id": "call-9" }
    ]
  }
  ```

- Persist only the smallest event records needed to demonstrate an anomaly. Every retained event has a stable source event ID and owning session; summaries refer to these using `(session_id, event_id)`. Deduplicate overlapping evidence by that identity. Do not retain every successful event merely because it shares a turn with an anomaly. If a contributing event cannot be assigned an ID reliably, fail extraction with a record-level diagnostic rather than save an untraceable summary.
- For a tool retry-streak anomaly, retain the ordered error sequence that satisfies the rule threshold. For a context-saturation jump, retain the before/after context samples; for a ceiling breach, retain the sample that crosses the ceiling. For subagent rework loops, retain only the minimal alternating child-session role-transition sequence when it can be reconstructed unambiguously from invocation parent links and the summary role pair/cycle count. Match the number of detected sequences to persisted summaries; if child-session evidence is absent or duplicate/mismatched sequences make attribution ambiguous, keep the summary without event evidence rather than guessing. Invocation facts may still support the summary without event evidence.
- Per-session tool facts carry calls/errors and provide denominators for ticket-level aggregate anomalies. The aggregate summary also references each failed event that contributes to its numerator; retain those failed event rows with their source IDs.
- Keep event fields payload-free and minimal: timestamp, turn, role/tool as needed for the rule, event type/status, and context or duration only where it supports the anomaly evidence.
- `DeriveMetricsLedger` must combine stored session-specific and ticket-level anomaly summaries for display and counts. It must not re-run event-sequence detectors over sparse evidence or recompute aggregate summaries and emit duplicate/incomplete anomalies.

The timeline will show anomaly evidence only when local source logs are unavailable; it will no longer show normal activity around an anomalous turn. Keep existing OpenCode SQLite hydration unchanged. Do not add Codex source-log hydration to the events API in this iteration. If OpenCode hydration is unavailable, return the sparse committed evidence with an explicit unhydrated/source indicator. Tool error totals remain available even when individual error events are not direct evidence for a detected anomaly.

## Alternative

Keeping every error and compaction event while removing completed anomaly-turn events would retain 143 of the current 619 rows before adding context/role evidence. This is simpler and preserves a broader offline error/compaction timeline, but it retains rows that do not directly evidence any detected anomaly. The recommended sparse-evidence rule is smaller and aligns storage with the user's request; choose the broader error timeline only if it is needed offline.

Minifying JSON would reduce line count but retain the same 209 KB of facts. It is a formatting workaround, not a data-size fix.

## Implementation Scope

1. Add compact session-owned summaries for session-specific anomalies and a ticket-level `anomalies` collection for cross-session aggregates. Derive both from complete in-memory events and per-session tool facts during Codex/OpenCode generation.
2. Retain a sparse session `events` list containing only direct evidence selected for those summaries. Remove turn-wide retention. Preserve all relevant tool counts/errors in the existing per-session tool facts.
3. Update anomaly aggregation to consume stored summaries for canonical sparse ledgers. Ensure generation from full transient events produces the same summaries and does not recompute from partial persisted evidence.
4. Keep existing OpenCode SQLite hydration unchanged. Update the events endpoint and ticket detail UI only as needed to label sparse committed events as anomaly evidence and unhydrated when OpenCode hydration is unavailable. Do not add Codex source hydration in this iteration.
5. Do not run a workspace-wide backfill/migration command and do not modify other tickets' ledgers. Update only the current AL-56 ledger during the normal final generation step, and only after the user authorizes that generation. Other-ticket ledgers remain for the user to migrate on the machine with their source data.
6. Keep any AL-56 ledger update atomic and idempotent. If a legacy event cannot be assigned to a session, fail with a record-level diagnostic rather than inventing an ID or silently retaining it.
7. Update schema docs and tests only for anomaly summaries, sparse event evidence, and timeline fallback. Avoid unrelated migration, API, UI, or telemetry cleanup.

## Acceptance Tests

- Canonical v4 serialization keeps a small per-session `events` array, session-owned event-sequence anomaly summaries, and ticket-level aggregate anomalies. Nested session summaries do not repeat session ID; ticket-level summaries contain stable references to the direct evidence events, and each retained event has a stable source event ID. No persisted event/summary contains prompt, reasoning, tool input/output, or error-message payloads.
- Codex and OpenCode derive identical summaries from full transient events and select the same direct evidence records.
- For each anomaly rule, test exact evidence selection and identity: ordered retry error sequence; context before/after pair and ceiling sample; minimal unambiguous role-transition evidence; and every failed call contributing to the cross-session high-tool-failure numerator. The aggregate summary includes `(session_id, event_id)` references, `tool_name`, errors/count, `total_calls`, `failure_rate_pct`, and `threshold`.
- Deriving metrics from full in-memory events and from the resulting compact ledger produces the same anomaly summaries, tool call/error totals, and invocation totals.
- Existing OpenCode source hydration provides full timelines when its local database is available. Codex event-API hydration is unchanged/out of scope. Without OpenCode hydration, the endpoint/UI returns only sparse evidence, clearly marked unhydrated, while showing the persisted anomaly summaries.
- Verify the AL-56 current ledger conversion during its authorized final generation: all invocations and per-session tool totals remain unchanged, repeat generation is idempotent, and every retained event supports a persisted anomaly summary. No other ticket ledger is touched and no workspace-wide migration command is run.
- Compare the AL-56 ledger against the current 7,094-line / 209,603-byte baseline. The output must retain at most 25% of the baseline event rows and reduce total file size by at least 40%; record exact counts and bytes in the test report.
- Event parsers and migration fail clearly on unassignable or unidentifiable evidence and do not fabricate session or event IDs. Rework evidence must be grouped through explicit invocation parent/child relationships; independent groups cannot be concatenated into a synthetic cycle.

## Deliverables

- Updated v4 session anomaly-summary and sparse-event contract, generation/aggregation/API consumers.
- AL-56-only ledger migration retaining direct anomaly evidence and removing turn-wide event expansion.
- Focused Go and TypeScript tests, full `make`, and before/after ledger-size comparison.
- Updated iteration review, ticket summary, PR description, and final telemetry in the existing workflow's required order.
