# AL-64 iteration 03 bounded test-effectiveness recheck

**Outcome: APPROVED** for the two correction paths and their regression oracles only.

## Revision and scope

- Baseline: `c68fbc9079d2f24c624d5de0fe87d313ed095b19`.
- Candidate: frozen LEAN-02, 399-path manifest `/tmp/al64-03-review/fingerprint-02.json`, SHA-256 `e40b0bb87a233eeeec0da9716cfc51311c1763ecc8752c95b280d6f92d7fc3b8`.
- Verified all 399 manifest paths against current worktree bytes: zero mismatches.
- Exact correction patch: `/tmp/al64-03-review/correction.diff`, SHA-256 `7ac350522d006214d89ba013719c2ccedeb2ca36fa1f5a8cef61665c1176d931`; it contains exactly five paths: `current_ledger_preparation.go`, `migration_runner_test.go`, `cost_decode.go`, `duration_decode.go`, and `engine_test.go`.
- Full candidate patch: `/tmp/al64-03-review/candidate-02.diff`, SHA-256 `929c879c490d2df5e4a6e1963fe92bd6ab1c555be4a4b800e7a3d4bc9488b88c`.
- Dispatch identity: LEAN-02 test recheck, Codex `gpt-6-luna`, requested medium reasoning effort.

## Findings

No new distinct blocking finding. No severity assigned because no new defect was substantiated.

LEAN-WRITE-01's oracle is durable and reaches the actual `MigrationRunner`, ordinary backfill and force-backfill publishers. The self-contained completed-metadata fixture contains a positive draft contribution and an explicit zero. Assertions require successful cleanup to remove retired duration metadata, retain total 10/provisional 7 and total 4/zero, preserve unrelated scalar lexemes, nulls, optional field presence and array multiplicity, and preserve the original migration backup bytes. The recursive raw comparison allows object-key reordering but checks all other fields and scalar encodings. The helper is not an implementation copy: it compares the published output with a separately assembled expected fixture. Source inspection confirms the metadata cleanup candidate now runs duration normalization after original input normalization and migration-state validation, before removal of migration metadata and publication.

LEAN-READ-01's oracle reaches `Engine.Migrate` for schema v1-v5, both flat and session-owned invocation layouts. Its 90 expanded cases cover lean, draft and equal-dual valid controls; ordinary and aliased conflicts; aliases of total/evidence; duplicate flat duration members; and a valid control for historical unknown fields and duplicate non-duration model keys. Valid controls assert owner/source/user, model, total, contribution, token input and unavailable-cost state. Rejection cases assert migration errors, and every case asserts input bytes remain unchanged. The correction guard inspects raw known invocation collections before typed decoding; inspection confirms it limits alias/duplicate treatment to duration members, preserving the old decoder's unrelated-field tolerance.

## Verification evidence and hypotheses

- Reused `/tmp/al64-03-review/fingerprint-02.json`, the exact correction/full-candidate patch hashes above, and the correction sections of iteration 03 `IMPLEMENTATION.md` and `TEST_REPORT.md`.
- Reused pre-fix overlay evidence in `/tmp/al64-03-test-corrections/pre-fix.log`: the final tests fail against saved original production files on 40 historical alias/duplicate-duration cases and all three cleanup publishers. This is direct before/after oracle evidence, not a mutation-equivalence claim.
- Reused corrected affected race evidence: `go test -race ./internal/domain/migration ./internal/adapters ./internal/adapters/backfill ./internal/adapters/capture` PASS (`affected-race.log`), and the selected CLI race command PASS (`cli-race.log`). The report records 90 historical cases and three cleanup variants green on the corrected code.
- Reviewed the actual production normalizers and migration/cleanup callers. No additional probe or broad suite was run.
- Flat `provisional_seconds: null` remains rejected by `DecodeProvisionalSeconds`; this is an existing numeric input contract. Nested null inside the legacy draft numeric object remains accepted as zero under the established decoder behavior. Treat that as accepted pre-existing LEAN-READ-02 debt: earlier baseline/current parity evidence is reused, and this correction keeps using the same decoder without introducing stricter legacy-null policy.

## Carried-forward scope and limits

The corrections touch backfill cleanup and migration input validation only. They do not alter source duration arithmetic/extraction, API/frontend projection behavior or UI code. Those unchanged areas carry forward from LEAN-01 review and verification with this dependency rationale: the five-path correction is downstream of the existing numeric facts and projection contract, and the frozen 399-path manifest proves those unrelated source and frontend paths did not change during this recheck. This report does not repeat or expand their prior conclusions.

This is a test-effectiveness recheck of the two named fixes, not final delivery approval. It does not replace parent-owned integrated `make`, standalone embedded-asset verification or final consolidated review. Synthetic fixtures establish the named production adapter/domain boundaries; they do not establish native harness runtime behavior. No broad Go suite result is claimed here.
