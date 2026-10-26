# AL-56 Iteration 08 Plan: Derive Anomaly Text

## Objective

Remove persisted anomaly `title` and `description` strings. Keep `rule_id`, `severity`, and structured `context`, and derive display text from those fields where anomalies are presented.

## Design

- Canonical v4 anomaly records contain only `rule_id`, `severity`, and `context`. They remain session-owned or ticket-level as established in iteration 05/07.
- Remove both string fields from the domain/persistence contract where feasible. If an API boundary needs presentation text, compute it from rule and context at response/presentation time; never write it back to the ledger.
- Support formatting for every current rule: `TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`, `CONTEXT_SATURATION_SPIKE` (spike and ceiling cases), and `SUBAGENT_REWORK_LOOP`.
- Keep anomaly detection, thresholds, severity, context values, counts, and rule identity unchanged. Provide a safe rule-name fallback if required context is absent.
- Normalize title/description fields away when reading existing schema-v4 ledgers; do not add a schema version bump or a historical compatibility layer.
- Update only AL-56's ledger during its final scoped generation. Do not run workspace-wide migration/backfill or change other ticket ledgers.

## Acceptance Criteria

- Canonical anomaly JSON contains no `title` or `description` fields and retains `rule_id`, `severity`, and `context`.
- CLI/API/UI anomaly displays derive concise readable text for every supported rule from structured context; malformed or incomplete context produces a safe fallback.
- Existing anomaly detection results, severity, and aggregate/session counts remain unchanged.
- Tests cover the current rule variants, including context spike versus ceiling breach and optional session/turn fields.
- Final AL-56 generation contains no persisted event/evidence/title/description data; invocation/tool totals and anomaly counts remain intact. No other ticket ledger changes and no workspace-wide backfill runs.
- Record exact AL-56 line/byte change from the current 1,691-line / 53,208-byte baseline; a serializer simulation estimates about 1,593 lines / 42,241 bytes if no other data changes.

## Deliverables

- Summary-only anomaly facts with derived display text.
- Focused formatter/schema coverage and full repository verification.
- Updated AL-56 summary, PR description, iteration review, and final scoped telemetry.
