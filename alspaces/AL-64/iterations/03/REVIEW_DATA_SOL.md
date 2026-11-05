# AL-64 iteration 03 independent data review

Verdict: CHANGES_REQUESTED. Two substantiated gaps remain in the approved lean numeric transition. This verdict covers data preservation, input normalization and write/recovery paths, not whole-product approval or merge authorization.

Reviewer: data integrity and failure handling; assigned/inherited model gpt-6.1-sol, reasoning high. This identifies the supplied execution configuration; separate runtime usage/cost metadata is unavailable. Phase: independent discovery. Candidate: LEAN-01, baseline c68fbc9079d2f24c624d5de0fe87d313ed095b19. Parent task label: lean_data_sol. Manifest /tmp/al64-03-review/fingerprint.json SHA256 b05c2065769609e7e29cbb7e4b8158d804fc2b38ca905cea9b6a2bdbd1f237b5. All 399 individual input hashes matched at the start and again immediately before this verdict. No sibling iteration 03 review conclusions were read.

Initial Git status was already dirty with AL-64 planning, implementation and tester changes, including staged planning paths and the untracked duration compatibility codec. Those changes were preserved. Product and test sources remained read-only; only this assigned report and disposable files under /tmp/al64-03-review/data-sol were written.

## Findings

### DATA-SOL-01: completed-migration cleanup republishes the retired duration object

Severity: 5/10. Classification: introduced integration omission against the new canonical-write contract, in an unchanged cleanup branch. Numeric facts are preserved; this is not a claim of numeric data loss.

Locations: backend/internal/adapters/backfill/current_ledger_preparation.go:49; backend/internal/adapters/backfill/migration_runner.go:240; backend/internal/adapters/backfill/engine.go:280.

Trigger: a valid current-v6 ledger contains compatible draft duration_evidence and completed migration metadata. Run explicit migration cleanup, or backfill without recomputation, including backfill with Force=true. The operation validates and decodes the numeric values correctly, removes migration metadata, and reports a successful target write.

Expected: the successful target write carries the flat positive provisional_seconds for every owned fact and removes duration_evidence, while preserving identity, numeric values and all unrelated raw facts/category presence.

Observed: prepareCurrentLedger constructs WithoutMigration from original content, deletes only migration, and formats that raw object. Both publishers select WithoutMigration rather than the normalized domain representation. The written invocation still contains duration_evidence with state, completed_seconds, provisional_seconds, cutoff and basis, and lacks the flat scalar. A subsequent metadata-free migrate is treated as up-to-date, so the successful cleanup does not finish the representation transition.

Evidence: a disposable Go overlay appended TestReviewDataCompletedDraftCanonicalWrites to the existing backfill external test file. Both migrate and backfill subtests executed their public runners against isolated valid files. Migration was actually removed and the decoded total/provisional remained exactly 10/7 with the same owner; both assertions that the retired object must be absent failed. Package test result: FAIL, 0.024s. Command: go test -overlay /tmp/al64-03-review/data-sol/overlay.json ./internal/adapters/backfill -run '^TestReviewDataCompletedDraftCanonicalWrites$' -count=1 -v, from backend/. Probe source: /tmp/al64-03-review/data-sol/migration_runner_test.go. An initial probe-only compilation error referenced an unavailable report field; it was corrected outside the repository before the behavioral run.

Required correction: normalize only duration members in the validated raw cleanup candidate before publication. Preserve original validation, non-duration raw values, optional-category shape, rollback bytes, and caller baseline. Replacing the entire raw candidate through a potentially lossy general DTO projection is not necessary.

### DATA-SOL-02: historical migration bypasses raw duration-key ambiguity checks

Severity: 4/10. Classification: introduced validation gap in the historical path's new lean/dual-form support. Go's general case-folded decoding predates this change; the failure to enforce the new duration equality/alias guarantee here is the regression. This is distinct from the supplied current-v6 alias fix.

Locations: backend/internal/domain/migration/cost_decode.go:133 and :254; compare the guarded reader in backend/internal/domain/migration/duration_decode.go:137.

Trigger: a schema-v5 invocation contains active_duration_seconds=10, provisional_seconds=6, Provisional_Seconds=7, and valid legacy duration_evidence whose provisional contribution is 7. Equivalent risks apply to the shared v1-v4 decoder; the executed reproduction uses v5.

Expected: reject the noncanonical duration alias and conflicting raw flat/legacy numeric contributions before migration or publication. Canonical 6 versus legacy 7 must not become equal through decoder overwrite.

Observed: decodeVersionedLedger unmarshals directly into invocationInput without durationObjectFields/NormalizeInvocationDurationJSON. Go's case-insensitive struct matching overwrites the canonical raw value 6 with the alias value 7. DecodeProvisionalSeconds therefore compares 7 with legacy 7 and accepts. The public migration runner successfully publishes a schema-v6 ledger with flat provisional_seconds=7, erasing the original ambiguity.

