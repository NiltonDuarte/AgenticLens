# GEN-25: Remove Windows support

AgenticLens will support macOS and Linux only. Planning is in progress; application implementation has not started and scope approval is pending. The request explicitly includes all remaining Windows support after its CI removal.

The configured ad hoc prefix is `GEN-`, with conventional type `chore`, and the configured workspace root is `alspaces`. Existing workspaces, the roadmap index, local refs and matching remote branch refs were checked before allocating GEN-25. No matching specification or roadmap entry exists or is required.

Baseline: `4e5c458322a2cddfdb5911b26a4569ea75f06a09`; the initial working tree was clean. The existing placeholder branch was renamed to `chore/GEN-25-remove-windows-support` in this worktree.

Read-only discovery found active Windows support in native installation/routing, binary format validation, build/release packaging, capture process/locking/publication adapters, hook command generation, OpenCode plugin transport, PowerShell setup, tests and current guidance. Historical ticket records and portable ledger facts are outside the removal scope.

The [contract](CONTRACT.md) and [iteration plan](iterations/01/PLAN.md) define the proposed implementation and checks. No runtime checks have been run and no implementation or review approval is claimed. Telemetry extraction is deferred to the final delivery step as required by repository policy.
