# feat(capture): add native Codex capture and a separate version manager (AL-79)

Install Codex capture once in the effective user home. Native callbacks resolve configured checkouts and linked worktrees, capture exact root/child sessions, and publish bound ticket snapshots without per-worktree installation.

Separate the native `agentic-lens` installation/version manager from `agentic-lens-engine`, which owns capture, generation, workers and the embedded dashboard. The manager selects the project's exact engine pin or valid development output, installs verified versions explicitly and updates its own installed executable independently. Ordinary commands and hooks remain offline. Remove the shell selector and duplicate dispatcher rather than maintaining the older installation layout.

Preserve whole-file local capture policy, authoritative source/session routing, independent child ticket assignments, immutable worker builds and portable ticket history. This repository uses `agenticlens.binary_dir: bin`; its fallback pin is unchanged. Installation and updates preserve ownership, recover interrupted transactions and retain user hook trust.

Affected installation/distribution and Go race checks pass. Correction reviews pass; integrated `make` and this machine's installation remain pending. Final evidence belongs in the summary. Earlier genuine normal-trust Linux Codex root/child capture proof applies within its recorded source-path limits. The Windows CI job was removed at the user's request; targets remain, Windows runtime is uncertified and existing manager replacement requires manual maintenance.

Scope: [CONTRACT.md](alspaces/AL-79/CONTRACT.md), [CI delta](alspaces/AL-79/iterations/02/PLAN.md), [approved manager/engine split](alspaces/AL-79/iterations/03/PLAN.md). Delivery and verification: [SUMMARY.md](alspaces/AL-79/SUMMARY.md).
