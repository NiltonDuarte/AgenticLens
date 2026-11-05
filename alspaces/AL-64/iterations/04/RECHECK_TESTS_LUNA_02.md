# AL-64 iteration 04 bounded test recheck

Outcome: **APPROVED** for the assigned test-effectiveness scope.

## Revision and scope

- Frozen baseline: `d8a7c14cfd9f0ee626fedc33c70f92bd928d9192`.
- Candidate: `/tmp/al64-04-review/candidate-02.diff`, SHA-256 `b426a55c735e39e4331ed5e1b03bd0095d7283c472bf64d484e2301b506486f9` (24 paths).
- Bounded correction: `/tmp/al64-04-review/correction.diff`, SHA-256 `c8cd7bf243b285d68b50d095538da34b9ddbaa68e4c230700fa73342f8848522` (exactly three paths: v2 extraction and two existing test files).
- The supplied 411-entry baseline manifest SHA-256 matches `512721e4192e55dcad98f2f1d7a080360178b73ce016c0225b50840dc69ff629`; every listed file hash matched (0 mismatches).
- Persona: reviewer-tests; dispatch-acknowledged identity gpt-6-luna / medium (parent attribution correction from the recorded execution response). Review is limited to correction tests/oracles and their production boundary; no product or test files were changed by this review.

## Findings

No new distinct finding.

## Hypotheses and evidence

- **CP-REV-01 idle overlay:** The v2 table exercises missing `time_idle`, NULL, positive old idle, equal-start, before/equal/after-cutoff idle, malformed text, fractional, zero and negative values. The assertions independently retain active duration 180, provisional duration 180 and input tokens 17, then require F03, F04 or silence as appropriate. The 179999ms valid and unusable-idle controls both remain silent, so rounded subtotal fallback cannot bypass known-short suppression. The optional lookup returns unavailable evidence on query/scan/iteration/close errors; the production path treats unavailable evidence as nonfatal and does not change numeric extraction. Failure-lookup behavior is source-traced; no separate injected SQL failure case was run.
- **CP-TEST-01 future supersession:** Both schema-family cases add a later user message beyond the operation cutoff, preserve active/provisional 180 and input 17, and require qualified fallback for the long case and silence for the short case. Reported disposable original-code mutant results fail both long assertions with F03 where F04 is expected; the unmutated focused race suite passes.
- **Original correction red evidence:** The preserved original v2 source overlay fails the eight introduced idle qualification cases, while active/null/old/short controls pass. This directly demonstrates discrimination for the production correction without modifying shared source.
- **Readonly stderr preflight:** The fixture opens stderr with `os.Open`, then requires the actual descriptor write to fail with zero bytes before invoking generation. The subprocess still asserts successful publication and supported ledger facts; this checks the relevant descriptor mode directly, not merely file permissions.
- Reused tester evidence: focused OpenCode/CLI race tests pass (1.713s/7.972s); full OpenCode race suite passes (6.156s). No broad suite was repeated. Parent-owned integrated `make` and standalone gates remain outstanding.

## Limits

This recheck validates deterministic adapter and CLI test oracles, not a native OpenCode caller/continuation cycle. Native caller and continuation remain unverified; the documented F03/F04 fallback is within the accepted scope. This is not approval of the complete feature or final delivery gates.
