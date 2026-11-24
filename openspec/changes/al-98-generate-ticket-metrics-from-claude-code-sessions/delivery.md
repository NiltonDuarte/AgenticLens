# feat(AL-98): generate ticket metrics from Claude Code sessions

A developer working in Claude Code can now publish ticket metrics with `generate --source claude` and refresh them with `backfill --source claude --recompute-db`. The ledger holds the root session and every subagent whose parent link is proven from the transcripts, with deduplicated usage, served models, skills, delegations, compactions and AL-64 duration. Invocation cost is reported as unavailable. Claude Code's own `cost-state` estimate is not used.

## Outcome

- **Source selection.** `claude` is accepted by `generate`, `backfill` and local `telemetry.default_source`. The store comes from `--claude-home` or the selected local `telemetry.sources.claude.home`; there is no fallback to `CLAUDE_CONFIG_DIR` or `~/.claude`. A store without a readable `projects` directory fails before any read or publication. `init --claude-home` writes the setting. Capture installation and automatic capture still reject `claude` (AL-99).
- **Backfill values.** `all` now covers OpenCode, Codex and Claude Code, and fails before any publication without a readable Claude Code store. `both` still means OpenCode plus Codex.
- **Counting.** Each API response counts once, keyed by `requestId` or `message.id`, even when written as several records. Copied fork history is excluded. Missing or contradictory usage makes the owner incomplete rather than zero.
- **Subagents.** A subagent becomes a child only when its `meta.json` `toolUseId` matches an `Agent` call in the same tree and its records agree on `agentId` and the root `sessionId`. Its role is its native agent type mapped through `role_mappings`. Unproven, unreadable and interface-launched sessions stay unresolved with a diagnostic.
- **Duration.** AL-64 turns are measured from real prompts. Local slash commands, Esc interrupts and unanswered bash-mode commands add no steering; an interrupt closes the open turn, and an answered bash-mode command is a non-steering turn. Resumed subagents start a new measured turn.
- **Refresh.** Only completely read owners are replaced. Codex, OpenCode, other-machine, cleaned-up and unavailable owners are kept, as is a child whose proven parent differs from the recorded one. A complete refresh with smaller totals names the owner with both totals. Backfill reads each tree once, and adds an unrecorded proven root only when it passes the project check.
- **Readers.** The backend ledger reader and the dashboard accept `source: "claude"`. Trace hydration and live views report Claude Code owners as an unsupported source.
- **Scope decisions.** On 2026-11-24 the user removed effort and requested-model preservation from this ticket, kept manual `generate` unbounded, made `all` include Claude Code and approved `source: "claude"` without a schema bump. See [design.md](design.md).

## Verification and review

The reviewed and tested candidate is `f1749596`, integrated with `origin/main` at `1161f43a`; main had not moved at delivery. The full `make` passed on `f1749596`: Go unit, race and integration suites, 231 frontend unit tests, 829 frontend integration tests, vet and lint, release, installation, capture-distribution and orchestration checks, and native builds for four platforms. An earlier `make` at `03f50ca2` lost one Codex capture-worker test to its 8-second process deadline while eight reviewers were running tests; it passed three times alone under the race detector and passed in the final run. Strict OpenSpec validation passes.

