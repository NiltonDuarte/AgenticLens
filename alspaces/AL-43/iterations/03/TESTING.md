# AL-43 Iteration 03 independent F1 verification

## Scope and tested diff

User-approved correction: retain main's execution classifier, keep content classification removed, and add the three supplied false-failure regression cases. This acceptance supersedes iteration 02's pre-feature execution restoration. Initial application head is e91cf0d05348537657c210601c6acc93c01f967d on chore/AL-43-revert-content-classification. Reviewed main reference is 3857fa22f49f825c82ced39c3f81f963b7447545.

Tested uncommitted application diff SHA-256: 482da10f38f0ddefada6a1871dfb46560c1226f4c50582f0fbeddea3fb542a7d, computed from git diff over backend/internal/domain/live.go, tool_execution_test.go, live_test.go, backend/internal/adapters/opencode/schema_v1_session_extraction_test.go, and backend/internal/adapters/live/parity_test.go. Parent/planner artifact edits were preserved. No application implementation, commit, telemetry, PR, merge, or delegated work was performed by the tester.

## Independent audit and test change

Compared the entire ClassifyToolExecution function text directly with git show from the reviewed main reference. Exact equality passed. The production diff changes only that function's output-failure rules and source-evidence comment, retaining all content-feature removal elsewhere. Existing explicit failure, nonzero exit, process crash, exact timeout/permission, no-match, escalation, and bounded-message cases remain present. The three approved successful read/grep/bash cases exercise the actual classifier and assert SeverityNone and absent ErrorMessage. The obsolete bare-panic crash expectation and inspection false-failure expectation are corrected to the superseding policy; they are not lost required failure coverage.

The existing temporary SQLite live/offline parity test covers all three source cases, exact original output, completed status and absent error evidence in both extraction paths, eight calls, and three genuine failures. The live trace verifies the final genuine failure remains medium after successful calls and no consecutive-error anomaly appears. Tester added only a small offline assertion to that same test: feed the extracted canonical sessions into production DetectAnomalies and reject a tool-retry-streak anomaly. This closes the offline anomaly projection boundary instead of inferring it solely from successful status rows. No new fixture or standalone probe was added.

Offline EventFact has no failure-severity field, so the live final-error severity assertion is a direct measurement of live streak reset; the offline assertion directly measures retry-anomaly output. The extractor's private per-tool accumulator reset is source-traced, not exposed as a separate measured counter. No HTTP measurement of these three source cases is claimed.

## Commands and results

Canonical embedded assets had been prepared by the coder with make web-assets. Independent verification on the tested diff:

- A Python source-body comparison of ClassifyToolExecution against commit 3857fa22f49f825c82ced39c3f81f963b7447545 passed exact equality.
- From backend: `go test -race ./internal/domain ./internal/adapters/opencode ./internal/adapters/live -run 'TestClassifyToolExecution|TestClassifyToolFailure_MessageTruncation|TestSchemaV1ExtractFacts_ContentInspectionTools_CompletedStatusAndMessageID|TestOfflineAndLiveToolExecutionParity|TestHydrateWorkspaceEvents_(ConsecutiveErrorsSpansTurns|MultiSessionIsolationAndStreakReset)' -count=1` passed all three packages. It includes the added offline anomaly assertion and existing cross-turn/session guards.
- From backend: `go test -race ./internal/adapters ./internal/adapters/codex ./internal/adapters/backfill ./internal/adapters/http ./internal/adapters/live -run 'TestCanonicalRetiredContentFieldsReadCompatibility|TestCodex(SuccessfulCommandOutputPreservesExecution|OutputEnvelopesPreserveRawEvidence|MCPOutputEnvelopeRespectsExecutionStatus)|TestBackfillEngine_(StrictSessionTrustRejectsUntrustedTitleFallback|GlobalRecoveryRequiresEveryKnownTree)|TestLiveHandlers_NilSlicesSerializeAsEmptyArrays|TestLiveStreamAndTraceRecoverInPlaceSQLiteUpdate|TestLiveTraceLegacyRevisionRecoveryAndBoundedOnDemandCache' -count=1` passed all five packages. These retain strict retired-field compatibility, Codex raw evidence/explicit outcomes, strict recovery, nil-array normalization and live update/cache coverage.
- `diff -qr web/dist backend/internal/adapters/web/dist` passed. Frontend sources are unchanged relative to the initial head, so existing frontend/AL-53 interaction verification remains applicable.
- `git diff --check` passed.

## Handoff and limitations

No blocking finding remains in the tested correction. Full make is intentionally deferred to the parent's required gate after implementation commit and main integration; this focused pass does not replace it. No frontend test rerun, browser, standalone serving, release, hosted-CI or merge claim is made. Existing iteration 02 evidence remains historical for its original classifier, not proof of F1 resolution. Raw execution logs and source sessions remain outside Git.

## Parent integrated full gate

Implementation commit 3828f537cd6c448b8b811688e2b0cd693b0a6cd5 was fetched/merged against actual origin/main 3857fa22f49f825c82ced39c3f81f963b7447545 (already up to date). Full make passed exit 0 with all dependency/check/test/build stages executed: Go vet/race suites, TypeScript typecheck/lint, 741 frontend tests (35 files), canonical assets and four platform builds. Exact main classifier source comparison passed again. Source/embedded asset trees compare identical and stay ignored; host version and command-help execution passed. Existing unchanged frontend/standalone-serving evidence is reused, without a new browser/geometry or F1 HTTP claim. This closes the previously pending parent gate. Subsequent commits change ticket artifacts/telemetry only. Integrated application diff fingerprint against main over backend and web/src: 5d45600be782aee45c3b243c8e49f35456f88f1b4134cdcf82addb00c403c13c.
