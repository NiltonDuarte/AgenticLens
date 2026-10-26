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

## Implementation Follow-up After Round 2

A subsequent tester check found that retry evidence selection could bind a summary to the last same-tool error run on the same turn, even when the detector had emitted that summary for an earlier, longer run. The selector now consumes exact event indexes for that fallback and resolves retry summaries by their source references. A regression covers a five-error streak followed by a two-error run on the same turn. `PruneEventsForPersistence` now propagates selection errors.

Round 3 then identified cross-session interleaving as another source of retry boundary mismatch. Retry detection now partitions by session and sorts each session chronologically; emitted summaries carry exact `(session_id,event_id)` refs resolved against persisted events. A regression interleaves three session A failures with session B tool calls and verifies only A's three calls are retained. The full `make` suite passed after these changes; the follow-up still awaits a separate reviewer pass.

## Round 3 Verdict: Changes requested

The same-turn retry mismatch is fixed: the detector carries exact event indexes for each run, and persistence associates each generated summary with one distinct run. The regression with five failures, a success separator, and two later failures on the same turn verifies that only the first five events are referenced. Propagating errors from `PruneEventsForPersistence` also removes its previous silent-empty fallback.

One ordering mismatch remains for concurrent sessions. `PersistSparseAnomalyEvidence` derives retry summaries with `DetectAnomalies(ledger)`, which reads `ledger.EventFacts()` in session-grouped order. It then computes retry run indexes from `fullEvents`, globally sorted by timestamp. The detector keeps one active streak and treats another session's tool call as a break, so interleaving events can change which runs it sees. For example, three same-tool failures in session A at times 1, 3, and 5 are contiguous in grouped input and produce a summary; failures from session B at times 2 and 4 interleave those rows in the sorted selector input, so no matching three-event candidate is found and persistence errors. Detection and index selection need the same per-session chronological stream, with indexes mapped back to selected events, plus an interleaved-session regression.

The ledger and backfill checks still show no metric ledger changes; no backfill or generation was run.

## Round 4 Verdict: Approved

The cross-session ordering finding is resolved. Retry detection now partitions events by owning session, sorts each partition chronologically, and emits exact source references with each streak. The selector uses those references to resolve retained rows in the globally ordered event set. For all-missing IDs, the index fallback reaches the existing stable-ID validation and fails extraction; partially identified evidence also fails closed because the reference count cannot satisfy the detected streak.

The added regression has session A failures at t1/t3/t5 and session B successful calls at t2/t4. It verifies the A streak survives interleaving and only its three identified rows are retained. The same-turn five-error/separator/two-error regression remains and checks exact refs for the first run. `PruneEventsForPersistence` now returns selection errors to its caller.

I found no further material issue in this retry-evidence follow-up. The working tree contains no ledger changes, and no generation or backfill was run. The tester report records the full `make` suite passing with 722 frontend tests. Real AL-56 ledger conversion, idempotence, and real before/after size remain for the authorized final generation step.

## Final AL-56 Generation Check

After round 4 approval, the parent ran the scoped Codex generation for AL-56 only. The resulting ledger has 231 retained events; all have unique IDs, all 231 evidence references resolve, and every retained event is referenced. No other ticket ledger changed and no workspace-wide migration/backfill ran. For the 20 sessions shared with the prior ledger, event rows fell from 619 to 124 (80.0%), event JSON from 114,993 to 31,610 bytes (72.5%), and serialized session facts from 129,048 to 56,865 bytes (55.9%). The full ledger is 176,299 bytes (15.9% smaller) and includes four additional sessions. These generation measurements were collected after the code review and are reported separately from the reviewer’s implementation approval.
