# AL-98: Technical design

## Scope and boundaries

This change adds Claude Code as a third manual telemetry source. A developer runs `generate --source claude` or `backfill --source claude --recompute-db` and publishes a ticket ledger. The ledger holds the root session and its proven native subagents, their usage, served models, skills, delegations and AL-64 duration. Cost stays explicitly unavailable. The [delta spec](specs/claude-code-session-telemetry/spec.md) owns acceptance. The [roadmap ticket](../../../alspaces/plan/to-do/AL-98%20generate%20ticket%20metrics%20from%20claude%20code%20sessions.md) owns intent.

The work follows the existing Codex rollout adapter wherever the two sources behave the same. Selection is explicit and never falls back to another store or harness. Owner snapshots are replaced only after a complete read. Reconciliation goes through the existing generate and backfill merge paths. The new adapter supplies facts. It does not add a second publication, merge or budget mechanism.

Claude Code differs from Codex in three ways that drive the decisions below:

- A subagent's records carry the root's `sessionId`, so a child is identified by its `agentId`.
- One API response is written as several records that repeat the same usage.
- Forks copy earlier records into a new file, marked with `forkedFrom`.

Excluded: automatic capture and `capture install --source claude` (AL-99), trace hydration and live views for Claude Code owners, persisting effort, requested model or cache-creation tokens (see Steering), API-equivalent pricing (AL2-6), effort display (AL2-8) and this repository's own Claude Code instructions (GEN-30). The ledger schema version stays at 6.

The backfill engine and `main.go` are also being changed by AL-94 (refresh coverage) and AL-78 (completion feedback). Whichever change merges second must reconcile the shared source dispatch rather than overwrite it.

## Evidence and resolved questions

Inspected read-only on 2026-11-24 with Claude Code 2.1.331 under `~/.claude`. Only record shapes, keys and counts were read, never content. The sessions were roots `b918d950-b1d1-42d6-8d61-a688c68333da`, `56c1469e-2ac2-4b81-a0e9-011575bed5c4` and three others, and subagents `a6bbc56a9508cd15e` (planner) and `a8f40de74cf9c4ce2` (web-fetch).

- **Tool names.** Skills use the `Skill` tool, with input `skill` and `args` and a result containing `commandName` and `success`. Subagents use the `Agent` tool, with input `subagent_type`, `description`, `prompt` and `run_in_background`. Its result carries `agentId`, `agentType`, `status`, `resolvedModel` and the child's own `usage`. The string `"Task"` appears in the binary, but no transcript records a `Task` call, so this change classifies only `Agent`. The child's usage inside the parent's Agent result is never counted for the parent; the child transcript is the only source of child usage.
- **Parentage.** Each meta file's `toolUseId` matched exactly one parent `Agent` tool_use, and the parent result's `agentId` matched the file name. Every record in a subagent file had `isSidechain: true`, the file's `agentId` and the root `sessionId`. An unfinished subagent's meta file has no `finished` field, so the adapter does not depend on it.
- **Duplicate usage.** Nine responses were written as several records each, and every one repeated identical usage. Each `requestId` mapped to a single `message.id`. The usage object reports input, output, cache read, cache creation (5m and 1h), `output_tokens_details.thinking_tokens` and server tool use.
- **Forks and compaction.** Binary strings show that fork copies keep their original `uuid` and `requestId` and add a `forkedFrom` field with the source `sessionId` and `messageUuid`. Compaction appends a `system` record with `subtype: "compact_boundary"`, `compactMetadata` and `logicalParentUuid` to the same transcript. No native compaction or fork was observed. Their verification is planned, and both cells stay unverified until it runs.
- **Turns.** Turn-start prompts carry `turnPosition`. People's prompts showed `promptSource: sdk`, and background task notifications showed `promptSource: system` with `turnOrigin: task_notification`. Turns end with `stop_reason: end_turn`.
- **Entrypoints.** Every inspected record has `entrypoint: "t3-code"`, an SDK host. The binary also names `cli`, `sdk-cli`, `sdk-ts`, `sdk-py` and `claude-vscode`.
- **Cost.** `cost-state` holds `totalCostUSD` and per-model usage, and two of the five roots had one.

Resolutions to the ticket's open questions:

- **`cost-state` and AL2-6.** It is Claude Code's own estimate, not observed spend. This change does not persist it, sum it or feed it into AL2-6. AL2-6 can evaluate it separately.
- **Supported entrypoints.** Support is claimed for the transcript layout, regardless of entrypoint. The capability matrix marks `t3-code` as natively verified. The approved `claude -p` probe in task 7.4 covers `sdk-cli`. Any other entrypoint is NOT RUN.
- **Compaction.** Per the binary evidence, compaction keeps the same file. Each `compact_boundary` counts as one compaction. The approved probe in task 7.4 observes it natively.
- **Exact tool names.** `Skill` and `Agent`, as described above.

