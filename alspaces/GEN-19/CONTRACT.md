# GEN-19: Release candidates target the upcoming release

## Accepted iteration 03 supersession

The user's "make it simplier" instruction explicitly approves the immediately preceding proposal: retain successful Release Please sequencing, immutable release-PR manifest capture, validated version/suffix metadata and one exact stable-target check before publication. Remove generated-commit-parent assumptions, repeated main-head freshness checks, historical stable-release scans and the second PR identity comparison. A newer main commit or release-PR update after capture no longer invalidates the built candidate. These explicit relaxed guarantees supersede iteration 02's freshness/provenance/history requirements. Exact candidate SHA checkout, ordinary API/invalid-version failure handling, build verification and prerelease/retention behavior remain. See iterations/03/PLAN.md. No renewed permission is needed for this already approved reduction.

## Accepted iteration 02 supersession

On 2026-11-03 the user explicitly approved the simpler approach in iterations/02/PLAN.md: read the version already calculated in the open Release Please PR, sequence automatic RC execution after Release Please finishes, and remove the independent engine and its dependencies. This supersedes the iteration 01 requirements to calculate via a pinned library and never consult a PR. The version source is the root manifest at the immutable head of the same-repository generated release PR; freshness/provenance, stable-target rejection, metadata and publication guards remain required. Stable publishing workflows and historical releases remain outside scope. The original contract below records the approved baseline; the explicit accepted delta governs the replacement.

Status: Approved by the user's explicit "Approve this scope" response on 2026-11-03. Implementation is authorized within these boundaries.

## Problem and evidence

The Release Candidate workflow currently runs git describe, removes prerelease/build suffixes, and appends a new RC suffix. This reuses the previous release number rather than determining the upcoming release. At baseline 036d808, .release-please-manifest.json records 2.0.0. The parent verified that v2.0.0 is published as stable and Release Please PR #55 targets 2.1.0, while recently published candidates still use v2.0.0-rc.

## Accepted behavior proposed for approval

Calculate the upcoming version using Release Please's read-only candidate-building API, the existing .release-please-config.json and .release-please-manifest.json, and the same release rules used for stable release proposals. For the current release train, produce v2.1.0-rc.<run-number>+<short-sha>. Patch, minor, major, and configured release-as decisions must follow Release Please rather than a new independent increment algorithm.

The calculation must describe the checked-out candidate commit. Since Release Please reads branch history through GitHub, require main's remote head to equal the candidate commit immediately before and after calculation. A superseded queued run must skip publication with an explicit explanation. Recheck eligibility before publication so a stable release or advancing main cannot cause an obsolete candidate to be published silently. Do not pretend a SHA is a supported branch parameter without verifying the selected library's API.

Use the structured candidate version, never human-readable CLI logs, PR title parsing, the most recent RC tag, or a possibly stale open PR. Validate one unambiguous stable semantic target for the root package and require it to exceed published stable versions. A calculation with no release-worthy candidate skips explicitly; malformed/missing configuration, ambiguous results, API failures, and unsupported API behavior stop publication with actionable errors. There is no fallback to the previous stable version or a guessed version.

Preserve the existing RC run-number/SHA suffix, build metadata propagation, archives, verification, prerelease classification, and retention behavior. Do not rename or delete existing releases, merge release PRs, publish a stable release, change stable-release version policy, or add a manually maintained second next-version setting.

## Implementation boundaries

Changes are limited to RC version resolution, its CI-only dependency setup, .github/workflows/release-candidate.yml, proportionate executable tests and their Makefile entry, and directly relevant contributor documentation. The coder must establish a compatible pinned Release Please dependency and demonstrate consistency with the existing stable proposal engine. Changing the stable publishing workflow or product code is outside this proposal and needs renewed scope approval.

Go domain/adapters, schemas, frontend behavior, and production distribution remain outside the implementation scope. Any helper is CI tooling; Node is already present in the release runner and must not become a production runtime dependency. Preserve FCIS, static typing, zero any in Go/TypeScript, zero emojis, and unrelated checkout changes.

## Completion boundaries

Prove actual version-resolution behavior for patch/minor/major and explicit version policy, the current 2.0.0-to-2.1.0 case, stale-run guards, prior RC contamination, no-candidate behavior, and failure without publication. Reuse existing workflow tests where coherent. Run the repository make gate after integration and verify build metadata reaches the existing release build boundary. Local fixtures are not proof of successful hosted publishing; report any unexecuted hosted boundary explicitly.

Public API and branch-bound behavior must be established before implementation is accepted. If a compatible read-only API cannot meet these constraints within the stated CI-only scope, return to planning rather than silently substituting PR scraping or custom semver rules.
