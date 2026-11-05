# Architecture recheck: AL-64 iteration 03

## Assessment

**APPROVED** for the bounded LEAN-WRITE-01 and LEAN-READ-01 correction scope. This is a focused architecture recheck, not consolidated ticket approval.

- Candidate: frozen LEAN-02 manifest, 399 paths; every path's SHA-256 matches `/tmp/al64-03-review/fingerprint-02.json` (manifest SHA-256 `e40b0bb87a233eeeec0da9716cfc51311c1763ecc8752c95b280d6f92d7fc3b8`).
- Supplied correction diff SHA-256: `7ac350522d006214d89ba013719c2ccedeb2ca36fa1f5a8cef61665c1176d931`; it covers exactly the five listed correction/test paths. Full candidate diff SHA-256: `929c879c490d2df5e4a6e1963fe92bd6ab1c555be4a4b800e7a3d4bc9488b88c`.
- Runtime identity supplied by dispatch: Codex, gpt-6-luna, medium effort. No commit was created for this candidate; the manifest and diff hashes identify the reviewed state.
- Inspected: `current_ledger_preparation.go`, `duration_decode.go`, `cost_decode.go`, migration decode/validation path, migration/backfill regression tests, approved plan, implementation and correction test reports, plus formatter ownership and the relevant review and portability rules.

## Findings

No new distinct architecture or correctness finding in the assigned scope.

The cleanup projection remains layered appropriately: the backfill adapter validates original bytes and migration status, obtains a pure JSON duration projection from the migration/domain package, removes only migration metadata, then applies the established runtime validation and formatting boundaries. The duration helper returns original bytes when no duration rewrite is required and propagates that no-change result through session and ledger traversal. When a rewrite is needed, it edits the duration fields while carrying unrelated raw members through the projection. Cleanup still operates on a separate candidate; the correction does not mutate caller-owned input or the rollback source.

Historical duration validation is a pre-decode guard in `decodeVersionedLedger` for schema versions 1–5. It inspects only the established root `invocations` and `sessions[].invocations` collections, validates duration members through the shared numeric compatibility logic, and leaves typed historical decoding responsible for the rest of each schema. The historical reader tolerates unrelated duplicate fields while rejecting duration duplicates and aliases before Go's case-insensitive typed matching. This maintains the existing historical policy boundary without adding a general raw JSON policy or duplicating duration business rules.

The assigned correction is compatible with the pure-core boundary: normalization and validation perform deterministic JSON/value transformations without I/O, clocks, or adapter dependencies. The adapter retains persistence and publication ownership. I found no new passthrough layer or speculative interface.

## Hypotheses checked and evidence

- **Metadata-only cleanup can publish draft or zero-valued duration metadata:** correction changes `prepareCurrentLedger` to project canonical duration before constructing its metadata-free candidate. The correction tests exercise migration, ordinary backfill, and force backfill; they check total and positive-only provisional values, absence of retired duration fields, unrelated JSON payloads and optional collection presence, exact rollback bytes, and repeat cleanup stability. The tester reports all three publishers failing against saved pre-fix production overlays and passing after correction.
- **Historical typed decode can let a case alias override a canonical duration member:** the guard runs before typed historical decoding and covers flat plus session-owned forms. The correction tests exercise schema v1–v5 with lean, draft, equal-dual, conflicting, alias, and duplicate-duration cases, plus valid controls for unrelated historical duplicates and unknown fields. The tester reports 40 historical failures against the original-code overlay and green corrected race suites.
- **Accepted nested draft numeric null semantics were tightened accidentally:** `DecodeProvisionalSeconds` retains the existing typed draft decode behavior; the new rejection is specific to flat scalar `provisional_seconds: null`. The correction test report records this as existing compatibility debt, not a new policy. I reused the supplied baseline/current parity evidence; I did not independently rerun that separate probe.
- **No-change byte preservation, cleanup rollback, or unrelated history could regress:** the early return in `NormalizeInvocationDurationJSON`, and unchanged propagation at session/ledger levels, avoids re-serializing already-canonical data through this duration projection. The supplied correction tests cover raw scalar lexemes, nullable values, multiplicity, optional-field absent/empty/nonempty states and exact backup bytes. Existing formatter behavior still owns final cleanup layout.

## Verification reused

Reused the correction report and raw logs, without repeating broad suites:

- `go test -race ./internal/domain/migration ./internal/adapters ./internal/adapters/backfill ./internal/adapters/capture`: PASS on the frozen correction candidate.
- Selected CLI race command for registered cache/history and refresh paths: PASS.
- Correction red/green overlay evidence: original-code overlay produced 40 historical and 3 cleanup failures; corrected cases passed.
- The broader `make` gate and standalone embedded-asset verification remain parent-owned and are not claimed here.

## Scope and limits

The report is limited to cleanup ownership, duration normalization boundaries, and the historical duration-only guard on the frozen 399-path candidate. The unchanged source interval arithmetic/extraction semantics, API projections and frontend behavior are carried forward from the earlier review because this correction changes none of those 399-path contents outside the supplied correction. Their correctness is not newly re-certified by this architecture recheck. Parent integration, final `make`, standalone asset verification, and consolidated approval remain outstanding.
