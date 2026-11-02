# GEN-15 installation boundary verification

## Scope and test ownership

Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`; verification targets the working tree installation scripts and the new `scripts/tests/test_installation.py`. Tester owns only that suite and this report. No application, runtime script, Makefile, telemetry, Git staging, or publication changes were made by the tester.

The focused suite executes production installation, repair, and launch scripts in disposable adopting repositories. A deterministic prebuilt executable fixture and controlled curl/wget transports exercise the shell boundary without external downloads. The fixture binary is not evidence of a real release or native compiled-binary behavior.

## Executed verification

- `python3 scripts/tests/test_installation.py`: twelve tests passed after review corrections; no skips. The initial nine passed before review. Cases cover the full installed setup/operator dependency closure, repository-owner Git tracking/ignore policy, unchanged Git index and shared configuration, a real second Git worktree, missing binary diagnosis without network calls, explicit repair in paths containing spaces, argument/current-directory/exit-status propagation, a worktree missing setup assets and its bootstrap path, pinned remote file and binary URLs, one-time latest-release resolution, failed download preservation and temporary-file cleanup, offline installed execution, conflicting skill preflight, custom binary destinations, legacy wrapper invocation, explicit replacement, dry-run platform artifact names, and missing option values.
- `python3 scripts/tests/test_release_revision.py`: six existing release-workflow tests passed; no skips.
- `git diff --check`: passed during this test pass.

Regression proof: in a disposable source bundle only, replaced the new downloader with the baseline downloader from `git show 7bc670a:.agents/skills/agentic-lens-setup/scripts/download.sh`. Running `test_install_and_repair_preserve_owner_git_policy_and_configuration` then failed specifically because the baseline downloader created a `.gitignore` in the adopting repository. The mutation was outside the checkout and discarded; the production suite remains unchanged.

## Review correction regressions

Three coherent cases were added for SOL-01-F1 and SOL-01-F2 and their compatibility boundary:

- A native file changed to mode 0644 makes the launcher refuse execution. The test parses and executes the launcher's exact printed repair command, with both ordinary remote-repair selection and an explicit local-prebuilt override. It requires successful subsequent launch, restored execute permission, unchanged file bytes, and zero transport requests.
- A real second worktree inherits scripts while the owner ignores only the generated release marker and binaries. Default remote repair and dry-run must fail with explicit version guidance before network or directory creation. Explicit version repair must select only the chosen pinned URL. An offline local-prebuilt repair remains available without a marker or network; no marker is created or forced into Git.
- The contributor-facing script is copied with its production dependencies to a disposable source tree without a release marker. Its established latest default still works for dry-run and explicit binary download. This isolates the source entrypoint compatibility case from installed-repository missing-pin behavior and from a developer's live generated files.

Before/after proof: archived production scripts from candidate `d7a07eb` into a disposable tree and ran the two new finding regressions against that snapshot. Both detected their corresponding defect: execute permission remained missing after a reported-success repair, and a markerless worktree selected moving latest instead of refusing an unspecified release. No checkout files were restored. Both regressions and the compatibility case pass with the corrected production scripts. The complete twelve-case installation suite and six existing release tests were rerun successfully; `git diff --check` passed.

## Evidence limits and integration handoff

Host execution is Linux amd64. macOS, Windows, and alternate CPU execution were not performed. The dry-run assertions verify supported artifact naming, not execution on those platforms. No live GitHub release availability, release checksums, native PowerShell execution, or actual telemetry-source hydration is claimed. Machine-local configuration safety is verified as absence/preservation and argument forwarding; the fixture does not validate CLI telemetry semantics.

The parent owns the required full integrated `make` gate after integration, final review attribution, source revision/fingerprint, telemetry, commits, and PR delivery. Tests do not read the live ticket workspace or developer source-session stores.
