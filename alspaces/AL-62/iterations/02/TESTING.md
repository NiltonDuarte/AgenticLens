# AL-62 iteration 02 verification

## Scope, actor and current status

Authoritative scope: CONTRACT.md current-acceptance section and iterations/02/PLAN.md. Current runtime/API/persistence is schema6 with session-owned occurrences only; explicit legacy migration drops ownerless rows with counts. Canonical facts belonging to unavailable owners remain portable. Baseline: 19d70574d0e22b00f185c32d048baaf62792743e.

Tester: gpt-6.1-sol/high, inherited dispatch configuration. Tester owns backend tests/fixtures only; backend coder owns production, frontend coder owns all web source/tests/fixtures. The parent owns final review, integration, race/make, telemetry and publication. All regression source fixtures are self-contained synthetic native records; they are not evidence of real historical activity. No tests read live ticket workspaces or private harness logs.

All eight independent discovery runs completed against their frozen pre-correction candidate: seven returned CHANGES_REQUESTED; test-effectiveness reviewer gpt-6-luna/medium returned APPROVED with no findings. The reporting reviewers' probes/source investigations established the failures below. Passing development tests were not approval for those uncovered triggers. The corrected broad backend non-race run completed with all internal packages passing; its sole CLI association-guard regression was corrected and the affected cases passed. No known unresolved test failure remains. Private log: /tmp/al62-final-backend-all.log. No integrated make/race or final review approval is claimed here.

## Correction verification

| Finding | Permanent regression and oracle |
| --- | --- |
| DATA-01 | TestGenerateMalformedSkillArgumentsPreservePreviouslyOwnedSnapshot invokes raw CLI over an initially valid native load, then corrupts arguments, uses numeric call_id42, omits call_id or uses a numeric native name. Every case verifies an incomplete-source diagnostic and unchanged accepted ledger bytes. TestCodexMalformedRecognizedSkillOccurrenceMarksSourceIncomplete covers malformed skill arguments; TestCodexMalformedOuterResponsePayloadMarksSourceIncomplete adds typed outer failures and missing/blank native fields, source-line diagnostics and a warning-free positive control. |
| DATA-02 | TestGenerateExactSourceWithoutPrivateRegistrationReplacesPortableOwner and TestGenerateCopiedLedgerPreservesUnavailableOwnedFactsWithoutPinnedDestination invoke CLI directly without fixture.args auto-registration. Exact complete source replaces its owner across worktree paths; copied supported history and other owners survive without private receipts or pinned destination bytes. |
| DATA-03 | Existing OpenCode snapshot/global/v2 tests now require independently complete parents to refresh while unavailable/failed children remain. TestBackfillCodexCompleteParentRefreshesWhileMissingChildFactsSurvive uses actual native source, Execute and save/reload twice. Owner replacement helper also preserves unavailable descendants. |
| BEH-01 | TestCodexDistinctSkillAttachmentsPreserveNativeMultiplicityAndTimeline covers adjacent distinct attachments and attachments separated by unrelated user content; existing direct selection plus its adjacent attachment remains one activation. |
| BEH-02 | Registered generation tests require complete equal/older/restored/shorter exact snapshots to replace their own facts. Capture reconciliation contrasts successful smaller restore with missing/malformed reads and budgets, which continue preserving accepted facts. |
| BEH-03 | TestTaskNamespaceChildResolutionRemainsReadProjectionAcrossPersistence follows native source through CombineTranscripts, SaveFactLedger/reload and newly ambiguous child metadata. Canonical namespace-only facts never gain a derived UUID. |
| BEH-04 | TestTaskChildResolutionUsesOnlyNativeParentAndAgentPath includes known explicit child metadata disagreeing with a zero-match namespace, plus ambiguous namespace and explicit UUID controls. Raw IDs/outcomes remain unchanged and contradiction is diagnosed. |
| BEH-05 | Existing OpenCode v1/v2 fixtures exercise exact CaptureSessionSnapshot as well as ordinary extraction. TestCaptureManualAuxiliaryPublicationSurvivesUnavailableSource also performs registered-only OpenCode generation while the source remains available. |
| BEH-06 | Native parser tests require actual skill_activation events. Native source->CombineTranscripts->SaveReload retains minimal matching activation/delegation metadata while generic unrelated events/content are stripped. Actual producer/HTTP DTO output was handed privately to frontend for its own decoding/DOM regressions; no invented event fixture or live source dependency. |
| DATA-04 | TestCodexOccurrenceOutcomesRequirePositiveSourceEvidence covers opaque/empty/unknown-shape/task-reference-only responses and explicit success/status/error controls for skill and delegation calls. Unknown outcomes never enter a measured success denominator. |
| BEH-07 | TestCodexDelegationKeepsOpaqueAgentSeparateFromSessionIdentity preserves child_agent_id independently, prefers explicit session_id and leaves session unset for an opaque-agent-only response. |
| ARCH-01 | TestCurrentTaskLinksRequireExplicitOutcome covers missing/empty/unknown/succeeded/failed/unsupported native task outcomes. Valid current fixtures carry explicit outcomes. |

