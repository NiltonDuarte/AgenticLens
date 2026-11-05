# AL-64 iteration 03 behavior correction recheck

Verdict: APPROVED for the bounded LEAN-02 behavior correction scope. No new distinct correctness finding. This is not whole-repository approval or permission to publish.

Reviewer: behavior and integration reviewer; gpt-6.1-sol/high inherited runtime identity supplied by the parent. Run: LEAN-02 / behavior Sol correction recheck. Baseline: c68fbc9079d2f24c624d5de0fe87d313ed095b19. Acceptance boundary: approved iteration 03 PLAN.md and CORRECTION_BRIEF.md; original CONTRACT.md is baseline context.

## Revision integrity

- Verified all 399 entries in `/tmp/al64-03-review/fingerprint-02.json` against the checkout: zero mismatches. Manifest SHA256: `e40b0bb87a233eeeec0da9716cfc51311c1763ecc8752c95b280d6f92d7fc3b8`.
- Full candidate diff SHA256: `929c879c490d2df5e4a6e1963fe92bd6ab1c555be4a4b800e7a3d4bc9488b88c`; exactly matches `git diff c68fbc90 -- backend web docs/OWNED_OCCURRENCES.md`.
- Correction diff SHA256: `7ac350522d006214d89ba013719c2ccedeb2ca36fa1f5a8cef61665c1176d931`. Comparing LEAN-01/02 manifests yields exactly the five paths in `correction-paths.json`: current ledger preparation, historical cost decoding, duration normalization and their two regression test files. No source, arithmetic, frontend or projection file changed.
- Saved pre-fix production overlay files each hash-identically match LEAN-01. Thus the supplied red/green evidence exercises the actual prior candidate rather than a reconstructed approximation.

## Findings and dispositions

| Finding | Severity | Recheck disposition |
| --- | ---: | --- |
| LEAN-WRITE-01: completed cleanup re-emitted rich duration | 5 | Fixed in the inspected candidate. |
| LEAN-READ-01: historical typed aliases bypassed duration equality | 4 | Fixed in the inspected candidate. |
| LEAN-READ-02: legacy nested numeric null becomes zero | 3 | Existing accepted debt, unchanged and nonblocking. Owner: backend input compatibility maintainer; follow-up unassigned. |

There are no new findings requiring a numeric severity. Existing finding IDs and original discovery attribution remain owned by the parent's consolidated records; this report records verification by the reviewer identified above.

## Consequential behavior checked

**Successful explicit cleanup.** Traced MigrationRunner and backfill from their public entry points into `prepareCurrentLedger` and the `WithoutMigration` publication branch. Preparation validates original input through `NormalizeFormerGenerationMetadata`, validates domain duration and migration status, then applies duration normalization to the raw cleanup candidate before removing migration metadata. Both publishers consume that candidate; neither substitutes the original rich bytes. Draft positive contribution becomes the positive flat scalar; draft and explicit scalar zero disappear. Total is preserved. No source discovery or private receipt is required for cleanup.

The durable cleanup test exercises actual migrate, backfill and force-backfill branches with total/portion 10/7 and 4/0. Its recursive raw comparison allows object order changes while independently preserving nullable costs, the `1.2500` lexeme, identity, source, tokens, tools, duplicate skill occurrences and absent/empty/nonempty optional categories. It separately checks no retired duration members, positive-only output and exact original migrate backup bytes. The inspected pre-fix log fails on all three publishers; affected corrected race suites pass.

**Original validation and recovery.** The new normalization remains after original duplicate-member validation and the completed/pending/partial consistency checks. A temporary metadata-free candidate is runtime-validated; the original recovery metadata is not deleted from the input/domain ledger. Existing `TestMigrationCommandsValidateRawCleanupBeforePublication` rejects invalid status, duplicate root members and missing owner tools without rewriting the ledger or creating `.bak`/`.tmp`; omitted/null completed step arrays remain accepted. Existing completed-cleanup coverage retains dry-run immutability, idempotence, audit reporting and exact backups. MigrationRunner still validates/prepares before backup, then publishes by temporary file and atomic rename. These paths are included in the supplied passing full backfill package race run. I found no new rollback or recovery-state bypass.

**Historical input.** `decodeVersionedLedger` invokes the raw guard before either v1-v4 or v5 typed decode. The guard traverses both supported flat and session-owned invocation layouts and validates duration names, duplicate duration members, integers/ranges and old/new equality before Go can case-fold values. It does not transform rows, ownership or input bytes. Historical unrelated duplicate/unknown fields remain tolerated through `readDurationObjectFields(..., true)`; current-object parsing retains its existing stricter behavior. The typed historical decoder still owns flat owner extraction, cost policy and other migration behavior.

The 90-case durable historical table covers versions 1-5 and both layouts. Lean, draft, equal-dual and unrelated-field controls preserve independently specified owner/source/user/model, 10/7 duration, token 31 and unavailable cost; conflicts, aliases and duplicate flat members reject. Input immutability is asserted. The actual LEAN-01 overlay fails on 40 alias/duplicate-duration cases while valid controls remain intact; corrected affected race verification passes. Guarding collection rows before typed decode addresses the demonstrated overwrite consequence without imposing a new unrelated historical field policy.

**Nested-null compatibility.** `DecodeProvisionalSeconds` and the legacy evidence struct are unchanged by this correction. I reran the supplied isolated `TestReviewLegacyDurationNullProbe` overlay on LEAN-02: both complete/null-provisional and provisional/null-completed inputs still return runtime and migration errors of nil. The intentionally stricter rejection probe therefore exits 1, as it did in the supplied exact-baseline parity evidence. This is observed preservation of accepted historical behavior, not a passing rejection test. Flat scalar null remains independently rejected. No new legacy-null policy is required here.

## Verification reused and limits

Inspected `/tmp/al64-03-test-corrections/pre-fix.log`, `affected-race.log`, `cli-race.log`, overlay originals, correction test bodies and TEST_REPORT.md correction attribution. Reused the tester's passing race commands for migration/adapters/backfill/capture and selected registered-history/cache/refresh CLI regressions on the frozen correction. No broad suite was duplicated. Executed only the bounded nested-null compatibility overlay described above and revision/hash verification.

LEAN-01 duration arithmetic, source extraction, frontend/API projections and six original boundary assertions carry forward because their files hash-identically match LEAN-01 and this diff adds no producer, arithmetic or consumer behavior. Shared duration normalization is a changed dependency; its cleanup preservation, canonical output, raw compatibility and registered-cache/history callers are specifically covered by the supplied affected suites and this source trace. Native refresh evidence remains valid for unchanged source/replacement behavior; it is not evidence that the earlier binary exercised this new completed-metadata cleanup branch.

Parent still owns integrated `make`, final standalone embedded-asset verification, final approval and delivery. The earlier broad Go race run's CLI installation timeout remains recorded separately; the isolated retry and this scoped approval do not convert it into a clean whole-suite run. No new native harness experiment, platform matrix, product/test edit, telemetry, commit or publication was performed.
