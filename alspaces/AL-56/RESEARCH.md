# AgenticLens and OpenAI Codex integration review

Review date: 2026-10-25
Codex source inspected at upstream `main` on 2026-10-25. Earlier parser work was based on a shallow clone at `970111f6cb62cc02fe36a82ea7d71b37a2889c97`; the findings below also include current upstream source checked during this research pass.

## Executive recommendation

Add Codex as a separate ingestion adapter and retain the current domain, ledger, aggregation, and dashboard where their meanings match. For the user's existing interactive Codex sessions, start with a `CODEX_HOME` history reader; it is the nearest passive equivalent to OpenCode's local database reader. Add two higher-fidelity paths for new sessions:

1. **History and current sessions:** stream rollout JSONL under `CODEX_HOME/sessions/` and `archived_sessions/`, associate child threads, and project safe facts into `AgenticFactLedger`. Keep Codex's state SQLite database out of the contract; it is a private discovery cache, while rollouts are the source of session history.
2. **Highest-fidelity opt-in:** when a session is launched with `CODEX_ROLLOUT_TRACE_ROOT`, consume Codex's local rollout-trace bundle (prefer reduced semantic state and never ingest raw payload files by default). This has first-class inference, tool, compaction, and multi-agent lifecycle objects and is the strongest match for OpenCode invocation/tool facts. It is a sensitive diagnostic facility, not default telemetry.
3. **Live managed runs:** have AgenticLens launch `codex exec --json`, stream its structured turn/item events, and track process exit status. This is the clearest live contract when AgenticLens owns the run.
4. **Live interactive integration:** use app-server JSON-RPC when AgenticLens is the Codex client or is deliberately attached to a managed server. It does not provide ambient observation of unrelated Codex clients.

The best practical parity design is a passive history adapter plus an optional live/traced mode. Make the history adapter useful by itself, and expose which metrics were directly observed versus unavailable. `CODEX_ROLLOUT_TRACE_ROOT` can close gaps in tool/runtime detail for sessions where it is enabled, but requiring it would miss existing sessions and may collect sensitive diagnostic payloads.

## How AgenticLens works today

AgenticLens has a good adapter seam. `backend/internal/adapters/opencode` resolves and extracts offline OpenCode facts. `backend/internal/adapters/live` opens the OpenCode SQLite database read-only and polls active sessions. Both map records into shared `backend/internal/domain` structs and are consumed by CLI/HTTP/dashboard layers. Git history is a separate adapter, and JSON ledgers are workspace artifacts.

The dependency that does not transfer is the source connector: the OpenCode extractor and live watcher are written around OpenCode tables and record semantics. The pure aggregation/anomaly domain can transfer if the new source can produce the same facts. Important current metrics include token/cache/reasoning usage, tool lifecycle/errors, session timing, roles/models, and git-derived code facts. Several current value concepts are OpenCode-specific and should not be inferred from Codex fields unless they have the same semantics.

## Codex source surfaces

### 1. `codex exec --json` event stream

The current CLI parser defines `--json` as “Print events to stdout as JSONL”; `--experimental-json` remains an alias in the inspected source. The event contract in `codex-rs/exec/src/exec_events.rs` includes thread started, turn started/completed/failed, item started/updated/completed, and top-level error events. Completed turn usage includes input, cached input, cache-write input, output, and reasoning-output token counts. Items include agent messages, reasoning summaries, command executions, file changes, MCP calls, collaboration calls, and web searches.

This is the best live path for batch or managed Codex jobs that AgenticLens starts itself. The stdout stream gives start/end transitions and turn usage. The process exit status should be retained as a separate run outcome. `codex exec --json` does not by itself discover or subscribe to existing TUI sessions. Its JSON event types currently include message and tool payloads, so the adapter must explicitly project only safe facts and must not persist raw prompts, assistant text, command arguments, MCP arguments, outputs, or reasoning summaries by default.

