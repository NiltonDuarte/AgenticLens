# AL-68 build/release discovery review

Verdict: CHANGES_REQUESTED. Scope: installation, distribution, upgrade, rollback and removal surfaces only. Codex runtime evidence also remains incomplete; this review does not grant a complete-support claim.

## Run and revision

- Run ID: release-sol-01; phase: independent specialist discovery.
- Requested configuration: codex / gpt-6.1-sol / high. These are requested settings, not independently observed runtime model/effort metadata. Actual model, effort, usage and cost were not exposed to this reviewer and are not inferred.
- Baseline HEAD: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Candidate: `c7dade3313a1d65a4269d595c782750655be8a044c020dc1f3ac502464cf325a`, defined by CANDIDATE.json as sorted tracked/nonignored untracked regular files outside alspaces, path/NUL/content/NUL. Independently recomputed and matched before probes. This is a working-tree candidate, not a commit SHA.
- Initial status contained the parent's existing AL-68 staged/unstaged/untracked changes. Product, tests and sibling reports remained read-only. Only this assigned report was written. No staging, commits, telemetry, trust changes, publishing or release network actions occurred.
- Probe native binary: `bin/agentic-lens-linux-amd64`, SHA256 `5ca77acea605841fccd518e3f854d52be4ae52666af9b6c18b863677c39b20e7`. Disposable fixtures: `/tmp/al68-release-sol-qm_mm32w`.

## Findings

### RLS-1: P1 - Release replacement can strand active hooks on an incompatible handler

Locations: `scripts/download-agentic-lens.sh:328`, `scripts/download-agentic-lens.sh:374`, its packaged copy under `.agents/skills/agentic-lens-setup/scripts/download.sh`, and `scripts/agentic-lens-wrapper.sh:90` / `:163` plus its packaged copy. Compatibility is checked only later in `backend/internal/adapters/capture/install.go:153`, after the downloader has already replaced the old executable; wrapper update performs no capture protocol check at all.

Trigger: a repository already has an active lifecycle installation, then an explicitly selected release has no matching capture handler/protocol. Both downloader `--force --version TAG --capture-source ...` and wrapper `update --version TAG` are advertised operations. The same failure is reachable through downloader `--local-bin`.

Observed consequence: the executable at the pinned hook path is replaced before compatibility validation. Both downloaders then fail while retaining the old active hook configuration and `active: true` manifest. The former compatible executable is lost; subsequent capture calls fail. Wrapper update reports success in this state. Advising status/install afterward cannot preserve the working installation and does not roll back the failed replacement.

Evidence: actual scripts and actual candidate handler first installed hooks in temporary repositories. A local transport fixture supplied a legacy executable that supports version output but rejects capture; no external download occurred. For each downloader, the real `--force --version v-fixture` branch exited 1, printed `Capture setup failed`, preserved hook bytes and `active: true`, while the installed handler's `capture protocol` exited 2. The local-bin branch reproduced the same state for both copies. Wrapper `update --version v-fixture` exited 0 and printed `Successfully updated 1 binary(ies)` while leaving unchanged active hooks and a handler whose protocol command exited 2.

Required correction: stage and validate the replacement against active capture requirements before replacing the working handler, and preserve/restore the prior executable on setup failure. Cover the downloaders and wrapper update together, including explicit older releases. Preserve user hook/plugin edits throughout. Add a regression that starts with a working active installation; the current failed-download/fresh-install tests do not reach this transition. Binary replacement existed before AL-68; the new active lifecycle installation and its approved upgrade/rollback contract make this an unresolved integration defect in this delivery.

### RLS-2: P2 - The documented manual Unix/PowerShell setup path omits lifecycle installation

Locations: `docs/SETUP_MANUAL.md:81` through the manual setup steps, especially `:113` and `:116`; entrypoint `README.md:13` identifies this manual as the complete setup/manual installation guide. Its agentic setup sequence also lacks the new lifecycle step.

Trigger: an adopter follows the documented direct manual path, including its PowerShell commands, rather than the setup skill.

