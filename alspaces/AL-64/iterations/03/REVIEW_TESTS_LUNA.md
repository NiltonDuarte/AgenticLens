# AL-64 iteration 03 test-effectiveness review

## Assessment

**CHANGES_REQUESTED** for the frozen LEAN-01 candidate, baseline `c68fbc9079d2f24c624d5de0fe87d313ed095b19`, branch `fix/AL-64-active-turn-duration`. Requested reviewer identity: gpt-6-luna/medium. Frozen manifest SHA256 `b05c2065769609e7e29cbb7e4b8158d804fc2b38ca905cea9b6a2bdbd1f237b5` verified: all 399 listed paths exist and match their recorded SHA256. Scope is test effectiveness and the directly implicated legacy numeric input boundary; this is not a whole-repository review.

One introduced compatibility defect is not caught by the tests: JSON `null` for the legacy draft's numeric component fields is silently accepted as zero by Go's `encoding/json` when unmarshalling into `int64`. The flat scalar's null case is tested, but equivalent null cases in the draft component object are absent. This conflicts with strict numeric normalization and allows malformed historical input to be canonicalized as a real zero contribution.

## Finding

**Severity 5/10 — Legacy draft null components pass numeric validation and can be canonicalized.**

- **Location:** [duration_decode.go](/home/niltonduarte/.t3/worktrees/AgenticLens/t3-727b7383/backend/internal/domain/migration/duration_decode.go:44), with the consistency check at line 54; test gap in [schemas_test.go](/home/niltonduarte/.t3/worktrees/AgenticLens/t3-727b7383/backend/internal/adapters/schemas_test.go:1671).
- **Trigger:** A draft `duration_evidence` object contains `state:"complete"`, `completed_seconds:10`, `provisional_seconds:null`, and a nonblank basis alongside `active_duration_seconds:10`.
- **Expected:** Reject the malformed legacy component as nonnumeric before migration/DTO normalization; retain the explicit distinction between zero and invalid/missing input.
- **Observed:** `json.Decoder.Decode` into the `int64` fields succeeds and leaves `ProvisionalSeconds` at its zero value. The sum, complete-state, basis and overall duration checks then pass. A disposable Go probe of the same standard-library decoding behavior returned `err=<nil>`, `completed=10`, `provisional=0`. The actual production path uses that struct decode at lines 45–48, then proceeds through those checks.
- **Consequence:** `ValidateRuntimeFactLedger`, DTO normalization, and current-schema migration can accept a malformed legacy ledger and write canonical output that removes the invalid legacy blob while presenting the provisional contribution as zero. Existing compatibility tests reject null for the flat scalar but do not cover null for either legacy component. Add legacy `completed_seconds:null` and `provisional_seconds:null` regressions, including a complete-state case that otherwise sums to the total, through the shared runtime-validation/DTO/migration path.
- **Classification:** Introduced input-validation defect in the compatibility decoder; not a preexisting issue or optional suggestion.

## Hypotheses investigated

- **Legacy and lean normalization / current migration bypass:** The shared decoder validates negative, fractional, string, over-total and unsafe flat values, draft state/cutoff/basis/sum constraints, equal and conflicting dual forms, explicit flat-zero conflict, and duration-key aliases. `NormalizeLedgerDurationJSON` routes current-schema migration input through the normalizer. The native evidence records a real earlier failure in this path and successful correction; the correction is supported by the current migration route and compatibility assertions.
- **Retained history and cache publication:** The retained-history test asserts source baseline immutability, owner and numeric preservation, canonical output, and reload equality. The registered cache test asserts read-only decoding preserves identity/numbers and exact cache bytes. The registered-history test asserts canonical publication. These are useful boundary tests; they do not substitute for the null component case above.
- **Refresh/completion/repeat/unavailable history:** The backfill test executes three source families with active snapshot growth, fixed-cutoff repeats, completion replacing provisional contribution, exact owner count/parent identity, unavailable same-harness retention, other-harness retention, aggregate totals, and force-only rewrite. CLI generation checks completion and stable repeated snapshots. Native Codex evidence adds authentic old-provisional completion, source refresh/repeat and owner preservation, while explicitly not claiming fixed-cutoff or other-harness coverage from that probe.
- **Source boundary regressions:** Existing Codex rollout tests retain completed/open, cutoff, future, abort, reversed/malformed completion, and malformed/missing start assertions. OpenCode v1/v2 tests assert their respective open snapshot and completion semantics, including their schema-specific streaming cases. I found no new test-effectiveness defect in these reviewed duration boundaries; prior tester verification is reused for broad affected package execution.
- **Rollups, averages, API and UI:** The test report identifies assertions for workflow/fleet/comparison/progression rollups, HTTP scalar propagation, Markdown labels, frontend decoders and average denominator/filter behavior. Frontend tests, typecheck and lint are unchanged since the reported frozen frontend fingerprint. No broad duplicate frontend run was needed for this focused review.

## Verification and limits

- Manifest verification: 399/399 path hashes match.
- Focused run: `go test -race ./internal/adapters ./internal/domain/migration ./cmd/agentic-lens -run 'TestDurationEvidenceRoundTripLegacyAndValidation|TestRetainedDraftDurationNormalizesWithoutMutatingBaseline|TestRegisteredDraftDurationCacheNormalizesWithoutReadMutation|TestCaptureCommittedHistoryPreservesSupportedStoredEvidence|TestBackfillActiveDurationReplacesSnapshotsAcrossSourceFamilies|TestCLIGenerate_SessionRefresh' -count=1` from `backend/`: PASS. The migration package reported no matching standalone test; migration is exercised by the adapter compatibility test.
- Null behavior probe: disposable `/tmp/al64-03-null-int-demo.go` confirmed `encoding/json` accepts `null` into an `int64` struct field without error. Production-path impact is source-traced through `DecodeProvisionalSeconds`; no production or test file was modified for this probe.
- Reused evidence: frontend 868 tests / 38 files, typecheck and lint pass; post-alias affected Go-race checks pass; native Codex refresh evidence uses real source read-only and a disposable ledger. Broad pre-alias Go race had one CLI install deadline failure, with unchanged isolated retry passing. These results are recorded with their revision limits in `TEST_REPORT.md`; they do not replace the parent's required integrated `make` gate.
- `git diff --check`: PASS.

This review did not read sibling iteration-03 review reports on first pass, run a broad new matrix, mutate source files, or inspect/mutate the live ledger. Parent integration, fixes, consolidation and required integrated verification remain outstanding.
