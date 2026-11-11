# AL-81 iteration 01: Separate manual extraction from capture admission

Approval: Explicit user instruction “$agentic-lens-dev plan is approved” on 2026-11-11 authorizes this contract and implementation plan. Baseline: `4e5c458322a2cddfdb5911b26a4569ea75f06a09`; branch `AL-81/metrics-without-capture-registry-access`; initial working tree clean. Configuration selects `alspaces`; parent owns setup/index and summary artifacts. Follow the [contract](../../CONTRACT.md).

## Implementation and ownership

After explicit scope approval, the coder owns affected production Go paths in `backend/cmd/agentic-lens/` and `backend/internal/adapters/config/`, plus directly needed source-validation helpers and canonical manual-generation documentation. The tester owns affected/new test files; parent owns integration and delivery artifacts. All roles share this checkout and preserve other edits.

1. Give `runGenerate` a manual configuration path that retains selected source/path policy without requiring `decodeCapturePolicy`, `ResolveGenerationCapture`, configuration capture admission or state allocation. Keep automatic loading/admission strict. Manual `--capture-dir` remains accepted without touching its path; reject invalid explicit budget values while making capture budgets unnecessary for manual extraction.
2. Restrict registry reconciliation, association checks, retained-history diagnostics and publication acknowledgements to automatic operation. Split source/project validation out of `validateCaptureSelection` before bypassing its registry mutations; preserve trusted portable owner refresh and child ancestry rules. `classifyGenerationBaseline`, snapshot replacement and owner-preserving serialization must remain registry independent.
3. Preserve `capture.AcquirePublication` as the common canonical-output lock, acquired before baseline/source reads, together with destination validation and the existing publication journal. No lock relocation or new persistence schema is planned. Keep worker `--capture-auto` invocation and ensure explicit session/nonce options cannot route automatic requests around registry admission/binding.
4. Update canonical operator/setup documentation and any shipped equivalents that still describe manual capture-state requirements. Keep automatic installation/ownership guidance intact.

## Verification and review

Select extraction/refresh, portable ledger, configuration portability and publication failure boundaries from [regression prevention](../../../../docs/REGRESSION_PREVENTION.md). Extend existing owners (`main_codex_test.go`, `main_refresh_test.go`, `owned_refresh_regression_test.go`, `capture_isolation_test.go`, capture worker/history tests, config admission tests and publication tests) only where assertions are missing. Reclassify prior manual-binding expectations as automatic-only where required; preserve their distinct protections.

Use isolated Codex rollout and OpenCode SQLite fixtures: omitted capture policy; missing/conflicting/corrupt/inaccessible state with no initialization or registry byte/metadata changes; foreign projects and true aliases; missing/invalid sources; root/child and second root; complete lower/zero refresh, unchanged repeats, unavailable/partial same-source and other-harness history. Deterministically exercise concurrent manual/manual and manual/automatic publication for lost updates/duplicates, interruption recovery and unsafe destinations. Verify disabled, unassigned and conflicting automatic requests, including explicit selectors and actual worker invocation.

Run focused affected checks, then integrated `make` after merging main; record revision, actual execution and limits. At final extraction timing, compare the same real Codex root/child source under isolated registry-absent and registry-present conditions without changing native capture state. Record native IDs, parent linkage, source/config selection, output facts and any source growth; raw transcripts remain outside Git. Missing native evidence is incomplete, not fixture success.

Freeze the candidate for four independent focus reviews: behavior/integration, data/failure handling, architecture and test effectiveness, each with Sol 6.1/high and Luna 6/medium per canonical review policy. Parent assigns unique reports before dispatch and consolidates findings. No unresolved product decision is identified; native source availability remains an evidence prerequisite.

## Review assignments

The frozen candidate will be recorded before dispatch. Each reviewer owns only its unique report under this iteration; product and test edits pause during review. Every cell reads the same contract, baseline/candidate diff and raw test evidence independently.

| Focus | Risk question | Completed evidence |
| --- | --- | --- |
| Behavior and integration | Do manual sources work without capture state while worker and explicit-selector automatic paths retain admission? | `REVIEW_SUMMARY.json` and ticket `SUMMARY.md` |
| Data integrity and failure handling | Are incomplete/unavailable owners preserved and writes serialized and recoverable across both paths? | `REVIEW_SUMMARY.json` and ticket `SUMMARY.md` |
| Architecture and maintainability | Is configuration/source validation separated without relaxing automatic policy or scattering authority? | `REVIEW_SUMMARY.json` and ticket `SUMMARY.md` |
| Test effectiveness | Do real adapter fixtures and failure oracles cover registry independence, refresh and automatic guards? | `REVIEW_SUMMARY.json` and ticket `SUMMARY.md` |

Configured profiles: `gpt-6.1-sol` / high and `gpt-6-luna` / medium. Runs use capacity-limited waves. Parent owns consolidation, telemetry and publication.

Candidate frozen for independent review: `cfdc01ae4049c0cd34cb6bc5859e7786c6766a4a5e069f3d60e5cc2f430d9981`, computed from sorted changed/untracked `.agents` and `backend` paths followed by NUL, full file bytes, NUL. Baseline unchanged. Focused affected checks passed before dispatch; full CLI run was pending. Review reports record their own evidence and scope.

## Review corrections within approved acceptance

Independent data reviews identified two pre-existing paths that defeat the accepted partial-source preservation/reporting guarantee: grouping warnings could erase an accepted child link, and nonce scanning discarded an already identified child after a later scanner error. The coder owns rejection before publication for both paths; tester owns root/child ledger-and-summary regressions and unrelated-corrupt-source controls. No new scope or authenticity policy is introduced: known selected-family failures are actionable, while wholly unattributable/unrelated files cannot be assigned to a ticket by inference. The architecture review also identified unused manual admission/recovery branches, now removed without changing automatic policy. Verification requires corrected candidate tests and affected focus re-checks, followed by integrated `make`.

Corrected candidate frozen for affected focus/model re-checks: `5de23fa986778b4e8048c9a42f13321e950e2e4d5b3c0ab36bdc7ee06e65e62b`, same complete-file fingerprint method. New grouping/discovery regressions pass; restoring baseline discovery in a disposable overlay fails only the selected nonce-child case. All four focuses re-check the small correction set under both original profiles; reports append revision-scoped outcomes.

## Completion evidence

Reviewed source was committed as `39356cee20080152906e61ef9c668e0548628adb`; origin/main integration was already current, so all eight rechecks remain applicable without a product diff. Integrated `make` passed every lane, including full CLI race tests and actual frontend integration. Standalone version/help and embedded-asset equality passed. Native source root/eleven-child extraction matched registry-absent/present with unchanged state; [native evidence](NATIVE_EVIDENCE.json) records IDs, source hashes, facts and cutoff/runtime limits. Consolidated review is APPROVED. Completed reports are compacted into [retained findings](REVIEW_SUMMARY.json), with scope, actors and initial/final outcomes retained in ticket SUMMARY.md. Final telemetry follows this completion account immediately before the artifact commit.
