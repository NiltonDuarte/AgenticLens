# GEN-19 iteration 02: Reuse the existing release PR version

Status: User explicitly approved the simpler approach on 2026-11-03. Implementation authorized within these revised boundaries.

## Steering delta

Source: the user's follow-up says iteration 01 is overbuilt and asks to reuse the mechanism already calculating the version in the open release PR. This supersedes the independent pinned-engine calculation and the contract's prohibition on consuming the release PR. New acceptance: obtain the version already written by Release Please, remove the duplicate engine and dependencies, and order automatic RC execution after Release Please succeeds. Preserve version correctness, explicit skip/error behavior, stale-run guards, suffix format, artifact metadata, existing stable workflows, and prerelease retention. The parent records the accepted contract delta after approval; prior verification remains historical evidence, not proof of this replacement.

## Grounding

Baseline e639f33; existing PR #98 remains open. Fresh parent evidence identifies open release PR #55 on main, same repository, expected root release branch, and autorelease: pending label. Its immutable head 1300741296d069b1282d89a6192219bf13fbb4f9 contains root manifest version 2.1.0. Current .github/workflows/release-candidate.yml invokes scripts/release/rc-version.cjs, and Makefile installs that engine before release tests.

## Proposed implementation

Replace automatic push triggering with successful completion of the existing Release Please workflow on main using workflow_run. Preserve workflow_dispatch. Both jobs explicitly check out the upstream run's head_sha for workflow_run; manual dispatch uses its selected SHA and permits publication only when that equals current main. Read the checked-out commit message through git log for release-commit skipping because workflow_run has no push head_commit message. Preserve source-change detection otherwise.

Discover exactly one open same-repository root Release Please PR targeting main, using its known branch and label rather than title parsing. Read .release-please-manifest.json at the discovered immutable head SHA. Require a strict stable root version and append the existing RC run-number/SHA suffix. Check the generated PR head's source-main provenance against the candidate SHA; establish the simple generated-commit parent check against actual Release Please output before relying on it. A completed workflow alone cannot prove a concurrently updated PR describes this candidate.

Skip explicitly for no release PR, stale PR, or superseded main. Fail closed on API errors, ambiguity, malformed manifest/version, or unsupported provenance. Compare numeric version components with existing runner tools to reject targets already published stable. Recheck current main, PR identity/head/state, and stable-target eligibility before publication. No guessed fallback, new version calculator, new package dependencies, or published release is part of this work.

Remove scripts/release, its dependency installation from Makefile and workflows, and tests specific to the deleted engine. Preserve existing stable revision tests and replace their engine assertions with proportionate executable shell coverage for the new authority and guards.

## Ownership, verification, and limits

Coder owns RC workflow, engine removal, Makefile cleanup, and directly affected guidance. Tester owns scripts/tests/test_release_revision.py. Parent owns integration, review consolidation, artifacts, telemetry, commits, and PR delivery; all preserve shared-checkout changes.

Execute actual workflow shell with disposable Git/manifest fixtures and a fixture gh command. Cover 2.1.0 selection, immutable manifest reads, event SHA selection, missing/stale PR, ambiguity/API/malformed failures, stable-target rejection, and publication guards. Review behavior, failure handling, architecture, and test effectiveness under the configured review protocol. Run affected tests and required integrated make; local fixtures do not establish hosted workflow scheduling or upload success.

GitHub's workflow_run documentation confirms completion triggers and event-specific SHA semantics: https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#workflow_run. Stable publishing workflows remain unchanged.

## Review roster and frozen evidence

Base e639f3303c9d92236edce85befccda66617d7cb8; frozen source/deletion fingerprint 903c512467c9100cbb6b80756e501280efb9a3ab08f3466c0f4890cae92ece1e. Four responsibility owners independently review under both configured gpt-6.1-sol/high and gpt-6-luna/medium profiles, in capacity-limited waves: behavior/integration owns event/SHA/version-to-artifact wiring; data/failure owns stale/error/stable-target rejection; architecture owns simplicity/dependency removal and maintenance; test effectiveness owns the regression oracles and removed assertion mapping. Reports use unique REVIEW-<focus>-<sol|luna>.md paths in this iteration; parent alone owns run/finding/consolidated records. Product/test edits remain frozen during the pass. Detailed records are working artifacts until consolidated approval and compaction.
