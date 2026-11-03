# GEN-19 behavior and integration review

Verdict: APPROVED for the assigned behavior/integration scope. No blocking correctness findings. Parent retains consolidation and delivery ownership; this is not authorization to publish a release.

- Local run: GEN-19-behavior-sol-01, independent first pass.
- Requested identity: gpt-6.1-sol / high. Actual model and reasoning-effort runtime metadata were not exposed to this reviewer; requested identity is not asserted as independently observed identity.
- Baseline/current HEAD: 036d808b0c77bec400f23ca77a5cf056a0b2078b plus the frozen uncommitted candidate.
- Fingerprint: 3a5ea9f89c48818862332b7da06019145544ec9ab2edcb830cae9a28b752765b. Independently SHA-256 checked all nine CANDIDATE.json files; every file matched.
- Read CONTRACT.md, iteration PLAN.md, assigned persona/shared procedure, technical standards and applicable release regression guidance. No sibling review reports or conclusions read. Product/test paths remained read-only.

## Investigation and evidence

1. Upcoming-version authority: traced `.github/workflows/release-candidate.yml:97` through `scripts/release/rc-version.cjs:32` and pinned Release Please 17.3.0 `Manifest.fromManifest(...).buildPullRequests()`. Inspected the installed manifest's release-baseline lookup, conventional-commit parsing, root Go strategy and branch-bound commit iterator. The checked-in root configuration/manifest remain the stable engine's inputs. Candidates use the structured version field. Neither git describe, the newest RC nor open PR text selects the numeric target. Actual-engine fixtures distinguish feature 2.1.0, fix 2.0.1, breaking 3.0.0, commit Release-As 2.3.0 and configured release-as 4.0.0.

2. Real integration countercheck: executed the exported production resolver with the real GitHub adapter, authenticated reads only, on Node v20.20.2. The checkout/expected commit was 036d808b0c77bec400f23ca77a5cf056a0b2078b; result was `{status:"ready",version:"2.1.0"}`. This exercised actual API history and changed-file backfill, rather than only injected fixtures. Credentials stayed in process memory and were not printed. No API writes or publication occurred.

3. Superseded/no-candidate/error distinction: traced initial and post-calculation head reads, the additional ready guard and final workflow eligibility step. Tests exercise stale before/during resolution, stale before publication, no release-worthy commit, missing/invalid configuration, API errors and targets equal to/below published stable releases. RCs/drafts are excluded from the independent stable-release check; pagination reaches stable releases after 100 RC records. Invalid outputs fail without publication permission. Both publication and pruning depend on the final eligibility output; failed preceding steps remain subject to GitHub's default success condition.

4. Release identity propagation: `candidateMetadata` writes the existing `v<target>-rc.<run>+<short-sha>` format, commit and UTC build date to GITHUB_ENV. Workflow lines 117-149 pass them to unchanged release-build/package targets and use the same VERSION as prerelease tag/name. Inspected Makefile linker flags and domain/CLI version consumers. Ran the actual Makefile in a disposable source copy at `/tmp/gen19-behavior-0nvr6t48`: linux/amd64 and windows/amd64 release-build both succeeded. The host binary's `version --json` exactly returned version `v2.1.0-rc.114+036d808`, commit `036d808`, build_date `2026-11-03T12:00:00Z`. The Linux archive with that version in its name contained the expected binary, README.md and LICENSE. The probe used copied existing embedded assets and did not rebuild frontend assets.

5. Preservation: inspected unchanged stable workflow, release configuration/manifest, build/package/verify consumers and pruning shell. Stable release policy and historical releases are not mutated by this change. Prerelease classification, asset patterns and existing keep-three retention remain intact. Source-change detection still uses tags only to decide whether to build, not as version authority.

## Nonblocking observation

Severity 2/10; classification: introduced dependency support mismatch; confidence high for declared mismatch, no observed runtime failure. `scripts/release/package-lock.json` resolves `@google-automations/git-file-utils@3.2.1`, whose engines.node declares `>=22`, while `.github/workflows/release-candidate.yml:87` uses Node 20. The full resolver fixture suite and real current-repository resolution both succeeded on Node v20.20.2, including real changed-file backfill. Therefore this review does not claim CI currently fails. Aligning the supported Node declaration/dependency graph would remove the unsupported-version maintenance risk; this is not a demonstrated blocking behavior defect.

## Verification and limits

- Executed `node --test scripts/release/*.test.cjs` on Node 24.22.1: 13 passed, zero skipped.
- Executed `python3 scripts/tests/test_release_revision.py`: 8 passed, zero skipped.
- Executed `npm exec --yes --package=node@20 -- node --test scripts/release/rc-version.test.cjs`: 13 passed, zero skipped; separately confirmed v20.20.2.
- Executed live read-only production resolution on Node v20.20.2 as described above: ready / 2.1.0.
- Executed disposable `make release-build PLATFORMS="linux/amd64 windows/amd64" VERSION=v2.1.0-rc.114+036d808 GIT_COMMIT=036d808 BUILD_DATE=2026-11-03T12:00:00Z`: passed, with exact runtime metadata assertion.
- Disposable `make release-package` produced the Linux tar but stopped at Windows ZIP creation with `/bin/sh: zip: Permission denied`; `release-verify` was consequently not executed by this probe. This is an environment limitation, not an introduced source regression. No successful ZIP/checksum verification is claimed.
- Reused the brief's successful make build evidence only for its stated frontend embedding/build boundary. Full integrated make remains the parent's delivery gate.
- Hosted workflow execution, tag creation, upload/download and historical retention deletion were not exercised. The final remote read is best effort; a stable release/main advancement after that read is not made atomic by this implementation, as the accepted brief explicitly acknowledges.
