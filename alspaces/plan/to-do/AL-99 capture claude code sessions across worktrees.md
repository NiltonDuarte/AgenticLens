# AL-99: Capture Claude Code sessions across worktrees

## Goal

Install the Claude Code integration once at user scope. Claude Code root and subagent sessions in configured base checkouts and worktrees then reach the shared collector and publish ticket metrics automatically. This is the same capture Codex has from AL-79 and OpenCode has from AL-80.

## Context

AL-98 provides extraction for `--source claude`. Automatic capture still needs a native trigger. Claude Code 2.1.331 names these lifecycle hook events: `SessionStart`, `UserPromptSubmit`, `Stop`, `SubagentStart`, `SubagentStop`, `PreCompact`, `PostCompact` and `SessionEnd`. Hook payloads refer to `transcript_path`, `agent_id`, `agent_type` and `agent_transcript_path`. Hooks are configured in user, project or local `settings.json`. Policy can disable them, or allow only managed hooks. These names come from binary strings; how they behave at runtime is not yet verified.

## Acceptance criteria

- **Install.** `capture install` accepts `claude` and adds AgenticLens-owned hooks to the effective user settings, honoring `CLAUDE_CONFIG_DIR`. It keeps unrelated hooks and settings, never edits committed project settings or user-owned skills and agents, and supports dry run, reinstall, upgrade, disable and removal of only the entries it owns.
- **Hooks only signal.** Hooks tell the existing worker to refresh one exact session. They do not replay event payloads or do long extraction work in hooks. Payload values are treated as data and are never interpolated into a shell command. When a hook fails, coding keeps working and diagnostics show capture as degraded.
- **Routing and source checks.** Routing uses the validated session cwd, not the hook process directory. Configuration resolution, required local enablement and first-event initialization follow AL-79. A mismatch between the configured store and the runtime store is reported, and capture stops without reading another store.
- **Binding.** Binding to a ticket and publication follow AL-79 and the [AL-96 pending snapshot contract](../../../openspec/changes/al-96-verify-automatic-ticket-metrics-capture-end-to-end-for-v3/specs/automatic-ticket-publication/spec.md). A native subagent with proven parentage inherits its parent's ticket unless it was bound explicitly. Sessions launched independently bind independently. Unbound sessions stay local and unassigned.
- **Publication.** Turn, subagent and session checkpoints publish coalesced, honest provisional snapshots without double counting. Retry is bounded, unresolved work is kept, and other harnesses and contributors are preserved.
- **Real verification.** Run with a real installed hook, normal trust, a root bound in a linked worktree, a native subagent and source-backed publication. Record the version, payload semantics, IDs and limits. Disabled, managed-only or untrusted hook policy reports degraded capture and is never bypassed. Codex, OpenCode and manual paths still work.

Out of scope: a live observability service, and editing repository-owned Claude Code instructions (GEN-30).

## Implementation notes

Open questions for the developer contract:

- Is `SessionEnd` reliable enough, or are `Stop`/`SubagentStop` checkpoints enough on their own?
- How do `SubagentStart`/`SubagentStop` payload IDs map to subagent transcript IDs?

Reuse the AL-79 runtime. Do not reimplement it per harness.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply.
