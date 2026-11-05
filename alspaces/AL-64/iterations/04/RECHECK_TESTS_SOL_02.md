# AL-64 iteration 04: bounded test-effectiveness correction recheck

Verdict: APPROVED for this assigned correction scope. New distinct findings: none. This does not approve integration, native capabilities or publication.

Run: checkpoint_tests_recheck02; phase: CP-02 correction recheck. Assigned profile: gpt-6.1-sol / high, inherited from parent; no independent runtime model/usage metadata was queried. Baseline/HEAD: d8a7c14cfd9f0ee626fedc33c70f92bd928d9192. Initial shared Git status was already dirty/staged; those changes were preserved. Product and tests remained read-only; repository writes were limited to this report. Disposable probe files are outside Git.

## Frozen revision and scope

Verified all 411 entries of `/tmp/al64-04-review/fingerprint-02.json`: zero missing or mismatched hashes. Its SHA-256 is `512721e4192e55dcad98f2f1d7a080360178b73ce016c0225b50840dc69ff629`. Independently reproduced the 403-path tester fingerprint `29b4e1215134acdc5baa39caed9bc88d68f329b9eff5b17ab49ef2444553b6ba` using the documented path/NUL/bytes/NUL algorithm.

Reproduced the exact full candidate diff (24 paths) from baseline: SHA-256 `b426a55c735e39e4331ed5e1b03bd0095d7283c472bf64d484e2301b506486f9`. Reproduced the exact three-path correction diff: SHA-256 `c8cd7bf243b285d68b50d095538da34b9ddbaa68e4c230700fa73342f8848522`. Comparing the CP-01 and CP-02 manifests identifies only `schema_v2_session_extraction.go`, `opencode/checkpoint_test.go` and `cmd/agentic-lens/checkpoint_generate_test.go` as changed inputs.

Read the approved iteration plan, own persona/shared review rules, canonical review procedure, applicable regression/testing/portable-ledger rules, implementation and final tester/capability evidence. No sibling persona report or conclusions were loaded during this investigation.

## Correction hypotheses and dispositions

- CP-REV-01 (severity 4/10), shared independent discovery by behavior/integration and data/failure reviewers, both gpt-6.1-sol / high: resolved within this scope. The 13-case durable test uses real SQLite, public strict source extraction and the actual domain selector. Each case asserts extraction success, one owner/model row, active180, provisional180 and input17 before asserting selection. Absent-column, NULL and positive idle strictly before message start retain F03. Equal-start, pre/equal/post-cutoff idle, text, fractional, zero and negative idle select precisely F04 with the expected source/owner/model/row/subtotal. Both 179999ms controls remain silent despite rounded180. The optional read only affects SourceOpen; running-state/DurationInterval/token/cost accumulation is unchanged. Correction author: coder gpt-6.1-sol / high; durable regression verification: tester gpt-6.1-sol / high; this reviewer independently checked the oracles and source trace.
- The original-source red evidence is discriminating: the tester's pre-idle source bytes match the staged CP-01 production bytes (SHA-256 `75a1337442bcc21a4f6c35f48d0ababa2319da6de7e83ced1a6ee79fab1674ee`). `/tmp/al64-04-tester-corrections/pre-fix.log` records exactly eight failures at the F03-versus-F04 assertion, after numeric assertions passed; it is not an unrelated extraction/compiler failure. The active/old/absent/short controls do not fail.
- CP-TEST-01 (severity 3/10), discovery by test-effectiveness reviewer gpt-6.1-sol / high: resolved. Both source-family tables now insert a separate future user record after cutoff, preserve active180/provisional180/input17, require F04 at180000ms, and require silence at179999ms. Inspection confirms the supplied mutants remove exactly v1 currentMessages qualification and v2 last-message qualification from CP-01. The recorded mutation log fails both long supersession controls at the feedback assertion, with F03 instead of the required F04; neither test stops at an unrelated error. No product change was needed for this gap.
- Optional metadata failure remains conservative and nonfatal. In a disposable Go overlay, only the new optional query projection was changed to request a nonexistent column; the test overlay independently preflighted that exact SQL error and changed NULL/old-idle expectations to F04 under lookup failure. All 13 extraction/selection cases passed, including numeric facts and known-short suppression. Query/scan/iteration/close failures statically return a nil map; missing entries prevent SourceOpen without removing interval timing. The execution probe specifically exercises query failure, not every database driver failure branch.
- CP-OBS-01 stderr caveat is refuted. `os.Open` supplies an O_RDONLY descriptor; the new preflight requires zero bytes and a nonnil write error before the real child command starts. This does not depend on chmod or effective UID. The existing successful publication and ledger/pending-journal assertions then establish that diagnostic write failure does not fail generation. Reused focused CLI race evidence includes this subtest.

## Executed and reused evidence

Executed the disposable query-failure probe from backend:

`go test -overlay /tmp/al64-04-review/recheck-tests-sol-02/query_failure.json ./internal/adapters/opencode -run '^TestOpenCodeV2CheckpointIdleBoundaryPreservesNumericFacts$' -count=1 -v`

PASS (13 subcases). Exact injected error and results: `/tmp/al64-04-review/recheck-tests-sol-02/query_failure.log`. These are synthetic fault-injection results, not native runtime proof or a product modification.

Read and reused final correction logs under `/tmp/al64-04-tester-corrections/`: `unit.log` (public units, including190 frontend tests), `green-focused-race.log` (OpenCode1.713s/CLI7.972s), `opencode-race.log` (full OpenCode race6.156s), `pre-fix.log` (eight expected red oracles), and `supersession-mutant.log` (two expected red oracles). No duplicate broad suite was run.

## Carry-forward rationale and limits

Initial selector, configuration, exact renderer copies, owner/row grouping, portable history and CLI publication evidence carries forward because the frozen delta changes only one private v2 metadata qualification and two focused test files. No selector/config/renderer/ownership/persistence/API/frontend implementation changed. The correction has no new ledger authority or accounting gate. Carry-forward is revision/dependency reasoning, not newly executed whole-feature or native verification.

Generic trustworthy caller/current-interval handoff and natural completion followed by next-turn refresh remain unverified. Production F03/F04 fallback is the accepted supported boundary; pure F01/F02/F05 tests do not certify installed native support. Full make, standalone assets/binary, integration, consolidated approval and delivery remain the parent's gates. No children, staging, commits, telemetry or PR actions were performed.
