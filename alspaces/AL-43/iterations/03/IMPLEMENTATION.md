# AL-43 Iteration 03 implementation: F1 correction

User-approved scope: retain main's execution classifier while removing content classification, and add the three manager regression cases. Starting application commit: e91cf0d05348537657c210601c6acc93c01f967d. Reviewed main reference: 3857fa22f49f825c82ced39c3f81f963b7447545. Parent/planner artifact changes were preserved. No commits, telemetry or PR actions were performed by the coder.

## Changes

- Replaced only ClassifyToolExecution in backend/internal/domain/live.go with the exact function from reviewed main. A direct source-body equality check passed after gofmt. No content fields or classifier were restored, and no new execution policy was introduced.
- Extended the existing tool_execution_test.go table with successful read output `const timeoutMs = 5000;`, grep output `logs/example.txt: panic: saved example`, and bash output `exit code 0` followed by a newline and `fixture says permission denied`. These expect no execution failure, SeverityNone and nil ErrorMessage. The existing table now checks that diagnostic-message presence matches the execution-failure result for every case. Explicit failure, nonzero exit, crash, timeout, no-match, severity escalation and bounded-message guards remain.
- Adapted the existing domain live crash row from a bare panic substring to an actual segmentation-fault signature recognized by the approved main classifier. Adapted the OpenCode v1 inspection case to keep saved panic text completed with zero errors. These expectations supersede iteration 02's false-failure policy; the production source parsing and message-ID assertions remain intact.
- Strengthened the existing real SQLite TestOfflineAndLiveToolExecutionParity rather than adding a separate probe. Its eight calls include two genuine consecutive failures, the three successful manager cases, and a final genuine failure. Checks establish three true failures in both offline and live paths, eight calls, exact completed/error status parity, unmodified output evidence and absent failure metadata for successful outputs. The final failure remains medium severity and no consecutive-error anomaly appears, demonstrating streak reset and non-escalation after the successful calls.

Only five assigned backend application/test paths changed. The prior feature-removal, strict retired-field decoding, raw Codex evidence, AL-55 recovery and optional-array normalization, AL-47 and AL-53 implementations were left untouched.

## Verification

Tested uncommitted application diff against e91cf0d: SHA256 `2cb545ee2cc6f6af68c2a25102d232e9a77acca40350655f3093c74c482be478`, computed from git diff over the five assigned backend paths.

- `make web-assets` passed, including TypeScript compilation, frontend production build and copy to the Go embedded asset tree.
- Direct comparison confirmed ClassifyToolExecution's entire function body exactly matches commit 3857fa22f49f825c82ced39c3f81f963b7447545.
- From backend, `go test -race ./internal/domain ./internal/adapters/opencode ./internal/adapters/live -run 'TestClassifyToolExecution|TestClassifyToolFailure_MessageTruncation|TestSchemaV1ExtractFacts_ContentInspectionTools_CompletedStatusAndMessageID|TestOfflineAndLiveToolExecutionParity|TestHydrateWorkspaceEvents_(ConsecutiveErrorsSpansTurns|MultiSessionIsolationAndStreakReset)' -count=1` passed all three packages. This includes the existing cross-turn/session streak guards alongside the strengthened parity case.
- `git diff --check` passed.

Independent tester/reviewer verification and the integrated full make gate remain parent-owned. No browser geometry, release or hosted-CI claims are made by these focused checks.
