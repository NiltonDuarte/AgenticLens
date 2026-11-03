# GEN-19: Align release candidate versions with the upcoming release

## Request and baseline

- User requested the next available GEN ticket for candidates labeled v2.0.0 after stable v2.0.0, when the intended next release is v2.1.0.
- Allocated GEN-19 after checking local workspaces, the ticket index, remote workspace tree and matching remote branches. The generated initial branch name referenced an already used ticket and was not used for allocation.
- Initial working tree: clean. Baseline: 036d808. Ticket branch: chore/GEN-19-rc-target-version.
- Configuration: workspace root alspaces; GEN conventional type chore.

## Investigation

- The Compute Release Candidate Metadata step in .github/workflows/release-candidate.yml selects a tag through git describe and removes prerelease/build suffixes. It never computes the upcoming release version.
- GitHub release list on 2026-11-03 confirms stable v2.0.0 and subsequent v2.0.0-rc candidates, including v2.0.0-rc.113+b1742e9.
- Open Release Please PR https://github.com/NiltonDuarte/AgenticLens/pull/55 targets 2.1.0. Main's .release-please-manifest.json records the last released version, 2.0.0; it is not an upcoming-version source.
- Existing scripts/tests/test_release_revision.py executes stable Release Please workflow shell in disposable repositories. It does not cover RC target selection.

## Status and verification

Planning complete: see [contract](CONTRACT.md) and [iteration 01 plan](iterations/01/PLAN.md). Proposed resolution reuses Release Please's read-only upcoming-version calculation, preserves the RC suffix, skips superseded/no-candidate runs, and rejects invalid or already stable targets. The planner authored these artifacts; the parent reviewed them against the request. Compatibility of the pinned library API remains an implementation acceptance requirement.

The user explicitly approved the proposed scope on 2026-11-03. Implementation and regression testing are delegated with separate writable-path ownership. The release identity change receives the four core review responsibilities under both configured model profiles; this corrects the initial narrow-review suggestion without changing product scope.

Environment preparation: npm ci in web succeeded, and make build prepared the embedded frontend and standalone CLI. Full integrated verification remains pending. No release has been published, retagged, or deleted. Telemetry is deferred until final iteration delivery.

## Implementation and focused verification

- CI-only scripts/release/rc-version.cjs uses pinned Release Please 17.3.0, matching the stable action v4 engine inspected during implementation. The read-only Manifest API applies existing repository configuration and conventional commit rules; no independent bump algorithm or manual next-version setting was introduced.
- The RC workflow consumes validated metadata, skips superseded/no-candidate runs, rejects already stable targets, and rechecks before publication. Added explicit pull-request read permission for the engine's GraphQL reads.
- Coder live read-only probe against main/036d808 returned ready target 2.1.0. No remote release/PR mutation occurred.
- Tester: make test-release-workflow passed, 8 Python tests and 13 Node tests, zero skips. These preserve stable revision guards and exercise real engine bump rules plus isolated GitHub boundary and workflow guards.
- Parent: the same 13 Node tests passed under Node v20.20.2, matching CI's Node 20 major (host default is v24.22.1). git diff --check passed; built binary version --json ran successfully; web/dist and the embedded asset destination matched exactly.
- Full integrated make and consolidated review remain pending. Local test fixtures and the live calculation do not establish hosted publication/download success. Independent stable/RC workflows cannot make final eligibility reads and publication atomic; documented in scripts/release/README.md.
