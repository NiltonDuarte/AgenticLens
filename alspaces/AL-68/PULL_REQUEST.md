# feat(capture): automatically publish lifecycle session metrics (AL-68)

Install additive Codex/OpenCode lifecycle adapters that ask the agent for explicit ticket, title and type. A bounded native worker then keeps the ticket's canonical metrics up to date without the agent running generate. Manual generation remains a recovery path.

Concurrent roots and source-proven children in the same configured directory safely contribute to one ledger. Different project directories remain isolated even when they share a ticket and source store. Publication coordinates automatic/manual writers, preserves accepted facts on source failures, and recovers interrupted ledger/report/summary updates through an authorized journal. Snapshot completion describes observability, independently of ticket completion.

Configure cadence, work budgets, state location and optional detailed reports through shared/local configuration and CLI overrides. Project-local recoverable state remains the default; explicit external state is directory-scoped. Installation preserves user hooks, provides state-drift diagnostics, and disable/uninstall stops automatic publication. Ledger schema v5 is preserved. No mandatory Git hooks or CI changes.

Observed runtime support is Linux Codex0.166.0 and OpenCode1.19.5. Codex SubagentStart delivery, interrupted children before Stop, and other platforms/interfaces remain unverified. Raw transcripts stay outside Git.

Scope: [CONTRACT.md](https://github.com/NiltonDuarte/AgenticLens/blob/feat/AL-68-automatic-session-capture-hooks/alspaces/AL-68/CONTRACT.md) and [iteration02 plan](https://github.com/NiltonDuarte/AgenticLens/blob/feat/AL-68-automatic-session-capture-hooks/alspaces/AL-68/iterations/02/PLAN.md). Implementation and limitations: [SUMMARY.md](https://github.com/NiltonDuarte/AgenticLens/blob/feat/AL-68-automatic-session-capture-hooks/alspaces/AL-68/SUMMARY.md). Compact findings, reviewer/model attribution and review outcomes are retained in each iteration's FINDINGS.json and REVIEW.md; detailed historical reports remain in Git history.

Integrates complete skill bootstrap, explicit worktree binary repair and offline dispatch from main while preserving native capture protocol checks and rollback. Setup accepts the workflow's release-candidate tags; retained review summaries preserve the finding and verification credit. Iteration03 remains planning only.
