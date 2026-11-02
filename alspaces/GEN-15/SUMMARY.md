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
