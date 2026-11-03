# AL-76 iteration 01 testing

Test Engineer: gpt-6.1-sol / high (inherited execution settings). Reviewed the implementation diff against baseline `122e983f8c365a2aa248b1b6d492252555e922ae`, contract, plan and the persistence/configuration/compatibility/testing checklist. Product code was unchanged during this testing pass; only tests and this report were edited.

## Changes and regression purpose

- Extended formatter assertions to cover nested compact cost evidence, legacy tools and rejection of unsupported invocation evidence/token and anomaly shapes. Existing lexeme oracle uses `json.Compact` on original and output, preserving raw scalar spelling and member order independently of the implementation's syntax tree. Unknown slash-containing member names exercise the structural-path collision correction.
- Added real filesystem/publication contention coverage. A context wrapper signals only after the kernel lock acquisition observes contention; the current publisher then writes newer facts and releases ownership. The final ledger must retain those newer facts, detecting a stale pre-lock read without arbitrary sleeps.
- Added preservation checks for retired capture reports, SUMMARY.md and an unselected ledger; selected hard-link refusal must leave both originals intact.
- Added CLI coverage for a non-default configured workspace root, operation with absent capture state, and distinguishing malformed-input exit 2 from formatting-difference exit 1.
- Extended existing fact writer, backfill refresh and migration batch cases with compact token line assertions. These would fail with the previous expanded MarshalIndent output. No tests were removed or silently skipped.

## Executed verification

Passed after final edits, from `backend/`:

`go test -race ./internal/adapters ./internal/adapters/backfill ./cmd/agentic-lens -run 'TestFormatLedger|TestExecuteFormat|TestSaveFactLedger|TestSaveMetricsLedger|TestBackfillEngine_ExplicitRecomputeDBDoesNotSkipCompletedLedger|TestMigrationRunner_Batch' -count=1`

All three packages passed. `git diff --check` passed. Tests use disposable ledgers/configuration; no live ticket workspace, source store or network input is used. The parent owns the integrated full `make` gate.

## Limits

The first pass omitted deterministic late-I/O failure injection. Independent test-effectiveness review identified a real-filesystem synchronization fixture that reaches the branch without product hooks; the correction below closes that gap. Platform-specific directory fsync/rename behavior remains an explicit review/runtime limit. This report does not certify the integrated upstream revision or hosting CI.

## Test-review corrections on product revision c32c893

Test-effectiveness reviewer gpt-6.1-sol / high reported TS-1 (late replacement failure coverage, severity 4) and TS-2 (computed-export compact anomaly coverage, severity 3). The tester read the supplied report and implemented test-only corrections; this is verification/correction work, not independent rediscovery.

- TS-1: A/B/C fixtures are preflighted in canonical order while C ownership is held. The existing contention-signaling context establishes A/B finished preflight; the test moves B's directory to a preserved path and places a regular file at its old directory location, then releases C. It requires A/C successful Written flags, B failure with Written false, a nonnil overall error naming B, stable target order and unchanged preserved B bytes. No sleeps, production hooks or permission-based assumptions are used.
- TS-2: The existing computed-writer persistence test now serializes a nonempty anomaly with nested tool/count context, asserts its exact one-line representation inside an expanded array, checks semantic values and repeats persistence for byte idempotence. Reverting only SaveMetricsLedger to the expanded printer would fail this assertion.

Parent-reported full CLI-suite limits from the coder: TestCaptureAutomaticSharedStoreAndTicketStayInExactProjects shared-store shutdown timeout; TestCaptureInstallerUsesSharedLocalAndExplicitStatePolicy shared/local/cli and TestCaptureCLIOpenCodePluginOwnershipAndMissingHandler pending-installer native-version unavailable/changed failures; TestCaptureNativeWorkerPublishesMultipleRootsAndRefreshesWithoutGenerate idle shutdown timeout. These are recorded as reported failures, not reproduced or waived by this focused pass. The parent owns integrated make and upstream coordination.

Correction verification passed from `backend/`: `go test -race ./internal/adapters -run 'TestFormatLedger|TestSaveMetricsLedger' -count=1`. `git diff --check` passed. Only `ledger_format_test.go`, `repository_test.go` and this report changed in the correction pass; no product code, full make run or publication.
