# GEN-15 iteration 01 independent review — LUNA-01

## Assessment

**APPROVED for the assigned installation/distribution scope**, at candidate `d7a07eb40f1a66b98da7f6bd9252e5a8c434f1ad` against baseline `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. Requested profile: `luna_6_medium` (Codex gpt-6-luna/medium); this records the requested identity, not independently verified runtime metadata. Narrow-role exception combines installation behavior, integration, failure preservation, dependency ownership, and test effectiveness. No Go/domain, ledger, extraction, API, or UI change was found in the candidate diff.

## Findings

No blocking introduced installation/distribution defect found in the inspected paths.

Plan record correction: `PLAN.md`'s clarification paragraph says “implementation approval remains pending,” while its status and later “Approval and delivery” section record the user's approval (“approved, follow with the implementation”). This is internally inconsistent and may mislead a later handoff. The final approval record and contract agree that scope was approved; parent should reconcile the stale sentence. This does not change the implementation verdict.

## Failure hypotheses investigated

- **Worktree repair may overwrite shared/local project configuration or impose Git policy.** Traced `install.sh` and `download.sh`: neither runs Git or touches project configuration; the downloader no longer creates `.gitignore`. The isolated suite checks unchanged configuration/index, no generated ignore file, and a second worktree whose owner-selected policy is preserved.
- **Ordinary launch might silently fetch, lose argument boundaries/cwd, or mask the CLI exit code.** `agentic-lens.sh` only resolves a local binary and uses `exec "$binary" "$@"`; absent binaries produce a repair instruction. The test exercises a missing-binary invocation without transport access and then checks spaced arguments, literal shell metacharacters, caller cwd, and exit code 37.
- **Pinned installation dependencies might mix release revisions or silently upgrade.** `install.sh` resolves `latest` once (or accepts an explicit tag), fetches all bundle paths from that tag, records `RELEASE_VERSION`, and preflights differing files before copying. `download.sh` defaults to that recorded version, skips existing binaries unless `--force` is supplied, and uses temporary files before replacement. Controlled-transport tests assert all fetched paths use one release and preserve an existing executable after failed forced replacement.
- **Repair failure might leave a corrupt executable or report success.** Download/local-copy paths stage to temporary files and move only after successful transfer/copy and chmod. The suite checks failed transfer leaves no new binary/temp file and preserves an existing binary. Wrapper replacement is separately gated by absence or explicit `--force`.
- **Legacy command and source build paths might break.** The legacy destination command remains installed by `download.sh`; the fixture invokes it. The root contributor downloader delegates to the bundled implementation, and the Makefile still builds/copies `scripts/agentic-lens-wrapper.sh` as `$(BINARY)`. The bundled and root legacy wrappers are byte-identical in this candidate.
- **Owner tracking choice or custom paths might be assumed.** Script inspection shows no staging/index or ignore manipulation. Tests cover committed scripts with ignored binaries, a worktree with setup assets absent, a custom destination, and a repository path containing spaces.

## Evidence and verification

- Reviewed `CONTRACT.md`, `iterations/01/PLAN.md`, `TECHNICAL_GUIDELINES.md`, relevant downloader/portability/testing sections in `docs/REGRESSION_PREVENTION.md`, the setup/repair skill, operator guidance, both launcher/download flows, legacy wrapper, Makefile, contributor downloader, public docs, and focused installation test.
- Diff reviewed: `7bc670a..d7a07eb`; current candidate commit is `d7a07eb40f1a66b98da7f6bd9252e5a8c434f1ad`.
- Independently ran `python3 scripts/tests/test_installation.py`: 9 tests passed on the Linux host (3.863s).
- `git diff --check 7bc670a..d7a07eb` passed.
- Parent reports the integrated `make` passed at this candidate, including 9 installation tests, 6 release tests, Go vet/race, 35 frontend files / 781 tests, and build. This is parent-reported verification, not rerun independently here.

## Limits

No live release download or native release executable was exercised. Controlled transport and fixture executables establish script behavior only. Review host was Linux; native macOS, Windows PowerShell, and Git Bash behavior remains unverified. The integrated gate passed per the parent, but does not replace those platform/live-release boundaries. These limits do not leave an unresolved defect in the scoped source-level and disposable-fixture investigation, but must remain explicit in the consolidated handoff.
