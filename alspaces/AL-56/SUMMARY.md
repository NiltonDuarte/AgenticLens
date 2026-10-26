# AL-56 Summary

Research branch: `research/codex-openai-metrics`.

## Result

AgenticLens can derive meaningful metrics from a Codex rollout without keeping transcript payloads. The implementation adds a Codex JSONL adapter that projects observed thread/model/token usage, turn timing, tool lifecycle outcomes, and compaction events into source-neutral facts, then calls the existing schema-v3 metrics derivation pipeline.

`agentic_metrics.json` is the committed session ledger generated from the real local Codex rollout using `generate --source codex --nonce`. It contains observed session/model/token/tool facts and pruned diagnostic events. Ticket metadata was supplied as AL-56. A literal nonce lookup correlated the rollout thread ID with this conversation; details are in `DATA_PROVENANCE.md`. The raw rollout remains local, and parser fixtures remain synthetic for repeatable tests.

The real rollout exposed a mapping detail: current records carry command lifecycle facts in completed `CommandExecution` items and custom tool calls in paired `response_item` records. The parser handles both while retaining no command/input/output content. It also applies a turn's `turn_context` model to its earlier `task_started` event, avoiding a spurious `default` model invocation.

CLI ledgers apply the shared event-retention policy: errors, compactions, and events from anomaly-linked turns. Codex event context uses the last request's input count; using cumulative session usage had caused false context anomalies and excessive event retention. Full tool/token totals are preserved before pruning. Regression tests cover request-level and cumulative-fallback token records, retained anomaly/error/compaction events, and ledger pruning.

## Source research

The passive Codex history path is rollout JSONL under `CODEX_HOME`; it is the nearest analogue to AgenticLens's OpenCode history/live readers. The CLI discovers sessions from active and archived rollout directories by a unique nonce or explicit thread ID, and includes child threads linked through rollout metadata. Ticket text and working-directory heuristics were removed as unreliable selectors. The existing dashboard consumes the CLI-generated ledger. Opt-in rollout traces offer richer inference, tool, compaction, and child-thread lifecycle data but carry sensitive payloads. `codex exec --json` is the clearest source for live runs AgenticLens launches. App-server notifications are suitable when AgenticLens owns or explicitly attaches to the Codex client, not for ambient collection from unrelated interactive sessions. Full comparisons and source links are in `RESEARCH.md`.

The local database audit found that `state_5.sqlite` can discover threads and rollout paths, while `thread_history_1.sqlite` offers a useful but rebuildable structured projection of turns/items/status/timing. The rollout JSONL remains the detailed canonical input for tokens and fallback when the projection is behind. `logs_2.sqlite` is diagnostic rather than a metrics event store; goals/memories/queue databases are not general session sources.

## Limits

- Nonce discovery fails when a nonce appears in multiple roots; explicit thread ID is available as a deterministic selector. Discovery supports `.jsonl.gz` but does not tail active rollouts.
- The local audit found 88 `world_state` and 27 `inter_agent_communication_metadata` records. Those known non-metric types are explicitly skipped; any unrecognized type or missing type now fails extraction. Missing child resume markers also fail rather than emit undercounted metrics.
- All records in the audited 21 rollout files had timestamps, but this is not guaranteed by the private format. Missing timestamp values remain unavailable and should not be interpreted as zero duration.
- Cost is unavailable in observed rollout data. The current schema stores numeric zero, which is ambiguous with free usage; a nullable representation should be designed across the shared schema, aggregation, and UI.
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

Code metrics correction: `generate` defaults to the explicit `--code-ref HEAD`; Git discovery no longer searches ticket IDs in branches or commit subjects. Backfill requires an explicit code ref when recomputing Git facts. Codex root user messages supply SDLC iterations. Regression coverage uses isolated Git repositories and synthetic rollouts, including ticket-free branch names, invalid refs, duplicate message representations, and child delegation.

