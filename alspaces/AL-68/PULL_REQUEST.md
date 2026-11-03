# chore(release)!: declare AL-68 configuration compatibility break

Declare the breaking configuration requirement delivered by [AL-68 / PR #93](https://github.com/NiltonDuarte/AgenticLens/pull/93), whose original merge commit did not include a breaking-change marker. This documentation-only follow-up supplies the conventional commit metadata for Release Please to propose 3.0.0 from the recorded 2.0.0 release.

BREAKING CHANGE: AL-68 requires an explicit, valid telemetry.capture.state_dir in existing configuration; operational commands reject configurations that omit it. Use init --update-capture-config to insert the home-state default when the key is absent. Existing invalid values require correction, and relocating existing capture state requires an explicit migration.

The ! in this PR title and this footer must be retained in the eventual squash commit. The release manifest and changelog remain owned by Release Please. This PR does not publish or merge a release.

Scope: [iteration 05 plan](alspaces/AL-68/iterations/05/PLAN.md), following [AL-68's approved configuration requirements](alspaces/AL-68/iterations/04/PLAN.md).
