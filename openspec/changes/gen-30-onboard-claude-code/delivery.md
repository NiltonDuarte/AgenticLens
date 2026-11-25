# chore(GEN-30): onboard Claude Code

## Outcome

Claude Code now loads the repository policy, the shared skills and the development subagent personas. `CLAUDE.md` imports `AGENTS.md`. `.claude/skills/` links to `.agents/skills/`, and `.claude/agents/` wraps the personas in `.agents/agent/`, as `.codex/agents/` does for Codex. `AGENTS.md` now says that a Claude Code session has no supported AgenticLens telemetry source. The decisions are in [design.md](design.md).

## Verification

- In a live Claude Code 2.1.331 session in this worktree, the six `agentic-lens-*` skills, the six `openspec-*` skills and the `planner`, `coder`, `tester`, `reviewer` and four focused reviewer subagents became available. A headless `claude -p` run reported section 1 of `AGENTS.md` from its context through the `CLAUDE.md` import. That run also reported that `.claude/settings.json` applies only once the workspace is trusted.
- In OpenCode 1.19.5, `opencode debug skill` listed each shared skill once, with no duplicate or conflict output, after the `.claude/skills/` links were added.
- `git diff --check` and `scripts/openspec.sh validate gen-30-onboard-claude-code --strict` passed.
- The integrated `GOMAXPROCS=6 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 make -j2` exited 0, and `python3 scripts/tests/openspec_workflow_test.py` passed its 3 tests. Both ran on the content of `0d4a8b57` over main `db4477fa`, which differs from the commit only in this delivery file and `tasks.md`.

## Limits

- Telemetry was backfilled after AL-98 added the Claude Code source, using the pinned release `v3.0.0-rc.169+6c41095e`. AL-98 also replaced the `AGENTS.md` sentence this change added. The ledger has two root sessions: the orchestrating session and the headless `claude -p` discovery check. The orchestrating session also coordinated GEN-31, GEN-32 and GEN-33, recovered the Pages publication and delivered this backfill. Its usage cannot be split by ticket, so all of it is counted here. Each delegated ticket's own session is in that ticket's ledger. The orchestrating session was still open at extraction, so its duration is provisional. Cost is unavailable because Claude Code transcripts record no cost.
- The setup bundle for adopting repositories still configures only Codex and OpenCode.
- Branch protection, Windows symlink checkout and subagent model choices were not tested.

## Pull request

Onboard Claude Code as an agent harness for this repository. Add `CLAUDE.md`, which imports `AGENTS.md`. Link `.claude/skills/` to the shared skills, and add `.claude/agents/` wrappers for the planner, coder, tester and reviewer personas. Add a shared `.claude/settings.json` allowlist and ignore personal Claude settings. `AGENTS.md` and the OpenSpec workflow guide now state that Claude Code has no telemetry source yet, so its sessions record telemetry as unavailable.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 14
- **Active Agent Duration:** 23m 58s (recorded duration including 159s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 239.1k tokens (53% of 450k healthy ceiling, 0 compactions) | Subagents peak 25.7k (Web-fetch)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Web-fetch 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 27,088 reasoning tokens (23.5% of generated output)
- **Tool Health & Self-Correction:** 140 tool calls (6 self-corrections, 95.7% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 2 | 326 | 86,953 | 21,493,379 | 100.0% | 239.1k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **web-fetch** | 1 | 4 | 1,038 | 4,110 | 99.9% | 25.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **3** | **330** | **87,991** | **21,497,489** | **100.0%** | **239.1k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **claude-haiku-5-5** | 2 | 6 | 1,464 | 4,759 | 99.9% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **claude-opus-5-5** | 1 | 324 | 86,527 | 21,492,730 | 100.0% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **3** | **330** | **87,991** | **21,497,489** | **100.0%** | **Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations)** |

Retained recorded history: 1 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
