# AL-56 Iteration 06 Review

## Round 2 Verdict: Approved

The two findings from the first review are addressed. Rework detection and evidence selection now group child events by their explicit invocation parent; an adversarial fixture interleaves independent parent groups and verifies that they do not form a synthetic loop. The Codex/OpenCode adapter fixture now compares anomaly scope/rule/count and evidence cardinality, checks the selected events resolve through each adapter's source-specific IDs, and confirms equivalent retained-event counts.

The revised evidence contract is implemented across the inspected paths: retained events have stable IDs; summaries reference exact `(session_id, event_id)` pairs; high-failure summaries retain and reference each failed tool-call contributing to the numerator; DTO validation checks uniqueness and reference semantics; and unidentifiable required evidence fails generation/migration with diagnostics. UI matching checks exact evidence references before applying older heuristic matching. Tool/invocation facts remain preserved, event payload fields are stripped, and normal turn-wide retention is gone.

## Verified in Round 2

- `PersistSparseAnomalyEvidence` derives summaries before reducing events and sanitizes prompt, reasoning, tool input/output, and error-message payloads. Direct evidence selection is fail-closed when event identity is missing.
- Retry sequences, context samples, child-role transitions, and cross-session tool failures have focused exact-reference coverage. Aggregate summaries carry the failure count, denominator, rate, threshold, and references to all numerator events; persisted rows are validated as failed tool calls for the named tool.
- Session summaries are nested without serialized `session_id`; ticket summaries are root-scoped. `DeriveMetricsLedger` uses stored summaries instead of re-detecting against sparse evidence. Repeated persistence is covered for stored summaries.
- The synthetic size fixture retains 132/619 event rows (21.3%) and reduces the reported JSON from 135,551 to 38,130 bytes (71.9%). It represents a fixture, not conversion/idempotence of the real AL-56 ledger; real-ledger generation remains deferred as directed.
- Invocation totals and per-session tool totals are checked through persistence. UI tests cover exact reference matching, unrelated same-tool failures, committed-evidence labeling, and local SQLite hydration.
- Codex and OpenCode use source-specific stable IDs; the parity test verifies both adapters' summaries/evidence resolve to their own retained events and compare the shared anomaly facts. Source IDs are appropriately not required to be byte-identical across harnesses.
- `git diff --name-only -- 'alspaces/*/agentic_metrics.json'` and the corresponding status query are empty. The test report states that no workspace-wide migration/backfill was run.

## Verification source

Reviewed the revised iteration 06 plan and test report, the current diff in the domain event/anomaly and aggregation code, schema DTOs and validators, OpenCode/Codex ID assignment and adapter tests, migration normalization, UI correlation, and `TECHNICAL_GUIDELINES.md`. The tester's report records full `make` and `git diff --check` as passing; this review did not rerun tests, generate telemetry, mutate ledgers, or run backfill. No material implementation finding remains. The only unperformed verification is the real AL-56 ledger conversion/idempotence and real before/after size comparison, which remain for the explicitly authorized final generation step; the report does not present synthetic results as real-ledger results.
