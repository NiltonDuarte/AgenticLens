# AgenticLens

> Full-Lifecycle Telemetry, Real-Time Observability, and Anomaly Detection for Autonomous Agentic Software Engineering.

AgenticLens is a specialized observability and engineering intelligence platform built for multi-agent software development workflows. Operating directly against OpenCode execution traces, AgenticLens transforms autonomous coding agents from unpredictable black boxes into observable, accountable, and high-velocity engineering swarms.

## Install AgenticLens with Your Coding Agent

Ask your coding agent to install and configure AgenticLens in the repository. For an OpenCode agent, use this prompt:

```text
Read https://raw.githubusercontent.com/NiltonDuarte/AgenticLens/main/.agents/skills/agentic-lens-setup/SKILL.md and follow it to install and configure AgenticLens in this repository. Explain the proposed changes and ask before modifying agent policy files.
```

The setup skill guides the agent through downloading the platform binaries, configuring the repository, and integrating the AgenticLens operator skill. Review the [Platform Setup Manual](docs/SETUP_MANUAL.md) for the complete setup process and manual installation options.

---

## Why AgenticLens?

Autonomous AI coding agents are transforming software engineering, but unmonitored agentic execution introduces critical failure modes:

- **Runaway Token Costs**: Autonomous agents can burn through hundreds of dollars in API credits within minutes if left unmonitored.
- **Context Window Saturation**: As conversational context expands, model reasoning decays, introducing subtle bugs and hallucinations.
- **Tool Execution Death Spirals**: Agents repeatedly retry broken bash commands or file operations instead of self-correcting.
- **Subagent Rework Loops**: Hierarchical agents (e.g. planner, coder, tester) bounce incomplete tasks back and forth in endless thrashing cycles.
- **Zero Engineering Visibility**: Traditional CI/CD and APM tools capture build status, but cannot measure agent reasoning effort, turn velocity, test-to-prod code ratios, or prompt cache efficiency.

AgenticLens solves this by providing real-time telemetry streaming, deterministic execution anomaly detection, and longitudinal cost ledgers inside a single, zero-dependency executable.

---

## Core Capabilities

### 1. Real-Time Execution Observability
- Stream live agent turns, active tool invocations, and token growth directly from OpenCode's local SQLite database.
- Monitor active multi-agent sessions with sub-second latency via built-in Server-Sent Events (SSE).
- Track subagent delegation transitions (Orchestrator -> Planner -> Coder -> Tester -> Reviewer) live.

### 2. Deterministic Anomaly Detection
An embedded pure-functional configurable rule engine flags dangerous execution patterns before budgets and contexts collapse:

- **Tool Retry Streaks (`TOOL_RETRY_STREAK`)**: Identifies repeated consecutive failures on the same tool (e.g. warning at 3 failures, critical alert at 5).
- **Elevated Tool Error Rates (`TOOL_HIGH_FAILURE_RATE`)**: Surfaces unreliable tools with failure rates (e.g. exceeding 20% to 50%).
- **Context Saturation Spikes (`CONTEXT_SATURATION_SPIKE`)**: Detects hazardous token deltas (>40k/turn) and context window exhaustion (e.g. >80% of model ceiling).
- **Subagent Thrashing Loops (`SUBAGENT_REWORK_LOOP`)**: Uncovers recursive handoff ping-pong between subagent roles.

### 3. Multi-Agent SDLC Analytics
- **Role Effort Distribution**: Break down token consumption, wall-clock time, and cost across specialized agent personas.
- **Model Economics & Cache Efficiency**: Track prompt cache hit ratios, input/output token balance, reasoning token share, and exact USD costs.
- **Code Velocity & Test Ratio**: Quantify test lines versus production lines delivered per session (`Test:Prod Ratio`), files touched, and iteration cycles.
- **Tool Health & Self-Correction**: Measure invocation counts, self-correction frequency, and tool success percentages.

### 4. Automated Session Ledgers
- Automatically compile audit-ready markdown telemetry blocks into session summaries (e.g. `SUMMARY.md`).
- Ensure every work session includes verifiable proof of agent reasoning, context health, anomaly status, and test coverage.

### 5. Universal Backfill and Schema Migrations
- Recompute historical workspace facts from git commit history and SQLite event traces.
- Multi-version schema migration engine with strict seven-point mathematical invariant verification.

### 6. Single Self-Contained Binary
- Written in Go with an embedded modern TypeScript/React web dashboard.
- Zero external runtime dependencies: no Node.js, no Docker, and no external database required in production.

---

## Visual Platform Showcase

The following captures demonstrate AgenticLens operating against real multi-agent engineering traces from 63 completed software development workspaces in this repository:

### 1. Fleet Overview & Anomaly Reliability Profile
Macro-level operational telemetry aggregating $520+ in model inference, 3.08B tokens, prompt cache efficiency, active anomaly rule distributions, and brittle tool rankings:

