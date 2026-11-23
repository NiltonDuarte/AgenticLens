# GEN-31: Planning outcome

## Goal and context

Claude Code is now used to work on AgenticLens, but it is not a supported telemetry source. `--source`, `capture install` and `telemetry.default_source` accept only `codex` or `opencode`. A Claude Code session therefore cannot produce the ticket ledger that AGENTS.md rule 7 requires; it can only be recorded as unavailable. This planning task adds roadmap tickets so that a later development session can deliver a Claude Code source adapter. It covers both manual generation and automatic capture, at the same level the existing Codex and OpenCode adapters reached.

The tickets are based on the current code and on local Claude Code 2.1.331 data, checked read-only:

- **Code paths.** The code paths are source selection in `backend/internal/domain/enums.go`, `backend/internal/adapters/config/source.go`, `backend/cmd/agentic-lens/main.go` and `capture.go`, and the Codex and OpenCode adapter and capture packages.
- **Transcript shape.** We inspected the record types and fields of three transcripts. Content was not copied.
- **Duplicate usage.** Streamed responses repeat identical usage once per content block for the same `requestId` and `message.id`, so usage must be deduplicated.
- **Cost.** No record has an invocation-level cost. A session-level `cost-state` record holds Claude Code's own estimate.
- **Settings and hooks.** The Claude Code binary names `CLAUDE_CONFIG_DIR`, a per-session `subagents/` folder, `cleanupPeriodDays`, forked sessions, and these hook events with their payload fields: SessionStart, Stop, SubagentStart, SubagentStop, SessionEnd and others.
- **Subagents.** No native subagent transcript was available on this machine.

## Roadmap outputs

The work is split into two tickets, each delivering a usable result end to end, in order:

1. [AL-98: Generate ticket metrics from Claude Code sessions](../../../alspaces/plan/to-do/AL-98%20generate%20ticket%20metrics%20from%20claude%20code%20sessions.md) is Ready. It covers manual generate and backfill with `--source claude`: discovery across worktrees, usage counted once, model and effort, skills and delegations, native subagent ownership, AL-64 duration, explicit cost coverage, portable ledger preservation, a row in the capability matrix, updates to guidance and AGENTS.md rule 7, and verification against a real Claude Code install. On its own it meets rule 7 for Claude Code sessions.
2. [AL-99: Capture Claude Code sessions across worktrees](../../../alspaces/plan/to-do/AL-99%20capture%20claude%20code%20sessions%20across%20worktrees.md) is Blocked by AL-98. It adds user-scope lifecycle hooks that feed the shared AL-79 runtime, which binds sessions, publishes them and keeps installation reversible. This matches the AL-79 (Codex) and AL-80 (OpenCode) pattern.

We split the work because AL-79 and AL-80 already separated extraction from native capture, and putting both in one ticket would be too large for one development session. GEN-30 (Claude Code instructions for this repository) is related work only.

## Open decisions

- **`cost-state` estimate.** Should Claude Code's own estimate feed AL2-6's API-equivalent estimates? Until that is decided, invocation cost stays unavailable.
- **Supported entrypoints.** Which entrypoints are claimed as supported: the CLI only, or also SDK hosts such as T3?
- **Unverified behavior.** How native subagent transcripts are linked, how compaction affects files, and the exact tool names for skills and subagents all need runtime verification in AL-98.
- **Hook reliability.** How reliable `SessionEnd` is, and how subagent hook IDs map to transcript IDs, are left to AL-99.
