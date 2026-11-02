# chore(setup): make installed skills recoverable in worktrees (GEN-15)

Fresh worktrees can lack installed skills, launchers, or binaries depending on the adopting repository's Git policy. Install the complete setup/repair skill alongside operator and provide a portable launcher outside binary storage. The launcher reports missing executables without downloading; explicit repair installs the current platform binary in that worktree.

The bundle installer retrieves dependencies from one selected release and preserves conflicting local files. Installation leaves configuration, Git staging, and ignore rules to the repository owner. Operator guidance distinguishes missing binaries, missing setup scripts, and missing machine-local settings. Existing binary wrappers remain supported.

After completed review approval, retain a concise findings summary with discovering actor/model and verification credit, and remove detailed review working files. Align repository and skill guidance with that lifecycle.

Scope and delivery status: [CONTRACT.md](CONTRACT.md), [SUMMARY.md](SUMMARY.md), and [iteration plan](iterations/01/PLAN.md).
