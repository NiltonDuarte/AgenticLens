## Purpose

Record Claude Code root sessions and their native subagents in portable ticket ledgers through manual generation and explicit backfill. Every fact must come from the selected local transcript store, and the history recorded for other owners, harnesses and machines must be preserved. Product intent lives in the [AL-98 roadmap ticket](../../../../../alspaces/plan/to-do/AL-98%20generate%20ticket%20metrics%20from%20claude%20code%20sessions.md).

## ADDED Requirements

### Requirement: Accept the Claude Code source

`generate` SHALL accept `claude` from `--source` or, when the flag is absent, from the selected local `telemetry.default_source`. `backfill` SHALL accept `--source claude`. Backfill `--source all` SHALL include Claude Code together with OpenCode and Codex. `--source both` SHALL keep meaning OpenCode plus Codex. Any other value SHALL fail and list the supported values.

#### Scenario: Backfill selects all sources
- **WHEN** `backfill --source all --recompute-db` runs with readable OpenCode, Codex and Claude Code stores
- **THEN** recorded owners from all three harnesses are refreshed from their own stores

#### Scenario: Backfill all without a Claude Code store
- **WHEN** `backfill --source all --recompute-db` runs and the Claude Code store is not configured, cannot be read, or has no readable `projects` directory
- **THEN** the command fails before any publication with an error that names `--claude-home` or `telemetry.sources.claude.home`, and every ledger and summary stays unchanged

#### Scenario: Backfill both is unchanged
- **WHEN** `backfill --source both --recompute-db` runs with no Claude Code store configured
- **THEN** only OpenCode and Codex owners are refreshed, and recorded Claude Code owners keep their facts

#### Scenario: Configured default source
- **WHEN** `generate` runs without `--source`, and the selected local YAML sets `telemetry.default_source: claude` and `telemetry.sources.claude.home`
- **THEN** the configured Claude Code store is the only source read

#### Scenario: Unsupported source value
- **WHEN** `generate --source`, `backfill --source` or `telemetry.default_source` names a value the command does not support
- **THEN** the command fails and lists the values it does support

### Requirement: Resolve the Claude Code store without fallback

The store SHALL resolve from `--claude-home` first, then from the selected local `telemetry.sources.claude.home`. Relative and `~/` paths SHALL follow the existing Codex home rules. Neither `CLAUDE_CONFIG_DIR` nor a default home directory SHALL be read implicitly. A store that is missing, not a directory or unreadable SHALL stop the command before any publication. The command SHALL NOT read another store or harness.

#### Scenario: Flag takes precedence over local configuration
- **WHEN** local configuration selects `default_source: codex` and `telemetry.sources.claude.home: /a`, and `generate` runs with `--source claude --claude-home /b`
- **THEN** only the Claude Code store at `/b` is read, and no Codex or OpenCode source is opened

#### Scenario: Store missing or not configured
- **WHEN** `claude` is selected and no store path is configured, or the configured path does not exist, is a file or cannot be read
- **THEN** the command exits nonzero and names `--claude-home` or `telemetry.sources.claude.home`
- **AND THEN** no ledger or summary is created or changed, and no other harness store is read

### Requirement: Write Claude Code settings during local setup

The `init` local setup writer SHALL accept `--source claude` with an explicit `--claude-home`, and SHALL preserve valid existing local values.

#### Scenario: Local setup writes Claude Code settings
- **WHEN** `init --local-only --source claude --claude-home <path>` runs without existing source settings
- **THEN** the local YAML records `telemetry.default_source: claude` and `telemetry.sources.claude.home`

### Requirement: Keep automatic capture unchanged

Automatic capture SHALL keep supporting only Codex and OpenCode. A local default source of `claude` SHALL NOT change how registered owners are published automatically.

#### Scenario: Worker runs with a Claude Code default source
- **WHEN** the local default source is `claude` and the capture worker runs `generate --capture-auto` for a registered Codex or OpenCode owner
- **THEN** automatic publication of that owner behaves as before

#### Scenario: Capture installation request
- **WHEN** a user runs `capture install --source claude`
- **THEN** the command rejects `claude` as an unsupported capture source

### Requirement: Discover the root across project folders

Discovery SHALL search every project folder under `<store>/projects`, including folders that belong to linked worktrees. It SHALL select the root by a literal `--nonce` or an exact `--session-id`. A supplied subagent agent ID SHALL resolve to its proven root. A nonce matching several roots, a session ID found in several folders, or no match SHALL fail without publishing.

