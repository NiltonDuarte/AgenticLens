# AL-68 iteration 05: major release declaration

The user explicitly requested a dummy PR to declare a major change pointing to AL-68. This is a documentation-only release follow-up to merged PR #93; application behavior and release workflows are outside scope.

Baseline: 122e983; initial Git status clean; branch AL-68/major-version-change. Existing AL-68 workspace and history are retained. The previous PR is merged, so delivery requires a new PR.

Acceptance: publish a normal PR with a conventional chore(release)! title and BREAKING CHANGE footer explaining that existing configurations must explicitly set telemetry.capture.state_dir. Link merged PR #93 and the existing requirements. Release Please should classify the eventual squash commit as a major increment from the recorded 2.0.0 release. Do not edit the manifest or changelog manually, merge the PR, or publish a release.

Verification: check title/footer, references and documentation diff; integrate origin/main and run the required make gate. No new tests are warranted for release metadata documentation. Generate Codex telemetry last, preserving the prior multi-session ledger or explicitly recording extraction limitations.
