# GEN-20 iteration 01

1. Inspect Git worktrees and T3 thread ownership; record the baseline.
2. Preserve dirty checkouts and live top-level thread checkouts. Confirm no other provider session is active in candidate worktrees.
3. Preserve local capture data outside Git, then remove clean candidates without force, retaining branches.
4. Verify remaining worktrees and branch preservation; record results and publish the operational record per repository delivery rules.

Authorization: the user explicitly requested cleanup. No application implementation or delegated work is required. T3 tool metadata explicitly documents that worktrees remain after thread deletion. Its rationale is undocumented; data protection is a possible explanation, not a verified design rationale.

Verification: inventory and removal exit codes, remaining registrations, preserved dirty statuses and branch refs. The full make gate is a repository delivery requirement; no application behavior changes.
