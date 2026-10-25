# AL-56 Summary

Research branch: `research/codex-openai-metrics`.

## Result

AgenticLens can derive meaningful metrics from a Codex rollout without keeping transcript payloads. The implementation adds a Codex JSONL adapter that projects observed thread/model/token usage, turn timing, tool lifecycle outcomes, and compaction events into source-neutral facts, then calls the existing schema-v3 metrics derivation pipeline.

`agentic_metrics.json` is the committed session ledger generated from the real local Codex rollout using `generate --source codex --nonce`. It contains observed session/model/token/tool facts and pruned diagnostic events. Ticket metadata was supplied as AL-56. A literal nonce lookup correlated the rollout thread ID with this conversation; details are in `DATA_PROVENANCE.md`. The raw rollout remains local, and parser fixtures remain synthetic for repeatable tests.

The real rollout exposed a mapping detail: current records carry command lifecycle facts in completed `CommandExecution` items and custom tool calls in paired `response_item` records. The parser handles both while retaining no command/input/output content. It also applies a turn's `turn_context` model to its earlier `task_started` event, avoiding a spurious `default` model invocation.

CLI ledgers apply the shared event-retention policy: errors, compactions, and events from anomaly-linked turns. Codex event context uses the last request's input count; using cumulative session usage had caused false context anomalies and excessive event retention. Full tool/token totals are preserved before pruning. Regression tests cover request-level and cumulative-fallback token records, retained anomaly/error/compaction events, and ledger pruning.

## Source research

The passive Codex history path is rollout JSONL under `CODEX_HOME`; it is the nearest analogue to AgenticLens's OpenCode history/live readers. The CLI now supports `generate --source codex`, discovers sessions from active and archived rollout directories, resolves by nonce/session ID/ticket text, and includes child threads linked through rollout metadata. The existing dashboard consumes the CLI-generated ledger. Opt-in rollout traces offer richer inference, tool, compaction, and child-thread lifecycle data but carry sensitive payloads. `codex exec --json` is the clearest source for live runs AgenticLens launches. App-server notifications are suitable when AgenticLens owns or explicitly attaches to the Codex client, not for ambient collection from unrelated interactive sessions. Full comparisons and source links are in `RESEARCH.md`.

The local database audit found that `state_5.sqlite` can discover threads and rollout paths, while `thread_history_1.sqlite` offers a useful but rebuildable structured projection of turns/items/status/timing. The rollout JSONL remains the detailed canonical input for tokens and fallback when the projection is behind. `logs_2.sqlite` is diagnostic rather than a metrics event store; goals/memories/queue databases are not general session sources.

## Limits

- Automatic discovery scans rollout files to match ticket text; nonce or explicit session ID gives more precise selection. It supports `.jsonl.gz` but does not tail active rollouts. Ticket matching scans raw JSONL content in memory without retaining it.
- Cost is unavailable in the observed rollout data. Schema-v3 stores numeric zero; this means unavailable, not measured free usage.
- A rate-card proxy is feasible: estimated Codex credits for credit-billed use, or API-equivalent USD only when API-key billing is known. Neither is actual spend, and subscription usage cannot be reliably priced from tokens alone. Keep estimates separate from schema-v3 `cost_usd`; details and current official rates are in `RESEARCH.md`.
- Code facts are computed by the existing Git analyzer. Codex transcript/file-change data is not treated as Git truth.
- Codex rollout persistence is a private/version-sensitive format; fixture coverage is not an upstream compatibility guarantee.

## Verification

- Codex regeneration replaces the prior telemetry snapshot instead of adding it again. A CLI regression test runs generation three times, changes the model between runs, and verifies both models' token usage without duplicate invocation, tool, or event facts.
- `go test ./...` — passed.
- `go vet ./...` — passed.
- `npm test` — passed.
- `npm run typecheck` — passed.
- `npm run lint` — passed.
- `npm run build` — passed; embedded frontend assets refreshed.
- `git diff --check` — passed.
- Real-session metrics are committed as lifecycle evidence; automated tests use independent synthetic fixtures.


## Publication privacy review

The PR includes real session IDs, usage snapshots, timestamps, and the generated metrics table as development lifecycle evidence. Raw rollouts and machine-specific database inventories remain local. Working-directory paths are excluded from exported summaries; transcript and tool payloads are tested for omission.

## Ticket rename coupling correction

The ticket rename incorrectly updated research identifiers embedded in tests and in the former import UI/API defaults. Those runtime defaults were removed with the import workflow. Remaining Codex fixtures now use synthetic identifiers unrelated to this ticket. The CLI refresh test also inherited the checkout working directory and configuration, violating FIX-4 isolation; it now runs in temporary projects with their own configuration and checks two independent ticket prefixes and workspace layouts. Ticket renames must be scoped to tracking documents and artifact paths, not application or fixture values.

