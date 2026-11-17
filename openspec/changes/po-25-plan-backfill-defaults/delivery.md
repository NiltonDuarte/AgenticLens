# docs(roadmap): plan source-backed backfill follow-ups (PO-25)

## Pull request

Turn the recommendation in [AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145) into four roadmap tickets. AL-94 covers write-free repeats and requested summary repair. AL-95 covers retained owners and incomplete refresh reporting. Both are Ready for scope review. AL-96 depends on both and introduces an opt-in preview. AL-97 depends on preview evidence and can conclude with deferral or retention of the current default.

The product owner still needs to choose the preview and maintenance-only interface, acceptable batch cost, compatibility and release policy, and whether to approve a switch. Implementation requires separate authorization.

## Verification

Checked allocated IDs against roadmap files, worktrees and Git history. Verified all four ticket links and reciprocal dependencies against the index, and checked the PO planning artifacts. PR #145 is the base of this planning delivery. Verification here concerns the plan; the ticket criteria describe future runtime checks.

Strict OpenSpec validation and all three repository workflow tests passed. Application and build inputs match the discovery base; its integrated checks and retry limits are recorded in PR #145.

## Session telemetry

The planning handoff used the existing AL-65 chat. The engine rejected final Codex generation for PO-25 because the recorded source owner belongs to the AL-65 worktree. There is no separately attributable PO session ledger or cost. The original AL-65 telemetry remains in PR #145.
