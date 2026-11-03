# FIX-16 iteration 01: Isolate source-build assertions from release metadata

Status: Approved. Explicit scope approval received before coder dispatch.

## Grounding and ownership

Restore release candidate publication by correcting the check-stage failure recorded in GitHub Actions. Initial Git status was clean, branch FIX-16/restore-release-publishing, baseline dc44619ae15586fe1194f872743033f523635c92. There was no existing FIX-16 workspace or matching ticket-index entry found. Configuration resolves workspaces.dir to alspaces and FIX to conventional type fix.

Planner owns only CONTRACT.md and this plan. Coder owns scripts/tests/test_installation.py and any justified narrowly scoped shared-fixture adjustment in scripts/tests/distribution_fixture.py. Tester owns verification and may amend missing regression assertions in the same test after coder completion, with sequential ownership handoff. Parent owns SUMMARY.md, PULL_REQUEST.md, review consolidation, integration, telemetry and publication. All agents share the checkout and must preserve others' edits.

## Confirmed GitHub evidence

Read .github/workflows/release-candidate.yml, .github/workflows/release-stable.yml, GEN-19 CONTRACT.md and iteration 03 PLAN.md, Makefile, scripts/tests/test_installation.py, scripts/tests/distribution_fixture.py, scripts/tests/test_release_revision.py, TECHNICAL_GUIDELINES.md testing rules and docs/REGRESSION_PREVENTION.md release/build checklist.

[Release Candidate run 41733288460](https://github.com/NiltonDuarte/AgenticLens/actions/runs/41733288460), at the baseline revision, completes metadata computation and fails Static Analysis. Its job records show that release compilation, packaging, verification and publication are skipped after that failure. The failed-step log records VERSION=v3.0.0-rc.122+dc44619 and an assertion failure at scripts/tests/test_installation.py:33: actual v3.0.0-rc.122+dc44619 differs from expected v2.0.0-rc.115+65e0eeb-7-gdc44619.

The [installation test at the run's exact revision](https://github.com/NiltonDuarte/AgenticLens/blob/dc44619ae15586fe1194f872743033f523635c92/scripts/tests/test_installation.py#L17-L39), also retrieved through the GitHub contents API, invokes make verify-build with the inherited environment and compares the executed binary's version to unconditional git describe output. [Makefile](https://github.com/NiltonDuarte/AgenticLens/blob/dc44619ae15586fe1194f872743033f523635c92/Makefile#L14-L20) honors explicit VERSION through its conditional default and linker flags. The [candidate workflow](https://github.com/NiltonDuarte/AgenticLens/blob/dc44619ae15586fe1194f872743033f523635c92/.github/workflows/release-candidate.yml) exports release metadata before make check. These records establish the failed assertion and its environment mismatch. They do not establish the result of downstream stages that never ran.

An initial local reproduction attempt stopped earlier because web/node_modules and the generated backend embedded asset tree were absent. It did not locally reproduce the version assertion; the hosted traceback is the direct mismatch evidence. Install dependencies and generate canonical assets after scope approval before repeating the local probe. Prior candidate run 41720631449 also failed Static Analysis according to hosted job metadata, but its logs were unavailable; do not attribute its exact failure without evidence.

GEN-19 intentionally computes candidate metadata before checks. Stable Release computes metadata after checks. Preserve those workflows and their accepted guarantees. Do not restore superseded GEN-19 freshness/history restrictions.

## Implementation slice

1. After user approval, have the coder adapt the existing source-build test, avoiding a new copied build script or weakened assertion. Control VERSION, GIT_COMMIT and BUILD_DATE only in the subprocess environment needed for the test. Do not mutate the parent process environment or remove release metadata from the actual workflow.
2. Exercise the existing real make verify-build invocation with an explicitly clean default-version environment and with an explicit, valid candidate VERSION different from the installed pin and the checkout-derived version. Reuse the existing setup and assertion owner with subcases or a small coherent helper. Make each oracle independent of the inherited VERSION.
3. Keep the native identity assertion, printed verification output, installed native inode/bytes preservation, selected pin launcher behavior and absence of a source wrapper copy. Preserve existing assertions rather than replacing them with a return-code-only check. Do not add exhaustive metadata matrices unrelated to the observed regression.
4. If inspection shows that broad fixture sanitization would affect unrelated installation contracts, keep the change local to this source-build test. Do not touch application code, release selection, quality-gate ordering or publication guards.

## Verification

The existing source-build installation case owns coverage for the observed failure: exported release VERSION reaches the binary while the test expects the Git-derived default. The production path exercised is Makefile verify-build/build, actual Go linker metadata and direct native version execution. Temporary installation/cache fixtures supply the separate invariant that source development must not alter the selected release cache.

Tester records commands, tested diff or commit, results and limits. Run the focused source-build case first with metadata unset, then with the real observed candidate metadata shape exported. Run the entire installation suite under exported candidate metadata and the release-workflow/distribution suites. Ensure both default and explicit override subcases run in each environment; no silent skips. A deterministic explicit override is required even when the outer environment is clean, so normal CI continuously protects the escaped regression.

Run the integrated make gate with valid candidate metadata exported to cover the same environment that failed remotely, retaining ordinary-default evidence from the focused case. Canonical dependency installation and frontend asset generation are prerequisites. Inspect web/dist and backend/internal/adapters/web/dist and use existing native/embedded-asset assertions; generated assets remain ignored. This test-only correction does not change frontend source or embed mechanics, so no new asset test or destructive clean-asset experiment is required. Reuse valid existing release-shell tests; they cannot prove remote publication.

When feasible, execute existing release-build, release-package and release-verify targets for the host platform with explicit candidate metadata, keeping artifacts ignored or disposable. This supplies packaging/checksum/native evidence beyond the failing check stage; record any unexecuted platform matrix or hosted upload boundary explicitly.

## Focused review

This is a narrow test-environment correction with no production or workflow behavior change. Use the narrow-review exception in the canonical parallel-review protocol: independent test-effectiveness review covers oracle independence, retention of cache assertions, both metadata modes and shell environment inheritance; behavior/integration review covers actual Makefile override behavior and unchanged workflow metadata propagation. Separate data/architecture personas are omitted because no persistence, failure policy, application boundary or architecture changes. Expand both selected focuses into all model profiles required by the current model policy; parent loads that policy before dispatch. If scope reaches runtime or workflow changes, reassess coverage before implementation.

Test-effectiveness reviewer owns only iterations/01/REVIEW_TESTS_<profile>.md; behavior/integration reviewer owns only iterations/01/REVIEW_BEHAVIOR_<profile>.md. Freeze the same diff fingerprint or commit for all reviews, pass contract, plan, baseline, current diff, relevant source pointers and raw verification evidence. Reviewers have read-only product/test access and use temporary probes. Parent alone writes consolidated review and approves delivery after required integrated checks.

## Delivery and limits

Prepare contract-linked PR and summary artifacts, commit only task-owned paths, synchronize origin/main with a clean tree, run the required gate and resolve only demonstrated findings. No matching plan specification requires lifecycle movement at present. Preserve all unrelated historical workspaces.

Generate real Codex telemetry only at the final pre-commit step with a literal nonce, FIX-16 and type fix; confirm actual CLI options and root/child session extraction. Record unavailable source facts without fabricating a ledger. Publish the feature PR and link it immediately using the host integration. No release PR merge or stable publication is authorized. A successful hosted candidate run and uploaded assets, if subsequently observed, should be recorded separately from local verification; otherwise hand off with that boundary explicitly unverified.
