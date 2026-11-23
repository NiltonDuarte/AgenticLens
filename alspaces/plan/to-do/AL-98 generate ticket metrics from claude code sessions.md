# AL-98: Generate ticket metrics from Claude Code sessions

## Goal

A developer working in Claude Code can run manual `generate` and explicit `backfill` with `--source claude` and publish a ticket ledger. The ledger holds the root session, its native subagents, usage, models, effort, skills, delegations and duration. Cost coverage stays explicit. This meets AGENTS.md rule 7 without falling back to another harness.

## Context

Today `--source`, `capture install` and `telemetry.default_source` accept only `codex` or `opencode`, so a Claude Code session can only be recorded as unavailable telemetry. These are the facts we have about the source. Checked on 2026-11-23 against Claude Code 2.1.331 (transcript shape and binary strings only):

- **Where sessions live.** Each session is a JSONL file at `<store>/projects/<encoded cwd>/<sessionId>.jsonl`. The store is `~/.claude` by default and `CLAUDE_CONFIG_DIR` overrides it. The encoding replaces both `/` and `.` with `-`, so you cannot decode a folder name back to its cwd. Every record carries its `cwd` and `gitBranch` instead.
- **Duplicate usage.** Assistant records carry `requestId`, `message.id`, `message.model`, `requestedModel`, `effort` and `message.usage`. Usage covers input, output, cache read, and cache creation with a 5m/1h split, plus thinking tokens and server tool use. One API response is written as one record per content block, and every one of those records repeats the same usage. That held for all 35 multi-record requests we inspected.
- **No per-invocation cost.** No assistant record has a cost field. A session-level `cost-state` record, which is not written for every session, holds Claude Code's own per-model `costUSD` estimate.
- **Subagents.** Records carry `isSidechain`, `parentUuid` and `uuid`. The binary refers to a per-session `subagents/` folder. We did not observe a native subagent transcript on this machine.
- **Settings that affect history.** Transcripts are removed after `cleanupPeriodDays`. Forked or resumed sessions can copy earlier history into a new file.
- **Interface-launched sessions.** Sessions launched through an interface (for example T3 delegation) are separate root sessions with no native parent link.

## Acceptance criteria

- **Choosing the source.** `claude` is accepted by `generate`, `backfill` and local `telemetry.default_source`. The store location comes from an explicit flag or the selected local configuration, with the same precedence and no-fallback rules as the Codex home. If the store is missing or unreadable, the command stops with an actionable error. It never returns empty successful telemetry and never reads another harness.
- **Finding sessions.** Discovery finds the root by a literal nonce or an exact session ID across all project folders under the store, including linked worktrees. It works within the existing discovery budget and reports when that budget is exhausted. Worktree and project identity come from record fields, never from decoded folder names.
- **Counting usage once.** Each API response contributes its usage exactly once, even when it is split across content blocks, copied into forked or resumed transcripts, or seen again on refresh. Missing usage categories stay unknown rather than zero. Native model and effort values are kept as they are, including requested versus served models.
- **Subagents and delegations.** A native subagent becomes a child session only when the source proves its parentage. Its role comes from its native agent type through `role_mappings`. Skill and delegation occurrences are owned by the session that recorded them, following the [owned-occurrence rules](../../../docs/OWNED_OCCURRENCES.md). Children that cannot be read, and sessions launched independently, stay explicitly unresolved and never get an inferred parent.
- **Duration.** Duration follows AL-64: completed turns are measured; an open turn counts as provisional time up to the snapshot cutoff and is replaced on refresh. Time between turns is never counted.
- **Cost.** Invocation cost is reported as unavailable. The `cost-state` estimate is not treated as observed spend (see open questions).
- **Shared ledgers stay intact.** Refresh replaces only owners whose sources are complete. It keeps Codex, OpenCode and other-machine sessions, owners whose transcripts are unavailable or were cleaned up, and unchanged repeat runs. The [portable ledger contract](../../../TECHNICAL_GUIDELINES.md#portable-ledger-contract) applies.
- **Capability matrix.** The [source capability matrix](../../../openspec/changes/al-63-verify-source-capabilities-across-harnesses/CAPABILITIES.md) gains a Claude Code entry with the tested version, the evidence behind each claim and any unsupported cells.
- **Guidance.** Setup, operator and developer guidance, the AGENTS.md rule 7 command and source references in `docs/OPENSPEC_WORKFLOW.md` name the `claude` source and its retention caveat.
- **Real verification.** Run against an installed Claude Code with a real root session and a native subagent. Record the version, source IDs, the deduplicated totals checked against the records, and the limits. Synthetic fixtures do not count as native evidence.

Out of scope: automatic capture (AL-99), live trace or observability services, API-equivalent pricing (AL2-6), effort display (AL2-8), and this repository's own Claude Code instructions (GEN-30, related but not a dependency).

## Implementation notes

Open questions for the developer contract:

- Should the `cost-state` estimate feed AL2-6?
- Which entrypoints (CLI and SDK hosts) are claimed as supported?
- Does compaction keep the same session file?
- What exact tool names does the native version use for skills and subagents?

Coordinate the shared backfill and CLI paths with AL-94 and AL-78.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply.
