# Dashboard Reference

Start the embedded dashboard with `agentic-lens server`. The current CLI defaults to port `8085` (or valid `PORT`); `--port` overrides it. Open `http://localhost:<port>/`. The server embeds the dashboard in the Go binary, so a separate Node runtime is not needed.

## Fleet Overview `/`

Use for a cross-ticket summary and to find a workspace:

- Aggregate tokens, costs, workflow duration, and average iterations.
- Context health distribution across healthy, elevated, bloated, and critical.
- Search/filter ticket inventory by ticket ID, title, status, or prefix.

Open a ticket row for its ticket detail page. Investigate the underlying ledger when a displayed value is surprising; the UI presents persisted metrics and does not repair them.

## Live Observability `/live`

Use for active local OpenCode sessions:

- Active session cards show agent personas, model, turns, tokens, and elapsed duration.
- Trace viewer streams tool calls, parameters, and output.
- Alerts surface stalls, high tool error rates, and context exhaustion.

This view depends on the local OpenCode database selected by server config or `--db-path`. An empty view can mean no active sessions or unavailable/wrong local source; inspect local config and server startup output before concluding there is no activity. Do not expose trace contents beyond the user's requested diagnostic scope.

## Ticket Detail `/ticket/:id`

Use for one ticket's history:

- Iteration timeline and role/turn progression.
- Tool invocation counts and failures.
- Model, token, and cost breakdown.
- Artifact viewer for configured workspace documents, including `CONTRACT.md` and `SUMMARY.md` when present.

For migration diagnosis, inspect the underlying `agentic_metrics.json` migration status and pending steps. The ticket view is not a team-wide migration coordinator.

## Comparisons `/comparisons`

Compare selected tickets for token efficiency, cost, tool reliability, and code changes. Treat differences in project configuration, model pricing, and ticket scope as potential confounders.

## Metric Progression `/progression`

Use for longitudinal trends such as cost per ticket, turn velocity, test-to-production ratio, and context health. Verify the time range and ledger coverage before interpreting a trend.

## Dashboard Troubleshooting

1. Confirm the server process is running and use its actual port.
2. Confirm `.agenticlens-config.yaml` loads and `workspaces.dir` points to an existing directory.
3. For live data, confirm `telemetry.sources.opencode.database_path` or `database.path`, or pass `--db-path`; paths are machine-local.
4. Distinguish an empty live feed from missing historical ledgers: live data comes from the local source, historical views come from workspace ledgers.
5. Treat all config edits, server restarts, or ledger changes according to the user's authorization.
