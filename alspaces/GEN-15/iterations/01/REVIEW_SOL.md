# GEN-15 installation/distribution review: SOL-01

Verdict: CHANGES_REQUESTED.

## Identity and scope

- Phase: independent discovery, iteration 01. Requested policy profile: `sol_6_1_high` (`gpt-6.1-sol`, high). Actual runtime model/effort metadata and usage were unavailable to this reviewer; the requested identity is not an attestation of runtime identity.
- Candidate: `d7a07eb`; base: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`. Reviewed that diff, frozen contract/plan, reviewer/shared persona, review skill, technical guidelines and relevant regression checklist.
- Narrow approved coverage: installation/launch behavior, failure preservation, release dependency closure, Git/config ownership, legacy/source build compatibility and test effectiveness. No product/test edits, publishing, telemetry or subagents. Initial working tree was clean; parent subsequently updated SUMMARY.md and PULL_REQUEST.md and those changes were preserved. Only this report is reviewer-owned.
- Did not read sibling review reports or personas. Parent owns consolidation and delivery.

## Blocking findings

### SOL-01-F1 — P2 / medium: the recommended repair falsely succeeds for a non-executable binary

Introduced regression in the new launcher workflow. Locations: `.agents/skills/agentic-lens-operator/scripts/agentic-lens.sh:24` and `.agents/skills/agentic-lens-setup/scripts/download.sh:249` (also local-prebuilt skip at line 230).

Trigger: a native executable exists as a regular file but lacks executable permission. This can follow copying/restoring an executable or owner-selected Git mode. The launcher explicitly recognizes this state and prints a download.sh repair command. That command considers any regular file already installed, skips it, reports successful installation and exits zero. The next ordinary invocation fails identically. The retained legacy wrapper instead attempts chmod before execution, so its behavior does not rescue the new direct-launch path.

Disposable production-script reproduction on Linux amd64:

1. Install the bundle with `install.sh --source-dir <checkout> --version v9.8.7 --dest <consumer>` and use `download.sh --local-bin <executable-fixture>`.
2. Set the consumer native file mode to 0644.
3. Invoke the launcher: exit 1, diagnostic says `missing or not executable` and recommends download.sh.
4. Run that recommended downloader without force: exit 0, `already exists ... (skipping)` and `Successfully installed AgenticLens wrapper`.
5. Invoke the launcher again: exit 1. No network requests were made.

Required correction: make the recommendation repair this diagnosed state, or fail clearly with the actual permission-repair action. Do not report a usable installation solely because a non-executable regular file exists. Preserve bytes and owner-controlled tracking policy. Add a coherent regression to the existing focused suite covering diagnosis, repair outcome and subsequent execution.

### SOL-01-F2 — P2 / medium: missing release metadata silently unpins worktree repair

Introduced gap in the new pinning/worktree contract. Location: `.agents/skills/agentic-lens-setup/scripts/download.sh:18-22`; related launcher repair selection at `.agents/skills/agentic-lens-operator/scripts/agentic-lens.sh:27` and generated marker in setup `scripts/install.sh:78`.

Trigger: the owner tracks setup/operator scripts but leaves generated `agentic-lens-setup/RELEASE_VERSION` untracked or ignored. A new worktree then contains all runtime scripts, but lacks that marker and native binaries. This is compatible with the requirement that every file's Git policy belongs to the owner. The launcher recommends ordinary local repair; download.sh silently defaults to `latest`, mixing scripts installed from a selected release with whichever binary is current when repaired. There is no missing-pin diagnostic, explicit version choice or bootstrap requirement.

Disposable actual Git-worktree reproduction:

1. Install bundle at v9.8.7 into a consumer repository.
2. Owner ignores `agentic-lens-setup/RELEASE_VERSION` and `agentic-lens-operator/bin/`, commits `.agents` and the chosen ignore file, then creates a second worktree.
3. Confirm download.sh exists and RELEASE_VERSION does not.
4. Execute that worktree's download.sh with controlled online curl/wget transport. It exits 0; the sole requested URL is `https://github.com/NiltonDuarte/AgenticLens/releases/latest/download/agentic-lens-linux-amd64`, rather than a v9.8.7 URL. Shared configuration remains unchanged.