After delivery, CI on `f5ad3cf3` hit a pre-existing race in the capture test fixtures: they rewrote the local configuration in place while a capture worker re-read it. The fixtures now replace the file atomically ([design.md](design.md#ci-flake-in-capture-test-fixtures)). This change touches test files only. The affected capture tests passed five times under the race detector, the other 32 tests that use the changed helpers passed under the race detector, and the full `cmd/agentic-lens` package passed. The full `make` was not rerun for this change.

Native verification ran against Claude Code 2.1.331; see [NATIVE_VERIFICATION.md](evidence/NATIVE_VERIFICATION.md). This ticket's own session tree (`t3-code` entrypoint, root and five subagents) matched an independent oracle exactly on tokens, tools, delegations, skills, turns and duration, and repeated runs were byte-identical. A `claude -p` probe (`sdk-cli`) confirmed one compaction in the same file. An interactive `cli` session driven through tmux confirmed Esc interrupts and bash-mode records. Native runs found two defects, both fixed and re-checked: local slash commands left an open turn, and answered bash-mode commands dropped their turn.

Behavior, data integrity, architecture and test effectiveness were each reviewed by claude-opus-5-5/medium and claude-sonnet-5-5/high, as main's review policy (GEN-33) requires. Round 1 at `809e23d6` requested changes in seven of eight cells; the architecture review by claude-opus-5-5 approved. The confirmed findings were interrupt and bash-mode turns, the store check, changed-parent retention, unreported shrinking refreshes, repeated tree reads, unchecked root adoption, `any` in tests, stale guidance and test gaps. Round 2 at `03f50ca2` approved in all eight cells. Round 3 re-reviewed the bash-mode fix and test additions at `f1749596` in the data, behavior and test cells and approved all six; architecture approval carried forward because only turn logic inside the existing parser changed. All 40 findings, reporters and dispositions are in [REVIEW_SUMMARY.json](evidence/REVIEW_SUMMARY.json). Four earlier Codex `gpt-6-luna` review runs failed at start on a Codex usage limit, before main switched the policy to Claude, and did no review work.

## Limits

- **Older engines.** Engines older than this change, including the pinned `v3.0.0-rc.156`, reject a ledger containing `source: "claude"`. An engine built from the base revision skipped only the affected ticket and kept serving the others.
- **AL-94 reconciliation.** `recoverClaude` mirrors `recoverCodex`. A trial merge with AL-94 conflicts in `backend/internal/adapters/backfill/engine.go`, `recovery.go`, `report.go` and `source_budget.go`. Whichever ticket merges second must port `recoverClaude` to AL-94's recovery structure, add Claude Code to AL-94's per-harness coverage and fold the Claude replacement notices into AL-94's per-owner coverage report. Resolving the conflicts so the code compiles is not enough. AL-78 merges cleanly.
- **Source limits.** Compaction requests and interrupted streaming responses write incomplete usage to the transcript, so those totals are lower bounds. Transcripts are deleted after `cleanupPeriodDays`. One unreadable or oversized transcript anywhere in the store stops nonce discovery, and makes backfill keep every Claude Code owner as stored while other harnesses still refresh.
- **Interrupt variants.** Interrupt records written by SDK hosts with a `promptSource`, and the "session ended" interrupt text, are not handled specially and were not observed natively.
- **Not observed natively.** Forks, nested subagents, subagent skill calls, failed skill calls, interrupt records written by SDK hosts and the `sdk-ts`, `sdk-py` and `claude-vscode` entrypoints. The [capability matrix](../al-63-verify-source-capabilities-across-harnesses/CAPABILITIES.md) records each cell.
- **Accepted debt.** Per-harness copies of the discovery budget type, the completeness rule and the changed-parent rule remain. Low-severity test gaps around bash-mode resets and record variants are listed in the review summary.
- **Reading the metrics.** Claude Code subagent roles are recorded as their native agent types, such as `reviewer-data`; the shared `role_mappings` maps only `reviewer`, so these types are not grouped yet. The subagent rework-loop anomalies reflect parallel subagents whose invocations interleave in time, not sequential coder and tester rework.
- **Roadmap Pages check.** The `publish` job of the Ticket roadmap pages workflow fails on this PR with "Artifact not found for name: github-pages". The job runs main's workflow (`pull_request_target`), so this branch cannot change it; AL-97's PR hit the same error.
- **Pending captures.** The four failed Codex review runs left unbound automatic-capture snapshots under the local `.agenticlens/pending/`; they are not committed and contain no review work.

## Pull request

### What changed

- Add a Claude Code source adapter that reads root and subagent transcripts, counts each response once, proves subagent parentage and measures AL-64 duration.
- Accept `--source claude`, `--claude-home` and local `telemetry.sources.claude.home` in `generate`, `backfill` and `init`, with no implicit store. Backfill `all` now includes Claude Code.
- Accept `source: "claude"` in backend and dashboard ledger readers.
- Update AGENTS.md rule 7, setup, operator and cost guidance, and the source capability matrix.
- Make capture test fixtures replace the local configuration atomically, so a running capture worker never reads a partly written file.

### Compatibility

- Engines older than this change reject ledgers that contain a Claude Code session.
- `backfill --source all` now requires a readable Claude Code store; use `both` for OpenCode plus Codex only.
- The AL-94 merge must port per-owner coverage to the Claude recovery path (see Limits).

### Verification

The reviewed and tested candidate is `f1749596`, integrated with `origin/main` at `1161f43a`; main had not moved at delivery. The full `make` passed on `f1749596`: Go unit, race and integration suites, 231 frontend unit tests, 829 frontend integration tests, vet and lint, release, installation, capture-distribution and orchestration checks, and native builds for four platforms. An earlier `make` at `03f50ca2` lost one Codex capture-worker test to its 8-second process deadline while eight reviewers were running tests; it passed three times alone under the race detector and passed in the final run. Strict OpenSpec validation passes.

After delivery, CI on `f5ad3cf3` hit a pre-existing race in the capture test fixtures: they rewrote the local configuration in place while a capture worker re-read it. The fixtures now replace the file atomically ([design.md](design.md#ci-flake-in-capture-test-fixtures)). This change touches test files only. The affected capture tests passed five times under the race detector, the other 32 tests that use the changed helpers passed under the race detector, and the full `cmd/agentic-lens` package passed. The full `make` was not rerun for this change.

Native verification ran against Claude Code 2.1.331; see [NATIVE_VERIFICATION.md](evidence/NATIVE_VERIFICATION.md). This ticket's own session tree (`t3-code` entrypoint, root and five subagents) matched an independent oracle exactly on tokens, tools, delegations, skills, turns and duration, and repeated runs were byte-identical. A `claude -p` probe (`sdk-cli`) confirmed one compaction in the same file. An interactive `cli` session driven through tmux confirmed Esc interrupts and bash-mode records. Native runs found two defects, both fixed and re-checked: local slash commands left an open turn, and answered bash-mode commands dropped their turn.

Behavior, data integrity, architecture and test effectiveness were each reviewed by claude-opus-5-5/medium and claude-sonnet-5-5/high, as main's review policy (GEN-33) requires. Round 1 at `809e23d6` requested changes in seven of eight cells; the architecture review by claude-opus-5-5 approved. The confirmed findings were interrupt and bash-mode turns, the store check, changed-parent retention, unreported shrinking refreshes, repeated tree reads, unchecked root adoption, `any` in tests, stale guidance and test gaps. Round 2 at `03f50ca2` approved in all eight cells. Round 3 re-reviewed the bash-mode fix and test additions at `f1749596` in the data, behavior and test cells and approved all six; architecture approval carried forward because only turn logic inside the existing parser changed. All 40 findings, reporters and dispositions are in [REVIEW_SUMMARY.json](evidence/REVIEW_SUMMARY.json). Four earlier Codex `gpt-6-luna` review runs failed at start on a Codex usage limit, before main switched the policy to Claude, and did no review work.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 7
- **Active Agent Duration:** 6h 1m (recorded duration including 819s provisional)
- **Context Window Health:** BLOATED — Root agent peak 342.1k tokens (76% of 450k healthy ceiling, 1 compactions) | Subagents peak 499.6k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.7 avg turns/run (peak: 5 turns) | Planner 2.0, Coder 4.0, Tester 2.0, Reviewer-architecture 2.0, Reviewer-behavior 3.0, Reviewer-data 3.0, Reviewer-tests 3.0
- **Test-to-Production Code Ratio:** 1.7:1 (+4,194 test lines / +2,527 prod lines)
- **Code Delivered:** +6,721 / -67 lines across 40 code files (excluding docs)
- **Reasoning Effort:** 431,379 reasoning tokens (24.5% of generated output)
- **Tool Health & Self-Correction:** 1,445 tool calls (14 self-corrections, 99.0% success rate)

### Anomalies & Health Flags
- **Status:** 19 execution anomalies detected (6 CRITICAL, 13 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.2k tokens) — Context tokens reached 450.2k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (451.0k tokens) — Context tokens reached 451.0k, breaching the 450k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (18 cycles) — Subagent roles 'coder' and 'tester' alternated for 18 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (5 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (23 cycles) — Subagent roles 'coder' and 'tester' alternated for 23 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.3k tokens) — Context tokens reached 251.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (254.3k tokens) — Context tokens reached 254.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.0k tokens) — Context tokens reached 250.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (254.3k tokens) — Context tokens reached 254.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (257.5k tokens) — Context tokens reached 257.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.4k tokens) — Context tokens reached 253.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.3k tokens) — Context tokens reached 250.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (252.9k tokens) — Context tokens reached 252.9k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.1k tokens) — Context tokens reached 250.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (258.3k tokens) — Context tokens reached 258.3k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (4 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> reviewer-tests (3 cycles) — Subagent roles 'tester' and 'reviewer-tests' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 182 | 87,438 | 17,109,915 | 100.0% | 314.1k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 488 | 246,687 | 60,336,661 | 100.0% | 477.7k | BLOATED (106%) | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 3 | 576 | 309,118 | 64,720,242 | 100.0% | 499.6k | BLOATED (111%) | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Unknown** | 1 | 452 | 142,059 | 31,692,355 | 100.0% | 342.1k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 2 | 198 | 70,273 | 11,954,527 | 100.0% | 223.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 432 | 160,455 | 38,777,337 | 100.0% | 322.8k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 2 | 278 | 148,506 | 23,625,541 | 100.0% | 309.7k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 320 | 166,603 | 34,920,087 | 100.0% | 379.0k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **15** | **2,926** | **1,331,139** | **283,136,665** | **100.0%** | **499.6k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **claude-opus-5-5** | 11 | 2,222 | 1,010,099 | 216,100,751 | 100.0% | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **claude-sonnet-5-5** | 4 | 704 | 321,040 | 67,035,914 | 100.0% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Total** | **15** | **2,926** | **1,331,139** | **283,136,665** | **100.0%** | **Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations)** |
<!-- agentic-lens:metrics:end -->