#### Scenario: Nonce selects one root in a worktree folder
- **WHEN** a literal nonce appears in exactly one root session or in that session's subagent transcripts, stored in the project folder of a linked worktree
- **THEN** that root and its proven subagents are selected, whichever checkout's folder holds them

#### Scenario: Nonce or session ID is ambiguous or absent
- **WHEN** the nonce appears in transcripts belonging to two different roots, the same session ID appears in two project folders, or no transcript matches
- **THEN** generation fails, asks for an exact `--session-id` where that applies, and leaves existing artifacts unchanged

### Requirement: Establish project identity from recorded fields

Project and worktree identity SHALL come from recorded `cwd` fields, never from decoded folder names. A newly selected root SHALL be accepted only when its recorded working directory belongs to the project running the command.

#### Scenario: Recorded working directory outside the project
- **WHEN** an exact session ID selects a root whose recorded `cwd` lies outside the project that runs `generate`
- **THEN** generation fails with a source ownership conflict and does not publish

### Requirement: Bound backfill discovery by the configured budget

Backfill discovery and reads for Claude Code SHALL consume `telemetry.capture.max_sessions` and `max_source_bytes`. When the budget runs out, backfill SHALL report incomplete recovery and SHALL preserve the stored facts of every owner it did not read. Manual `generate` SHALL NOT consume capture budgets, matching manual Codex generation.

#### Scenario: Backfill budget reached
- **WHEN** `backfill --source claude --recompute-db` reaches the session or byte budget before all recorded Claude Code owners are read
- **THEN** owners that were not read keep their stored facts, the report marks the refresh incomplete and names the budget, and owners read completely may be replaced

#### Scenario: Backfill finds an unrecorded proven root
- **WHEN** a ledger records only a delegated Claude Code child and backfill proves its root
- **THEN** the root and its proven children are added only if the root's recorded working directory passes the project check, each added owner is named in the output, and otherwise the root is skipped with a diagnostic

### Requirement: Count each API response once

Each API response SHALL contribute its usage exactly once, keyed by its native `requestId` or, when that is absent, by `message.id`. This SHALL hold when one response is written as several content-block records and when the ledger is refreshed.

#### Scenario: Content blocks repeat the same usage
- **WHEN** one response is written as thinking, text and tool_use records that each repeat identical usage
- **THEN** the invocation's input, output and cache-read tokens equal that usage counted once

### Requirement: Exclude copied fork history

Records marked with `forkedFrom` SHALL be treated as inherited history. They SHALL add no usage, turns, tools, skills, delegations or duration to the session that copied them, and SHALL NOT create a parent link.

#### Scenario: Copied history in a forked transcript
- **WHEN** a forked session's file begins with records copied from its source session and marked with `forkedFrom`, and both sessions are recorded in the ledger
- **THEN** each response is counted only for its original session, and the forked session has no parent link

### Requirement: Never record unknown usage as zero

A response whose records carry no usage, lack the input, output or cache-read category, or carry conflicting usage SHALL make its owner incomplete. Generation SHALL then preserve the stored ledger and summary instead of recording zero.

#### Scenario: Usage missing or contradictory
- **WHEN** an assistant record for a selected owner lacks its usage object, or two records of the same response carry different usage
- **THEN** generation reports that owner as incomplete and preserves the stored ledger and summary

### Requirement: Record the served model

Invocation rows SHALL be keyed by the served `message.model` value of their responses, recorded exactly as the source wrote it.

#### Scenario: Responses from two models
- **WHEN** a session's responses carry two different `message.model` values
- **THEN** the session records one invocation row for each served model, with each name kept verbatim

### Requirement: Report invocation cost as unavailable

Every Claude Code invocation SHALL record unavailable cost evidence. A session `cost-state` estimate SHALL NOT be treated as observed spend, persisted or summed.

#### Scenario: Transcript with a cost estimate
- **WHEN** a transcript contains a `cost-state` record with `totalCostUSD`
- **THEN** every Claude Code invocation records unavailable cost evidence, and the estimate appears nowhere in the ledger

### Requirement: Link subagents only by proven parentage

A subagent transcript SHALL become a child session only when all of the following hold:

