# GEN-15 iteration 01: Worktree-safe installation and repair

Status: Approved by the user: "approved, follow with the implementation." Implementation in progress under the corrected scope.

## User request and interpretation

Source: initial user request, asking how to solve setup not installing itself and ignored binaries disappearing in worktrees. This is the first scope proposal, with no superseded implementation decision. Preserve existing release-only adoption and approval policies. Verification must exercise an actual disposable Git worktree, rather than checking manifest strings alone.

Clarification source: the user stated, "This is on user, not on the platform to decide how to track the binary file," and then asked whether the contract and plan were updated. Superseded decision: requiring a tracked launcher and ignored binaries. Changed acceptance criteria: install a portable launcher outside binary storage, leave tracking/ignore choices to the adopting repository owner, and diagnose missing setup assets when those choices leave a worktree without repair scripts. Preserve per-worktree installation, complete setup dependencies, explicit release-only repair, and configuration safety. Verification must exercise owner-selected policies rather than assume scripts are committed. Record this clarification in iteration 01; implementation approval remains pending.

## Evidence inspected

- .agents/skills/agentic-lens-setup/SKILL.md, sections 1.2 and 1.3: installs the operator skill only and downloads the wrapper into its ignored bin directory. Recovery depends on the setup downloader only if already available.
- .agents/skills/agentic-lens-setup/scripts/download.sh: DEST_DIR defaults to setup/bin; install_wrapper uses a bundled copy or moving-main network fallback. It supports explicit version, dry-run, destination, and local-prebuilt inputs.
- .agents/skills/agentic-lens-setup/scripts/agentic-lens-wrapper.sh and scripts/agentic-lens-wrapper.sh: native dispatch and missing-binary downloading require the wrapper to exist first.
- scripts/download-agentic-lens.sh: contributor-facing downloader overlaps the bundled downloader and must retain coherent defaults/behavior.
- .agents/skills/agentic-lens-operator/references/cli.md: canonical path is the ignored binary; missing command instructions do not identify a local repair procedure.
- .gitignore: bin directories are globally ignored, with a setup-bin ignore-file exception. The launcher must not rely on force-adding generated files.
- Makefile build: copies scripts/agentic-lens-wrapper.sh into development bin; retain this behavior. Make check currently runs scripts/tests/test_release_revision.py, which covers another release boundary and is not an installation test owner.
- No existing installation or wrapper runtime test suite was found in the targeted tracked-file searches. A focused script integration suite is justified for this missing boundary.
- TECHNICAL_GUIDELINES.md testing policy and docs/REGRESSION_PREVENTION.md configuration/portability, downloader, and guidance-ownership sections govern this slice.

These are local source observations. No release download, platform execution, or worktree repair has been verified in this planning stage.

## Implementation sequence after approval

1. Coder establishes a complete installation file manifest: setup manifest and its downloader/wrapper dependencies, operator manifest/references, and portable launcher outside bin. Git tracking is the adopting repository owner's choice for every installed file. Do not stage, commit, force-add, or impose ignore rules in adopting repositories. Identify a single owner for executable resolution and retain existing installed command compatibility. Prefer bundled dependencies during repair; resolve remote dependencies consistently with the selected release instead of silently mixing moving main with a pinned binary.
2. Coder implements the launcher and bounded per-worktree recovery. Preserve caller directory, argument boundaries and exit code. Honor existing Git policy without assuming downloaded artifacts are ignored. Diagnose absent setup scripts and provide an explicit bootstrap route as well as the local binary repair route. Keep downloads under the explicit repair authorization policy. Installation repair must not modify shared project configuration, infer another worktree's local telemetry settings, or overwrite edited skill files.
3. Coder aligns setup destinations/defaults, CLI/operator guidance, setup's generated policy block, and affected README/contributor references. Document missing binary and missing machine-local config as separate diagnoses. Keep source-build wrapper behavior and old installations functional.
4. Tester executes production scripts in isolated consumer fixtures. Add one coherent focused installation suite under scripts/tests and wire it into an appropriate Make check target; do not add product tests that merely assert documentation literals.
5. Reviewers independently examine the same frozen diff and actual test evidence. Parent consolidates findings and routes runtime fixes through coder/tester. No child owns integration, telemetry, commits, or PR publication.

