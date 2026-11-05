# AL-64 routine main integration test report

Status: focused verification passed; parent-owned integrated make and independent integration review remain pending. Tester responsibility: adversarial integration assertions; no product implementation, staging, commits, publication or telemetry.

## Candidate and scope

Inputs: ticket head e9ca26767e2097883d6462fd4c5e56b46074b686, upstream main 40b16def, merge base a84adbe5. Checks ran on resolved shared working-tree contents before the parent created merge commit 1378860955deb1acc05924953b8b959b825d3ebc. Final application/test changed-content SHA256: 8499dabaa2ce9be0cf00e4a5d9aa2cfa0a8a7b5c61b671fc52da5a93ccbb44ee. Fingerprint input: sorted changed/untracked backend/ and web/ regular-file paths against ticket head e9ca2676, each followed by NUL, contents and NUL; 69 paths. The fingerprint was rechecked after the merge commit and remained identical.

The authorized scope is the routine main merge, combining AL-64 duration accounting with upstream AL-66 optional output and AL-77 session-overlap warnings. Iteration 02 remains proposed. These tests do not implement or establish hybrid liveness, native status probes or continuous process activity. Synthetic source fixtures establish source-state accounting only. No real ticket ledger, user source store or live workspace is a fixture.

## Coverage selection and test changes

Inspected the existing generation-policy CLI matrix, registered generation checks, both OpenCode schema-family active/completion tables, Codex lifecycle table, explicit fallback and family backfill matrix, persistence duration/collection read-boundary coverage and FleetOverview overlap lifecycle tests before adding coverage. Existing tests already establish absent/empty/nonempty category semantics, completed invocation preservation, unknown legacy coverage and replacement/retention. Duplicating those complete matrices was unnecessary.

- backend/cmd/agentic-lens/generation_policy_test.go: one new coherent matrix covers Codex, OpenCode v1 and OpenCode v2 through explicit and registered CLI generation. Disabling all optional categories must omit skills/delegations/anomalies without removing either active native owner, source identity, tokens or source-proven parent linkage. Both owners must retain provisional evidence, no fabricated completion and numeric duration rounded from the stored cutoff; the cutoff must fall within the generation operation and be identical across its owners. Real CLI subprocesses, JSON output, canonical runtime validation, native-format JSONL/SQLite and capture registry boundaries execute. This distinguishes the newly combined policy/cutoff wiring from the existing separate duration and policy tests.
- web/src/views/FleetOverview.test.tsx: extended the existing overlap fixture and meaningful lifecycle cases. Provisional duration disclosure remains visible alongside overlap navigation/filtering, corrected reload, failed inspection/recovery and rejected obsolete success/error responses. Existing accessibility, callback, full-dataset scope and cost assertions remain intact.

Only those two test paths were edited by this tester; no test assertions were removed or weakened. The registered Codex fixture initially omitted its required authoritative transcript path and queued child evidence, causing its new row to report one owner. Supplying those established capture prerequisites made all six rows pass. This was a test setup correction, not a confirmed product defect.

The three conflicted test files (backfill/engine_test.go, opencode/schema_v1_session_extraction_test.go and adapters/schemas_test.go) retain every Test function name from both input revisions, with no duplicates. Inspection and focused execution preserve all six prior review corrections: invalid Codex completion, v1 future progress timestamp, future newer message, fallback cutoff, tool-free streaming in both families, and invalid Codex start. No silent skips were introduced.

## Executed verification

From backend/:

- go test -race ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill -run 'TestRolloutActiveDurationCutoffAndTerminalBoundaries|TestSchemaV[12]ActiveDurationSnapshotAndCompletion|TestBackfillEngine_ExplicitFallbackResolvesWhenAllowed' -count=1: exit 0; Codex 1.029s, OpenCode 1.290s, backfill 1.150s. Includes all six retained review-correction oracles and both streaming shapes.
- go test -race ./cmd/agentic-lens -run 'TestGenerateOmissionPreservesActiveDurationAndDescendants|TestGenerateOptionalMetricsOmissionRetainedOwnersAndSummary|TestRegisteredGenerateHonorsDisabledAnomalies|TestOpenCodeGenerationPolicyAcrossSchemaFamilies' -count=1 -v: the three unchanged policy tests passed (all eight policy masks, registered omission/re-enable-empty, both OpenCode schema families). The command exited 1 because the initial registered Codex fixture lacked the capture prerequisites described above; this command is not represented as an overall pass.
- After that fixture-only correction: go test -race ./cmd/agentic-lens -run '^TestGenerateOmissionPreservesActiveDurationAndDescendants$' -count=1 -v: exit 0, all six subcases passed, package 2.737s. Valid prior results for unchanged policy tests are reused.
- Initial OpenCode-only development run of the new integration test also passed all four schema/selection combinations before Codex was added.

From web/:

- npm test -- --run src/views/FleetOverview.test.tsx: exit 0; 21 tests passed, 4.26s total.
- npm run typecheck: exit 0.

Repository checks: git diff --check passed; function-name preservation inventory against e9ca2676 and 40b16def found no missing or duplicate tests in the three conflicted test files. Environment: Linux amd64, Go 1.26.0, Node 24.22.1, npm 11.19.1. No test was silently skipped.

## Reused evidence and remaining limits

The implementation agent's INTEGRATION_IMPLEMENTATION.md records passing focused adapter/domain/source/capture/backfill/HTTP/migration checks, canonical frontend typecheck/build and seven affected suites with 387 tests. Those unchanged scopes were not repeated wholesale here. Generated embedded assets were already prepared before these direct Go commands. The earlier full iteration-01 suite and mutation evidence remain historical; they do not certify the merged candidate.

Retain the implementation agent's initial CLI timing failure as an unresolved reliability observation: TestCaptureLaterAgentChannelReportsBoundPublicationFailure reached its short worker-acquisition deadline despite acknowledged generation and successful publication logs; its isolated unchanged rerun passed. No deadline or assertion was changed here. The focused results above do not establish a clean full CLI suite or replace the parent's full integrated make, standalone embedded-asset verification and independent focused review.

No new product defect was confirmed. No new mutation experiment was performed for this merge-only coverage; prior iteration correction overlays remain the discrimination evidence for the six retained defects. Final integration verdict, merge commit and any later delivery remain parent-owned.