- its `meta.json` `toolUseId` matches an `Agent` tool_use in a transcript of the same tree
- its records carry the file's `agentId` and the root `sessionId`
- no parent result contradicts the agent ID

The child ID SHALL be the native `agentId`. Its parent SHALL be the session whose transcript holds the matching tool_use.

#### Scenario: Proven subagent
- **WHEN** a root transcript contains an `Agent` tool_use whose ID equals a subagent's `meta.json` `toolUseId`, and the parent's tool result reports the same `agentId`
- **THEN** the ledger records a child session whose ID is that agent ID and whose parent is the root

#### Scenario: Nested subagent
- **WHEN** a subagent's own transcript contains the `Agent` tool_use matching another subagent's `toolUseId`
- **THEN** the second subagent's parent is the first subagent's agent ID

### Requirement: Take a child's role from its native agent type

A child's invocation role SHALL be its native `agentType`, recorded verbatim. `role_mappings` SHALL apply to that role only in aggregation and presentation.

#### Scenario: Mapped child role
- **WHEN** `role_mappings` maps `planner` to another role, and a proven child has `agentType: planner`
- **THEN** the ledger records the role `planner`, and aggregates report it under the mapped role

### Requirement: Leave unprovable children unresolved

A subagent that cannot be read or proven, and a session launched independently of the native Agent tool, SHALL remain unresolved and SHALL never receive an inferred parent. A root that keeps sidechain conversation records inside its own transcript SHALL be reported as an unsupported layout.

#### Scenario: Unproven or unreadable subagent
- **WHEN** a subagent file lacks its meta file, its `toolUseId` matches no tool_use in the tree, its records name a different `sessionId`, or the file cannot be read
- **THEN** that subagent is not added as a child, the command prints a diagnostic naming it, and any stored snapshot for that agent ID is preserved

#### Scenario: Explicit child output without a transcript
- **WHEN** the parent's Agent tool result reports an `agentId` and no transcript exists for it
- **THEN** the parent's delegation keeps that agent ID as its explicit child reference, the child is reported unavailable, and its stored facts, if any, are preserved

#### Scenario: Interface-launched session
- **WHEN** a session was launched through an interface such as T3 delegation and carries no native Agent proof
- **THEN** it can be selected only as its own root, and it never becomes a child of another session

#### Scenario: Unsupported sidechain layout
- **WHEN** a root transcript contains `isSidechain: true` conversation records instead of separate subagent files
- **THEN** that root's extraction is reported incomplete, and its stored facts are preserved

### Requirement: Record skill activations by their owning session

Each successful `Skill` tool_use SHALL record one activation on the session that made the call. The activation SHALL use the `skill` input as its name, the owner's raw role as its caller role and the record timestamp as its invocation time. A failed call SHALL count only as a tool error.

#### Scenario: Skill activations owned by a subagent
- **WHEN** a subagent calls `Skill` twice with the same skill name, and both calls succeed
- **THEN** that child session records two identical activations, and the root records none for those calls

#### Scenario: Failed skill call
- **WHEN** a `Skill` tool_result is an error, or its tool result reports `success: false`
- **THEN** no activation is recorded, and the call counts as a tool error

### Requirement: Record delegations by their owning session

Each `Agent` tool_use SHALL record one delegation on the calling session. The delegation SHALL carry the native subagent type, the description and the record timestamp, and SHALL reference a child only by a proven or explicit native agent ID. The shared settings for optional categories SHALL apply.

#### Scenario: Optional categories disabled
- **WHEN** `generation.metrics.skills` or `generation.metrics.delegations` is false
- **THEN** that category is omitted for Claude Code owners, as it is for other sources

### Requirement: Measure duration under AL-64

A turn SHALL start at a user prompt record that is not a tool result, meta record, compaction summary or copied record. It SHALL end at the owner's assistant record with a terminal stop reason. Completed turns SHALL add measured time. A final open turn SHALL add provisional time up to the snapshot cutoff, and a refresh SHALL replace it. A turn superseded by a new prompt SHALL end at its last owned record. Time between turns SHALL NOT be counted.

#### Scenario: Completed and open turns
- **WHEN** a root has one completed turn of 120 seconds, an idle gap of 600 seconds and an open turn that started 30 seconds before the cutoff
- **THEN** its duration is 150 seconds, of which 30 are provisional

