# Independent test effectiveness review — A12

**Scope:** AL-70 iteration 03, test-effectiveness focus; baseline `54d58be40c74864d645c5168ba23820a21389c01`; candidate fingerprint `974d1449fc440115b563067461d43b9b6381baf60e178ff45e6854f884b10e6a`. I verified all 29 supplied manifest file hashes against the checkout (zero mismatches). This is a scoped review, not a whole-repository certification. Run ID and runtime model/effort metadata were not supplied.

## Verdict: APPROVED

I found no blocking test-oracle gap in the assigned A12 scope. The new SQLite-backed OpenCode tests cover absent, null, blank and observed model names in v1/v2 metadata-only extraction, with zero-token rows and message-derived turn/iteration counters; the v1 assistant-message cases check tokens, cost and duration, including separation from literal `default`. The v2 negative cases still require `model.id`, so the changed metadata fallback is not allowed to weaken assistant validation. Fixtures call the strict extraction entry point and check checkpoint identity and source-owner save/reload where applicable.

Codex tests distinguish absent map entries from a present empty model through late attribution, event turn lookup, and observed `default`/`unknown` turns. They assert row identity, tokens, duration, turns, usage totals and compaction/user-iteration counters. The live SQLite-to-HTTP test covers SQL NULL, blank and native/object names, non-null empty-string wire representation, node counts, and byte-for-byte database immutability. The repeated recorded-ledger HTTP test checks native historical names, per-model rows/counters/cost/tokens across two reads, and unchanged ledger bytes. Markdown and frontend tests retain row counts and ordering while presenting empty as Unknown; frontend API/UI fixtures exercise model and thread tables, live views and debug summaries without mutating source fixtures. These assertions would fail for the principal regressions in the A12 contract: collapsing empty with observed names, dropping metadata-only rows, losing late turn attribution/counters, emitting null/omitting live model fields, or rewriting stored history.

## Hypotheses assessed

- **Missing-model default collides with observed `default` or `unknown`: refuted by mixed v1 extraction, Codex identity cases and UI/summary formatter cases.**
- **Codex empty model is mistaken for an absent turn-map entry, or late context loses earlier evidence: refuted by dedicated presence and late-attribution tests with multiple turns and preserved totals.**
- **A12 removes metadata-backed zero-token aggregates or weakens v2 validation: refuted by v1/v2 metadata-only cases and required-model rejection cases.**
- **Display formatting mutates raw keys, merges rows, or alters counts/order: refuted by unchanged input assertions, distinct rendered row counts/percentages, API fixture immutability, and Markdown raw-group order/count checks.**
- **Reads relabel or rewrite historical facts: refuted by repeated API reads and byte comparisons for source DB and ledger fixtures.**

**Verification evidence:** Reviewed the frozen diff and new test bodies; independently checked manifest hashes. Reused tester evidence at `/tmp/al70-a12-tester-evidence.json`: affected backend packages, focused frontend unit/consumer suites, and `git diff --check` passed. No suites rerun. No material test-effectiveness limit found within this assignment; parent integration and broader delivery verification remain outside this verdict.

## Revision recheck — backfill A12-TEST-01

**Rechecked revision:** baseline `f6a2f5439a91c622341d72f7b8e838760b2dee0b`; r2 candidate fingerprint `60b88d9ec2ab79b0e92a21cc5acd06e2e7a182b27d622e556ea12c53d1b9178c`. All 30 r2 manifest hashes match the checkout; the original 29 candidate paths retain their r1 hashes. This recheck was limited to the affected backfill oracle and preservation around it. Runtime identity reported for this pass: `gpt-6-luna/medium` (source-observed).

The coordinator confirmed **A12-TEST-01 (severity 3)**: on a complete native Codex refresh whose metadata-only owner has no model, the old test expected `default`, while the approved representation is empty. The correction changes only that expected model and its adjacent explanatory comment. It still requires exactly one aggregate, zero turns and `CostUnavailable`; thus it does not weaken row, activity or cost-state behavior.

The surrounding table cases and assertions remain intact: two backfill passes, incomplete-refresh status where source/children are unavailable, exact equality of the unrefreshed remote owner, preservation of recorded-empty other-source shape and known unrelated findings, canonical validation and reload, no obsolete availability marker, explicit empty arrays for completely observed categories, unavailable aggregate health, honest summary wording, and finding visibility. The changed complete-owner case still checks that stale ordinary work is replaced only for that owner and that complete-empty analysis is recorded as empty arrays. The change does not mask the other-source/history/category or counter oracles.

**Verification:** the focused four-case test passed in 0.172s (`/tmp/al70-a12-backfill-focused.log`); the complete backfill package with race detector passed in 19.237s (`/tmp/al70-a12-backfill-race.log`). The supplied correction evidence records unchanged backend test input after checks. I did not rerun broader suites. The earlier assertion failure is resolved by aligning its expected model identity to A12; no remaining test-effectiveness finding in this affected oracle. Parent owns the integrated `make` result and final delivery gate.

**Recheck verdict: APPROVED.**
