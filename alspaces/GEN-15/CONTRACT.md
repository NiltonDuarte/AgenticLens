# GEN-15: Recoverable AgenticLens installation in Git worktrees

Status: Approved by the user: "approved, follow with the implementation." Approval includes the correction leaving Git tracking and ignore choices to the adopting repository owner.

## Problem and baseline

The user reports that setup installs the operator skill but does not install the setup skill itself. Git worktrees omit the ignored binary directory, leaving operators without the command or a local repair procedure.

Grounded baseline: commit 7bc670a7f5c9e93a220f3596f3d0e05cb6007332. Initial checkout status was clean. The configured ticket workspace root is alspaces. GEN-14 already has a remote branch; this task uses GEN-15. The parent owns branch chore/GEN-15-worktree-installation.

The setup manifest installs only the operator manifest and three references. Its downloader and bundled wrapper exist in the development repository but are not installed into adopting repositories. Both the wrapper and platform executables are placed under an ignored bin directory. The downloader defaults still reference setup/bin while setup instructions explicitly select operator/bin. The operator CLI reference identifies only the ignored executable path and says missing executables require setup. The wrapper can download a missing native executable only when the wrapper itself exists.

## Proposed acceptance boundaries

1. Setup installs a complete setup/repair skill, including every runtime script it requires, alongside the operator skill and its references. A consumer repository does not require the AgenticLens source tree or Go/Node toolchains for repair.
2. A portable launcher lives outside the binary storage directory. Proposed POSIX command path: .agents/skills/agentic-lens-operator/scripts/agentic-lens.sh. Setup and operator guidance consistently identify that entrypoint and the local repair command. Its placement does not require the adopting repository to track it.
3. Git tracking and ignore policy for all installed files, including scripts and native executables, belong to the adopting repository's owner. Setup/repair must not stage or commit installed files, force-add them, or impose ignore rules. Explain the worktree implications of the owner's choices: committed skills/scripts can be inherited, while untracked or ignored files may need installation in each worktree. Diagnose missing setup assets as well as missing binaries and provide an explicit bootstrap/repair route when local repair scripts are unavailable. Honor existing repository policy; do not assume native executables must be ignored or tracked.
4. Installation remains local to the current worktree. Launching or repairing must preserve the caller's current working directory, arguments, exit status, shared project configuration, ticket outputs, and explicit telemetry source selection. Paths containing spaces must work.
5. If binaries are absent, the operator identifies the missing local installation and explains the exact repair action and network/file effects. Existing authorization for that repair is sufficient; otherwise obtain approval before downloading. An ordinary launcher invocation, including help/version or other read-only operations, must not silently download a missing native binary through the legacy wrapper. Fetching is an explicit setup/repair action under the existing authorization policy. An installed local executable remains usable offline; unavailable repair fails clearly with a retry command and leaves existing executables intact.
6. Resolve the selected release consistently for downloaded installation dependencies and platform binaries. Preserve explicit version selection. Repair must not silently upgrade an existing installation or replace locally edited skill files. Setup still presents exact proposed policy text and honors its existing approval gate for adopting-repository changes.
7. Operator diagnosis distinguishes missing binaries from a missing ignored machine-local configuration in a fresh worktree. It explains explicit harness/source-path selection where needed; it must not silently copy the main worktree's local configuration, infer another developer's paths, or rerun init over existing shared configuration.
8. Preserve existing legacy bin command installations and source-project build behavior. Update the setup template, operator references, and active public instructions affected by the new entrypoint. Historical ticket records remain unchanged.

## Architecture and exclusions

This slice concerns imperative shell scripts, installation manifests, respect for user-owned Git policy, and operational guidance. No Go domain, telemetry schema, extraction, API, or TypeScript behavior changes are proposed. FCIS and existing single-binary distribution remain intact. Tests use isolated consumer-repository fixtures rather than live ticket workspaces. Zero emojis apply throughout.

## Clarification incorporated

Source: user correction, "This is on user, not on the platform to decide how to track the binary file," followed by the request to update the contract and plan. This supersedes the proposed requirement to track the launcher and mandate ignored binaries. Installation location, complete setup dependencies, per-worktree repair, release-only adoption, and configuration preservation remain proposed guarantees. Verification now covers owner-selected Git policies and missing setup assets instead of assuming installed scripts are committed. This is a correction to iteration 01; it does not establish implementation approval.

Use per-worktree binary storage for this slice. A shared user cache could reduce downloads but adds version selection, concurrent updates, cleanup, ownership, and offline fallback concerns. Global installation, cross-worktree binary discovery, machine-local configuration inheritance, automatic Git hooks, and cache management are excluded.

## Completion gate

The parent presents this contract and iterations/01/PLAN.md for explicit scope approval before dispatching implementation. Completion requires focused runtime regression evidence, independent review, the required integrated make gate, and parent-owned telemetry and PR delivery. A plan alone does not establish that the defect is fixed.
