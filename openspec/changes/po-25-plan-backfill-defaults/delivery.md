# docs(roadmap): plan source-backed backfill follow-ups (PO-25)

## Pull request

Turn the recommendation in [AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145) into two roadmap tickets. AL-94 makes explicit refresh repeatable, repairs requested summaries and reports retained owners independently of migration completion. AL-95 depends on AL-94 and delivers the opt-in preview and a documented rollout decision. Deferral and retention are valid outcomes.

The product owner still needs to choose the preview and maintenance-only interface, acceptable batch cost, compatibility and release policy, and whether to approve a switch. Implementation requires separate authorization.

## Verification

Checked allocated IDs against roadmap files, worktrees and Git history. Verified both ticket links and their reciprocal dependency against the index, and checked the PO planning artifacts. PR #145 is the base of this planning delivery. Verification here concerns the plan; the ticket criteria describe future runtime checks.

Strict OpenSpec validation and all three repository workflow tests passed. Application and build inputs match the discovery base; its integrated checks and retry limits are recorded in PR #145.

## Session telemetry

The planning handoff used the existing AL-65 chat. The engine rejected final Codex generation for PO-25 because the recorded source owner belongs to the AL-65 worktree. There is no separately attributable PO session ledger or cost. The original AL-65 telemetry remains in PR #145.
