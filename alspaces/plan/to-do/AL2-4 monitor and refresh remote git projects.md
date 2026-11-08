# AL2-4: Monitor and refresh remote Git projects

## Goal

Operators can monitor configured remote repositories and refresh them without manually cloning every project or interrupting current reads.

## Context

AL2-3 supplies local multi-project identity and scanning. This ticket adds remote retrieval and an HTTP registry/query interface that also serves local projects, enabling AL2-5’s dashboard.

## Acceptance criteria

* Configured remote projects or --project <id>=<url> accept validated HTTP/HTTPS/SSH Git locations. Use isolated shallow caches under the user cache (default ~/.cache/agenticlens/remotes/<project_id>/); credentials never appear in ledgers, API responses or diagnostics.
* GET /api/projects lists configured local/remote projects, connection state, workspace counts and last successful synchronization. Unknown versus unavailable projects are distinguishable.
* POST /api/projects/sync requests background refresh. Readers retain the last complete usable cache during refresh or network failure; failed/partial fetches do not replace it or advance the successful-sync timestamp. Failures remain visible.
* Workspace, fleet, comparison and progression queries accept project selection and aggregate only selected identities. Repeated project parameters and comma-separated values have equivalent meaning; omitted selection means all configured projects, invalid IDs fail clearly.
* Project isolation holds for duplicate ticket IDs, concurrent refresh/read and repository URLs containing credentials. Remote caches require no private source/capture receipts to expose committed telemetry.

## Implementation notes

Git/network/filesystem operations remain adapter-owned. Verify with temporary local bare repositories and controlled failures; do not require public-network access for regression checks. Remote read caches do not authorize mutation or commits in upstream repositories.

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md). [PO-19](../../PO-19/SPEC.md) preserves the earlier development reconciliation; historical observations require revalidation, not an assumption of current runtime support.
