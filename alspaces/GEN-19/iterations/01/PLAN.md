# GEN-19 iteration 01: Correct the RC target version

Status: User explicitly approved this scope on 2026-11-03. Implementation authorized; approval applies to CONTRACT.md and this plan.

## Baseline and grounding

Baseline: 036d808. Initial Git status was clean. Dedicated branch: chore/GEN-19-rc-target-version. Workspace allocation follows configured alspaces; GEN-19 was absent from the existing workspace, index, and remote branch search. The generated original branch name was not used as ticket allocation evidence.

The faulty block is Compute Release Candidate Metadata in .github/workflows/release-candidate.yml. It derives BASE_TAG using git describe, strips a suffix, and preserves the released numeric version. The source-change detector is separate and is not the version authority. .release-please-config.json configures one root Go package without component-prefixed tags; .release-please-manifest.json contains the last released version, 2.0.0. The Release Please workflow already computes the next release and its open PR #55 proposes 2.1.0.

Existing scripts/tests/test_release_revision.py executes stable-workflow shell in disposable repositories. It covers forced tag revision selection, retargeting rejection, invalid selections, and stable workflow guards; it does not cover RC target selection. Makefile target test-release-workflow currently invokes this file. Preserve these guards and extend the existing entry point for actual RC resolver tests rather than replacing them with workflow-text assertions.

## Design decision

Reuse a pinned Release Please library through its read-only Manifest candidate-building API and consume a structured root-package candidate version. Verify the exact API and compatible version before coding acceptance. Branch reads are mutable: constrain calculation to runs whose checked-out commit equals main before and after calculation, then recheck eligibility before publication. Superseded runs skip explicitly. Validate target greater than stable releases; errors fail closed; no release candidate means an explicit skip.

This avoids three defective alternatives: simply incrementing minor would mislabel patch/major releases; reading the open release PR introduces update races; a separate next-version file requires manual synchronization. Do not parse dry-run terminal output. Do not update the stable-release workflow under this scope. Record any future dependency compatibility concern rather than changing stable policy implicitly.

Primary-source grounding: Release Please's official troubleshooting documentation describes its manifest as the last released version and documents read-only CLI dry runs. Its manifest source exposes buildPullRequests, while GitHub.mergeCommitIterator reads a branch ref through GraphQL. These findings justify using the library and explicit branch-head guards; passing an arbitrary SHA as targetBranch has not been established as supported. Sources: https://github.com/googleapis/release-please/blob/main/docs/troubleshooting.md, https://github.com/googleapis/release-please/blob/main/src/manifest.ts, and https://github.com/googleapis/release-please/blob/main/src/github.ts. Implementation must verify the selected pinned version rather than treating main-branch source as its contract.

## Ownership and execution after approval

1. Parent records user approval against CONTRACT.md and this plan. Planner owns only these two documents; parent owns summary, PR delivery, telemetry, and integration.
2. Coder owns .github/workflows/release-candidate.yml and the minimal CI-only resolver/dependency files. Implement the read-only calculation, structured output, validations, and publication guards. Give the coder this contract, plan, baseline, and current diff; remind them the checkout is shared.
3. Tester owns the relevant release tests and narrowly necessary Makefile test-entry changes. Exercise the production resolver with disposable fixtures and the real pinned versioning engine. Do not implement product behavior or rely solely on mocked library results.
4. Use the four core review responsibilities under both configured model profiles (gpt-6.1-sol/high and gpt-6-luna/medium), in capacity-limited waves. This supersedes the initially suggested single-reviewer exception: release identity and publication guards are substantive release behavior. Behavior owns version authority/build-release integration; data/failure owns stale runs, error/no-candidate separation and stable-target rejection; architecture owns dependency compatibility and maintenance; tests owns regression oracle effectiveness. Each reviewer gets the same frozen revision and a unique report path under this iteration; the parent owns consolidated review and run/finding records. No implementation acceptance boundaries change.
5. Parent consolidates findings, assigns corrections, runs required integration make verification, records reviewed revision and environment limits, and retains compact approved review evidence. Preserve tested evidence for unchanged code; rerun affected checks after corrections.
6. Parent completes normal ticket delivery, including final-step Codex telemetry with literal nonce and ticket type chore, task-owned staging/commits, main integration, push and PR. No release publishing or release PR merge is authorized by this ticket.

## Regression evidence required

- Root version 2.0.0 with feature changes resolves to 2.1.0; prior v2.0.0-rc tags cannot supply the numeric target.
- Patch-only, breaking, and supported explicit version policies use the same engine as Release Please. Use the library's actual parsing/versioning behavior, including message bodies where relevant.
- Stale main before/during calculation and before publication never publish a mismatched candidate. An already stable target cannot be reused.
- Missing/malformed root configuration, ambiguous candidate output, unavailable API/dependency, and invalid target stop before publishing. No-candidate skips distinctly from errors.
- Metadata contains the checked-out SHA and preserved run-number format. Existing cross-compile, package, verification, and prerelease flow consume that value.
- Existing stable revision tests remain green; run the release test target and required make integration gate. Apply the release regression checklist's embedded-asset checks if shared build prerequisites change. Do not claim hosted upload or download verification from local workflow or shell tests.

## Approval boundary

The user explicitly approved the initial scope on 2026-11-03. If API feasibility requires another version authority, stable-workflow changes, or additional release policy, return a material steering delta and obtain renewed approval.
