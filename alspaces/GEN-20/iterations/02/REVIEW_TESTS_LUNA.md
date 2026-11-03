# GEN-20 iteration 02 test-effectiveness review

## Assessment

**Verdict: APPROVED** for the assigned test-effectiveness focus, against frozen revision `67d2a8bdd7039e7638e5dfb68fc435993252433b` (parent/baseline recorded as `dc44619ae15586fe1194f872743033f523635c92b`). Scope: backend capture changes and their tests, with relevant unchanged project, installation, worker, diagnostics, namespace, and relocation consumers. The baseline commit object is unavailable in this checkout, but the candidate commit records it as its direct parent and its complete patch is present in `git show`.

No concrete consequential test gap was established. The implementation report and passing suites were treated as claims, then the new fixtures, assertions, and production call paths were inspected. The report's unexecuted upgrade rollback fixture and numeric-path/protocol-upgrade assertions were intentionally removed under iteration 02; those tests exclusively asserted superseded compatibility behavior. Existing protocol-1 relocation/resume tests and current ownership, worker, and diagnostic tests remain meaningful recipients for preserved behavior.

Requested model/effort and run ID were not supplied in the review brief; runtime metadata is unavailable. The report filename reflects the assigned Luna review slot, not independently verified runtime identity.

## Findings

None.

## Test-oracle assessment

- `TestCoordinationTokensConcurrentFirstOpenAndStableReopen` releases eight real `OpenProject` calls together against a shared registry. The fixture includes two linked worktrees with matching administrative basenames in different repositories, alongside repeated opens of each scope. Assertions detect duplicate per-scope allocations, cross-scope token reuse, loss of Git-derived naming, and read-only reopen instability. The focused race commands reported by the coder exercise the race-sensitive path; I did not duplicate them.
- `TestCoordinationUUIDFallbackPersistsAcrossGitEnrichment` reaches a genuinely unallocated non-Git directory. It checks that preview leaves the state root absent and returns no guessed ID/path, then checks an allocated 16-hex token survives later Git initialization. This directly guards against preview claims and changing an established scope when Git metadata appears.
- `TestCoordinateAllocationSkipsReservedNamesAndBoundsCollisions` calls the production allocator with deterministic entropy. A pre-existing file forces a retry without overwrite, and a persisted-token collision drives the bounded 16-attempt exhaustion path. The test checks the retained file contents. This is a direct allocator oracle rather than a copied algorithm.
- `TestUnsupportedCoordinateLayoutFailsWithoutChangingState` covers both absent mapping table and absent registered mapping, for read-write and read-only admission plus preview. It checks actionable unsupported-layout context, registry database bytes, and a retained coordination sentinel. This is sufficient evidence for the specified non-destructive failure path; it does not claim arbitrary SQLite sidecar byte equality.
- `TestCoordinationTokenAliasesRejectUnknownAndRedirectedState` checks a valid persisted coordination alias, rejection of an unknown alias without creating it, and rejection of a symlink redirected persisted path without touching the external target. Existing namespace alias/containment tests remain in place.
- `TestCaptureInstallReportsPendingAllocationUntilCommittedManifest` runs two fresh CLI dry-runs and checks no registry appears, then performs a real install and requires the active manifest's state path to equal the committed location. A subsequent dry-run must report a resolved path and preserve it. This observes the CLI/install boundary and protects against writing a provisional path into a durable manifest.
- Unchanged consumers retain independent tests: namespace containment/aliases in `namespace_test.go`; installation and manifest ownership in installation and transaction tests; worker successor exact-state and drift behavior in `worker_successor_unix_test.go`; missing-state and read-only diagnostic behavior in diagnostics tests; and relocation interruption/resume and retained inspection in `relocate_test.go` and `relocate_install_unix_test.go`. Relocation fixtures were adapted to obtain production-resolved paths rather than hard-code numeric directory names.

The particular policy-equivalence edge—two unallocated preview inputs resolving to one validated project/worktree identity—is primarily covered by source-level flow through `checkCaptureInstallation` and `PreviewStateLocation`, while the new CLI test exercises repeated fresh dry-runs and committed-path stability. I found no production path in the inspected callers that generates independent provisional UUIDs: provisional locations have an empty coordination path and callers compare the validated registry identity until allocation. This is a bounded source-trace conclusion, not a dedicated adversarial policy fixture.

## Hypotheses investigated and disposition

| Hypothesis | Result and evidence |
|---|---|
| Concurrent first opens can allocate different names for one worktree or share one name across worktrees. | Refuted by synchronized real opens, per-project/per-path assertions, same-basename distinct-repository fixtures, plus reported passing focused race run. |
| A reserved filesystem entry or occupied persisted token can be overwritten, or collision retries can loop forever. | Refuted by deterministic production-allocator test and explicit bounded-exhaustion assertion. |
| Unsupported layouts can be silently repaired, adopted, or destructively rewritten. | Refuted for the required missing-table and missing-row cases by read-only/writable/preview rejection and unchanged DB/sentinel checks. |
| Alias handling can create state for an unknown coordinate or follow a redirected coordinate symlink. | Refuted by acceptance test; existing namespace tests retain broader containment coverage. |
| A dry-run can claim a UUID that a later install does not use, or installation can persist a provisional location. | Refuted by pending-allocation CLI assertions and exact manifest-to-committed-location comparison. |
| Removing iteration-01 upgrade/numeric-path assertions discarded useful supported behavior. | Not substantiated: those assertions encoded explicitly superseded upgrade, numeric preservation, and old-reader fencing. Current-schema allocation, rejection, and supported protocol-1 relocation/resume recipients remain. |
| Diagnostics, worker succession, or relocation can regress because consumers resolve different coordinates. | No concrete missing oracle found: inspected unchanged suites cover no-state/read-only diagnostics, exact worker state and drift, and relocation/resume; fixtures now use actual resolved coordinates. |

## Verification and limits

Reused the tester's and parent/coder's recorded evidence rather than rerunning suites:

- Tester ran the targeted `go test -list` discovery command, confirmed the listed acceptance tests, and ran `git diff --check` successfully. Test listing alone is not body execution.
- Coder reported both focused `go test -race` commands passed, including coordination/allocation/unsupported layout/installation acceptance and relocation cases.
- Parent reports integrated `make` passed on the frozen revision, including backend race tests, 790 frontend tests, static/Python checks, embedded assets, and platform builds; log: `/tmp/gen20-make.log`.

I did not independently rerun the test suites or inspect the integrated log; verification is reused evidence from the task brief/test report. No isolated mutation was needed because each high-risk acceptance case directly reaches the production boundary and has a specific failure oracle. This approval is limited to test effectiveness for the inspected candidate and is not a whole-system correctness or merge approval.
