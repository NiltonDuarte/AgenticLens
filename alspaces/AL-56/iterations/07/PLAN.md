# AL-56 Iteration 07 Plan: Keep Anomaly Evidence Local

## Objective

Remove persisted anomaly evidence references and event rows from the schema-v4 ticket ledger. Keep compact anomaly summaries, session invocation facts, and tool totals. When AgenticLens runs locally, rebuild full events and anomaly evidence from the selected harness source instead of reading event records from the ledger.

## Design

- Persist no `sessions[].events` rows and no anomaly `evidence` references. Keep session-owned and ticket-level anomaly summaries and the existing per-session tool call/error totals.
- Keep event facts and event IDs transient during extraction. Detect anomalies from the complete in-memory stream before serialization; do not derive them later from a partial event stream.
- Record the harness source once in the ledger so the local events endpoint can choose the right source without inferring from session IDs. Legacy ledgers without a source remain unhydrated unless the source can be selected explicitly; do not guess.
- Extend local hydration for Codex to locate rollouts from the ledger's root session/invocation IDs under `CODEX_HOME`, parse linked child rollouts, and rebuild event facts and anomaly associations for the API response. Keep the existing OpenCode SQLite hydration path.
- Treat hydrated events and any event-to-anomaly associations as request-local data. Do not write source events, evidence IDs, prompts, reasoning, or tool payloads back to the ledger.
- The v4 contract remains unreleased; no schema version bump or historical compatibility path is needed. Normalize existing v4 event/evidence fields away during the AL-56 generation/migration path.
- Update only AL-56's ledger during its final scoped generation. Do not run workspace-wide migration/backfill or change other ticket ledgers.

## Acceptance Criteria

- Canonical v4 JSON contains no event arrays/rows and no anomaly evidence references. It preserves anomaly summaries, invocation facts, and per-session tool totals.
- Codex and OpenCode generation derive the same anomaly summaries from complete transient event streams, while writing no event/evidence data to the ledger.
- The local events endpoint rebuilds Codex event facts from the selected rollout tree and OpenCode event facts from SQLite. It returns explicit hydrated/source metadata; missing or unknown source data fails closed to an empty unhydrated response while summaries remain available.
- Local anomaly detection can associate rebuilt events with stored summaries for UI highlighting without persisting those associations.
- Privacy checks confirm event payloads and transient evidence do not enter the ledger or unrelated ticket artifacts.
- AL-56 generation is idempotent, preserves invocation/tool totals, and changes no other ticket ledger. No workspace-wide migration/backfill command is run.
- Compare the resulting AL-56 ledger with the current 5,166-line / 176,299-byte baseline and report exact lines/bytes. The final output contains no event or evidence fields and is 1,691 lines / 53,208 bytes (67.3% fewer lines, 69.8% fewer bytes).

## Deliverables

- Session-owned schema-v4 anomaly summaries without persisted event/evidence records.
- Local Codex and existing OpenCode event hydration with request-local anomaly association.
- Updated AL-56 plan, verification report, review, summary, PR text, and final scoped telemetry.
