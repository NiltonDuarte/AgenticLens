# AL-64 iteration 04 bounded data/failure correction recheck

**Verdict: APPROVED** for the assigned CP-02 data-integrity and failure-handling scope. No new distinct numeric finding. This is not approval of the integrated delivery gates or unverified native caller/continuation behavior.

## Revision and identity

- Baseline and candidate HEAD: `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192`.
- Requested responsibility: reviewer-data, bounded CP-02 recheck. Parent-assigned identity: gpt-6-luna / medium; runtime metadata was not independently available, so no runtime attribution is asserted.
- `/tmp/al64-04-review/fingerprint-02.json`: all 411 SHA-256 entries verified; 0 missing, 0 mismatched. Manifest file SHA-256: `512721e4192e55dcad98f2f1d7a080360178b73ce016c0225b50840dc69ff629`.
- `/tmp/al64-04-review/correction.diff`: SHA-256 `c8cd7bf243b285d68b50d095538da34b9ddbaa68e4c230700fa73342f8848522`; exactly three paths: `backend/internal/adapters/opencode/schema_v2_session_extraction.go`, `backend/internal/adapters/opencode/checkpoint_test.go`, and `backend/cmd/agentic-lens/checkpoint_generate_test.go`.
- `/tmp/al64-04-review/candidate-02.diff`: SHA-256 `b426a55c735e39e4331ed5e1b03bd0095d7283c472bf64d484e2301b506486f9`.
- Read the correction brief, approved PLAN, assigned persona/shared context/canonical review procedure, implementation report and final test/runtime reports. No sibling persona/review conclusions were loaded for the initial assessment. Product and test sources remained read-only; this assigned report is the only checkout write.

## Findings

None. No severity score applies.

## Hypotheses and evidence

1. **Stale v2 running-tool evidence can still produce exact SourceOpen feedback despite native idle.** Refuted for the correction. In `schema_v2_session_extraction.go`, the bulk `time_idle` query is separate from numeric `activeSessions`; only `CheckpointInterval.SourceOpen` consumes its result. A native idle at/after message start suppresses exact SourceOpen, including idle after the numeric cutoff. A valid positive idle before message start preserves legacy support. Missing column preserves old-schema behavior; NULL is known and retains support. Malformed, fractional, zero or negative metadata cannot be treated as an active interval. The existing “latest message” and running/completion guards remain in place.
2. **Unavailable optional metadata can reject extraction or erase measured facts.** Refuted. Query, scan, iteration and close failures in `schemaV2CheckpointIdle` yield unavailable metadata rather than an extraction error. A missing map entry cannot establish SourceOpen. The mandatory numeric activity query remains a separate path. The adapter preserves active/provisional duration, tokens and cost independently of feedback qualification. Existing tester reports include isolated query/scan failure and missing-row probes; iteration/close failure behavior was inspected in the implementation, not dynamically injected here.
3. **Idle qualification changes known-short suppression, owner replacement, or publication/accounting.** No counterexample found. `DurationInterval` and accumulation precede checkpoint evidence construction; the correction does not alter active-session selection, invocation facts, owner/model/row identity, completion, or replacement. The CP-02 cases assert numeric 180 and input 17 remain intact for long qualified cases, while known 179999 ms cases remain silent. Existing reports establish full generate publication behavior; this recheck did not duplicate those subprocess tests.
4. **Later native records or optional-feedback failure can falsely imply currentness.** Existing latest-message qualification remains, and tester mutation evidence shows removing supersession guards fails both-family assertions. Optional feedback is not a success gate. The correction does not add persisted feedback, a portable-ledger dependency, a caller authority, or process-liveness claim.

## Verification

- Independently verified every pinned manifest entry and both diff hashes above.
- `git diff --check HEAD -- <three correction paths>`: PASS.
- From `backend`: `go test -race ./internal/adapters/opencode -run '^TestOpenCodeV2CheckpointIdleBoundaryPreservesNumericFacts$' -count=1`: PASS (1.225s).
- Reused final CP-02 tester evidence rather than rerunning broad suites: full OpenCode race package 6.156s, focused OpenCode/CLI race tests 1.713s/7.972s, public unit run including 190 frontend units, and expected red results for the original idle overlay and supersession mutants. Exact logs and broader unchanged-suite evidence are recorded in `TEST_REPORT.md`.

## Limits

Native generic-caller/current-interval handoff and natural-completion-to-next-turn refresh remain unverified. Production F03/F04 remain the accepted fallback; stronger native F01/F02/F05 support is not claimed. Parent owns integrated `make`, standalone binary/assets checks, final consolidation and delivery. No additional data/failure evidence gap blocks this bounded correction approval.