The controlled transport proves URL selection, not availability or content of any real release. The same missing-marker state is directly observable with `download.sh --dry-run`: `Version: latest` and a moving latest-download URL.

Required correction: diagnose missing installation pinning information and require an explicit release/bootstrap choice before repairing an adopted installation. Keep Git policy with the owner; do not force tracking the marker. Retain the deliberately supported source-project entrypoint behavior through an explicit distinction if needed. Test a real worktree missing only the release marker and ensure it cannot silently choose a newer release.

## Hypotheses investigated and evidence

- Complete dependency closure: installer lists both manifests, operator references, setup installer/downloader/bundled legacy wrapper and external-to-bin launcher, plus generated release marker. No runtime source-tree dependency is needed after complete install. The contributor downloader explicitly delegates to the bundled implementation; documentation identifies that it is not standalone. Missing-pin partial closure is F2.
- Download failure preservation: temporary-file download and move preserve existing native file bytes on failed transport. Installer stages and preflights all skill files before copying, rejects differing/symlink targets and does not silently replace edits. No substantiated additional failure-preservation finding.
- Ordinary operations: direct launcher contains no downloader or background updater call; exec preserves arguments, current directory and status. Legacy wrapper retains established automatic-fetch/update behavior, called out explicitly in current guidance. Native file permission recovery is F1.
- Git/configuration ownership: runtime installation scripts contain no Git commands, ignore writes, init or config-copy operations. Tests exercise owner-selected ignored binaries and a repository without ignore rules, preserving index and shared config. Guidance separates absent machine-local config from missing executables and explains explicit source/path overrides.
- Source build/legacy compatibility: unchanged Make build still copies scripts/agentic-lens-wrapper.sh into bin and compiles platform binaries there. Bundled legacy wrapper remains identical to source wrapper. Custom destination legacy execution is exercised by the focused suite.
- Plan versus intent: frozen boundaries and current approval preserve owner-controlled tracking, explicit per-worktree repair, no config inheritance, release-only adoption and legacy compatibility. Earlier paragraphs retaining the words approval pending are historical clarification wording superseded by the explicit current approval; no implementation scope contradiction identified. F2 exposes an omitted partial-installation case, rather than justification to require tracking.
- Test oracle: inspected all nine installation cases and transport/prebuilt helpers. They execute actual production scripts and a real Git worktree; assertions are behaviorally meaningful. Existing cases always keep the release marker with tracked skills and use executable native fixtures, so neither F1 nor F2 is caught. The transport intentionally substitutes source files/binary fixtures, and cannot establish live release existence, platform execution or checksums. Tester's disposable baseline-downloader mutation fails the owner-policy assertion, supporting that oracle.

## Verification and limits

Executed two bounded disposable Python probes using the focused suite fixture helpers, importing with importlib. The first established F1 and missing-pin dry-run behavior. The second created a real consumer Git worktree and established F2's requested URL. Temporary consumer trees were cleaned; no product/test files were modified.

Reused tester report: nine focused installation tests and six release-workflow tests passed, no skips; baseline-downloader mutation detects installer-imposed ignore policy. Parent reports integrated make passed candidate d7a07eb: installation/release tests, Go vet/race tests, frontend typecheck/lint/tests (35 files, 781 tests) and build. No duplicated full suite was needed to establish the two uncovered states.

Host probes ran on Linux amd64 with deterministic prebuilt and transport fixtures. No live release download/checksum verification, macOS/Windows/native PowerShell execution or other-CPU execution was performed. No claim is made about real telemetry extraction or source hydration. These acknowledged deployment limits do not explain away the reproduced local contract failures.
