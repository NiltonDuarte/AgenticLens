# chore(setup): make installed skills recoverable in worktrees (GEN-15)

Planning draft; implementation approval pending.

Fresh worktrees can lack installed skills, launchers, or binaries depending on the adopting repository's Git policy. The proposed change installs setup alongside operator, supplies a portable launcher outside binary storage, and documents explicit release-only bootstrap and repair paths. Tracking and ignore decisions remain with the repository owner; installation does not stage files or impose ignore rules.

Scope and delivery status: [CONTRACT.md](CONTRACT.md), [SUMMARY.md](SUMMARY.md), and [iteration plan](iterations/01/PLAN.md).
