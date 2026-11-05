# Data review — AL-64 iteration 03

**Reviewer:** data integrity, gpt-6-luna / medium
**Candidate:** frozen at `c68fbc9079d2f24c624d5de0fe87d313ed095b19`; branch `fix/AL-64-active-turn-duration`
**Scope:** numeric duration compatibility, migration/publication, retained/cache history and source replacement.
**Manifest:** `/tmp/al64-03-review/fingerprint.json`, SHA-256 `b05c2065769609e7e29cbb7e4b8158d804fc2b38ca905cea9b6a2bdbd1f237b5`; verified all 399 entries with zero mismatches.

## Findings

### DATA-LUNA-01 — raw cleanup publication retains the retired duration object

**Severity:** 7/10. **Classification:** introduced canonical-write defect.

**Location:** `backend/internal/adapters/backfill/current_ledger_preparation.go:20-70`, `backend/internal/adapters/backfill/migration_runner.go:240` (the same raw cleanup publication is used by backfill at `backend/internal/adapters/backfill/engine.go:292`).

**Trigger:** a current schema ledger has a valid legacy `duration_evidence` invocation and is accepted by migration or backfill. The shared migration normalizer canonicalizes the duration for typed decoding, but the publication path selects `WithoutMigration`, built from the original raw bytes with only migration metadata removed.

**Expected:** canonical target writes publish the positive flat `provisional_seconds`, remove `duration_evidence`, preserve total, owner identity, unrelated fields, and caller baseline.

**Observed/consequence:** migration/backfill can complete and preserve the typed numeric amount while writing the original rich blob back to disk. A later read can normalize it again, but the promised canonical representation transition has not occurred; old explanatory metadata remains in target output and repeated migrate can report current without completing cleanup.

**Evidence:** source trace through `prepareCurrentLedger` and both publication sites; raw `WithoutMigration` is selected instead of normalized facts. This review did not execute a disposable public-runner probe. The defect is directly observable from the selected bytes and is bounded to these raw cleanup paths.

**Required correction:** normalize only invocation duration members in the validated raw cleanup candidate before publication, preserving the raw path's unrelated-field/category and rollback guarantees.

### DATA-LUNA-02 — legacy schema duration aliases bypass raw conflict validation

**Severity:** 5/10. **Classification:** introduced compatibility validation gap in historical migration.

**Location:** `backend/internal/domain/migration/cost_decode.go:133-175, 245-263` (typed schema v1-v5 decoding); compare `duration_decode.go:120-165`.

**Trigger:** a v1-v5 invocation contains a canonical duration member plus a case-folded alias such as `Provisional_Seconds`, and the alias value matches the legacy blob while the canonical flat value conflicts. Go's case-insensitive JSON struct decoding can overwrite the canonical field before `DecodeProvisionalSeconds` compares values.

**Expected:** reject duplicate/noncanonical duration spellings and conflicting raw flat/legacy contributions before migration. The canonical spelling and the legacy representation must agree as supplied.

**Consequence:** ambiguous historical input can be silently converted into one numeric value, losing evidence of the conflict during canonical migration. This undermines the newly required equal-dual-form validation.

**Evidence:** source trace: historical `decodeVersionedLedger` unmarshals directly into typed v1-v5 inputs; its invocation path reaches `DecodeProvisionalSeconds` only after Go has collapsed case variants. The raw duration alias guard exists only in the current/numeric normalizer. No disposable probe was executed.

**Required correction:** run raw member spelling/duplicate validation on historical invocation records before typed decode for both session-owned and flat legacy layouts, without tightening unrelated historical fields.

## Hypotheses assessed

- **Ordinary current-schema numeric decode:** the new current-schema normalizer validates total, flat value and legacy draft object before strict decoding; dual values are compared and normalized. No loss path found in the inspected code.
- **Cache and retained owner preservation:** `decodeCurrentSnapshot` normalizes temporary owner bytes before strict decoding; cache storage remains untouched. DTO/history paths carry total and provisional as numeric fields, and clone-based retained publication avoids mutating the caller's baseline. No additional source receipt requirement found.
- **Refresh replacement/idempotence:** source extraction constructs total and provisional together; publication/reconciliation is owner-scoped. The provided native refresh evidence is relevant but not a substitute for parent integration verification. No separate loss/double-counting path found in this bounded trace.
- **Accepted source-state estimate:** open source intervals can accrue through the explicit cutoff; the stopped-with-open-marker overestimate is accepted. No liveness requirement or completion inference is raised here.
- **Unchanged registered Codex disabled-delegation cache behavior:** treated as pre-existing and outside this change; no new duration-specific consequence established.

## Verification and limits

- Frozen manifest: 399/399 path hashes match.
- Inspected the approved iteration 03 plan, canonical portable-ledger contract, migration duration codec, historical migration decoder, capture cache decoder, duration DTOs, source extraction and relevant publication call sites.
- Reused evidence stated in the brief: post-alias affected Go race checks pass; frontend 868 tests/typecheck/lint pass; prior broad race run had a CLI timeout and an unchanged isolated retry passed; native history rewrite/refresh evidence passed. These do not replace the required integrated `make` gate.
- No tests were run in this review. No repository product or test files were changed. Findings are source-traced and should receive bounded publication/decoder probes during correction. Parent owns fixes, integrated verification and consolidation.

## Verdict

**CHANGES_REQUESTED.** Two concrete data-boundary defects remain in canonical publication and historical input normalization. The verdict applies only to the frozen candidate and scoped data review; it is not a whole-repository assessment.
