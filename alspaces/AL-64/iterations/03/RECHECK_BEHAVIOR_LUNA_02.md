# Behavior recheck: LEAN-02 corrections

## Outcome

**APPROVED — bounded behavior recheck only.** Reviewed frozen baseline `c68fbc9079d2f24c624d5de0fe87d313ed095b19` with the supplied 399-path candidate fingerprint `e40b0bb87a233eeeec0da9716cfc51311c1763ecc8752c95b280d6f92d7fc3b8`. The fingerprint file hash matches; all 399 listed source-file SHA-256 values match the checkout. The exact correction diff contains the five reported paths. Its SHA-256 is `7ac350522d006214d89ba013719c2ccedeb2ca36fa1f5a8cef61665c1176d931`; `candidate-02.diff` SHA-256 is `929c879c490d2df5e4a6e1963fe92bd6ab1c555be4a4b800e7a3d4bc9488b88c`.

No new distinct behavior finding was demonstrated. The only new compatibility restriction observed is limited to ambiguous/case-aliased duration members, as approved. Historical duplicate/non-duration field tolerance remains. The older nested draft numeric-null interpretation remains accepted baseline behavior and was not tightened; new flat scalar null is rejected.

## Traced correction paths

### LEAN-WRITE-01: completed-migration cleanup

For a current-schema ledger, both `MigrationRunner.migrateWorkspace` and ordinary/force backfill call `prepareCurrentLedger` before their metadata-cleanup branch. That preparation first validates the original raw ledger with `NormalizeFormerGenerationMetadata`, decodes the current migration state, validates the cleanup facts, applies `NormalizeLedgerDurationJSON`, removes only the migration member from that candidate, validates the raw runtime ledger, and formats it. Migration publishes this candidate after its DTO validation and optional backup creation; completed metadata-only backfill publishes the same candidate under its existing publication ownership. Source refresh/recompute continues through the ordinary DTO projection path.

The correction therefore reaches the externally published cleanup output rather than only the in-memory projection. The positive draft value becomes flat `provisional_seconds: 7` alongside total `active_duration_seconds: 10`; explicit zero is omitted while total 4 remains. The raw cleanup test compares the full nested JSON while allowing object-key reordering, and retains scalar lexemes, nulls, array multiplicity, and optional categories in absent/empty/nonempty forms. It also checks exact original rollback bytes for the migration backup fixture. The existing `sourceVer >= targetVer` decision still permits completed metadata cleanup on subsequent migration invocations, while after cleanup the removed completed marker routes later calls to the up-to-date result. Ordinary backfill’s completed cleanup branch is likewise not a source refresh and leaves other facts intact.

### LEAN-READ-01: historical duration guard

`decodeVersionedLedger` runs `validateHistoricalDurationInputs` before historical typed unmarshalling for schema versions 1 through 5. The guard visits both root-level flat `invocations` and session-owned `sessions[].invocations`, and inspects only duration members. It rejects duplicate or case-aliased duration keys before Go's case-insensitive struct matching can collapse their values. It then leaves typed historical decoding and migration ownership/identity behavior unchanged. The supplied 90-case test spans all five versions and both layouts: lean, draft, and matching dual inputs preserve owner/source/user/model, total/provisional values, token facts and unavailable-cost state; conflicts and ambiguous duration keys reject; unrelated duplicate `model` members and unknown historical metadata remain accepted. The migration input remains byte-identical.

This preserves input compatibility for the established historical layouts while closing the equality bypass. It does not impose canonical-object duplicate policy on unrelated historical fields.

## Verification evidence and limits

I reused the tester's supplied correction evidence rather than rerunning broad suites: the original-code overlay records 40 historical-boundary failures and three cleanup publisher failures before correction; the corrected historical cases (90 expanded cases) and three cleanup command variants pass. The affected Go race packages (`migration`, `adapters`, `backfill`, `capture`) and selected CLI race tests are reported passing in `iterations/03/TEST_REPORT.md`, with raw logs under `/tmp/al64-03-test-corrections/`. This is review of the frozen source and those reported commands, not an independent test execution.

Carry-forward rationale: these two corrections are confined to cleanup publication and historical migration input decoding. They do not change duration arithmetic, source extraction/normalization, frontend/API projections, or the original six corrected boundary behaviors. The supplied frozen fingerprint and scope-specific call trace show no changed dependency in those areas, so the valid earlier evidence remains applicable to those unchanged behaviors. Parent-owned integrated `make`, standalone embedded-asset verification, and final cross-focus approval remain outside this scoped verdict.
