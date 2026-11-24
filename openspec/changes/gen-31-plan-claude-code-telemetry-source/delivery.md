# docs(GEN-31): plan Claude Code telemetry source tickets

## Outcome

Two development tickets now plan Claude Code as an AgenticLens telemetry source ([proposal](proposal.md)):

- **[AL-98](../../../alspaces/plan/to-do/AL-98%20generate%20ticket%20metrics%20from%20claude%20code%20sessions.md)** is Ready, with no dependencies. It covers manual generate and backfill with `--source claude`. On its own it unblocks AGENTS.md rule 7 for Claude Code sessions.
- **[AL-99](../../../alspaces/plan/to-do/AL-99%20capture%20claude%20code%20sessions%20across%20worktrees.md)** is Blocked by AL-98. It covers automatic capture through user-scope lifecycle hooks.

Both entries are in `alspaces/plan/TICKETS.md`. No development change was created, and no application code, tests, runtime scripts or developer contracts were written.

## Verification and limits

- **Checks run.** All 31 relative links in the new and changed documents resolve. `git diff --check` passed, strict OpenSpec validation of this change passed, and no emojis were found.
- **IDs.** AL-98 and AL-99 were free. We checked the `to-do` and `done` folders, the roadmap index, the remote branches and the open PRs.
- **Source facts.** These come from read-only checks on Claude Code 2.1.331: the record types, fields and duplicate-usage pattern of three local transcripts, and strings in the binary. No transcript content entered Git.
- **Not yet observed.** No native subagent transcript existed locally, and the hook payloads were not tested at runtime. AL-98 and AL-99 must verify both.
- **Open questions.** These are listed in each ticket: whether the `cost-state` estimate should be used, which entrypoints are supported, how compaction affects files, the exact tool names, and how reliable the hooks are.
- **Telemetry.** This session's telemetry is unavailable. It ran in Claude Code, which AgenticLens cannot yet read as a source; AL-98 adds that. No ledger was generated, and no other harness's session was substituted.

## Pull request

Add roadmap tickets for a Claude Code telemetry source. Today `--source` accepts only Codex and OpenCode, so Claude Code sessions cannot produce the ticket ledger that AGENTS.md rule 7 requires.

- **AL-98 (Ready):** generate and backfill with `--source claude`. It covers discovery across worktrees, counting each API response's usage once (local transcripts repeat usage once per content block), model and effort, skills and delegations, ownership of native subagents, AL-64 duration, explicit unavailable cost, portable ledger preservation, a capability matrix entry, updated guidance and rule 7, and verification against a real Claude Code install.
- **AL-99 (Blocked by AL-98):** automatic capture through user-scope Claude Code hooks on the shared AL-79 runtime, following the AL-79/AL-80 pattern.

These are planning artifacts only. They do not approve implementation. This session's telemetry is unavailable because it ran in Claude Code.