![AgenticLens Fleet Overview](docs/images/fleet-overview.png)

### 2. Ticket Forensic Autopsy & Context Saturation
Granular single-ticket autopsy (AL-4) measuring turnaround lead time versus active compute duration, test-to-production ratio, reasoning token volume, and context window saturation:

![AgenticLens Ticket Deep Dive](docs/images/ticket-deep-dive.png)

### 3. Chronological Multi-Agent Trace Waterfall
Interactive execution waterfall tracking turn-by-turn transitions between Orchestrator, Planner, Coder, and Tester roles, with sub-second execution timings and failure markers:

![AgenticLens Execution Trace Waterfall](docs/images/trace-waterfall.png)

### 4. Deterministic Anomaly Inspector
Contextual diagnostic drawer isolating tool execution failures, parameter payloads, subagent context, and error diagnostics:

![AgenticLens Anomaly Drawer](docs/images/anomaly-drawer.png)

### 5. Longitudinal Metric Progression
Temporal velocity tracking cumulative financial spend trajectories, token volume distributions, prompt cache hit rate baselines, and net code churn:

![AgenticLens Metric Progression](docs/images/metric-progression.png)

### 6. Cross-Ticket Benchmarking
Comparative matrix ranking tickets by active duration, turnaround lead time, iteration cycles, and cost:

![AgenticLens Ticket Comparisons](docs/images/ticket-comparisons.png)

---

## Powering the Agentic SDLC

AgenticLens brings rigor to the Autonomous Software Development Life Cycle:

```
+-----------------------------------------------------------------------------------+
|                           Autonomous Multi-Agent SDLC                             |
+-----------------------------------------------------------------------------------+
|                                                                                   |
|  [ Inception / PO ] ----> [ Plan ] ----> [ Code ] ----> [ Test ] ----> [ Review ] |
|         |                     |              |             |               |      |
|         v                     v              v             v               v      |
|  +-----------------------------------------------------------------------------+  |
|  |                           AgenticLens Telemetry Bus                         |  |
|  |  * Real-Time SQLite Polling                 * Anomaly Guardrails            |  |
|  |  * Role Effort Tracking                     * Token & Dollar Ledger         |  |
|  |  * Context Window Saturation Warning        * Test-to-Prod Code Facts       |  |
|  +-----------------------------------------------------------------------------+  |
|                                       |                                           |
|                                       v                                           |
|                   +---------------------------------------+                       |
|                   |  Observable Dashboard & Artifacts     |                       |
|                   |  - Embedded Web SPA (port 8085)       |                       |
|                   |  - Markdown SUMMARY.md Telemetry      |                       |
|                   |  - Automated Work Session Audit Trail |                       |
|                   +---------------------------------------+                       |
+-----------------------------------------------------------------------------------+
```

1. **Pre-flight & Inception**: Track architectural planning and contract formulation without runaway prompt spending.
2. **Implementation Guardrails**: Catch tool failure spirals and context saturation spikes before code corruption happens.
3. **Adversarial Verification**: Monitor testing velocity and ensure test-to-production ratios satisfy engineering standards.
4. **Handoff & Merge**: Stamp every work session with deterministic execution telemetry, anomaly verdicts, and cost ledgers.

---

## CLI Overview

AgenticLens provides a unified CLI for automation and local workflows:

```bash
# Start the web server and dashboard (default port: 8085)
./bin/agentic-lens server --port 8085

# Ingest active session telemetry and generate summary markdown block
./bin/agentic-lens generate --ticket-id AL-15 --ticket-type feat --json-output alspaces/AL-15/agentic_metrics.json --append alspaces/AL-15/SUMMARY.md

# Replace ses_abc123 with the Codex root session ID; child threads are discovered from it
./bin/agentic-lens generate --source codex --session-id ses_abc123 --code-ref HEAD --ticket-id AL-56 --ticket-title "Codex metrics research" --ticket-type chore --json-output alspaces/AL-56/agentic_metrics.json --append alspaces/AL-56/SUMMARY.md

# List all tracked workspaces with detected anomaly counts
./bin/agentic-lens --list

# Inspect detailed anomaly diagnostics for a specific ticket
./bin/agentic-lens --ticket-id AL-15 --anomalies

# Batch reconcile git facts and database telemetry across historical workspaces
./bin/agentic-lens backfill --source opencode --recompute-git --code-ref HEAD --recompute-db --migrate-schema --update-summaries

# Run stepwise schema migrations across workspace ledgers
./bin/agentic-lens migrate --backup --update-summaries
```

