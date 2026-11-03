# FIX-16: Restore release candidate publishing

Status: Approved. Explicit scope approval received before coder dispatch.

## Problem and grounded cause

The user reports that release candidates stopped publishing after GEN-19. At baseline dc44619ae15586fe1194f872743033f523635c92, GitHub Release Candidate run 41733288460 successfully computes v3.0.0-rc.122+dc44619, then fails during Static Analysis in make check. The failure is scripts/tests/test_installation.py, InstallationTest.test_source_build_verifies_new_native_without_changing_selected_release_cache: the newly built binary correctly reports the inherited release VERSION, while the test expects git describe's v2.0.0-rc.115+65e0eeb-7-gdc44619.

.github/workflows/release-candidate.yml exports VERSION, GIT_COMMIT and BUILD_DATE before its checks. Makefile intentionally uses conditional defaults for these values. The installation test copies the process environment through DistributionFixture and invokes make verify-build without controlling VERSION, then assumes the default is selected. This is an environment-dependent test oracle, not evidence of failed Release Please version selection. The stable-release workflow computes its metadata after checks and does not encounter this same ordering by construction.

## Accepted behavior proposed for approval

Make the existing source-build installation test independent of ambient release metadata while retaining its real native build, native execution, release-cache immutability and selected-release launcher assertions. Exercise both the ordinary git-derived source version and an explicitly supplied candidate version through the actual make verify-build path. The latter case must prove that release metadata is still honored, rather than suppressing a supported override to make checks pass.

The full check gate must succeed with valid candidate VERSION, GIT_COMMIT and BUILD_DATE exported, as well as in an ordinary development environment. Preserve GEN-19's release-PR version authority, candidate SHA, suffix, exact stable-target guard, publication and retention behavior and all existing quality gates.

## Implementation boundaries

Limit implementation to scripts/tests/test_installation.py and, only if needed for coherent fixture isolation, scripts/tests/distribution_fixture.py. Prefer local environment control in the owning test over global fixture changes. Ticket artifacts belong exclusively to alspaces/FIX-16. No workflow or Makefile change is currently necessary. A demonstrated separate publishing defect requires a recorded scope delta before its implementation.

Go domain, adapters, CLI, frontend, schemas, production assets and stable release policy are outside this correction. Preserve FCIS, strict typing and zero-any requirements where applicable, zero emojis, unrelated checkout changes and the existing installed-release cache contract. Do not merge release PRs or publish a stable release.

## Completion and evidence boundaries

Retain the observed hosted failure as the baseline reproduction. Verify the corrected existing test and relevant release/distribution suites with candidate metadata present and with its variables absent. Run the required make gate after integration and inspect embedded assets and native execution through existing canonical targets. Tests must use temporary fixtures, never a ticket workspace as a fixture.

Successful local checks establish removal of the observed check-stage blocker. They do not establish hosted GitHub publication, uploaded downloadable assets or successful downstream installation. Record any hosted boundary not exercised; do not call the pipeline remotely verified without an actual successful run and published assets. Parent owns review consolidation, integration, telemetry, commits and PR delivery.
