# AL-79 test effectiveness review — first pass

## Identity and scope

- Requested run identity: `AL79 independent TEST-EFFECTIVENESS first-pass discovery`, `gpt-6-luna/medium`.
- Acknowledged identity: this report records the requested model/effort. Runtime model metadata was not available to this reviewer, so actual model/effort is **unknown**.
- Candidate: AL79-C2; base `4df45967c801824500ac2ea39a8f52fa9cbdcf66`; source/guidance/test diff `/tmp/al79-review-coordinator-b67f2d/CANDIDATE_C2.diff`, SHA-256 `84aa66679b8b2b643ccdfd3c17148b77a369ea383a546eee372397778218a7de`. The observed SHA matches `CANDIDATE_C2.json`. Repository HEAD is the stated base with the candidate changes in the shared worktree; `alspaces` metadata was excluded from the candidate diff identity.
- Focus: test-oracle effectiveness for the frozen source/test candidate. Read the assigned test persona, shared procedure and review skill, contract, plan, PO-20 F1-F6/specification, regression and proportionality guidance, and raw `TEST_EVIDENCE.md`, `DISTRIBUTION_EVIDENCE.md`, `NATIVE_EVIDENCE.md`. Did not use sibling reports or consolidated review conclusions.

## Findings

No blocking test-effectiveness finding identified in the inspected candidate paths.

## Consequential hypotheses and evidence

| Hypothesis | Disposition and evidence |
|---|---|
| A test could pass while the active publisher silently switches from its admitted binary to a newly preferred build. | Refuted for Linux: `TestCaptureWorkerRetainsAdmittedBuildAcrossPreferredSwitch` exercises a real worker process, changes preferred selection, replaces bytes at the mutable path, and asserts later child executable identity remains the retained build; it also checks queued work drains and a subsequent selection uses the new build. Raw evidence records an initial failure and focused correction verification. |
| Installation tests only validate serialized hook text and miss ownership/recovery behavior. | Refuted for the Go transaction boundary: `user_install_test.go` exercises idempotent install, dry-run nonmutation, foreign hook preservation, owned removal, modified-owned refusal, dispatcher-only upgrade preserving hook bytes/group ownership, receipt SHA-256 validation, and refusal to replace a subsequently modified dispatcher without changing receipt/hooks. `test_capture_distribution.py` adds the shipped installer/dispatcher A-to-B, legacy/manual coexistence and caller-cwd/relative-argument boundary. Raw distribution evidence covers the actual cross-process path. |
| Config tests use synthetic directories and would miss incorrect base/worktree/nested-root resolution. | Refuted: `TestSelectedLocalOriginAcrossLinkedWorktrees` creates Git base and linked worktrees plus a nested cwd; exercises local origin selection, paths/custom shared policy and local-file no-fallback behavior. Its companions cover explicit boolean validation. |
| Native callback fixtures omit the consequential Codex SubagentStart parent-session/child-transcript shape. | Refuted: `TestNativeSubagentStartUsesProvenChildTranscriptWithoutRelabelingParent` supplies the native shape, verifies root callback identity is preserved, exercises source-proven child extraction/parent linkage, and rejects contradictory metadata. `TEST_EVIDENCE.md` records the real native callback defect, its correction, then focused repeat and affected-package verification; `NATIVE_EVIDENCE.md` records a clean normal-trust root/child publication. |
| Existing tests that asserted per-worktree receipts or capture config for manual generation were simply deleted, weakening supported manual-read behavior. | No gap established: those expectations conflict with the approved portable-history/manual-reader boundaries. The changed tests retain configured-registry rejection and explicit recovery preservation; the test evidence maps omitted machine-policy expectations to whole-file local-selection tests and manual compatibility fixtures. This is a policy transfer, not evidence of a missing oracle. |

## Verification reused and limits

Reused exact candidate raw evidence rather than rerunning broad suites: focused user-install/retained-executable Go tests passed; focused native callback Codex/CLI tests passed; final affected config/capture packages passed; the repeated lifecycle correction command passed three runs. Real native evidence on Codex 0.167.0/Linux amd64 proves normal-trust root/child callback, binding, source-backed automatic publication and separate worktree destination/history preservation. Distribution evidence covers shell/POSIX and scoped Go behavior. Full integrated `make`/race remains parent-owned and pending. Windows runtime execution was unavailable; cross-compilation/helper source is not runtime evidence. Native `SessionEnd` callback run status is not observable in the cited app-server event stream, although durable registration records the session end. These are explicit boundaries, not test-oracle defects in the inspected scope.

## Outcome

**APPROVED — scoped test-effectiveness review.** The consequential regression hypotheses above have reachable production paths and discriminating assertions or real runtime evidence. No generic count/coverage threshold was used. This does not certify whole-ticket correctness or replace the pending integrated gate and platform limits.
