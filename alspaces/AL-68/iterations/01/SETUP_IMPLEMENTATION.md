# AL-68 setup/distribution implementation

Baseline: 7bc670a7f5c9e93a220f3596f3d0e05cb6007332. Shared checkout; only the assigned distribution/setup paths and this report were changed. No commits, staging, final telemetry, trust changes or external publishing were performed.

## Delivered behavior

- Both downloaders accept opt-in `--capture-source`, require explicit absolute `--capture-store`, and accept repository/state overrides. Successful binary setup delegates to the selected host-native handler. Binary-only installs explicitly state that lifecycle setup was not performed. Dry run uses an available native handler for its own dry run, or explicitly reports the absent file preview.
- Wrapper capture dispatch is local: no background update, missing-binary download or self-heal. An executable wrapper accidentally supplied as a native artifact fails visibly instead of recursively dispatching to itself. Wrapper update advises a reviewed native capture setup/status check.
- Setup and operator references document actual CLI flags, per-session agent request/response, runtime-supplied OpenCode session identity, challenge binding, no heuristic attribution, local state/privacy, final digest-bound report, trust/degraded limits and owned recovery/removal. The proposed adopting AGENTS block is limited to binding and final publication.
- `make check` includes `test-capture-distribution`, which exercises both shipped script entrypoints.

## Verification

Owned-content fingerprint (SHA-256, sorted path/NUL/content/NUL): `94d072244e66292755c7aa8ec467700bac2fc7e709a4276d1b23695dddbe25a1`.

- `make test-capture-distribution`: PASS, eight tests. Real scripts execute in temporary repositories. Recording native fixtures check exact argument delegation, paths with spaces/shell-like text, explicit source validation, dry-run preservation, native setup failure, failed artifact download, advertised platform artifact plans, mistaken wrapper-as-native recursion rejection and network-free wrapper capture. Fixtures establish command/installer boundaries, not real agent/harness capture.
- `bash -n` on both downloader copies and `sh -n` on both wrapper copies: PASS.
- `git diff --check`: PASS at this handoff.
- Additional disposable integration probe against the actual newly built `bin/agentic-lens-linux-amd64` (SHA-256 `330d7fc342052c333c981ee8efa3af6bffcd50d8d2719eb0a849db7bcaee9c34`): PASS for each downloader. Existing Codex user hook preserved after install; absent-runtime status remained degraded; owned disable restored the original hook entries. Source store and repository were self-contained temporary directories with spaces. No real source sessions or trust settings were accessed/changed.
- A first probe accidentally selected `bin/agentic-lens`, which is a wrapper. It was terminated in its disposable environment; that failure motivated the recursion guard and regression test. The successful native probe above used the actual platform executable.

## Limits

No release network download or Windows/macOS native execution was performed. Advertised platform tests verify downloader artifact selection only. No hook callback, source session capture, model usage, parent/child association, trust approval or actual agent-binding exchange is proven by these script/setup checks. Those remain primary implementation/test smoke responsibilities and must be reported honestly. Full integrated `make` and independent review remain parent-owned.

## Owned paths

- `Makefile`
- `scripts/download-agentic-lens.sh`
- `.agents/skills/agentic-lens-setup/scripts/download.sh`
- `scripts/agentic-lens-wrapper.sh`
- `.agents/skills/agentic-lens-setup/scripts/agentic-lens-wrapper.sh`
- `.agents/skills/agentic-lens-setup/SKILL.md`
- `.agents/skills/agentic-lens-operator/references/cli.md`
- `.agents/skills/agentic-lens-operator/references/configuration.md`
- `scripts/tests/test_capture_distribution.py`

## Discovery corrections: CAP-08 and CAP-09

This entry supersedes the initial script implementation evidence for the corrected paths. Approval and scope remain iteration 01. Owned paths remain the original script/setup paths plus the parent-assigned `docs/SETUP_MANUAL.md`; no backend or Go test edits were made by this worker. Source fingerprint for the ten sorted owned paths above (path/NUL/content/NUL): `e0f3e3efb4bb09868da2108a3fd347bd1eaa3de0fc2621deb6418a2cdacc7aad`.

CAP-08 correction: both downloaders stage local and downloaded native candidates, require protocol1 for capture activation and retain the installed capture-aware handler's reported protocol before any replacement, regardless of custom installation-state locations. Same-directory atomic replacements retain prior native/wrapper backups until setup succeeds. Native setup failure or activation failure restores prior owned binary changes; the native installer separately restores owned configuration. Wrapper update stages candidates and checks the host handler protocol before swapping any working executable; mismatch fails without a success claim. A staged wrapper cannot pass its probe by dispatching through an existing native handler. Existing foreign-platform binaries are preserved explicitly under capture-aware upgrades because their runtime compatibility cannot be verified on this host. New foreign downloads remain binary-only until verified on their native host. No callback networking or trust changes were introduced.

CAP-09 correction: the public manual now separates binary availability from lifecycle readiness and includes native Unix and PowerShell install/status commands with explicit source/state/repository paths, canonical agent-binding/trust/removal links, guarded upgrade/recovery guidance and no inferred ticket/source context. Direct initial-download examples are explicitly not instructions to overwrite an active handler. Setup/operator guidance describes the matching preflight/rollback and native-host limits.

Verification of corrected slice:

- `make test-capture-distribution`: PASS, 14 tests. New production-script regressions cover incompatible local-binary and real download-branch `--force --version v-fixture` replacements, with and without another capture install; subsequent setup failure and wrapper restoration; incompatible/successful wrapper updates; injected activation rename failure with recovery; and preservation of unverifiable existing foreign handlers. These use a local transport and native-protocol executable fixtures, never a release network call or simulated harness evidence.
- Actual newly built Linux native-handler probe through each downloader: PASS. A working protocol1 executable with a distinguishable appended marker first installed real owned hooks in a disposable repository. A user-edited owned hook then forced subsequent native setup to fail after candidate preflight/activation. Each downloader restored the prior native bytes including the marker and preserved the edited hook bytes exactly. No real sessions or trust settings were changed. The appended marker distinguishes rollback from merely copying the same build back over itself.
- `bash -n` both downloaders, `sh -n` both wrappers and `git diff --check`: PASS after the final corrections.

The slice is frozen at this handoff. Parent-owned integrated make, revision-aware focused review and outstanding real harness evidence remain required. The scripts establish compatibility/rollback contracts; they do not establish a real agent/hook exchange or Windows/macOS runtime support. Hard termination that prevents signal cleanup may leave named prior-candidate backups for explicit recovery; no crash-recovery claim is made for a SIGKILL during binary replacement.
