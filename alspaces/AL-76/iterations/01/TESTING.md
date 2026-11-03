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

No deterministic end-to-end late replacement I/O-failure injection was added: the production adapter has no fault hook, and filesystem permission denial is not a reliable oracle when run with elevated privileges. Preflight failure preserves valid siblings, pending journal refusal, permissions, no-write modes and unchanged-file idempotence are covered. Partial-write reporting and platform-specific directory fsync/rename behavior remain explicit focused-review boundaries. This report does not certify the integrated upstream revision or hosting CI.