## Regression cases and evidence required

- Fresh worktree: in a disposable consumer fixture, explicitly choose to commit installed scripts and ignore native binaries, then add a second worktree. Prove the scripts are inherited, missing binaries have a local repair path, and the fixture's chosen ignore rules are preserved. Invoke the actual local repair/launch path with a deterministic prebuilt fixture and prove argument/cwd/exit propagation. This fixture policy is a test choice, not an installation mandate.
- Owner-controlled tracking: verify setup/repair does not modify ignore rules or the Git index under differing owner policies, including a repository that does not ignore binaries. Cover a worktree without inherited setup/launcher files: guidance must identify their absence and the explicit bootstrap route rather than invoke a nonexistent repair script. Do not require users to commit binaries or scripts to obtain support.
- Installed/offline: deny downloader access and execute an existing platform fixture successfully. Missing/offline: fail repair clearly, preserve an existing executable if applicable, avoid a misleading success message, and give the exact local repair command. Do not treat a fixture binary as a real release verification.
- Space-containing paths and non-default destination: execute production launch/download scripts in a consumer path containing spaces. Assert output is written to the current worktree's selected destination, with no writes into the original worktree.
- Version consistency/idempotence: use controlled shell-boundary download fixtures to observe selected source URLs and destination behavior. Repeated repair keeps existing assets unless replacement was authorized; failed download does not clobber a usable executable. A controlled transport proves script behavior, not live GitHub release availability.
- Configuration separation: distinct shared configuration and absent machine-local configuration must survive repair byte-for-byte; launching must retain explicit source/path arguments. Inspect guidance for the correct missing-local-config diagnostic rather than adding tests that claim script fixtures validate CLI telemetry semantics.
- Platform contract: retain linux amd64/arm64, darwin amd64/arm64, and Windows amd64 artifact naming. POSIX launcher execution is covered on available POSIX hosts; Git Bash and native PowerShell instructions require explicit review and host-specific evidence where available. Do not claim simulated platform selection establishes native execution. Unsupported platforms fail clearly.
- Compatibility: execute the legacy installed command and retain the Make-built wrapper dispatch. Run syntax checks for changed shell files, focused installation tests, existing release-workflow tests, git diff --check, then the required integrated make gate. Do not expand backend/frontend product tests for unchanged code.

## Ownership and review allocation

Parent owns integration and all unassigned ticket artifacts, including incorporation of this user clarification into the planning documents. Coder owns approved installation scripts, skill/reference documentation, removal of installer-imposed ignore behavior, and any necessary Make target wiring. Tester owns the focused installation test file; coordinate Make wiring with coder before edits. Agents share the checkout and preserve others' changes.

This is a bounded shell/distribution change with no product data-model or UI edits. Use an independent reviewer to inspect behavior, failure handling, architecture/dependency ownership, and test effectiveness in one consolidated report, with a recorded narrow-scope exception to separate focus agents. Assign its report path explicitly before dispatch. A discovered substantive product impact requires renewed planning/review coverage and scope approval.

## Approval and delivery

The user explicitly approved the corrected contract and plan with "approved, follow with the implementation." This satisfies the implementation gate without another confirmation. The parent records actual commands, revision, outcomes, platform limits, and skipped checks in REVIEW.md and SUMMARY.md. Parent performs final telemetry extraction with the actual Codex source, validates session coverage, and delivers the reviewed PR under repository policy.

Review roster: generic reviewer under both configured profiles, Sol 6.1/high and Luna 6/medium, with independent fresh contexts. Questions cover install/repair behavior, download failure preservation, dependency/version ownership, Git-policy preservation, and test effectiveness. Unique reports: REVIEW_SOL.md and REVIEW_LUNA.md. The approved narrow exception combines these shell-installation focuses because no domain, ledger persistence, extraction, API, or UI behavior changes are permitted; any expansion into those areas reopens coverage. Parent owns the consolidated REVIEW.md and attribution records.
