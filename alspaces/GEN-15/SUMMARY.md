# GEN-15: Portable setup and worktree operation

## Request and baseline

- Source: user reports that adoption installs the operator but not setup itself, and ignored binaries disappear in worktrees; asks how to solve it.
- Allocated GEN-15 after checking existing ticket workspaces, the ticket index, and local/remote branch references (including GEN-14).
- Initial Git status: clean. Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Branch: `chore/GEN-15-worktree-installation`.

## Grounded findings

- Setup installs the operator manifest and three references, but does not install its own manifest and downloader.
- Setup installs the wrapper alongside native executables in the ignored operator `bin/` directory. A committed adopting repository therefore does not carry that entrypoint into another worktree.
- The wrapper can fetch a missing native binary only when the wrapper itself exists.
- The operator CLI reference treats an absent executable as requiring setup, but the installed repository may lack that setup skill.
- Downloader defaults and wrapper source fallback still assume setup/source-checkout paths in places; adoption must be verified without the AgenticLens source tree.

## Status and verification

User clarification incorporated into CONTRACT.md and iterations/01/PLAN.md: the adopting repository owner decides whether to track or ignore any installed files. Superseded the tracked-launcher and mandatory-ignored-binary proposal. The launcher remains outside binary storage; setup/repair must preserve Git policy and diagnose missing setup assets as well as binaries. Regression planning now covers owner-selected tracking policies and a worktree without inherited repair scripts. This clarification does not establish implementation approval.

The user explicitly approved the corrected contract and plan with "approved, follow with the implementation." Implementation is in progress through coder and tester roles. No telemetry generated: ingestion remains deferred to final delivery.

Dependency preparation: `make install` passed. npm reported existing dependency audit advisories; dependency versions and lockfile were not changed by this ticket. Runtime, review, and integrated verification evidence will be recorded below before delivery.

## Implementation

- Added complete setup/operator skill bundle installation, including setup's own installer, downloader and bundled wrapper. A selected release is recorded in RELEASE_VERSION and used by repair.
- Added an offline launcher outside bin storage. It preserves caller directory, arguments and exit status, and prints explicit binary repair or skill bootstrap guidance for missing files.
- Binary repair uses worktree-local operator/bin by default, supports explicit destinations/prebuilt binaries, retains existing executables unless replacement is requested, and preserves an existing executable on download failure.
- Removed automatic ignore-file creation. Neither installation path stages files, commits files, nor imposes Git policy. Existing consumer configuration remains untouched.
- Consolidated contributor downloads onto the bundled implementation and aligned setup, operator, README and manual guidance. Legacy wrappers retain their established behavior.

## Verification

- Candidate: d7a07eb40f1a66b98da7f6bd9252e5a8c434f1ad; main integration was already up to date.
- Nine installation boundary tests and six existing release tests passed. Tests use real disposable Git worktrees and controlled transport/prebuilt fixtures, with no silent skips.
- Restoring the baseline downloader in a disposable bundle makes the Git-policy preservation test fail on its unwanted ignore-file creation.
- Parent `make backend-check` passed, including Go vet, race tests and canonical frontend embedding preparation.
- Integrated `make` passed candidate d7a07eb: six release tests, nine installation tests, Go vet/race tests, frontend typechecking and 35 test files / 781 tests, canonical embedded assets and platform builds.
- First-pass review: Sol 6.1/high found two confirmed P2 recovery defects (non-executable retained files and missing release markers); Luna 6/medium found a stale approval sentence. Both runtime defects and the sentence were corrected. Twelve installation tests now pass; final integrated verification and two-profile re-review target the corrected candidate. Detailed attribution remains in iterations/01/REVIEW_RUNS.json and FINDINGS.json.
- Host evidence is Linux amd64. macOS/Windows native execution and live release asset availability were not verified. Skill bootstrap requires a release containing the new installer; publication of this PR does not itself publish such a release.
