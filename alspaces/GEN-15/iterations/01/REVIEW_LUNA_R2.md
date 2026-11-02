# GEN-15 installation/distribution re-review: LUNA-02

Verdict: APPROVED (scoped).

## Identity and scope

- Run: `LUNA-02`, revision-aware correction pass, requested profile `luna_6_medium` (`gpt-6-luna`, medium). Runtime model metadata and usage are unavailable; this records the requested profile, not runtime attestation.
- Candidate: `2d540946c01851face786f15894b1e00b2451b19`; base: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`; initial reviewed candidate: `d7a07eb`.
- Reviewed the accepted contract and current plan, candidate diff and relevant unchanged launcher/build paths, installation tests and tester evidence. Scope: installation/repair/launch behavior, error preservation, release/dependency ownership, Git/config ownership, legacy/source-build compatibility, and test effectiveness. No domain, ledger, or UI changes were found in the candidate scope.
- Product and test files were read-only. Only this report is reviewer-owned. The pre-existing modified `REVIEW_RUNS.json` was preserved.

## Findings

No blocking findings in this scoped pass.

## Initial finding hypotheses and dispositions

- **SOL-01-F1: non-executable binary repair false-success — corrected.** The downloader now checks the executable bit on an existing platform binary and runs `chmod +x`, failing if the permission cannot be restored. Both the normal remote-repair path and local-prebuilt path use this retention check when replacement is not requested. The focused test makes the binary mode 0644, executes the launcher's printed repair command (including local-prebuilt override), then asserts executable permission, unchanged bytes, successful subsequent launch, and no transport requests. This addresses the observed failure without replacing the user's binary.
- **SOL-01-F2: absent release marker silently selecting latest — corrected.** The downloader resolves an installed `RELEASE_VERSION` or explicit `AGENTIC_LENS_VERSION`/`--version`; it rejects unpinned remote repair and dry-run before creating the destination or contacting transport. The launcher directs the user to supply an explicit tag when the marker is absent. The added real-worktree test covers an ignored marker and binaries, confirms no request/directory for unpinned repair, confirms offline `--local-bin` remains usable without inventing a pin, and confirms explicit-version repair requests only that release. The contributor source-project entrypoint separately retains its documented latest default when no installed marker is present.
- **LUNA-01 plan approval wording — corrected in the current plan.** The current plan records the user's approval of the corrected contract and says implementation is in progress. The contract's completion gate remains a delivery criterion, not an assertion that review/integration are already complete. No stale pending-approval sentence remains in the current plan text reviewed here.

## Independent behavior and dependency checks

- The setup installer stages the full setup/operator bundle, including its installer, downloader, bundled legacy wrapper, release marker, operator references, and external-to-bin launcher; it preflights existing files before copying and refuses differing files. This avoids partial overwrite of locally edited skill files. The installed launcher resolves a worktree-local binary, uses `exec` with quoted `"$@"`, and does not download. Missing setup assets yield a bootstrap instruction; the downloader uses bundled dependencies only and does not fetch moving `main` assets.
- Pinned installation resolves `latest` once and stores the resolved release marker; subsequent downloaded setup assets and repair binaries use that release. Explicit versions remain explicit. The missing-marker branch prevents accidental silent upgrade while preserving the tested source-project latest behavior.
- Installer and repair scripts do not change `.gitignore`, stage files, or copy local configuration. The tests assert owner-selected Git policy and unchanged index/configuration. The launcher preserves cwd, argument boundaries and exit status through `exec`; custom binary destination and the legacy command are exercised. Existing Make wrapper-copy behavior remains present in the candidate diff.
- Test assertions use actual production shell scripts, a deterministic executable fixture, controlled transport, consumer repositories, and a real secondary Git worktree. They test branch behavior and observable effects rather than documentation strings. The fixture does not establish release availability or binary provenance.

## Verification and limits

- Executed on the reviewed candidate: `python3 scripts/tests/test_installation.py` — 12 passed, no skips.
- Executed: `bash -n` on setup `download.sh` and `install.sh`, and `sh -n` on the operator launcher — passed.
- Reused the tester's candidate evidence: `scripts/tests/test_release_revision.py` — 6 passed; `git diff --check` passed. The parent reports the integrated `make` gate is being rerun for this candidate; this review does not treat that in-progress run as a pass.
- Host evidence is Linux amd64. macOS, Windows/Git Bash/native PowerShell, and alternate CPU execution were not performed. Dry-run artifact names are not native-platform execution evidence. No live GitHub release download/checksum verification or telemetry hydration is claimed.

This approval covers only the installation/distribution focus and this exact candidate. It is not a whole-repository or final delivery approval; the parent retains consolidation and integration gates.