#### Scenario: Resumed subagent
- **WHEN** a subagent with no open turn is resumed and its first new record is a coordinator meta prompt
- **THEN** that record starts a measured turn for the subagent and does not count as steering

#### Scenario: Completion after a provisional snapshot
- **WHEN** the open turn later reaches `end_turn` and `generate` runs again
- **THEN** the measured total replaces the earlier provisional snapshot without adding to it

### Requirement: Count steering iterations and compactions

Root steering iterations SHALL count prompts that a person sent. Prompts sent by the system, such as task notifications, SHALL count as turns but not as steering. Compactions SHALL count the owner's `compact_boundary` system records.

#### Scenario: Task notification turn
- **WHEN** a background subagent's completion notification starts a root turn
- **THEN** the turn's duration is measured, and the root's steering iterations do not increase

#### Scenario: Local slash command
- **WHEN** a session records a local command echo such as `/compact` followed by its local output before any assistant record
- **THEN** neither record starts a turn or counts as steering, and the session has no open turn afterwards

#### Scenario: Interrupt records
- **WHEN** a session records a harness interrupt marker
- **THEN** the marker ends the open turn at its timestamp, does not count as steering, and leaves no turn open

#### Scenario: Bash-mode command
- **WHEN** a session records a bash-mode input and its output
- **THEN** a model reply that follows makes the exchange a measured turn that does not count as steering, and without a reply before the next prompt or the end of the transcript the command adds no turn and leaves none open

### Requirement: Replace only completely read owners

Generation and backfill SHALL replace only the Claude Code owners that were read completely from the selected store. A complete read SHALL replace that owner's stored snapshot even when it holds fewer invocations. Whole-ledger aggregates SHALL be recomputed over refreshed and preserved owners together.

#### Scenario: Complete snapshot with fewer invocations
- **WHEN** an older but complete transcript for the same root session ID is restored with fewer responses than the ledger records
- **THEN** that root's snapshot is replaced with the smaller totals, a diagnostic names the owner with its previous and new totals, and other sessions are unchanged

### Requirement: Preserve owners outside the refresh

Refresh SHALL preserve Codex, OpenCode and other-machine sessions. It SHALL also preserve Claude Code owners whose transcripts are unavailable or were removed by `cleanupPeriodDays`, and owners whose parent link changed. Repeating a run on unchanged input SHALL produce an equal ledger with no duplicate sessions and no totals added twice.

#### Scenario: Mixed-harness ledger refreshed from one source
- **WHEN** a ledger holds Codex and OpenCode owners, a Claude Code owner from another machine and a local Claude Code root, and `generate --source claude` refreshes the local root
- **THEN** only the local root and its proven children change, and every other session keeps its recorded facts

#### Scenario: Transcript cleaned up
- **WHEN** a recorded Claude Code child's transcript no longer exists and its root is refreshed
- **THEN** the child's stored facts remain, and the run reports the child unavailable

#### Scenario: Recorded parent differs from the proven parent
- **WHEN** `generate` or backfill proves a recorded Claude Code child under a different parent than the ledger records
- **THEN** the child keeps its stored facts and parent link, and the run reports the changed link

#### Scenario: Repeat on unchanged input
- **WHEN** `generate` or `backfill --source claude --recompute-db` runs twice against an unchanged store
- **THEN** the second ledger has equal facts and the same session count, and no totals are added twice

### Requirement: Read ledgers with Claude Code sessions

Backend ledger decoding and the dashboard's response validation in the embedded engine SHALL accept `claude` as a session source. Recorded facts SHALL stay readable on a machine that has no local transcript store. Trace hydration for Claude Code owners SHALL report an unsupported source and SHALL keep the persisted facts visible.

#### Scenario: Dashboard workspace with a Claude Code owner
- **WHEN** the server reads a current-schema ledger whose sessions include `source: claude`
- **THEN** the workspace loads in the API and the embedded dashboard, and the owner contributes to aggregates

#### Scenario: Ledger read on a machine without transcripts
- **WHEN** a ledger generated with Claude Code owners is cloned to a different path on a machine without the `~/.claude` store
- **THEN** CLI, API and dashboard reads show the recorded facts and claim no new source verification

#### Scenario: Trace hydration request
- **WHEN** `/api/workspaces/{ticket}/events` covers a Claude Code owner
- **THEN** the response lists that owner as uncovered and gives an unsupported-source diagnostic, and the stored metrics remain visible