## Decisions and ownership

- **Configuration.** Add `domain.SourceClaude = "claude"` and `TelemetrySourcePaths.ClaudeHome`. `ResolveSource`, `ResolveClaudeHome` (through `resolveSourcePath`), the loader's `telemetry.default_source` and `telemetry.sources.claude.home` validation, and the `InitialLocalYAMLWithSettings` and `--local-settings` allowed keys accept the new value and key. `CaptureIdentity.Validate` and capture installation keep rejecting `claude`. `--capture-auto` keeps ignoring the selected default source. The persisted label is `source: "claude"`, matching the CLI value, with no schema bump. This follows the precedent set when `provisional_seconds` was added. Engines older than this change, including the pinned `v3.0.0-rc.156`, reject such ledgers. The delivery and release notes must say so.
- **Adapter.** Add a new package, `backend/internal/adapters/claude`, organized by responsibility. `store_discovery.go` lists `projects/*/<id>.jsonl` and `<id>/subagents/agent-*.{jsonl,meta.json}`, matches nonces literally and applies an optional budget modeled on `codex.DiscoveryBudget`. Manual `generate` passes no budget, matching Codex. Backfill passes the capture budget and reports when it runs out. `transcript_jsonl.go` parses records and ignores non-conversation record types. A record with no type, a malformed line or a foreign `sessionId` makes the owner incomplete, as in Codex. `subagent_proof.go` holds the parent proof. `session_facts.go` deduplicates responses by `requestId` or `message.id`, builds one invocation row per served model and one tool row per tool name, and produces lean occurrences, transient events, `CheckpointInvocationEvidence` and `SessionFact` values that honor `GenerationPolicy`. Token mapping follows `codex.Usage.ToInvocationTokens`:
  - input is the uncached `input_tokens`
  - cache read is `cache_read_input_tokens`
  - output is `output_tokens`
  - reasoning is `thinking_tokens`
  - peak context is the largest single-response sum of input, cache read and cache creation

  No other usage field is persisted. A response that lacks input, output or cache-read usage, or that carries conflicting usage, makes its owner incomplete. When `thinking_tokens` is absent, reasoning is 0, as Codex records it.
- **Unreadable children.** The ticket requires unreadable children to stay unresolved without blocking publication. This intentionally differs from Codex, where an unreadable selected child blocks extraction. The diagnostic is mandatory, the child is excluded from complete coverage, and its stored facts are kept.
- **Ownership check.** A new root's first recorded `cwd` must pass `capture.ValidateProject`. Proven children inherit the root's check, because isolated subagent worktrees can sit inside nested checkouts. No capture registration is written.
- **CLI wiring.** The `generate` branch for `claude` and owner validation live in `backend/cmd/agentic-lens/claude_generation.go`. `main.go` gains only the `--claude-home` flags (for generate, backfill and init), dispatch and usage text. Session users resolve through `git.ResolveSessionUser`, as for Codex.
- **Backfill.** `BackfillOptions.ClaudeHome` resolves and is stat-checked like the Codex home. `--source claude` selects Claude Code alone, and `all` now covers OpenCode, Codex and Claude Code. `all` fails before any read or publication when the Claude Code store is missing or unreadable. `both` still means OpenCode plus Codex. The help text and the operator reference describe all three. `engine.go` adds a Claude Code session bucket, which is protected from other harnesses and does not count toward `otherSessions`. A `recoverClaude` function in `recovery.go` mirrors `recoverCodex`, including retaining owners whose parent link changed and adding newly proven children. `source_budget.go` charges Claude Code reads against the shared budget.
- **Readers.** `fact_ledger_dto.go` accepts `claude` during canonical validation. `web/src/types/telemetry.ts` and `web/src/api/validation.ts` accept it too, and the embedded assets are rebuilt. The live hydrator already reports an `unsupported source`; that behavior stays.
- **Writable-path owners.** The coder owns the Go and web paths named above and the documentation listed in tasks section 6. The tester owns the new and extended tests named in tasks section 7. The native verifier, the tester in a separate run, owns the evidence under `evidence/`. The orchestrator owns integration, roadmap status, telemetry, `delivery.md` and the PR. The source capability matrix is part of AL-63's change, which is already delivered. The ticket names that file as the destination, so the coder appends to it without altering existing rows.

## Implementation approval

Approved by the user on 2026-11-24 through the orchestrator's approval question ("Approve, run the probe"). The approval covers this contract as amended by the user's five decisions of the same date: no effort or requested-model scope (see Steering); unbounded manual discovery; `all` including Claude Code while `both` stays OpenCode plus Codex; `source: "claude"` with no schema bump and the older-engine notice; and the native `claude -p` probe. The baseline was `9eeabb94`.

## Implementation approach

Build the adapter as pure parsing and grouping over readers, which can be tested against self-contained synthetic JSONL fixtures. Each fixture is labeled synthetic and carries no transcript content. Then wire configuration, generate, backfill and readers, and update guidance last.

