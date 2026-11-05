# AL-64 iteration 01 test report

Tester: gpt-6.1-sol/high, inherited runtime settings. Baseline a84adbe5be46c2ef52ede0987a417562715d38e1; branch fix/AL-64-active-turn-duration. This report records testing, not consolidated review approval.

## Tested boundary

Application/test changed-content SHA256 at full-suite launch: 8bcdc892323531d6da367c76c9189c1e266e790e500c4c39f781b331e8b6619c. Fingerprint input: sorted changed/untracked backend/ and web/ file paths, each followed by NUL, its contents, and NUL. Coder's narrow native v1 empty-bootstrap family-detection fix was present before the full-suite launch. Canonical frontend assets were built by the coder before Go checks. Environment: Go 1.26.0, Node 24.22.1, npm 11.19.1, Linux amd64.

The current candidate counts explicit open source intervals provisionally. The user clarification about persisted stale activity versus additional live-process proof remains pending; this test report does not decide that product policy or claim native process liveness.

## Distinct regression coverage added or adapted

- Extended backfill/engine_test.go with one family-parameterized actual-source matrix for Codex, OpenCode v1 and OpenCode v2. It uses self-contained files/SQLite fixtures and the production engine, never native runtime databases or live ticket workspaces. At fixed supplied cutoffs it verifies active 10 seconds, advancing to 20, unchanged repeat at 20, completion replacing the provisional contribution with 30, and unchanged completed repeat. Each stage saves and reloads the ledger; it checks exact root identity, unchanged completed 3-second child/parent linkage, retained 7-second unavailable provisional owner and immutable cutoff, other-harness 17-second facts, additive totals, and no duplicated sessions. A final Force-only rewrite must not masquerade as source recomputation.
- Adapted the existing CLI fresh-versus-refresh test. It preserves completed-stage and source-growth identity/tokens/turns/duration/completion assertions. Active Codex duration is checked against its own recorded cutoff; separate real-clock CLI runs compare every stable fact while excluding only advancing provisional values. OpenCode missing completion without lifecycle proof remains partial, with no fabricated completion timestamp. Fixed-cutoff idempotence is independently asserted by the engine matrix.
- Extended schemas_test.go with duration evidence persistence and current-schema migration decode/copy preservation, accepted legacy absence with unchanged numeric total and unknown coverage, and invalid component/cutoff/state/basis/unknown-member rejection through the real runtime validator.
- Extended schema_family_detection_test.go for native v1's complete tables plus correctly shaped empty session_message bootstrap. Accepted extraction preserves completed duration. Populated auxiliary, malformed auxiliary and empty partial owner-table variants remain rejected. Existing table/column-omission rejection coverage remains intact.
- Added coherent duration.test.ts coverage for ledger API decoding, component/state/cutoff validation, unknown legacy coverage, distinct immutable cutoffs and display labels. Extended TicketDeepDive's existing component suite with a provisional/unknown baseline display when optional source hydration fails.

No production files were edited by this tester. Tests remain pure in the domain and use concrete types. No historical ticket artifacts are fixtures. The existing family adapter tests and native runtime report provide complementary interval/shape evidence rather than being duplicated here.

## Commands and outcomes

Passed focused checks:

- go test ./internal/adapters/backfill -run TestBackfillActiveDurationReplacesSnapshotsAcrossSourceFamilies -count=1
- go test ./internal/adapters ./cmd/agentic-lens -run 'TestDurationEvidenceRoundTripLegacyAndValidation|TestCLIGenerate_SessionRefreshMatchesFreshExtraction' -count=1
- go test ./internal/adapters/opencode -run 'TestSchemaFamilyV1AuxiliaryBootstrapTable|TestSchemaFamilyRejects' -count=1
- npm test -- --run src/domain/duration.test.ts: 9 tests passed.
- git diff --check

Full checks passed on the unchanged application/test fingerprint above:

- go test -race ./...: exit 0; all 13 test-bearing packages passed, including CLI (123.937s), capture (82.581s), backfill (27.087s), adapters, HTTP, Codex, OpenCode, live, embedded web, domain and migration. filesystem has no test files and is not represented as an executed test suite.
- npm test: 37 files, 834 tests passed, exit 0; duration 43.55s.
- npm run typecheck: exit 0.
- npm run lint: exit 0.
- git diff --check: exit 0 after final tests.

No new tests were skipped. No product defects were confirmed in this independent pass; the initial v1 matrix failure was a self-contained fixture error (the preexisting helper generated November row timestamps beyond the fixed January cutoff). Correcting those fixture timestamps made it exercise the intended active boundary. The initial missing path in a local edit command was corrected without changing production files.

Private command logs are outside Git at /tmp/al64-tester-{go-race,frontend,typecheck,lint}.log. The final post-suite fingerprint exactly matches the launch fingerprint. Source-state policy clarification remains unresolved and therefore this is verification evidence, not final delivery approval.

## Limits and ownership

Native runtime observations and generated-output experiments belong to RUNTIME_EVIDENCE.md. These deterministic tests do not establish process liveness or claim an unavailable runtime check passed. Parent retains the required post-integration make/build/embedded standalone verification and independent focused review gate. No telemetry was extracted, no commits were made, and no PR was published by this tester.

## Review-correction verification: DUR-01 through DUR-05