Current-source references: [exec event definitions](https://github.com/openai/codex/blob/main/codex-rs/exec/src/exec_events.rs), [JSON event processor](https://github.com/openai/codex/blob/main/codex-rs/exec/src/event_processor_with_jsonl_output.rs), and [CLI argument definition](https://github.com/openai/codex/blob/main/codex-rs/exec/src/cli.rs).

### 2. App-server JSON-RPC

The app-server protocol has typed notifications for `thread/started`, `turn/started`, `turn/completed`, `item/started`, `item/completed`, and `thread/tokenUsage/updated` (plus other thread and item notifications). Thread items include command execution with duration/status/exit code, MCP calls with server/tool/status, collaboration tool calls with recipient thread IDs and agent statuses, file changes, and agent messages. Thread token usage includes input, cached input, cache-write input, output, reasoning output, totals, and optional context-window size.

This offers stronger live lifecycle fidelity and can model tool start/end, failures, context saturation, and some multi-agent relationships. It is suitable if AgenticLens is the orchestrator/client of Codex. The app-server is not an ambient telemetry daemon: observation is bounded by the app-server/client session relationship. The external adapter must follow app-server protocol versioning and gracefully handle unknown notifications. Some API areas are explicitly marked unstable/experimental in the protocol.

Current-source references: [app-server item notifications](https://github.com/openai/codex/blob/main/codex-rs/app-server-protocol/src/protocol/v2/item.rs), [turn notifications](https://github.com/openai/codex/blob/main/codex-rs/app-server-protocol/src/protocol/v2/turn.rs), and [thread notifications and usage](https://github.com/openai/codex/blob/main/codex-rs/app-server-protocol/src/protocol/v2/thread.rs).

### 3. Persisted rollout JSONL and state index

Codex source discovers rollouts under `CODEX_HOME/sessions/YYYY/MM/DD/` and `CODEX_HOME/archived_sessions/`. Each rollout is JSONL with a timestamp and serialized item; `codex-rs/rollout/src/lib.rs` owns the parser, and compression code handles compressed rollout files. Session metadata includes thread/session IDs, cwd, source, CLI version, provider, agent role/path/nickname, context-window metadata, and git metadata. Rollout items may include event messages, completed turn items, token usage records, and other internal records.

The Go adapter cannot depend on Codex's Rust structs, so it needs permissive wire decoding and fixtures tied to captured/synthetic examples. Unlike `codex exec --json`, rollout files are a persistence format rather than a standalone downstream analytics API. It should be treated as private/unstable, with parse coverage tracking and `codex_version`/schema hints when available. The state SQLite index is useful for discovery and rollout-path resolution, but the rollout remains the canonical detailed fact source; the hybrid recommendation below uses the index as a catalog, not as the metrics authority.

Current-source references: [rollout parsing and directories](https://github.com/openai/codex/blob/main/codex-rs/rollout/src/lib.rs), [rollout discovery and summaries](https://github.com/openai/codex/blob/main/codex-rs/rollout/src/list.rs), [rollout metadata/backfill](https://github.com/openai/codex/blob/main/codex-rs/rollout/src/metadata.rs), and [Codex thread metadata](https://github.com/openai/codex/blob/main/codex-rs/state/src/model/thread_metadata.rs).

### 4. Opt-in local rollout trace bundles

Current Codex source includes `codex-rs/rollout-trace`, enabled by `CODEX_ROLLOUT_TRACE_ROOT`. It writes a local bundle with an ordered event spine, manifest, payload files, and optional reduced `state.json`. The reduced graph models threads and parent-child edges, turns, inference calls with model/provider and usage, tool calls with lifecycle/status/timing, compactions, code-mode cells, terminals, and information-flow edges. `codex debug trace-reduce <bundle>` replays the trace into semantic state.

This is the closest Codex source to OpenCode's separate invocation and tool facts. It can provide stronger provenance for models, inference boundaries, runtime-only tools, and subagent relationships than normal rollouts. Codex explicitly describes this as opt-in local diagnostics rather than uploaded telemetry. Its raw bundle can contain prompts, responses, tool input/output, terminal output, and paths. Therefore an AgenticLens integration should default to derived facts only, avoid reading `payloads/`, and put trace storage behind explicit opt-in. The reduced state schema is still Codex-owned and should be version-checked.

Current-source references: [rollout-trace design and privacy](https://github.com/openai/codex/blob/main/codex-rs/rollout-trace/README.md), [trace event envelope](https://github.com/openai/codex/blob/main/codex-rs/rollout-trace/src/raw_event.rs), [reduced inference calls](https://github.com/openai/codex/blob/main/codex-rs/rollout-trace/src/model/conversation.rs), and [tool runtime model](https://github.com/openai/codex/blob/main/codex-rs/rollout-trace/src/model/runtime.rs).

## Metric mapping and limits

| AgenticLens concept | Codex evidence | Feasibility and caveat |
|---|---|---|
| Session/thread identity, start/end, cwd, source | Session metadata; `thread.started`; turn lifecycle | High. Map Codex thread to session, preserve source and parent/thread relationships when present. |
| Model/provider | Session metadata and thread settings; configuration | Medium-high. Session model may vary by turn; capture observed model per turn where possible, not just initial metadata. |
| Input/output/cached/reasoning tokens | `turn.completed` usage, rollout token usage records, app-server usage, trace inference usage | High for core counts. Preserve per-turn or per-inference deltas; never sum cumulative `total_token_usage` snapshots alongside per-response records. Cache-write is present in current `codex exec --json` usage and trace models, but may be absent from some persisted token-count records or upstream paths. |
| USD cost | Not generally supplied by local Codex CLI event stream | Low/conditional. Codex subscription usage and API per-token billing are not equivalent. Only compute estimates for known API providers/pricing with explicit provenance; do not imply exact spend for ChatGPT-plan usage. |
| Tool call counts, success/failure, duration | Item lifecycle, command execution, MCP, collaboration items; explicit `ToolCall` execution windows in traces | High in live/app-server and trace modes. History can support many completed item facts, but may not retain all intermediate timestamps or a matched start event. Normalize tool class and preserve source event identity. |
| Tool arguments/output | Present in structured events/rollouts | Technically available, sensitive. Exclude from default metrics; optionally store redacted/debug data only under explicit user configuration. |
| File changes/code churn/test-to-prod | Codex file-change items plus Git analyzer | Medium. Use Git as source of truth for line counts and test/prod classification; Codex file-change events are attribution hints and may be incomplete. |
| Subagent/delegation effort | Child session metadata has parent-thread/agent-role data; app-server collab items and trace graph have explicit parent-child interaction edges | High when metadata is present. Discover related rollout files and correlate by Codex thread ID, not prompt text. Trace mode preserves richer spawn/task/result/close edges. |
| Skills | Skill use may appear as user input/protocol or extension activity; no universal `skill` tool contract equivalent | Low-medium. Need source-specific detection and validation; do not count prompt text matches as reliable invocation facts. |
| Context utilization/saturation | Context window and token usage fields exist in some session/thread data | Medium-high. Use the observed window value where available and label missing values; distinguish cumulative thread usage from current prompt context after compaction. |
| Tool retry anomalies | Ordered tool item lifecycle | High. Reuse pure anomaly rules after source-neutral normalization. Define what counts as a retry per Codex item identity. |
| Role/persona comparisons | Codex agent-role/nickname metadata and collab states | Medium. Values differ from AgenticLens's planner/coder/tester taxonomy; preserve raw normalized role and do not infer fixed persona names. |
| Wall-clock active duration | Turn/item timestamps and process lifecycle | High for turns/processes; less exact for pauses across a long-lived thread. Keep active time separate from ticket lead time. |

### Price and cost proxies

The rollout does not record actual dollars charged or subscription-credit debits. It does record token usage and model information, which supports a **rate-card estimate** if the estimate is kept distinct from observed spend.

- **Best Codex-native proxy: estimated credits.** For credit-billed Codex usage, apply the official Standard-speed per-model rates to each inference: `(uncached input × input rate + cached input × cached-input rate + output × output rate) / 1,000,000`. The current GPT-6 Luna rates are 2.5 / 0.25 / 12.5 credits per million input / cached-input / output tokens. Codex credit billing has no separate cache-write charge. The official [Codex pricing page](https://learn.chatgpt.com/docs/pricing) says API-key use follows API pricing and that credit rates alone do not determine included subscription usage. Therefore this is a rate-card equivalent, not proof of the user's plan debit or subscription dollar cost.
- **Conditional proxy: API-equivalent USD.** If the run is known to use API-key billing, use the model's API rate card. Current GPT-6 Luna Standard short-context rates are $0.10 / $0.01 / $0.50 per million input / cached-input / output tokens ([official API pricing](https://developers.openai.com/api/docs/pricing), [model pricing detail](https://developers.openai.com/api/docs/models/gpt-6-luna)). This is still an estimate unless reconciled with the actual API usage/billing record. Long-context, processing mode, regional uplift, and cache-write rules can change the rate. Apply those per inference, not to a whole-session aggregate, because the long-context threshold is per request.
- **Unavailable from this local evidence:** the selected rollout and local thread databases do not establish whether billing used a ChatGPT plan or API key, the applicable plan/contract, per-request speed tier, or actual debited credits/dollars. Do not convert a credit equivalent to USD or label API-equivalent USD as actual cost.

For implementation, keep `cost_usd` as zero/unavailable when actual billing is not known. If adding estimates, expose separate fields such as `estimated_codex_credits` and `estimated_api_cost_usd`, alongside `estimate_kind`, pricing source/retrieved date, model, speed, and per-inference token counts. Do not change the meaning of schema-v3 `cost_usd` to hold a proxy. The session ledger's zero `cost_usd` remains correct; calculating a session-level estimate from its aggregated totals would conceal per-request context and pricing-mode uncertainty.

## Recommended architecture

- Introduce `internal/adapters/codex` alongside `opencode`; do not rename the existing OpenCode adapter. Use one mapper to convert either a rollout or trace-derived facts into `AgenticFactLedger`, then call the existing `domain.DeriveMetricsLedger` so the AgenticLens dashboard works with no parallel aggregation implementation.
- For parity with OpenCode offline extraction, scan `CODEX_HOME/sessions` and `archived_sessions`, decode plain and compressed rollouts, correlate child threads, and associate a Codex thread to a workspace/ticket through cwd plus an explicit mapping. Do not silently assign a ticket from the session ID.
- For parity with OpenCode live observability of existing sessions, tail only active rollout files and recompute a bounded in-memory summary. Codex does not expose an ambient SQLite change feed suitable for substituting into the current OpenCode watcher.
- Offer opt-in trace bundles as a high-fidelity mode and parse only an allowlisted set of metadata/events from `state.json`; do not read raw bundle payloads for standard metrics.
- Keep source parsing and filesystem traversal outside the domain. Convert Codex events to existing source-neutral facts in an adapter mapper.
- Add additive source/provenance metadata to facts only after the prototype reveals the actual cross-source gaps. Likely additions include cache-write tokens, Codex source/version, turn/thread IDs, and a metric confidence/provenance marker for cost.
- Refactor the live polling mechanism behind a source interface only when the app-server or process stream is implemented. The existing live service embeds assumptions about SQLite polling; duplicating the UI API payload and SSE event names initially keeps this spike small.
- Keep expensive or volatile facts (USD estimates, content-bearing tool payloads, inferred skills) clearly marked and optional.
- Preserve Git analysis as the source for delivered code churn; join Codex records to repositories/tickets by cwd, time, and explicit mapping rather than relying on Codex transcript text.

## Local CODEX_HOME and SQLite audit

Schema inspection identified the following Codex persistence surfaces. Database access used read-only SQLite connections with `mode=ro` and `PRAGMA query_only=ON`. Published findings omit machine-specific row counts, session identifiers, paths, literal nonces, and payload contents.

| Source | Available structure | Metrics use |
|---|---|---|
| `sessions/**/rollout-*.jsonl` and archived rollouts | Session metadata, token records, turn and tool events | Canonical detailed fact stream; parse incrementally and retain allowlisted facts. |
| `state_*.sqlite` | Thread IDs, rollout paths, cwd, source, model/provider, timestamps, archive metadata, spawn edges | Discovery catalog. `tokens_used` is the last observed usage, not an aggregate. Do not use `has_user_event` as a complete eligibility signal without validation. |
| `thread_history_*.sqlite` | Projected items, turns, statuses, durations, per-thread materialization cursors | Optional derived index. Check projection freshness against the rollout; item JSON may contain transcript/tool payloads. |
| `logs_*.sqlite` | Runtime diagnostic records | Exclude from ordinary metrics ingestion; diagnostic text can contain sensitive information. |
| `goals_*.sqlite` | Goal-specific records and budgets | Optional goal metadata, not a complete session metrics source. |
| `memories_*.sqlite`, `queue_*.sqlite` | Memory and queued-work records | Not metrics sources; avoid scanning content for session metrics. |

This differs from OpenCode's current AgenticLens path. AgenticLens opens OpenCode's `opencode.db` read-only and derives facts from `session`, `message`, and `part`; its live watcher polls the same database for changes. Codex has SQLite databases, but their responsibilities are split: `state_5.sqlite` is primarily discovery/metadata, `thread_history_1.sqlite` is a materialized history view, and the rollout JSONL is the best detailed canonical source. Current Codex source calls thread history a rebuildable SQLite view and warns that it may lag the durable rollout after projection failure ([state runtime](https://github.com/openai/codex/blob/main/codex-rs/state/src/runtime.rs), [thread history materializer](https://github.com/openai/codex/blob/main/codex-rs/thread-store/src/local/thread_history_materialization.rs), [live writer](https://github.com/openai/codex/blob/main/codex-rs/thread-store/src/local/live_writer.rs)). Upstream's `ThreadMetadata.tokens_used` is explicitly last-observed usage, not an aggregate ([ThreadMetadata](https://github.com/openai/codex/blob/main/codex-rs/state/src/model/thread_metadata.rs)).

For the nonce-matched current thread, read-only aggregation of 415 `token_usage_record` entries in its rollout yielded 55,838,282 total input tokens, including 55,058,048 cached input; uncached input is therefore 780,234. Output was 230,970 and reasoning output 108,076. AgenticLens cache hit rate is derived as cached input / total input, which is 98.6% for this snapshot; Codex does not persist that percentage directly. The history DB had 20 completed turns totaling 3,039,382 ms (about 50m39s) and one in-progress turn with no completed duration. This sum is elapsed turn span including tool waits, not pure model-compute time. The snapshot is specific to this local thread and may change as the session continues.

### Recommendation for OpenCode-style extraction

Use a hybrid read-only adapter:

1. Discover eligible sessions from `state_5.sqlite` metadata and resolve their current rollout paths. Query only IDs, times, source, model/provider, cwd, history mode, archive state, and Git metadata; never fetch display/content fields such as `title`, `preview`, `first_user_message`, or rollout bodies from the database catalog.
2. Treat rollout JSONL as the authoritative detailed fact stream. Stream it to obtain token/cache/reasoning usage, compaction, response/item events, and exact Codex-native fields; use file identity/metadata and partial-tail handling for live sessions.
3. Optionally use `thread_history_1.sqlite` as a low-latency index for turns and tool lifecycle. Check its per-thread byte cursor against the selected rollout's length and compare latest projected timestamps. If it is behind or mismatched, read the rollout suffix/full source rather than silently accepting stale DB metrics.
4. Tail active rollout JSONL files for passive live monitoring; use app-server notifications when AgenticLens owns the Codex client. Do not poll `logs_2.sqlite` as a substitute for protocol events.

The database filenames are runtime-versioned (`state_5`, `logs_2`, etc.) and schemas are Codex-owned migrations, so direct SQLite access is not the same stable contract as Codex's app-server protocol. Open read-only with WAL-aware SQLite settings; never run Codex migrations or writes from AgenticLens. Detect schema/table/column presence and retain JSONL fallback. The app-server itself may expose more stable live lifecycle data than scraping these private databases when AgenticLens can own or attach to that client.

## Probe result

The Go adapter parses rollout JSONL and projects privacy-minimized facts into the existing domain: thread/session identity, model, per-response token usage, turn timing, command lifecycle outcomes, paired custom tool calls, tool duration, and compaction events. It derives the schema-v3 ledger through `domain.DeriveMetricsLedger`. Codex input tokens include cached input, so the adapter subtracts cache reads before populating AgenticLens uncached input and keeps cached input in its separate field. Cache-write counts and context capacity remain source-summary fields because schema-v3 has no matching fact fields. A real local rollout showed that `CommandExecution` items carry command status/duration and `response_item` carries call/output pairs; tests now cover both shapes without retaining payload content.

CLI output and exported facts use `domain.PruneEventsForPersistence`: preserve errors, compactions, and events in anomaly-linked session/turns, then strip content fields. Successful tools from an anomalous turn may remain under this shared policy. Event `context_tokens` uses the last observed request's input count (or `last_token_usage.input_tokens` on the fallback path), never cumulative session usage. The initial probe used cumulative usage in event context, producing false context-saturation anomalies and retaining excessive successful events; this mapping has been corrected. Aggregate usage and tool totals still come from complete source facts before pruning.

`agentic_metrics.json` is the committed session ledger generated from the real local Codex rollout using `generate --source codex --nonce`. It contains observed session/model/token/tool facts and pruned diagnostic events. Ticket metadata was supplied as AL-56. A literal nonce lookup correlated the rollout thread ID with this conversation; details are in `DATA_PROVENANCE.md`. The raw rollout remains local, and parser fixtures remain synthetic for repeatable tests.

The CLI has an explicit `--source codex` discovery path that scans `CODEX_HOME/sessions` and `archived_sessions`, accepts `.jsonl` and `.jsonl.gz`, resolves by nonce, explicit thread ID, or ticket text, and includes descendants linked by `parent_thread_id`. Ticket matching prefers the current working directory and otherwise selects the most recently modified matching thread; if ticket text is absent, cwd fallback is allowed only when it identifies one root thread. Raw rollout lines are scanned in memory for matching only and are not retained in generated ledgers. The CLI does not yet tail active rollouts or join Codex-derived Git facts beyond the existing Git analyzer.

The CLI integration preserves the OpenCode default and keeps source selection explicit: `agentic-lens generate --source codex ...`. `CODEX_HOME` comes from `--codex-home`, then the environment, then `~/.codex`. The implementation uses filesystem discovery rather than Codex's versioned SQLite catalog; this avoids schema coupling but currently scans rollout files to resolve ticket text. `--nonce` takes precedence over ticket matching and is the most precise selector.

## Suggested delivery sequence

1. Add bounded tailing for active sessions without loading large histories into memory.
2. Improve ticket-to-workspace correlation and measure scan performance for large CODEX_HOME histories.
3. Evaluate opt-in reduced rollout traces as a higher-fidelity source; add managed `codex exec --json` and app-server collectors when lifecycle ownership is clear.

## Open product decisions for later

- Which session-to-ticket association should be default when several Codex rollouts share a cwd or one session works across tickets? For this user's research output, use the confirmed ticket ID AL-56.
- Is local transcript parsing acceptable given that rollout files contain raw prompts, responses, tool arguments, and command output?
- Whether to add separately labeled Codex credit-equivalent and API-equivalent estimates; neither should be represented as actual spend without billing evidence.
- Which workspaces should Codex threads map to when a thread's cwd changes or one Codex thread spans several tickets?

## Validation boundary

This review inspects Codex source; it does not claim a stable public contract for every persistence field. Parser tests validate mapping assumptions against synthetic fixtures plus a real local rollout extraction. Real-session validation ledgers and their summary are committed; raw rollouts remain local and automated tests use synthetic fixtures. Later integration tests should use deterministic local fixtures or a Codex mock and should not send a model request by default.


### Child rollout replay boundary

A subagent rollout may begin with child metadata, replay parent records with rewritten timestamps, then resume child-owned records at `thread_settings_applied` for the child thread. The parser preserves initial child metadata and uses that explicit ownership marker instead of timestamps. If the marker is absent, ambiguous continuation is excluded and a warning is emitted. This is safer than charging parent usage to the child, but can undercount unusual or future rollout formats; the warning must be surfaced to operators.
