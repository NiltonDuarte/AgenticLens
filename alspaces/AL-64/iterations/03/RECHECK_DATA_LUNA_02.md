# AL-64 iteration 03 bounded data/failure correction recheck

## Verdict: APPROVED

Scoped to LEAN-WRITE-01 and LEAN-READ-01 corrections on the frozen LEAN-02 candidate. This is not final ticket approval and does not cover parent-owned integration gates.

## Revision and evidence

- Baseline / checked `HEAD`: `c68fbc9079d2f24c624d5de0fe87d313ed095b19`.
- Frozen manifest: `/tmp/al64-03-review/fingerprint-02.json`; SHA-256 `e40b0bb87a233eeeec0da9716cfc51311c1763ecc8752c95b280d6f92d7fc3b8`. Independently recomputed all 399 listed path hashes; zero missing paths or mismatches.
- Supplied five-path correction diff SHA-256: `7ac350522d006214d89ba013719c2ccedeb2ca36fa1f5a8cef61665c1176d931`.
- Supplied full candidate diff SHA-256: `929c879c490c2df5e4a6e1963fe92bd6ab1c555be4a4b800e7a3d4bc9488b88c`.
- Inspected: `duration_decode.go`, historical migration dispatch/decoding in `cost_decode.go`, current cleanup preparation and both backfill/migration publication callers, correction tests, implementation and test reports, and accepted LEAN-READ-02 baseline evidence in `REVIEW_TESTS_SOL.md`.

## Findings

No introduced blocking data/failure finding in the bounded correction scope.

LEAN-READ-02 remains visible compatibility debt (severity 3): nested legacy draft numeric `null` is interpreted as zero. The cited disposable probe against both the candidate and archived c68fbc90 has identical outcomes. The correction leaves that legacy numeric interpretation unchanged; flat scalar `null` remains rejected. I do not infer a stricter old-wire policy from this recheck.

## Hypotheses checked

- **Completed-migration cleanup canonicalizes duration while retaining other persisted facts.** `prepareCurrentLedger` validates and decodes the original input, then uses `NormalizeLedgerDurationJSON` only for the cleanup projection. That normalization validates duration members before returning original bytes when no duration rewrite is needed; when draft metadata or zero scalar must be removed, only invocation duration representation is projected. The metadata cleanup then removes migration metadata, and existing `FormatLedgerJSON` handles the established metadata policy. Completed cleanup in both ordinary/force backfill and `MigrationRunner` consumes the same candidate. The new command-level test exercises all three publishers with a positive draft contribution and explicit zero, requires only total plus positive scalar in output, and recursively compares every other raw JSON fact, scalar lexeme, null, array multiplicity and optional-field presence. Migration backup is checked byte-for-byte against original input. This supports preservation of original/recovery bytes and non-duration persisted facts for these paths.
- **Historical aliases/conflicts reject before typed decoding without widening historical policy.** Schema 1–5 dispatch calls `validateHistoricalDurationInputs` before typed collection decoding. The guard visits flat and session-owned invocation arrays and validates raw duration keys/values, including alias rejection, duplicate duration fields, and conflicting dual forms. It does not rewrite input. `readDurationObjectFields(..., true)` retains duplicate tolerance for unrelated historical members; the regression control keeps duplicate `model` and opaque metadata accepted. The 90-case table spans versions 1–5 and both layouts with lean/draft/equal-dual controls, conflicts, aliases, and duplicate duration members. Accepted controls assert owner identity, numeric facts, input tokens and unavailable-cost state remain intact.
- **Original/recovery and raw-byte behavior:** migration backup preservation has a direct byte-equality assertion. The completed cleanup candidate avoids duration rewriting for already canonical positive/no-provisional inputs, and normalizes draft/zero inputs without changing unrelated JSON values. Existing backfill publication/recovery paths retain their validation-before-publish ordering; no new failure ordering or source requirement is introduced by this correction.
- **Unchanged iteration scope:** Numeric duration math, source-family extraction and completion replacement remain outside these two correction paths; the candidate manifest is unchanged and supplied affected race evidence covers the touched backend migration/adapters/backfill/capture paths plus selected CLI paths. Frontend/projection behavior has no dependency on these backend-only fixes and remains carried forward from the reviewed candidate evidence. The original six corrected boundary assertions and accepted LEAN-READ-02 parity evidence remain as cited in the test report; this recheck does not claim new independent source or frontend verification.

## Verification and limits

- Reused `/tmp/al64-03-test-corrections/affected-race.log`: migration, adapters, backfill and capture affected packages PASS.
- Reused `/tmp/al64-03-test-corrections/cli-race.log`: selected CLI preservation/refresh race cases PASS.
- Reused `/tmp/al64-03-test-corrections/pre-fix.log` and the original-file overlay description in `TEST_REPORT.md`: correction oracles fail against saved original files and pass on the corrected code, including cleanup publishers and historical duration alias/duplicate cases.
- Inspected the test report's exact correction assertions and the raw affected-race log; did not repeat broad suites.
- `git diff --check` passed at review time. No source or test files were edited by this reviewer.

Remaining limits: this scoped review does not establish final integrated `make`, standalone embedded assets, the broad Go race suite (the recorded broad run has one CLI installation deadline failure with a passing unchanged isolated retry), or the uninspected branches of all persistence workflows. Parent owns those gates and final consolidation. No distinct numeric or source-family finding was discovered in this bounded data/failure recheck.