Evidence: disposable TestReviewDataHistoricalDurationValidation exercised the real migration engine. Legacy draft, lean numeric and equal-dual controls all passed, preserving exact total/provisional 10/7, owner remote, source codex, user alice and input tokens 31. Ordinary canonical 6 plus legacy 7 was rejected as expected. The case-aliased conflicting input was accepted, so that rejection assertion failed (package FAIL, 0.008s). A separate TestReviewDataHistoricalAliasPublished reached backfill.NewMigrationRunner(nil).Run with target 6 and proved successful on-disk publication of the aliased conflict (package FAIL, 0.020s). Commands: go test -overlay /tmp/al64-03-review/data-sol/overlay.json ./internal/adapters -run '^TestReviewDataHistoricalDurationValidation$' -count=1 -v; go test -overlay /tmp/al64-03-review/data-sol/overlay.json ./internal/adapters/backfill -run '^TestReviewDataHistoricalAliasPublished$' -count=1 -v. Probe sources are under /tmp/al64-03-review/data-sol; repository tests were not changed.

Required correction: apply the same raw duration-member spelling/duplicate checks before typed historical invocation decoding, covering both session-owned and flat legacy layouts, without changing the established permissive policy for unrelated historical fields. Retain the existing valid legacy/lean controls and ordinary conflict rejection.

## Consequential hypotheses assessed

- Loss or double-counting in ordinary original-v6/draft/lean decoding: the shared codec validates total and provisional inclusion, accepts explicit flat zero but omits it on normalization, rejects flat null/noninteger/negative/unsafe/over-total values, and compares equal dual forms. Invocation DTO UnmarshalJSON applies this before strict decoding. Source inspection and the tester's post-alias current-schema regressions refute the ordinary-path hypothesis. DATA-SOL-02 identifies the distinct historical bypass.
- Current migration still bypassing draft normalization: Engine.Migrate now calls NormalizeLedgerDurationJSON before strict current-domain decoding. The supplied fix is present; valid current cases are covered by independent tester evidence. The new finding concerns publication of a separate raw cleanup candidate, not failure of that fix.
- Legacy numeric values silently lost during schema upgrade: independently executed engine controls preserved actual draft/lean total 10 and provisional 7, input tokens 31 and source/owner identity. InvariantValidator compares provisional and total across transformations. No new loss was observed in these valid controls.
- Retained history requiring private source receipts or mutation of the caller baseline: classifyGenerationBaseline accepts the validated portable record; SaveFactLedgerPreservingHistory requires exact owner-domain equality, clones/validates retained DTOs and writes their scalar representation. No added receipt prerequisite was found. Existing TestRetainedDraftDurationNormalizesWithoutMutatingBaseline and native force-only evidence cover this behavior. The original immutable history object is not rewritten by the clone path.
- Registered cache reads requiring source availability or rewriting registry bytes: decodeCurrentSnapshot normalizes draft values in returned temporary bytes, while decode retains raw registry storage. The independent registered-cache test checks 10/7 and exact stored record bytes with source removed. Unsupported retired non-duration payloads continue to require re-extraction. No new cache-write authority was introduced by the inspected normalization change.
- Partial/unavailable refresh dropping known owners or accumulating snapshots: recovery starts from stored sessions, rejects failed/incomplete source owners, and replaces only successfully accepted exact owners; backfill recombines selected and other-harness sessions. Changed duration fields copy total and scalar together. Existing source-boundary/replacement tests and native repeat evidence support retention and replacement without duplicate cumulative amounts. This was source-traced, not reimplemented as a second full source parser.
- AL-66 presence/policy regression from duration normalization: raw non-duration members remain in normalizer maps, and ordinary retained DTO publication preserves optional presence through existing projectors. Existing disabled-category/source tests and 868-test frontend run are reused. DATA-SOL-01's correction must preserve the raw cleanup branch's unrelated shape guarantee.
- Known registered Codex delegation-cache overwrite with disabled delegations: pre-existing unchanged upstream issue from the brief; not reported as introduced by this scalar transition. No new changed consequence was established here.

## Verification and limits

Executed: three bounded overlay tests above, including valid controls and two public migration/backfill publication paths. The overlays use repository production sources at the frozen candidate and self-contained temporary ledgers. They were not race runs and do not replace the parent integration gate. The observed failures are candidate correctness failures, not inaccessible evidence or hypothetical cases.

Reused evidence, with revision distinctions preserved: tester post-alias affected DTO/migration/retained/cache/source checks passed; frontend 868 tests across 38 files plus typecheck/lint passed on unchanged frontend sources. The earlier broad Go race run passed 12 non-CLI test-bearing packages but failed the untouched CLI installation deadline test; its unchanged isolated retry passed in 2.833s. The broad run is not represented as an all-package pass. Native copied authentic 19-owner history force-only rewrite preserved all session objects modulo lean duration projection, and source refresh discovered a legitimate twentieth owner with no lost/duplicate owners. Those native runs used intermediate corrected executables before the final alias guard; they do not certify the final frozen binary or the cleanup branches tested here.

Accepted boundary: source-established open intervals estimate through an explicit transient cutoff. Stopped processes with open markers may overestimate; that risk is accepted. Missing/zero provisional values do not certify completion. No liveness, source-receipt, rich reasoning model, new config or frontend legacy parsing requirement was imposed.

No broad suite was duplicated, no native source store was mutated, and no telemetry, staging, commit or publication occurred. Final integrated make, final standalone assets and consolidation remain parent-owned. Fixes and correction verification for the two findings are pending.