The earlier no-defects result describes the initial independent test pass, not the later focused review. Focused reviewers subsequently discovered the five findings retained in REVIEW_SUMMARY.json; this tester verified supplied findings and added their permanent oracles, without taking discovery credit.

Only existing coherent tests were extended:

- DUR-01: the Codex lifecycle/cutoff table now includes a reversed completion and an unparsable completion timestamp after a valid measured 4-second turn. Both must preserve those 4 seconds, report partial coverage with a limitation, and contribute unknown coverage to the summary.
- DUR-02: the v1 snapshot/completion test now includes a native running-tool start preceding cutoff while the generic part progress/update timestamp follows cutoff. It must still count the interval through the supplied cutoff.
- DUR-03: that same table includes a newer user message created after all tested cutoffs. It must not suppress the earlier interval at those cutoffs.
- DUR-04: the existing explicitly enabled fallback test retains its missing-exact-owner and discovered-owner assertions while now using an unfinished tool fixture and a fixed operation cutoff. The persisted fallback contribution must be exactly 3 seconds with that cutoff, not time.Now().
- DUR-05: the existing family-specific active/completion tests now cover v1 native timed text and timed reasoning without tools, and v2 unfinished text with time_idle NULL and no tool. They assert advancing and fixed-cutoff snapshots, absent fabricated completion, completion replacement and unchanged completed repeats.

Before/after evidence was obtained without changing shared production files. Pre-correction copies of the three affected production files were saved under /tmp/al64-tester-prefix and supplied to Go through an overlay. The retained new tests compiled and failed for both DUR-01 cases, DUR-02, DUR-03 and DUR-04. The same tests passed against the corrected production files. The reviewer's isolated streaming-removal overlay was also reused: the retained new v1 text/reasoning and v2 idle-null streaming cases each fail with zero duration when the corresponding production proof path is removed, and pass on the control. Private logs: /tmp/al64-tester-prefix/results.log and streaming-mutant.log. These expected failing overlay commands are regression-discrimination evidence, not failures of the corrected candidate.

Commands from backend:

- go test -overlay /tmp/al64-tester-prefix/overlay.json ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill -run 'TestRolloutActiveDurationCutoffAndTerminalBoundaries|TestSchemaV[12]ActiveDurationSnapshotAndCompletion|TestBackfillEngine_ExplicitFallbackResolvesWhenAllowed' -count=1: expected exit 1 on all four corrected root causes.
- The same command without -overlay: exit 0 on all three packages.
- go test -overlay /tmp/al64-tests-sol/streaming-overlay.json ./internal/adapters/opencode -run 'TestSchemaV[12]ActiveDurationSnapshotAndCompletion/(streaming_text_without_tool|streaming_reasoning_without_tool|streaming_without_tool=true)$' -count=1: expected exit 1 for each removed no-tool streaming path.

Corrected app/test content fingerprint at affected-suite launch, using the same algorithm above: e5c640fe2bfd002e93dc0d809335e80eec94587ad3a3c3d1e4bc0ed18709306b. Full affected package race checks passed: go test -race ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill -count=1, exit 0. Codex 1.128s, OpenCode 6.350s, backfill 15.001s. Final diff whitespace check passed. Post-suite application/test content fingerprint: e5c640fe2bfd002e93dc0d809335e80eec94587ad3a3c3d1e4bc0ed18709306b. Earlier full Go race and frontend checks are reused for unchanged scopes. Parent retains post-integration make and revision-aware review. Pending process-liveness policy is unchanged by these corrections.

## DUR-06 invalid-start correction verification

Architecture reviewer gpt-6.1-sol/high discovered this distinct invalid-start path during candidate 02 correction review (retained as DUR-06 in REVIEW_SUMMARY.json). This tester verified the supplied issue; discovery credit remains with that reviewer.

Extended the same Codex lifecycle/cutoff table with malformed-start and missing-start timestamps, each followed by a valid matching completion after an earlier measured four-second turn. The assertions require the known four-second subtotal, partial coverage with explicit limitation, unknown summary coverage, no invented provisional cutoff/time, and completed_at retained at the earlier proven completion (00:00:05). Existing valid/invalid completion and interruption cases remain in the same table.

A disposable Go overlay under /tmp/al64-tester-dur06 removes only the DUR-06 start-error limitation branch, restoring its prior behavior without touching shared production. Both new cases compile and fail under this pre-fix overlay because their supported subtotal is falsely certified complete. The corrected control passes the entire Codex race suite.

- go test -overlay /tmp/al64-tester-dur06/overlay.json ./internal/adapters/codex -run 'TestRolloutActiveDurationCutoffAndTerminalBoundaries/(malformed_start|missing_start)$' -count=1: expected exit 1, both named cases fail for the coverage assertion.
- go test -race ./internal/adapters/codex -count=1: exit 0, 1.148s.
- git diff --check: exit 0.

Private logs: /tmp/al64-tester-dur06/{pre-fix,race}.log. Final app/test content fingerprint using the established algorithm: 894207b62c28077b7fdf23b72c42aec720a3b1e2b2bb66ae0724295030810cdb. No other test or product paths were changed by this tester in this correction. Prior verification remains reusable for unchanged scopes; integrated make, scoped review and policy clarification remain parent-owned.