Validation: race-enabled CLI and Codex adapter suites pass with fresh test execution. No runtime source change was necessary for this correction.

Ticket index maintenance: restored the baseline index and removed its Done section. Updated repository and OpenCode workflow instructions to remove completed entries instead of accumulating them, retaining completed specifications in `plan/done/` and downstream dependency updates.

Code metrics correction: the research branch lacked a ticket ID, so default ticket-based Git discovery returned zero. The Codex path also omitted human steering counts and reused one cached SDLC iteration. `--code-ref HEAD` now selects committed branch changes explicitly; Codex root user messages supply SDLC iterations. Regression coverage uses isolated Git repositories and synthetic rollouts, including ticket renames, stale-cache rejection, duplicate message representations, and child delegation.

Fixture classification correction: `testdata/` directory components now classify as test code, including JSONL fixtures. Explicit configuration overrides retain precedence. Isolated classifier and Git aggregation regressions cover fixture totals and similarly named non-test directories.

Configuration correction: removed implicit test-directory/name and production-extension classification. The project YAML now declares both policies. Recursive glob matching respects directory boundaries, configured test paths outside source roots participate in metrics, and automatic ticket discovery applies the same source filtering as explicit code refs. Regression tests use independent configured paths and verify empty rules do not infer a category.

Subagent accounting probe and fix: spawned `telemetry_hi`, which returned `Hi!`. The probe exposed object-valued source metadata, inherited root metadata replay, and root/child ID collisions. The parser now preserves first child identity and parent, accepts nested spawn provenance, assigns `UnassignedSubAgent` when no role is provided, and resumes child parsing only at a matching `thread_settings_applied` marker. Parent replay is excluded from child usage, turns, tools, events, and duration. A missing marker produces a warning and conservatively excludes the ambiguous continuation. Synthetic adapter and isolated CLI regression tests passed with race detection.

Role mapping probe and fix: the spawned `coder` child had null top-level and nested `agent_role`, while its persisted `agent_path` ended in `/coder`. The CLI now uses only that final persona component as a lookup key in project `role_mappings`; explicit role metadata takes precedence. The actual probe maps to `coder`, and the raw path is not exported. Regression tests cover custom mappings, explicit-role precedence, root identity, and path privacy.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 51
- **Active Agent Duration:** 1h 34m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 247.7k tokens (124% of 200k healthy ceiling, 4 compactions) | Subagents peak 160.1k (Unassignedsubagent)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 1.5 avg turns/run (peak: 2 turns) | Unassignedsubagent 1.5
- **Test-to-Production Code Ratio:** 0.6:1 (+894 test lines / +1,408 prod lines)
- **Code Delivered:** +2,302 / -134 lines across 12 code files (excluding docs)
- **Reasoning Effort:** 147,687 reasoning tokens (29.3% of generated output)
- **Tool Health & Self-Correction:** 490 tool calls (40 self-corrections, 91.8% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (2 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+137.2k tokens) — Context tokens surged by 137.2k tokens in turn 1 (session: 01a174a5-1271-7791-8474-55e6bf0e84fd, total: 137.2k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.3k tokens) — Context tokens reached 161.3k in turn 3 (session: 01a17362-baa4-72b3-934d-3cb040c899ea), breaching 160k critical ceiling
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+43.4k tokens) — Context tokens surged by 43.4k tokens in turn 14 (session: 01a17362-baa4-72b3-934d-3cb040c899ea, total: 148.4k tokens, threshold: 40000)

### Agentic Execution Metrics

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **rootagent** | 1 | 1,635,510 | 348,640 | 92,827,008 | 98.3% | 247.7k | BLOATED (124%) | $0.00 |
| **unassignedsubagent** | 4 | 20,381 | 8,315 | 4,662,528 | 99.6% | 160.1k | ELEVATED | $0.00 |
| **Total** | **5** | **1,655,891** | **356,955** | **97,489,536** | **98.3%** | **247.7k** | **BLOATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-astra** | 4 | 384,986 | 53,378 | 23,834,496 | 98.4% | $0.00 |
| **gpt-6-luna** | 2 | 1,108,046 | 287,165 | 69,411,072 | 98.4% | $0.00 |
| **gpt-6.1-sol** | 1 | 162,859 | 16,412 | 4,243,968 | 96.3% | $0.00 |
| **Total** | **5** | **1,655,891** | **356,955** | **97,489,536** | **98.3%** | **$0.00** |
