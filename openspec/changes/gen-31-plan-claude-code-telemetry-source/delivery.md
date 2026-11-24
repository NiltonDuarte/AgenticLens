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
- **Telemetry.** Telemetry was backfilled after AL-98 added the Claude Code source, using the pinned release `v3.0.0-rc.169+6c41095e`. The PO session was selected by its session ID, and the ledger has one complete root session with no subagents. Cost is unavailable because Claude Code transcripts record no cost.

## Follow-up: session telemetry and engine pin

After AL-98 merged, the user steered this ticket to add the telemetry the GEN-30 to GEN-33 sessions could not record at the time, and to pin the release that reads Claude Code sessions. That work merges separately from the planning PR, [#165](https://github.com/NiltonDuarte/AgenticLens/pull/165).

- **Engine pin.** The shared `agenticlens.version` moves from `v3.0.0-rc.156+9308ac7` to `v3.0.0-rc.169+6c41095e`, the pre-release built from AL-98's merge. Older engines reject ledgers that contain a Claude Code session, so the new ledgers need it. Its Linux amd64 engine matched the release's `SHA256SUMS`, and the manager ran that exact build from this checkout. Because the repository sets `binary_dir: bin`, each checkout still needs the engine in its local `bin/`.
- **Ledgers.** Each ticket's sessions went to that ticket's own change folder. GEN-30, GEN-31 and GEN-33 are in this follow-up, and GEN-32 is on its open PR, [#166](https://github.com/NiltonDuarte/AgenticLens/pull/166). Every delegated ticket ran in its own worktree, so each ledger was generated in that worktree with its session ID, then copied here. Each ticket's code metrics therefore come from that ticket's own branch. The GEN-30 ledger also counts this follow-up's work, as its [delivery](../gen-30-onboard-claude-code/delivery.md) explains.
- **Verification.** On main `6c41095e` with only this follow-up's changes, the integrated `GOMAXPROCS=6 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 GOFLAGS=-timeout=35m make -j2` passed every target except two capture-worker timing tests in `cmd/agentic-lens`. `TestCaptureLateTicketMustPublishBeforeIdle` reported "native worker lifetime exhausted", and `TestNativeCaptureRequestAndConfigUpdateCannotStrandBinding` reported a worker still running at fixture removal. The machine's load average was 11 to 15 at the time, and both tests passed when rerun alone. This follow-up changes no Go code. Those targets include frontend unit (231) and integration (829) tests, the Go unit and race suites, vet, lint, the build, and the release, installation, capture-distribution and orchestration checks. The pinned engine lists the new ledgers. The OpenSpec workflow tests and strict validation of the GEN-30, GEN-31 and GEN-33 changes passed.

## Pull request

Add roadmap tickets for a Claude Code telemetry source. Today `--source` accepts only Codex and OpenCode, so Claude Code sessions cannot produce the ticket ledger that AGENTS.md rule 7 requires.

- **AL-98 (Ready):** generate and backfill with `--source claude`. It covers discovery across worktrees, counting each API response's usage once (local transcripts repeat usage once per content block), model and effort, skills and delegations, ownership of native subagents, AL-64 duration, explicit unavailable cost, portable ledger preservation, a capability matrix entry, updated guidance and rule 7, and verification against a real Claude Code install.
- **AL-99 (Blocked by AL-98):** automatic capture through user-scope Claude Code hooks on the shared AL-79 runtime, following the AL-79/AL-80 pattern.

These are planning artifacts only. They do not approve implementation. Telemetry was backfilled after AL-98 shipped.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 5m 39s (recorded duration)
- **Context Window Health:** HEALTHY — Root agent peak 125.6k tokens (28% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 6,968 reasoning tokens (21.4% of generated output)
- **Tool Health & Self-Correction:** 37 tool calls (0 self-corrections, 100.0% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 72 | 25,535 | 3,297,469 | 100.0% | 125.6k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **72** | **25,535** | **3,297,469** | **100.0%** | **125.6k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **claude-opus-5-5** | 1 | 72 | 25,535 | 3,297,469 | 100.0% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **72** | **25,535** | **3,297,469** | **100.0%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |
<!-- agentic-lens:metrics:end -->
