# feat(capture): add native Codex capture and a separate version manager (AL-79)

Install Codex capture once in the effective user home. Native callbacks resolve configured checkouts and linked worktrees, capture exact root/child sessions and publish bound ticket snapshots without per-worktree installation.

Separate the native `agentic-lens` manager from `agentic-lens-engine`, which owns capture, generation, workers and the embedded dashboard. The manager selects the exact project engine or valid development output, explicitly installs verified versions and updates itself independently. Ordinary commands and hooks stay offline. Remove the shell selector and duplicate dispatcher; older installation compatibility was explicitly waived.

Preserve whole-file local capture policy, authoritative session routing, independent child assignments, immutable workers and portable schema v6 history. Recovery uses private undo journals and shared typed discovery references with complete resource leases. This repository uses `agenticlens.binary_dir: bin`; the fallback pin is unchanged.

Full `make` passes on a013dcfa, including race, frontend, installation/distribution/orchestration checks and five-target compilation. Affected reviews approve both configured model profiles. The native manager and companion engine are installed and verified locally; actual home hooks/configuration and trust are unchanged. Earlier genuine Linux normal-trust capture proof remains revision-bounded. Windows CI was removed by request; Windows runtime is uncertified and existing manager replacement requires manual maintenance. No release publication or merge is requested.

Scope: [contract](alspaces/AL-79/CONTRACT.md), [CI delta](alspaces/AL-79/iterations/02/PLAN.md), [approved split](alspaces/AL-79/iterations/03/PLAN.md). [Delivery and evidence](alspaces/AL-79/SUMMARY.md).
