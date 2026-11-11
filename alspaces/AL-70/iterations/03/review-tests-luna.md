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