Additional consequence guard: TestSparseOccurrenceEventsPreserveFullSourceAnomalySummariesAcrossReload supplies a real generic-tool failure streak plus owned skill failures interrupted by generic successes. Actual save/reload marks the retained occurrence-only timeline events_are_partial, preserves the full-source cached anomaly set and does not invent a skill retry streak. It also rejects persistence of source content. TestCLIGenerate_HighFailureAggregateKeepsSummaryAndUnavailableEventMetadata verifies ticket aggregation over unavailable historical owners without discarding their accepted event metadata. Fresh unrelated-event compaction is asserted separately from unavailable-history preservation.

## Completed affected checks

- Corrected Codex/OpenCode/domain/adapter/backfill targeted regressions passed, including native multiplicity, required outcomes, exact-parent recovery and minimal event persistence.
- Raw CLI no-private-registration/copied-ledger/malformed-arguments and complete-source capture checks passed4.003s.
- After correcting residual source/store/project guards, affected CLI source-isolation/history/unsafe-source tests passed except the ticket aggregate case; that aggregate correction and its affected tests subsequently passed4.157s.
- Sparse occurrence, namespace and current outcome adapter checks passed0.016s after the final aggregate/partial-stream corrections.
- Capture reconciliation, actual source/ticket association conflicts and local physical registry ownership checks passed0.203s.
- Backend coder reported production go build ./... and git diff --check passed before the final broad run. This is compilation evidence, not an independent correctness verdict.

## Test-oracle and removal rationale

- Current fixture schemas advanced to6, with explicit cost/outcome evidence, preserving unrelated numerical, malformed-data, path and UI assertions. Historical cost presence and older transition semantics remain under explicit migration tests; ordinary old-schema reads assert migration required without rewriting files.
- Ownerless/root/unresolved compatibility assertions were superseded by the user's explicit deletion/drop-compatibility instruction. Explicit migration tests retain exact dropped counts, reliably owned relocation, cross-owner native multiplicity, cost evidence and idempotence.
- Private HEAD/receipt authenticity assertions were removed with the unused authorization API. Actual source/ticket conflicts now have TestSessionAssociationChecksActualConflictsWithoutPrivateHistoryGate and existing CLI store-alias/cross-ticket isolation recipients. Physical operational-state ownership remains in TestWorktreeRepositoryProofEnrichesOnlyUnchangedOwnership. Valid copied/edited supported history is preserved without source re-verification claims.
- Unsafe raw source paths still fail receive/bind/reconcile. Generation may continue using accepted baseline facts rather than rejecting the entire ledger; the test asserts unchanged owner facts and no unsafe source import. Unsupported raw body fields are rejected in the actual working baseline, not by inspecting a different committed blob.
- Complete source authority replaces obsolete chronology/byte-length gates and whole-tree refresh vetoes. Missing/failed/malformed/unflushed source and budget controls remain explicit negative recipients. The unused CaptureSnapshotAccept helper test was removed after its obsolete production helper was removed.
- An earlier development oracle incorrectly treated plain load output as success. DATA-04, discovered by reviewer-data/gpt-6-luna/medium, superseded that oracle: opaque output remains unknown; positive controls now provide explicit source success/status evidence. This correction is not credited as new tester discovery.

## Limits

Private reviewer reports/probes and test logs remain working evidence outside committed deliverables. Temporary DTO fixture exporters were removed; only sanitized fixture output was handed to the frontend owner under /tmp. Prior pre-review passing runs are revision-scoped evidence, not approval for this corrected diff. Final corrected source fingerprint and targeted closure are recorded below. The parent still owns integrated make/race, standalone embedded assets, independent re-review and final telemetry.

