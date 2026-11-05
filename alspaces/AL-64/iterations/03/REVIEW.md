# AL-64 iteration03 active review

Status: INCOMPLETE pending integration checks. Frozen candidate LEAN-01; baseline c68fbc90;399-input manifest b05c2065769609e7e29cbb7e4b8158d804fc2b38ca905cea9b6a2bdbd1f237b5. Scope is approved numeric-only duration, not superseded reasoning/liveness plans.

## Roster

| Responsibility | Model / effort | Phase | Outcome | Report |
| --- | --- | --- | --- | --- |
| reviewer-behavior | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [REVIEW_BEHAVIOR_SOL.md](REVIEW_BEHAVIOR_SOL.md) |
| reviewer-behavior | gpt-6-luna / medium | discovery | APPROVED | [REVIEW_BEHAVIOR_LUNA.md](REVIEW_BEHAVIOR_LUNA.md) |
| reviewer-data | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [REVIEW_DATA_SOL.md](REVIEW_DATA_SOL.md) |
| reviewer-data | gpt-6-luna / medium | discovery | CHANGES_REQUESTED | [REVIEW_DATA_LUNA.md](REVIEW_DATA_LUNA.md) |
| reviewer-architecture | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [REVIEW_ARCHITECTURE_SOL.md](REVIEW_ARCHITECTURE_SOL.md) |
| reviewer-architecture | gpt-6-luna / medium | discovery | APPROVED | [REVIEW_ARCHITECTURE_LUNA.md](REVIEW_ARCHITECTURE_LUNA.md) |
| reviewer-tests | gpt-6.1-sol / high | discovery | APPROVED | [REVIEW_TESTS_SOL.md](REVIEW_TESTS_SOL.md) |
| reviewer-tests | gpt-6-luna / medium | discovery | CHANGES_REQUESTED | [REVIEW_TESTS_LUNA.md](REVIEW_TESTS_LUNA.md) |

## Findings and attribution

- LEAN-IN-01, severity7, fixed before formal review: current-schema migration bypassed normalization/validation. Shared implementation discovery by native runtime and adversarial test engineers (tester gpt-6.1-sol/high); coder gpt-6.1-sol/high fixed it.
- LEAN-IN-02, severity4, fixed before formal review: current input duration-key aliases bypassed dual-value comparison. Discovered by coder gpt-6.1-sol/high; tester gpt-6.1-sol/high verified red/green aliases.
- LEAN-WRITE-01, severity5, open: completed-migration cleanup publishes original rich duration blob. Shared independent discovery by behavior, data and architecture reviewers, each gpt-6.1-sol/high, with actual runner probes.
- LEAN-READ-01, severity4, open: historical migration decoding permits alias-hidden conflicting provisional values. Discovered by data reviewer gpt-6.1-sol/high, including end-to-end publication and valid controls.
- LEAN-READ-02, severity3, pre-existing/nonblocking: nested legacy null components remain interpreted as zero. Shared discovery by behavior and test-effectiveness reviewers, both gpt-6.1-sol/high. Exact baseline archive probes confirm parity; canonical new flat nulls correctly reject. Preserve numeric compatibility; no introduced regression or newly approved stricter historical-wire policy is inferred. Owner: backend input compatibility maintainer; follow-up unassigned.

Private structured evidence remains /tmp/al64-03-review. All reporter tuples and dispositions will survive compact retention. All selected first-pass cells completed before the correction freeze was lifted.

## Verification

Frontend868 tests/typecheck/lint and affected post-alias Go-race checks pass. Broad pre-alias Go-race failed one capture-install deadline; unchanged isolated final-source retry passed. Raw failure remains recorded. Authentic19-owner draft normalization and real Codex refresh/repeat passed, original live ledger unchanged. Final integrated make/standalone verification pending; these cells do not authorize publication or claim final approval.

No actual source-state hang/liveness machinery is required; the user accepted provisional estimates. Known initial INT-DATA-01 upstream registry concern remains separately documented in iteration01 and is not silently fixed or newly attributed here.

Luna data review independently confirmed LEAN-WRITE-01 and LEAN-READ-01 through source tracing (reviewer-data, gpt-6-luna/medium); preserve shared discovery with the Sol reporters. Luna behavior scoped APPROVED without findings. Luna architecture and test-effectiveness subsequently completed on unchanged LEAN-01.

All eight first-pass discovery cells completed on unchanged LEAN-01. Luna architecture approved; Luna test-effectiveness reported the same legacy nested-null interpretation (shared discovery with Sol behavior/test-effectiveness). This is verified baseline-compatible behavior, not new flat-null acceptance: new flat null remains rejected. Parent retains it as a nonblocking legacy-compatibility follow-up; no new stricter historical-wire policy is introduced. The two confirmed introduced gaps remain LEAN-WRITE-01 and LEAN-READ-01. Product/test freeze may now be lifted only for those corrections and their regression oracles.

## Correction candidate LEAN-02

Product and tests frozen after two corrections. The 399-path manifest SHA256 is e40b0bb87a233eeeec0da9716cfc51311c1763ecc8752c95b280d6f92d7fc3b8. Exactly five source/test paths changed from LEAN-01: current cleanup preparation, historical cost decoding, shared duration decoding, migration runner tests and migration engine tests. Duration normalization preserves original bytes when no projection is needed; historical checks preserve unrelated-field tolerance and legacy-null interpretation.

Independent tester gpt-6.1-sol/high reproduced 40 historical invalid-input failures and three public cleanup failures against original production overlays; corrected 90-case historical controls and three cleanup commands pass. Affected race checks are being finalized. Bounded correction reviews cover all four focuses on both profiles; unchanged source arithmetic, API/frontend consumers and six original regressions carry forward because none of those inputs changed. Final integrated make and standalone asset verification remain pending.

All eight bounded correction cells APPROVED on LEAN-02: behavior, data/failure, architecture and test effectiveness, each gpt-6.1-sol/high and gpt-6-luna/medium. No new distinct finding. The Sol test reviewer independently passed 97 race-enabled cause/publication/compatibility probes; other rechecks reused the affected tester evidence and source trace, with Sol behavior/data also observing unchanged nested-null behavior. Reports are RECHECK_<FOCUS>_<PROFILE>_02.md. Required final integration and standalone gates remain pending, so consolidated status is INCOMPLETE for those gates only; no blocker remains in the reviewed correction candidate.