Consequence: the complete advertised procedure ends after binary retrieval, init and dashboard startup. It never invokes native capture installation, explains binary-only readiness, chooses capture state/source identity, or links to binding/trust/removal instructions. Consequently that installation surface never configures automatic session capture. The setup skill's newly added PowerShell example does not update or redirect the separate public manual path.

Evidence: read the unchanged manual, README entrypoint, updated setup skill and both downloader help/implementation surfaces. The manual contains no lifecycle capture installation section. This violates the explicit contract/plan requirement to keep the documented direct and PowerShell paths consistent, rather than a demand for duplicated prose.

Required correction: update the public manual with the binary-versus-lifecycle distinction and a canonical link or concise native-handler setup sequence covering Unix/PowerShell, explicit paths, approval/trust, actual binding verification and owned removal. Keep the detailed operational instructions at their owning source.

## Hypotheses and verification

- Native installer post-write rollback: PASS in an independent failure-injection probe. Existing user Codex hooks were written into a temporary repository; the state directory was readable but not writable. Native install wrote hook configuration and its ignore entry, then failed creating the manifest temporary file. It returned an error, restored the original hook bytes exactly, removed the newly added .gitignore and left no manifest. This fills the post-write rollback gap identified in TEST_REPORT.md for this exercised transaction. Created empty directories are not asserted to be removed.
- Idempotence, edited hook/plugin preservation, disable/uninstall and missing-handler preflight: source traced through `capture.Install`, its manifest ownership comparisons and reverse application. Reused actual CLI tests recorded in TEST_REPORT.md and distribution checks in SETUP_IMPLEMENTATION.md; did not rerun equivalent suites without a new concern. Removal retains local state by documented design.
- Callback release independence: native hooks point directly at the executable; the OpenCode template is embedded with `go:embed plugin.ts`. Manual wrapper capture dispatch occurs before update/download checks. Runtime callback argument values are not interpolated into executable shell expressions; native static path arguments are quoted, and the plugin uses an argument-vector child process. The real OpenCode plugin execution is supported by SMOKE.md, separately from fixture results.
- Distribution consumers: inspected Makefile build/release targets, stable/candidate workflows, setup/manual entrypoints, both scripts/wrappers and embedded frontend prerequisite. Release publication includes platform native executables; the plugin needs no extra distributed file or new adopter toolchain. Existing wrapper/direct installer retrieval and release publishing were not executed against GitHub.
- Windows compile: PASS, independently ran `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /tmp/al68-release-sol-qm_mm32w/agentic-lens-windows-amd64.exe ./cmd/agentic-lens` from backend. This establishes compilation with embedded assets, not Windows execution or hook command correctness.
- Reused parent `make` PASS from VERIFICATION.md at the exact frozen candidate: Go vet/race suite, frontend checks/tests/build, release/distribution tests and Linux/Darwin builds. No repeated full make was needed for these read-only probes. No release-package/release-verify run or genuine release download is claimed.

## Evidence limits and remaining gates

Read the contract, approved plan and ticket specification, including the required actual hook-to-agent interaction and prohibition on inferred ownership. Read raw TEST_REPORT.md, SMOKE.md, SETUP_IMPLEMENTATION.md and VERIFICATION.md. Did not read sibling discovery reports or use their conclusions to create these findings.

SMOKE.md records actual OpenCode 1.19.5 Linux CLI independent roots and a native child, binding responses and source reconciliation. That is reused author-supplied runtime evidence, not an independent harness run by this reviewer. It explicitly leaves Codex trust/callback execution, native Codex child and independent-root capture unverified. No trust bypass or new permission request was attempted here. Codex support cannot pass the approved real-runtime acceptance boundary on version/features/configuration observations alone.

No Windows/macOS native runtime, T3 Codex child path, OpenCode v2 runtime, other harness version, resumed real-session interaction or interrupted real hook recovery was established. The Windows cross-build and synthetic release/rollback probes do not establish those boundaries. Resolve the two findings and obtain the outstanding required runtime evidence, or an explicitly approved revised acceptance boundary, before delivery approval.
