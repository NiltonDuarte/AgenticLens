# GEN-20 iteration 02 test evidence

Status: Test-source investigation complete; coder race results reused. The parent owns the integrated make gate and final review.

## Tested boundary and revision

Approved iteration 02 removes the registry upgrade, old-binary fence and numeric-path preservation requirements from the historical contract. Current version-2 schema contains coordinate mappings directly. Missing current mappings must fail explicitly without modifying retained state.

Baseline: dc44619ae15586fe1194f872743033f523635c92. Candidate is the uncommitted implementation plus new acceptance tests. SHA256 over every changed or untracked backend file, sorted by relative path and encoded as path, NUL, file bytes, NUL: faf8a6646625195d59ffc902b061037bc48a915c8fc28659ea5ed8f14ad28a4e (17 files). Workspace documents are excluded from this source fingerprint.

## Test effectiveness

- TestCoordinationTokensConcurrentFirstOpenAndStableReopen synchronizes eight actual OpenProject calls across same and distinct linked worktrees. Two repositories have matching administrative basenames. It asserts stable same-scope paths, distinct cross-scope paths, Git-derived naming and persisted read-only reopen.
- TestCoordinationUUIDFallbackPersistsAcrossGitEnrichment asserts an absent scope preview has no allocated path or SQL ID and writes no state; real allocation is a 16-hex token, retained after Git initialization. Existing ownership-negative cases remain intact.
- TestCoordinateAllocationSkipsReservedNamesAndBoundsCollisions drives the actual allocator with deterministic entropy, verifies retry around an unknown file, persisted-name collision exhaustion and unchanged unknown file contents.
- TestUnsupportedCoordinateLayoutFailsWithoutChangingState covers a missing mapping table and missing registered mapping. It requires explicit recovery context from writable/read-only opens and rejects preview, then compares database bytes and a retained scope sentinel. It does not establish arbitrary filesystem snapshots or SQLite sidecar byte equality.
- New TestCoordinationTokenAliasesRejectUnknownAndRedirectedState exercises a committed token alias, rejects unknown token preview/open without creating that path, and rejects a persisted token redirected by a symlink while leaving the external destination empty.
- New TestCaptureInstallReportsPendingAllocationUntilCommittedManifest exercises two fresh CLI dry-runs that disclose pending allocation and create no registry. Real install must expose a committed location and active manifest using that exact path; a subsequent dry-run has no pending allocation and retains the committed path. Existing callback tests deliberately keep their registry-path contract.
- Existing scoped registration/publication/diagnostic/worker-isolation, namespace containment and supported protocol-1 relocation/resume tests remain the independent recipients for those guarantees. Relocation fixtures now resolve real destination paths; no pre-existing meaningful oracle was removed.

The removed iteration-01 upgrade test asserted exclusively superseded behavior: preserving version-2 numeric coordinates, upgrading to a new protocol, and rolling back that upgrade. The tester's unexecuted upgrade-rollback fixture was removed after user steering. These removals do not discard accepted current-schema or supported protocol-1 relocation requirements.

## Commands and results

Executed by tester from backend: go test ./internal/adapters/capture ./cmd/agentic-lens -list 'TestCoordination|TestCoordinateAllocation|TestUnsupportedCoordinateLayout|TestCaptureInstallReportsPendingAllocationUntilCommittedManifest'. Passed; listed five adapter tests and the CLI installation acceptance test named above. This is discovery evidence, not execution of their bodies. Every new declaration is matched by the coder's focused race command, so there is no omitted new case requiring a duplicate run.

Reused coder verification after iteration-02 simplification, executed from backend and reported passed by the coder via the parent; not independently re-executed by the tester:

- go test -race ./internal/adapters/capture -run 'TestCoordination|TestCoordinateAllocation|TestSharedNamedRegistry|TestNamespace' -count=1
- go test -race ./internal/adapters/capture ./cmd/agentic-lens -run 'TestCoordination|TestCoordinateAllocation|TestUnsupportedCoordinateLayout|TestCaptureInstallReportsPendingAllocationUntilCommittedManifest|TestCaptureInspectionWithoutStateDoesNotInitialize|TestRelocation' -count=1

Tester executed git diff --check successfully. A targeted search found no remaining upgradeRegistryV2 helper, removed V2 upgrade tests, numeric backfill statement or version-3 transition in the affected current-schema files.

## Limits and handoff

No production edits by tester. Both new acceptance files are frozen, and no further test edits are pending. No unrelated domain/frontend tests were added. The tester did not rerun unchanged expensive suites or claim the required full make gate; the parent will execute it after integration. Independent review and publication remain pending. Tests use disposable repositories and state only, with no silent skips introduced.

## Final coordinator verification

Corrected revision: 13a6ff38c97e9c9d3019d76a0944ce46afe1814e. Coordinator: gpt-6.1-sol/high. New TestCoordinationStateBaseUnderWorktreesRemainsAStateBase failed on the reviewed source for both direct roots and physical aliases before the fix. After correction, from backend, go test -race ./internal/adapters/capture ./cmd/agentic-lens -run 'TestCoordination|TestCoordinateAllocation|TestNamespace|TestSharedNamedRegistry|TestCaptureInstallReportsPendingAllocationUntilCommittedManifest|TestRelocation' -count=1 passed. The test-reviewer's isolated randomized-Git-token overlay now fails the strengthened existing concurrency test as intended. Final make passed after integration: backend races, 790 frontend tests, static and Python checks, embedded assets and configured platform builds. Raw build/probe logs remain outside Git.

## Directory-name terminology correction assessment

Reviewed the six-file capture source/test diff against e2df9649de644f7dde1406757a85e9441a9a14d6 under the explicitly authorized review terminology correction in PLAN.md. The persisted column is now directory_name; schema creation, lookup, collision query and insert consistently use that column. A backend-wide search found no old coordinateTokenTx, validCoordinateToken, preferredCoordinateToken or randomCoordinateToken callers. Existing request challenge entropy and the exact request-token binding error in store.go remain unchanged; no authentication, session or AI token behavior was renamed.

The diff changes local names, helper/test names, SQL identifiers and explanatory text only. Validation characters and length, identity hashing inputs, entropy size, retry count, reservation/insert sequence, containment checks and assertions retain their previous behavior. Existing alias and concurrent-allocation test declarations were renamed in place; no unnecessary new test was introduced. The unchanged TestCoordination selection prefix still covers the renamed tests, including the previously added state-base regression. The unreleased schema does not require compatibility migration under the accepted user scope.

Tester ran git diff --check successfully and inspected all worktree_coordinates references in backend. No production or test edits were made during this assessment. Reused the coder-reported passing command from backend, not independently re-executed: go test -race ./internal/adapters/capture ./cmd/agentic-lens -run 'TestCoordination|TestCoordinateAllocation|TestUnsupportedCoordinateLayout|TestCaptureInstallReportsPendingAllocationUntilCommittedManifest' -count=1. Reported package durations: capture 3.845s, CLI 4.572s. The parent owns the required integrated make gate for this corrected candidate. Earlier full-make evidence above applies to the prior revision until that gate completes.
