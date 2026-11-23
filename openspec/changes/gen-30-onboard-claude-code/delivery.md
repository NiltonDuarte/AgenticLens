# chore(GEN-30): onboard Claude Code

## Outcome

Claude Code now loads the repository policy, the shared skills and the development subagent personas. `CLAUDE.md` imports `AGENTS.md`. `.claude/skills/` links to `.agents/skills/`, and `.claude/agents/` wraps the personas in `.agents/agent/`, as `.codex/agents/` does for Codex. `AGENTS.md` now says that a Claude Code session has no supported AgenticLens telemetry source. The decisions are in [design.md](design.md).

## Verification

- In a live Claude Code 2.1.331 session in this worktree, the six `agentic-lens-*` skills, the six `openspec-*` skills and the `planner`, `coder`, `tester`, `reviewer` and four focused reviewer subagents became available. A headless `claude -p` run reported section 1 of `AGENTS.md` from its context through the `CLAUDE.md` import. That run also reported that `.claude/settings.json` applies only once the workspace is trusted.
- In OpenCode 1.19.5, `opencode debug skill` listed each shared skill once, with no duplicate or conflict output, after the `.claude/skills/` links were added.
- `git diff --check` and `scripts/openspec.sh validate gen-30-onboard-claude-code --strict` passed.
- The integrated `GOMAXPROCS=6 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 make -j2` exited 0, and `python3 scripts/tests/openspec_workflow_test.py` passed its 3 tests. Both ran on the content of `0d4a8b57` over main `db4477fa`, which differs from the commit only in this delivery file and `tasks.md`.

## Limits

- This session ran in Claude Code, which AgenticLens cannot extract yet, so this change has no `agentic_metrics.json`. No ledger was generated from another harness's session. GEN-31 ([PR #165](https://github.com/NiltonDuarte/AgenticLens/pull/165)) plans the Claude Code source as AL-98 (extraction) and AL-99 (automatic capture).
- The setup bundle for adopting repositories still configures only Codex and OpenCode.
- Branch protection, Windows symlink checkout and subagent model choices were not tested.

## Pull request

Onboard Claude Code as an agent harness for this repository. Add `CLAUDE.md`, which imports `AGENTS.md`. Link `.claude/skills/` to the shared skills, and add `.claude/agents/` wrappers for the planner, coder, tester and reviewer personas. Add a shared `.claude/settings.json` allowlist and ignore personal Claude settings. `AGENTS.md` and the OpenSpec workflow guide now state that Claude Code has no telemetry source yet, so its sessions record telemetry as unavailable.