Selected [regression boundaries](../../../docs/REGRESSION_PREVENTION.md):

- **Telemetry extraction and refresh:**
  - a root with children, plus a second root
  - repeated content-block records
  - forked copies
  - an unavailable child
  - an unrelated harness
  - growth followed by an unchanged repeat
  - malformed fixtures that reach their intended branch
- **Portable ledgers:**
  - Developer B clones a ledger without the store
  - B refreshes only B's own root
  - a restored older complete transcript
  - the same session in two tickets
  - an equal repeat of a backfill
- **Configuration:**
  - non-default store paths through the real loader
  - flag-over-config precedence
  - no `CLAUDE_CONFIG_DIR` fallback
  - `init` writer round-trip
  - an unchanged `--capture-auto` path
- **Frontend embedding:** canonical build, current `backend/internal/adapters/web/dist` and a standalone-binary read of a ledger with a Claude Code owner

Native verification uses the installed Claude Code 2.1.331 and this ticket's own root session with its native subagents, including the planner `a6bbc56a9508cd15e`. A disposable oracle outside Git sums usage per unique response straight from the transcripts. The recorded evidence compares those sums with the ledger. Synthetic fixtures do not count as native evidence. The native verifier also runs the approved `claude -p` probe: a disposable session with a subagent and a `/compact` in an isolated fixture repository. It verifies the `sdk-cli` entrypoint and compaction, and keeps its transcripts outside Git. The final ledger is generated with the built candidate, because the pinned release has no `claude` source.

Review coverage:

- **Data:** deduplication, refresh preservation and the forked-history rule
- **Behavior:** CLI precedence, the no-fallback rule, backfill `all` and `both`, `--capture-auto` and dashboard reads
- **Architecture:** adapter boundaries, how much `main.go` grows, and coordination with AL-94 and AL-78
- **Tests:** whether oracles are independent, and whether malformed and partial fixtures are reachable

## Steering

### Effort and requested model removed from scope

- **Source:** The user's answer to planning decision 1 on 2026-11-24.
- **Superseded acceptance:** The ticket asked to keep effort values and to preserve requested versus served models. Both are removed from this ticket's scope, along with any cache-creation field. The planned capability-matrix cells for these values are dropped as well.
- **Current acceptance:** Invocation rows record only the served model, and no new ledger fields are added.
- **Preserved guarantees:** Responses are counted once. Missing required usage still makes the owner incomplete rather than zero. Cost stays unavailable.
- **Verification impact:** No tests, documentation or matrix cells cover effort or requested model. Verification checks only that rows are keyed by the served model.

### Implementation corrections within approved intent

- **Source:** Coder findings against the real Claude Code 2.1.331 store on 2026-11-24, accepted by the orchestrator as routine corrections that keep AL-64 and usage-once intent.
- **Resumed subagents:** A subagent resumed through `SendMessage` starts with a coordinator meta prompt (`isMeta: true`, `origin.kind: "coordinator"`). It now starts a turn when the subagent has no open turn, and never counts as steering. Without this, all work after a resume went unmeasured. Root meta records still start nothing.
- **Client error notices:** Assistant records with model `<synthetic>` contribute no usage and no invocation row, but may end a turn.
- **Preserved guarantees and verification:** Duration still excludes time between turns. The tester covers both cases with synthetic fixtures, and the native run exercises a resumed subagent from this ticket's own session.

### Review round 1 corrections

- **Source:** Consolidated findings of the eight round 1 discovery reviews of `809e23d6` (reviewer-data, reviewer-behavior, reviewer-architecture and reviewer-tests, each on claude-opus-5-5/medium and claude-sonnet-5-5/high), accepted by the orchestrator on 2026-11-24 as corrections within the approved intent. Native verification earlier found the local slash-command defect (F1), fixed in `264550f3`.
- **Turn boundaries:** Local slash-command echoes withdrawn by their own output, harness interrupt markers and bash-mode records start no turn and never count as steering. An interrupt ends the open turn. These shapes come from Claude Code 2.1.331 records (local commands) and binary strings (interrupts, bash mode), pending native confirmation.
- **Store check:** A Claude Code store must contain a readable `projects` directory before backfill reads or publishes anything.
- **Refresh preservation:** `generate` now retains a child whose proven parent differs from its recorded parent, matching backfill and the existing requirement. A complete refresh with smaller totals names the owner and both totals.
- **Backfill adoption:** Adding an unrecorded proven root matches Codex, but the root must pass the project check and every added owner is named.
- **Accepted debt:** `recoverClaude` mirrors `recoverCodex`, and the discovery budget type and completeness rule are duplicated per harness. AL-94 rewrites the Codex recovery path; whichever ticket merges second must port AL-94's per-owner coverage to the Claude path. `recovery.go` merges without a textual conflict, so the delivery notes name it explicitly.

