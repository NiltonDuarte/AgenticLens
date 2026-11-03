# chore(release)!: declare AL-68 configuration compatibility break

Declare the breaking configuration requirement delivered by [AL-68 / PR #93](https://github.com/NiltonDuarte/AgenticLens/pull/93), whose original merge commit did not include a breaking-change marker. This documentation-only follow-up supplies the conventional commit metadata for Release Please to propose 3.0.0 from the recorded 2.0.0 release.

BREAKING CHANGE: AL-68 requires an explicit, valid telemetry.capture.state_dir in existing configuration; operational commands reject configurations that omit it. Use init --update-capture-config to insert the home-state default when the key is absent. Existing invalid values require correction, and relocating existing capture state requires an explicit migration.

The ! in this PR title and this footer must be retained in the eventual squash commit. The release manifest and changelog remain owned by Release Please. This PR does not publish or merge a release.

Scope: [iteration 05 plan](https://github.com/NiltonDuarte/AgenticLens/blob/AL-68/major-version-change/alspaces/AL-68/iterations/05/PLAN.md), following [AL-68's approved configuration requirements](https://github.com/NiltonDuarte/AgenticLens/blob/main/alspaces/AL-68/iterations/04/PLAN.md).

Verification: release markers/reference and git diff --check pass; Go vet/race, frontend typecheck/lint, 14 release, 10 installation and 15 capture-distribution checks pass. make build passes for five targets. Full make fails twice on the unchanged TicketDeepDive child-navigation test exceeding 5000 ms (783 other frontend tests pass); the same case times out in isolation. No application/test changes are included.

Merge ownership: human_only release marker, supervised by MGT-2. Human review/merge is required; neither this thread nor MGT-2 will merge this PR or Release Please PR #55.

Telemetry limitation: the canonical ledger's provenance guard rejected historical-owner association in this checkout. All 199 prior sessions are preserved unchanged; a separate iteration05 ledger records this follow-up's verified Codex root. No historical provenance recovery is claimed.