Cost follow-up: created AL-57 through the PO ticket workflow to carry unavailable USD cost as nullable values across the shared schema, aggregates, migrations, CLI, and dashboard. It is Ready and can proceed independently of AL-55; no PO workspace was created.

Fixture classification correction: `testdata/` directory components now classify as test code, including JSONL fixtures. Explicit configuration overrides retain precedence. Isolated classifier and Git aggregation regressions cover fixture totals and similarly named non-test directories.

Configuration correction: removed implicit test-directory/name and production-extension classification. The project YAML now declares both policies. Recursive glob matching respects directory boundaries and configured test paths outside source roots participate in code-ref metrics. Regression tests use independent configured paths and verify empty rules do not infer a category.

Subagent accounting probe and fix: spawned `telemetry_hi`, which returned `Hi!`. The probe exposed object-valued source metadata, inherited root metadata replay, and root/child ID collisions. The parser now preserves first child identity and parent, accepts nested spawn provenance, assigns `UnassignedSubAgent` when no role is provided, and resumes child parsing only at a matching `thread_settings_applied` marker. Parent replay is excluded from child usage, turns, tools, events, and duration. A missing marker fails extraction. Synthetic adapter and isolated CLI regression tests passed with race detection.

Role mapping probe and correction: the spawned `coder` child had null top-level and nested `agent_role`, while its persisted `agent_path` ended in `/coder`. OpenCode extraction stores the source agent name as the raw fact and does not resolve project `role_mappings`; Codex now follows that behavior. It extracts only the final persona component as the raw role, gives explicit role metadata precedence, and keeps the path private. The CLI regression sets `coder: tester` in project config and verifies the persisted invocation role remains `coder`.

## Iteration 01: OpenCode Compatibility & Non-Breakage Hardening

Conducted cross-cutting audit to ensure Codex support introduced in AL-56 does not regress OpenCode workflows, historical ledgers, or metric derivations:

- **Domain Layer**: Pure functional core in `internal/domain` preserved with zero modifications, zero I/O, value semantics, and full backward compatibility.
- **Adapters Layer**:
  - `internal/adapters/git/classifier.go`: Added `DefaultTestPatterns` and `DefaultProdPatterns` fallback logic when project configuration arrays are empty or unconfigured, preventing test-to-production code ratio collapse to `0:0` in legacy or minimal OpenCode repositories. Explicit configuration overrides maintain full precedence.
  - `internal/adapters/config/writer.go`: Updated `GenerateDefaultYAML` so that new configurations scaffolded via `agentic-lens init` include standard pattern defaults rather than empty lists.
  - `internal/adapters/backfill/engine.go`: Enforced mandatory `--ticket-id` requirement when `--recompute-git` is passed, preventing batch backfills across entire workspace directories from accidentally overwriting every historical ticket with a single commit diff from `--code-ref`.
  - `internal/adapters/repository.go`: Enhanced ledger loader to preserve cached code facts even when invocation or tool slices are empty.
- **CLI / Entrypoint Layer**:
  - `cmd/agentic-lens/main.go`: In `runGenerate`, added graceful fallback to cached code facts from existing JSON ledgers with an explicit warning when `ComputeCodeFactsAtRef` fails (e.g. unresolvable base branches in minimal environments). Enforced mandatory `--ticket-id` for `backfill --recompute-git`. Verified `--source` strictly defaults to `"opencode"`.
- **Verification**:
  - Full backend race tests (`go test -race ./...`) passed across all packages.
  - Frontend typecheck, linter, and Vitest suite (30 files, 715 tests) passed.
  - Multi-platform compilation via `make build` succeeded.
  - Adversarial regression tests covering empty patterns, cached code fallbacks, and backfill scoping invariants passed.

## Iteration 02: Strict Zero-Fallback Classification & Warning Alignment

Refined file classification and operational diagnostics based on user steering:

- **Strict Zero Runtime Fallback Policy (`classifier.go`)**:
  - Reverted runtime classification fallbacks. `ClassifyFile` strictly evaluates user-configured patterns (`TestPatterns`, `ProdPatterns`, `Include`, `Exclude`). Unmatched files evaluate strictly to `CategoryOther`.
  - Added `CheckUnconfiguredPatterns` to inspect project configuration and emit actionable `stderr` warnings during `agentic-lens generate` when `code_metrics.test_patterns` or `code_metrics.prod_patterns` are empty.
- **Template-Only Default Scaffolding (`writer.go`)**:
  - Maintained `DefaultTestPatterns` and `DefaultProdPatterns` strictly inside `adapters/config/writer.go` for template generation during `agentic-lens init`, ensuring users explicitly own and commit pattern definitions on disk.
- **CLI Resilience & Operational Guardrails (`main.go`, `engine.go`, `repository.go`)**:
  - Reaffirmed CLI entrypoint resilience: when `ComputeCodeFactsAtRef` encounters Git errors in shallow or detached environments, it falls back to cached code facts from existing ledgers with an explicit warning instead of failing.
  - Retained `--ticket-id` enforcement on `backfill --recompute-git` to prevent multi-ticket Git diff contamination.
  - Retained non-destructive ledger loading in `repository.go`.
- **Verification**:
  - Race-enabled Go tests and Vitest suites passed cleanly (0 failures, 0 skips).
  - Verified warning emission in `main_test.go` and zero-fallback classification rules in `classifier_test.go`.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 2
- **Active Agent Duration:** 2h 7m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 143.5k tokens (72% of 200k healthy ceiling, 1 compactions) | Subagents peak 176.7k (Coder)
- **Role Effort Distribution (Subagents):** Planner 18.6%, Coder 56.9%, Tester 3.9%, Reviewer 20.5%
- **Subagent Turn Velocity & Thrashing:** 49.1 avg turns/run (peak: 127 turns) | Planner 37.0, Coder 103.0, Tester 13.5, Reviewer 43.0
- **Test-to-Production Code Ratio:** 0.8:1 (+1,224 test lines / +1,502 prod lines)
- **Code Delivered:** +2,726 / -1,018 lines across 19 code files (excluding docs)
- **Reasoning Effort:** 112,677 reasoning tokens (60.2% of generated output)
- **Tool Health & Self-Correction:** 513 tool calls (10 self-corrections, 98.1% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (1 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.2k tokens) — Context tokens reached 161.2k in turn 49 (session: ses_e8878898fffeAahBOyEwQAMD4f), breaching 160k critical ceiling
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: mcp-gateway_mcp-exec (3 consecutive errors) — Tool 'mcp-gateway_mcp-exec' encountered 3 consecutive execution failures (session: ses_e88858e3affeaQz4frllatuA9L, turn: 48)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 397,545 | 11,139 | 3,397,822 | 89.5% | 91.8k | HEALTHY | $0.69 |
| **coder** | 2 | 1,107,041 | 32,374 | 12,296,934 | 91.7% | 176.7k | ELEVATED | $2.10 |
| **tester** | 2 | 130,703 | 2,066 | 477,210 | 78.5% | 39.1k | HEALTHY | $0.15 |
| **reviewer** | 2 | 565,367 | 6,498 | 3,622,994 | 86.5% | 70.4k | HEALTHY | $0.76 |
| **general** | 1 | 1,702,923 | 22,468 | 8,153,618 | 82.7% | 143.5k | ELEVATED | $2.04 |
| **Total** | **9** | **3,903,579** | **74,545** | **27,948,578** | **87.7%** | **176.7k** | **ELEVATED** | **$5.73** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 9 | 3,903,579 | 74,545 | 27,948,578 | 87.7% | $5.73 |
| **Total** | **9** | **3,903,579** | **74,545** | **27,948,578** | **87.7%** | **$5.73** |
