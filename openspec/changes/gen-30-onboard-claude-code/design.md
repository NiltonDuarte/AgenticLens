# GEN-30 Claude Code onboarding

The user asked for Claude Code to be onboarded and approved the configuration-only part of the plan in this session. The scope is repository instructions and harness configuration. It changes no application code, tests, runtime scripts or schemas.

## Decisions

- `CLAUDE.md` imports `AGENTS.md` and repeats no policy, so the shared policy keeps a single source.
- `.claude/skills/<name>` are relative symlinks to `.agents/skills/<name>`, so the skills stay in one place. OpenCode also reads `.claude/skills/`. It deduplicates by skill name, and each skill was listed once in `opencode debug skill` (1.19.5).
- `.claude/agents/*.md` are thin wrappers that point to `.agents/agent/*.md`, in the same way `.codex/agents/*.toml` does. The OpenCode permission frontmatter in the shared personas is not valid Claude Code subagent frontmatter. The wrappers inherit the parent session's tools and permission mode, like the Codex agents. No orchestrator wrapper exists because the development skill drives orchestration.
- `.claude/settings.json` allows the commands the coder persona already allows, plus read-only Git and the pinned OpenSpec wrapper. Claude Code applies it only in a trusted workspace. `.claude/settings.local.json` stays out of Git.
- `AGENTS.md` rule 7 now states that a Claude Code session has no supported telemetry source. Its telemetry is recorded as unavailable instead of being generated from another harness's session.

## Preserved guarantees

The Codex and OpenCode wiring and the shared skill and persona content are unchanged. No AgenticLens source adapter is added or implied.