## Broad correction run and affected regression closure

Command: go -C backend test ./... -count=1. All internal packages passed: adapters1.870s, backfill4.150s, capture22.678s, Codex0.124s, OpenCode0.362s, HTTP0.665s, live2.513s, domain0.034s, migration0.020s and remaining packages. The filesystem package has no tests.

CLI ran72.883s with one failure: TestCaptureCopiedBaselineCannotCrossTicketEvenWithCorrectedHeader/true. A native owner locally bound to TEST-100 was copied into a baseline whose header was changed to TEST-200; registered generation for TEST-200 with no fresh owners bypassed the actual source/ticket association check. This is a real guard regression, not a request to restore private HEAD/receipt authenticity. The rejection oracle remains. Backend coder corrected the bounded positive local source/ticket association check without restoring private HEAD/receipt authenticity. All other CLI cases passed against that implementation; the affected closure below covers the final guard revision.

## Prior tester handoff after initial corrections

Prior backend/runtime-fixture source SHA256: e36f99c309ea57ad9ceb0223daab45bff3c8c1859cc300f340d3c9a2c11d6463, 286 Go/module/JSON/JSONL files under backend, testdata/workspaces and testdata/parity_workspaces; ignored embedded assets excluded. This is a working-tree fingerprint, not a commit. No application/test edits remain in progress by the tester.

Final affected command passed3.297s: go -C backend test ./cmd/agentic-lens -run 'TestCaptureCopiedBaselineCannotCrossTicketEvenWithCorrectedHeader|TestCaptureGenerationResolvesStoreAliasesWithoutConflatingDistinctStores|TestGenerateExactSourceWithoutPrivateRegistrationReplacesPortableOwner|TestGenerateCopiedLedgerPreservesUnavailableOwnedFactsWithoutPinnedDestination|TestGeneratePreservesUnavailableOwnedAuxiliaryFactsAcrossMachines|TestCaptureManualSelectorsRequireExactSourceProject|TestCaptureDescendantOwnershipComesFromEachNativeSession' -count=1. It covers both the restored actual source/ticket guard and the portability/store/project controls that must remain accepted.

Together with the final broad run's all-internal pass and every other CLI test pass, the affected rerun leaves no known unresolved test failure. This does not claim a second green full CLI or race run after the guard correction. Parent integrated make must provide the complete current-revision gate. git diff --check passed; raw/private logs remain outside Git. Tester is paused and relinquishes backend test ownership for the review freeze.


## Residual DATA-01 outer-payload correction

The subsequent correction review by data-integrity reviewer gpt-6.1-sol/high found the outer typed-decoder variant still open: numeric call_id42 silently removed a previously recorded skill on a successful actual CLI refresh. This is continued DATA-01 discovery credit, not a new tester finding. The prior fingerprint above does not include this final warning-path correction or its test extensions. Backend coder added privacy-safe warnings for typed outer payload failures and missing/blank native call fields; the tester changed tests only.

The permanent raw CLI regression now independently establishes a valid native occurrence before each malformed argument/call_id/name variant. It requires failed generation, incomplete-owner and specific source diagnostics, plus exact equality of all previously accepted ledger bytes. Parser cases require codex-rollout-line:2 diagnostics for numeric outer fields and missing/blank type, call_id and name; the same parser helper first proves the valid native call remains warning-free and produces the expected occurrence.

Passed on the resulting working diff: CLI1.258s; Codex0.012s.

```text
go -C backend test ./cmd/agentic-lens ./internal/adapters/codex -run 'TestGenerateMalformedSkillArgumentsPreservePreviouslyOwnedSnapshot|TestCodexMalformedRecognizedSkillOccurrenceMarksSourceIncomplete|TestCodexMalformedOuterResponsePayloadMarksSourceIncomplete|TestCodexOwnedOccurrencesUseNativeCallsAndActivationEvidence|TestCodexDistinctSkillAttachmentsPreserveNativeMultiplicityAndTimeline|TestCodexOccurrenceOutcomesRequirePositiveSourceEvidence|TestCodexDelegationKeepsOpaqueAgentSeparateFromSessionIdentity' -count=1
```

No broad suite was repeated after this bounded correction. The parent owns the new review freeze, remaining paired correction reviews and final integrated make/race gate. Tester is paused and relinquishes backend test ownership. Private reviewer overlays remain outside Git.
