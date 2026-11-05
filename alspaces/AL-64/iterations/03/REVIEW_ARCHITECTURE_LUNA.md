# AL-64 iteration 03 architecture review

## Assessment

**APPROVED — scoped architecture focus.** Reviewed candidate baseline `c68fbc9079d2f24c624d5de0fe87d313ed095b19` on `fix/AL-64-active-turn-duration`, requested reviewer identity `gpt-6-luna/medium`. The frozen manifest `/tmp/al64-03-review/fingerprint.json` has the brief's SHA-256 `b05c2065769609e7e29cbb7e4b8158d804fc2b38ca905cea9b6a2bdbd1f237b5`; all 399 listed paths exist and match their recorded hashes. This is a bounded architecture review, not a whole-repository certification.

No blocking architecture finding. The approved representation is carried as one numeric `ProvisionalSeconds` fact alongside the total. Duration interval arithmetic remains pure in `internal/domain`; source adapters supply the explicit cutoff and source-established activity. No duration reasoning configuration, persisted coverage model, aggregate cutoff graph, or liveness protocol was introduced.

## Findings

None (severity 0).

## Hypotheses investigated

- **Duplicated duration semantics across layers:** Refuted. `domain.InvocationFact` owns the numeric scalar and `domain.ValidateDuration` enforces its relation to the total. `migration.DecodeProvisionalSeconds` is the compatibility boundary for original, draft, and lean JSON. DTO unmarshalling, current-schema migration, and capture-cache snapshot decoding invoke normalization before their respective strict readers; these are separate ingestion paths, while outbound DTO/domain conversion serializes the same numeric field. Legacy shape validation is not repeated in the frontend.
- **Domain or dependency-boundary leakage:** Refuted. `domain/duration.go` retains deterministic interval and rounding calculations without I/O or outer-package imports. Codex/OpenCode adapters collect source facts and pass explicit times into it. `domain/migration` owns JSON compatibility and time-string validation at the wire boundary. No `any`/`interface{}` duration model or duration-specific config/CLI switch appears in the inspected production paths.
- **Duplicated or misleading aggregate state:** Refuted. Workflow, fleet, comparison, and time-bucket calculations sum the scalar independently while retaining the existing total duration. HTTP projection DTOs and TypeScript types carry only the scalar; UI labels use positive provisional seconds and do not interpret absence as completed coverage. I found no second completed-seconds field or persisted explanation graph in the reviewed source paths.
- **Potentially redundant compatibility abstraction:** The normalization helper is used by current-ledger migration, DTO decoding, and cached snapshot decoding. Although this adds a small JSON normalization pass at each owning boundary, the common helper avoids three independent interpretations of the retired draft form and preserves each reader's own strict validation for non-duration fields. No concrete maintenance or correctness harm was established.

## Inspected paths and evidence

Inspected the duration domain and its aggregations; invocation and ledger DTO conversion; migration input normalization and current-schema handling; capture-cache snapshot decoding; Codex and OpenCode source accumulators; metric/API projection DTOs; frontend duration validation, types, and progression aggregation; plus the approved iteration plan, implementation/test/runtime evidence, portable-ledger contract, and shared review procedure. The candidate source was checked against the complete frozen manifest before the review.

The source trace supports the lean contract: old draft evidence is validated and reduced to a numeric amount at backend readers; canonical output uses the scalar; preserved history/cache paths retain numeric facts and owner records; downstream aggregates do not add provisional seconds to the total a second time. The OpenCode `durationIncomplete` accumulator is transient source-adapter state controlling whether a completion timestamp is emitted; it is not persisted as a new duration coverage claim.

## Verification and limits

Reused the frozen candidate evidence in `TEST_REPORT.md`: frontend tests (868 across 38 files), typecheck and lint passed; affected post-alias Go race checks passed; the isolated final-source CLI installation retry passed after the broad pre-alias Go race run had one timeout. The broad Go race run remains recorded as failed, and the isolated retry does not establish a clean broad run. Native Codex history rewrite/refresh evidence is reviewed as bounded source evidence, including its initial compatibility failure and corrected candidate observations. I did not rerun full suites or integrated `make`, which remains the parent-owned required gate.

Review limitations: no independent native OpenCode runtime probe was run for this numeric representation; unchanged OpenCode boundary evidence and deterministic source tests are reused. This verdict covers architecture and maintainability on the frozen candidate only; it does not replace behavior/data/test reviews, integrated verification, or delivery approval.
