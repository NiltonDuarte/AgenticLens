# AL-79 test effectiveness correction re-review — C5

## Identity and scope

- Requested run: `tests-luna-c5`; requested identity `gpt-6-luna/medium`. Runtime effort metadata is unavailable; this records the requested tuple only.
- Candidate: `AL79-C5`; approved baseline `4df45967c801824500ac2ea39a8f52fa9cbdcf66`; candidate source diff SHA-256 `e6e82cade1b229edd2abd478366c9bf024cc94b34d71f423f4a832d25be78c70`. Test correction diff SHA-256 `55645dc842edcd2edcd4db5db34abf2233106b261c3a2ef9de448c10b20ed421`.
- Inspected paths: `backend/cmd/agentic-lens/capture_history_test.go`, `backend/internal/adapters/repository_test.go`, and `backend/internal/adapters/backfill/migration_runner_test.go`. Production is unchanged in this correction. The checked-out committed HEAD is `9524a28762e85d69827bb59694ecfdebdf0663a1`, with C5 implementation commit `fb0d3dccc5e2da3f1004a973cf5a31acf3995e3` as its parent; the intervening integration is documentation-only. The candidate manifest names `fb0d3d...` as integrated head; this report bounds its review to the stated C5 source/test diff and tested source inventory `b3621fcec3c18284f272e8d46b91e87368ae01dd59ee894ae8e832a1cde89eae`.
- I reused unchanged C4 approvals for the prior test paths and reviewed the approved contract, plan and C4 report. No source/test files were edited; no broad suite was repeated.

## Findings

No blocking test-effectiveness gap found in the requested C5 correction scope.

## Correction oracles inspected

| Contract behavior / stale assertion | Oracle and assessment |
|---|---|
| Preserve a supported external `ParentID` on its owning session without importing the external parent's facts. | `TestGenerationPreservesValidCopiedHistoryAndRejectsMalformedOwnership` now has a valid `external parent` case. It generates from an unavailable copied history, then compares the complete expected session owner to the result and checks the other checkout's registration/cursor/ack state is unchanged. Blank, whitespace-only and padded identifiers remain rejection cases; they require the `parent_id` diagnostic, preserve destination ledger bytes, and preserve foreign state. Existing token/model/user/source and destination/history ownership cases remain in the same test. This distinguishes external ancestry from malformed identifiers at the CLI generation boundary. The independent migration current-schema test also now accepts a portable external session parent while retaining other DTO rejection cases. |
| Read shared-only portable history without a capture machine state prerequisite. | `TestScanWorkspacesRejectsInvalidStateConfiguration/shared-only_history_needs_no_machine_state` supplies a shared config and a self-contained remote-owner ledger, with no local config. It requires the owner to be returned and both input files to remain byte-identical; it also checks no local config was created. The malformed selected-local control uses `state_dir: []` and requires a fatal diagnostic with no ledger result and unchanged local config. |
| Migrate under shared-only config while still rejecting invalid selected local state before mutation. | `TestMigrationRunnerRejectsExistingInvalidStateConfigWithoutMutation` includes shared-only and malformed-local branches. Success requires one scanned and migrated ledger, no errors, a changed output decoded at `CurrentSchemaVersion`, and no local config creation. The invalid-local branch requires an error naming `telemetry.capture.state_dir`, nil report, byte-identical ledger/config and unchanged shared config. `enabled: false` is present in the malformed local fixture, so the validation guard is exercised even when capture is disabled. |

## Verification and carry-forward

The tester's targeted `go test -race` run passed all named cases without skips: CLI 7.819s, adapters 1.029s, backfill 1.043s (`/tmp/al79-tester-integrated/targeted.log`). The evidence records isolated fixtures, and the source content inventory hash above bounds the tested source. The correction diff passes `git diff --check`. These results establish targeted oracle execution, not mutation equivalence. Parent-owned full `make` remains in progress and is not claimed here.

Unchanged C4 assertions and their carry-forward rationale remain applicable. This correction adds no production behavior, schema or authenticity policy. Native Windows runtime remains an unexecuted CI-stage limit; this test-only review supplies no new native Codex/trust evidence.

## Outcome

**APPROVED — scoped C5 test-effectiveness correction re-review.** The corrected tests exercise the specified external-parent ownership, malformed-parent rejection/nonmutation, shared-only scan and migration support, and invalid selected-local rejection before mutation. This does not replace parent consolidation, the integrated `make` gate, or Windows runtime verification.