`agentic-lens init` creates shared `.agenticlens-config.yaml` and machine-local `.agenticlens-config.local.yaml`. The local file is ignored by Git; check `.gitignore` before committing and keep credentials out of both files. Configure `telemetry.default_source`, `telemetry.sources.opencode.database_path` (OpenCode SQLite), and `telemetry.sources.codex.home` (Codex rollout directory). The tracked `.agenticlens-config.local.example.yaml` provides sanitized portable defaults. `generate --source` overrides the local default; otherwise `telemetry.default_source` is used, then OpenCode. Source locations resolve from CLI path override, local harness path, legacy shared OpenCode `database.path`, then adapter defaults. A missing selected source is an error and never triggers a switch to the other harness.

Use `--nonce <literal>` to select the exact Codex rollout containing a unique run nonce, or `--session-id <thread-id>` to select a known thread. `--codex-home <path>` overrides local Codex home, `$CODEX_HOME`, and `~/.codex`.

Code metrics always compare an explicit Git ref with the configured `code_metrics.base_branch`; `generate` defaults to `HEAD`, and `backfill --recompute-git` requires `--code-ref`. Empty or invalid refs fail instead of trying to infer a branch or commits from the ticket ID. Comparisons include committed changes and exclude the uncommitted working tree. Codex session discovery requires `--nonce` or `--session-id`; ticket text and working-directory heuristics are not reliable session identifiers. Codex SDLC iterations count root-session user messages, not folders, commits, model requests, or child-agent prompts. Schema-v4 sessions own invocation records, per-session tool summaries, and retained events; cross-session totals are derived for display. Migration reports and removes legacy aggregate tool rows without reliable session attribution. Re-extract from source logs to recover those counts when available.

#### Recovering a legacy ledger with unscoped tool totals

Generation validates an existing output ledger before merging. Therefore `--json-output` must point to a new file for recovery; writing directly to the old unscoped ledger fails validation. Keep the old ledger as a backup, generate each known root session into the same new output path, verify the recovered session IDs and tool rows, then replace the old file. For Codex, selecting a root session also discovers its child rollouts. For OpenCode, repeat once per root session. Repeated generations to the new path merge session facts by `(session_id, tool_name)`.

```bash
cp workspaces/<TICKET_ID>/agentic_metrics.json workspaces/<TICKET_ID>/agentic_metrics.unscoped-backup.json

# Run once for each root session, always reusing this new output path.
./bin/agentic-lens generate --source codex --session-id <ROOT_THREAD_ID_1> --ticket-id <TICKET_ID> --ticket-title "<TICKET_TITLE>" --ticket-type chore --json-output workspaces/<TICKET_ID>/agentic_metrics.recovered.json
./bin/agentic-lens generate --source codex --session-id <ROOT_THREAD_ID_2> --ticket-id <TICKET_ID> --ticket-title "<TICKET_TITLE>" --ticket-type chore --json-output workspaces/<TICKET_ID>/agentic_metrics.recovered.json

# Check the distinct session IDs and per-session tool rows against the source sessions.
jq -r '.sessions[].id' workspaces/<TICKET_ID>/agentic_metrics.recovered.json | sort -u
jq '[.sessions[].tools[]] | length' workspaces/<TICKET_ID>/agentic_metrics.recovered.json

# Replace the active ledger only after the recovered contents are verified.
mv workspaces/<TICKET_ID>/agentic_metrics.recovered.json workspaces/<TICKET_ID>/agentic_metrics.json
```

Substitute `--source opencode` and each OpenCode root ID when recovering OpenCode data. Confirm the expected root and child session IDs (for Codex) and expected tool rows/counts before replacement. If source logs are unavailable, the old aggregate tool count cannot be split into session facts or recovered; keep the backup and do not fabricate session IDs.

---

## Deep Dive Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)**: Comprehensive deep dive into the Functional Core, Imperative Shell (FCIS) pattern, domain models, adapter layers, anomaly rule engine, and schema migration mechanics.
- **[CONTRIBUTING.md](CONTRIBUTING.md)**: Developer setup, local development with hot reload, testing guidelines, complete Makefile target reference, and release packaging.
- **[TECHNICAL_GUIDELINES.md](TECHNICAL_GUIDELINES.md)**: Coding standards, typing rules (zero `any`), data race prevention, and invariant enforcement.
- **[docs/SETUP_MANUAL.md](docs/SETUP_MANUAL.md)**: Complete guide for integrating AgenticLens into target repositories via autonomous agent skills or direct binary deployment.

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

Code classification is configured under `code_metrics` in `.agenticlens-config.yaml`:
`test_patterns` identifies tests and fixtures; `prod_patterns` and `include` identify production files. There are no implicit test directory, test filename, or production extension rules. Empty patterns leave files unclassified (`Other`), which can contribute to total lines without contributing to test/production subtotals.

Patterns containing `/` are relative to the repository root; a `**` path component matches zero or more directories. Patterns without `/` match basenames. Workspace exclusions and `exclude` win, followed by test patterns, then production/include patterns. The existing lockfile/dotfile exclusions apply to otherwise unmatched files. Both Git discovery modes monitor the required `code_metrics.source_roots` plus configured test, production, and include patterns, allowing fixtures outside source roots.
