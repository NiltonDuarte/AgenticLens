# FIX-17 iteration 01

Baseline: 86bea597afe7208cb1836f7cc5747083f7b95f47. Initial working tree clean. Ticket workspace alspaces/FIX-17; feature branch fix/FIX-17-telemetry-capture. CONTRACT.md defines the approved boundary.

## Steering and approval

Initial request: fix telemetry capture. Subsequent user evidence identifies failure before extraction caused by the configured registry lacking worktree_coordinates despite metadata version 2; a fresh --capture-dir is also blocked. Root and eleven child source sessions are available. This evidence supersedes the unspecified failure hypothesis; it does not authorize modifying GEN-21.

The user subsequently explicitly approved recovery through an independent healthy/fresh explicit --capture-dir while preserving old state, portable history and legitimate installation isolation checks. No reset, migration or configuration change is authorized. Verification must prove the recovery and retain real installed-state rejection checks. The scope gate is satisfied for this bounded implementation.

## Grounded design

1. Coder owns backend/cmd/agentic-lens/capture.go and narrowly necessary CLI helpers plus corresponding focused regression tests. Inspect main.go generation admission and checkCaptureInstallation, then distinguish explicit independent state selection from default/same-state selection before previewing the incompatible configured coordinate schema. Prefer the smallest change; do not broadly catch and ignore schema errors.
2. Continue validating the selected state's schema and admission. Retain capture.CheckInstalledState from backend/internal/adapters/capture/installation.go, which discovers installed states and validates manifests/integrations without opening the configured registry. Separate the irrelevant registry preview from checks that still protect installed paths. Explicit drift retains existing warning semantics; actual installed ownership/provenance/integration failures remain failures.
3. Reuse the existing capture fixture and coherent peers: capture_corrections_test.go's TestCaptureInstalledStateDriftIsDiagnosedAndRecoveryPreservesOldRegistry, capture_isolation_test.go, capture_runtime_integration_test.go, and capture_history_authority_test.go. Add only missing assertions that detect this failure: incompatible configured coordinate layout plus independent fresh/healthy selection succeeds; no override and selecting bad state fail; configured registry stays unchanged; a real installed-state failure is still rejected under the override. Reuse existing portable-history cases unless a changed path needs one focused addition.
4. Tester independently checks the new regressions against baseline or a disposable targeted reversal, runs affected CLI/capture tests, and verifies unavailable/other-harness history and selected-state isolation through existing coverage. Record commands, tested revision/diff, outcomes and limits. No live source-store dependency or raw transcript artifacts.

## Review assignments and ownership

Freeze the implementation revision before review. Parent dispatches four distinct focuses across both configured profiles, gpt-6.1-sol/high and gpt-6-luna/medium, in capacity-limited waves. Each reviewer reads the contract, this plan and its own persona; product files are read-only and checkout changes by others must be preserved. Parent consolidates findings and owns publication.

| Focus | Responsibility | Separate report paths under this iteration |
| --- | --- | --- |
| Behavior | Explicit/default/same-state selection and native capture strictness | REVIEW_BEHAVIOR_SOL.md; REVIEW_BEHAVIOR_LUNA.md |
| Data and failure handling | Old registry preservation, failed admission, portable owner/history semantics | REVIEW_DATA_SOL.md; REVIEW_DATA_LUNA.md |
| Architecture | Narrow shell change and installation guard preservation | REVIEW_ARCHITECTURE_SOL.md; REVIEW_ARCHITECTURE_LUNA.md |
| Test effectiveness | Regression detects baseline, isolated fixtures and retained guard assertions | REVIEW_TESTS_SOL.md; REVIEW_TESTS_LUNA.md |

Record each actual reviewer model/effort, frozen revision, findings and disposition under the canonical parallel-review protocol. Correct findings through coder/tester and re-review affected focuses. Incomplete evidence does not pass. After approved consolidation retain REVIEW_SUMMARY.json and concise actor/verification evidence in SUMMARY.md; remove completed detailed reports only after retaining that evidence.

## Verification and delivery

Applicable regression checklist: telemetry refresh preserves exact owners and unrelated sources; portable ledgers do not depend on local receipts; explicit state/config selection preserves isolation; failures do not masquerade as empty extraction; fixtures isolate all writable state. No frontend changes are planned, so frontend-specific additions are unnecessary, but the integrated make gate remains required.

Commit reviewed implementation on the feature branch, integrate origin/main and run required make. Preserve any unrelated changes and stage only task-owned files. Update SUMMARY.md and the PR draft, then generate this ticket's actual Codex telemetry at the final step using a literal nonce and fix classification; verify root/children and existing-harness preservation before committing. Publish the reviewed tested PR and head for the GEN-21 owner; do not modify that ticket's artifacts or claim its real extraction has passed.

Remaining limits: origin of the incompatible registry is unknown. Automatic migration and historical registry recovery are out of scope. Real GEN-21 root/eleven-child generation occurs in its owner's thread after integration. Planning has not executed tests or claimed a runtime fix.

## Implementation and verification steering

The user explicitly authorized a final real-source recovery run using this FIX-17 native Codex root and a fresh capture override, preserving the incompatible registry. This replaces fixture-only evidence as the final recovery acceptance; it does not authorize GEN-21 edits. The user also authorized completion notification to the controlling GEN-21 thread, and subsequently requested an early immutable reviewed-source handoff for parallel integration. Source commit d4a9e6edb7b36247a5c35db15c2e0e64a050d775 was sent; publication readiness was explicitly withheld pending the remaining gate and telemetry. Final notification must include PR/head, dependencies and runtime limits.

Review corrections remain within iteration 01: preserve healthy configured-local manifests when discovery files vanish; then preserve manifests at literal configured coordination paths even when mappings are unsupported. Known unsupported layout errors gained an adapter sentinel; unrelated errors remain blocking. Final source diff fingerprint: 6cc147ec3e399a59e5589d7a19df5882cd98571b0e9db848c1cdc73772aeb439. Both fixes were red/green verified; all eight focused model/persona cells approved. No frozen contract or portable-history authority changed.
