# AL-56 Summary

Research branch: `research/codex-openai-metrics`.

## Result

AgenticLens can derive meaningful metrics from a Codex rollout without keeping transcript payloads. The implementation adds a Codex JSONL adapter that projects observed thread/model/token usage, turn timing, tool lifecycle outcomes, and compaction events into source-neutral facts, then calls the existing schema-v3 metrics derivation pipeline.

The `Codex Import` view lets a user select one uncompressed `.jsonl` rollout (up to 64 MiB), provide the AgenticLens ticket ID/title, preview calculated metrics, and download a standard `agentic_metrics.json` ledger. The upload is parsed in memory. The response omits prompts, assistant/reasoning text, command arguments, and tool input/output.

`agentic_metrics.json` (gitignored) is the local session ledger generated from the real local Codex rollout using `generate --source codex --nonce`. It contains observed session/model/token/tool facts and pruned diagnostic events. Ticket metadata was supplied as AL-56. A literal nonce lookup correlated the rollout thread ID with this conversation; details are in `DATA_PROVENANCE.md`. The raw rollout remains local, and parser fixtures remain synthetic for repeatable tests.

The real rollout exposed a mapping detail: current records carry command lifecycle facts in completed `CommandExecution` items and custom tool calls in paired `response_item` records. The parser handles both while retaining no command/input/output content. It also applies a turn's `turn_context` model to its earlier `task_started` event, avoiding a spurious `default` model invocation.

CLI ledgers and import downloads now apply the shared event-retention policy: errors, compactions, and events from anomaly-linked turns. Codex event context uses the last request's input count; using cumulative session usage had caused false context anomalies and excessive event retention. Full tool/token totals are preserved before pruning. Regression tests cover request-level and cumulative-fallback token records, retained anomaly/error/compaction events, and HTTP download pruning.

## Source research

The passive Codex history path is rollout JSONL under `CODEX_HOME`; it is the nearest analogue to AgenticLens's OpenCode history/live readers. The CLI now supports `generate --source codex`, discovers sessions from active and archived rollout directories, resolves by nonce/session ID/ticket text, and includes child threads linked through rollout metadata. The dashboard file picker remains available for manual single-rollout extraction. Opt-in rollout traces offer richer inference, tool, compaction, and child-thread lifecycle data but carry sensitive payloads. `codex exec --json` is the clearest source for live runs AgenticLens launches. App-server notifications are suitable when AgenticLens owns or explicitly attaches to the Codex client, not for ambient collection from unrelated interactive sessions. Full comparisons and source links are in `RESEARCH.md`.

The local database audit found that `state_5.sqlite` can discover threads and rollout paths, while `thread_history_1.sqlite` offers a useful but rebuildable structured projection of turns/items/status/timing. The rollout JSONL remains the detailed canonical input for tokens and fallback when the projection is behind. `logs_2.sqlite` is diagnostic rather than a metrics event store; goals/memories/queue databases are not general session sources.

## Limits

- Automatic discovery scans rollout files to match ticket text; nonce or explicit session ID gives more precise selection. It supports `.jsonl.gz` but does not tail active rollouts. Ticket matching scans raw JSONL content in memory without retaining it.
- Cost is unavailable in the observed rollout data. Schema-v3 stores numeric zero; the preview explains that this does not mean the run was free.
- A rate-card proxy is feasible: estimated Codex credits for credit-billed use, or API-equivalent USD only when API-key billing is known. Neither is actual spend, and subscription usage cannot be reliably priced from tokens alone. Keep estimates separate from schema-v3 `cost_usd`; details and current official rates are in `RESEARCH.md`.
- The single-file probe does not join AgenticLens Git analysis, so its code facts are zero. Codex transcript/file-change data is not treated as Git truth.
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
- Real-session validation artifacts remain local. Only synthetic fixtures are committed.


## Publication privacy review

Real session IDs, usage snapshots, timestamps, and the generated metrics table are excluded from the PR. The local ledger and `SESSION_METRICS.local.md` are gitignored. Published notes describe extraction methods without session identifiers, literal correlation nonces, or machine-specific database inventories. Uploaded working-directory paths are excluded from API responses; transcript and tool payloads are tested for omission.
