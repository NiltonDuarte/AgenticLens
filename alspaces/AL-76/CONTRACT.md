# AL-76 technical contract

Approved scope source: `alspaces/plan/to-do/AL-76 compact ledger json formatting and format command.md`. The user explicitly stated “plan is pre-approved. full send it”; this satisfies the implementation approval gate for the specification. Baseline: `122e983f8c365a2aa248b1b6d492252555e922ae`, initially clean working tree, ticket branch `feat/AL-76-compact-ledger-format`. Workspace root is configured as `alspaces`.

## Behavior

Ship one deterministic JSON-aware layout for persisted ledger writers and a source-independent `format` CLI command. Expanded containers use two spaces and one final newline. Compact recognized tool records, invocation tokens, cost_evidence and supported evidence objects, and root/session anomalies including nested context. Match structural paths deliberately, including supported historical root-owned shapes; unrelated same-named keys retain expanded layout.

Formatting changes whitespace outside string literals only. Preserve every token lexeme, member/array order, unknown member, null, absent field, schema version, timestamp and fact. Reject malformed JSON, duplicate decoded member names, unsupported shapes and missing explicit targets without replacing affected originals. Do not reconstruct existing bytes through domain DTOs, floating-point objects, source extraction, hydration, migration or calculations.

Require exactly one of `--all`, `--ticket-id`, `--ticket-ids`; allow explicit `--workspaces` or the normal configured resolver. Trim comma-separated IDs, reject empty items, deduplicate and prevent traversal. `--dry-run` and `--check` are mutually exclusive and never write ledgers or change timestamps. Check differences return nonzero with distinct diagnostics from input/I/O errors. Default writes changed targets only. Report changed, unchanged and failed targets deterministically; explicit missing targets fail, empty all succeeds explicitly.

Restrict physical paths to the selected workspace root and expected agentic_metrics.json targets. Coordinate supported aliases and deduplicate physical targets. Validate selection and parse selected inputs before writing; acquire existing canonical publication ownership before reads and hold it through replacement, preventing stale reads from overwriting a concurrent publisher. Preserve permissions and the existing recoverable publication/journal contract. Report partial results on later I/O failure without claiming fleet-wide atomicity.

The baseline has retired capture_report.json: preserve obsolete sidecars without recreating them or claiming refreshed coverage. Reject interrupted publication journals with actionable recovery guidance unless existing authorized recovery can safely be reused. No automatic digest-pair maintenance is required for retired reports.

## Architecture and scope

Pure lexical formatting belongs in a focused serialization adapter, free of I/O and source dependencies. Filesystem selection, locking, validation and publication remain adapters; CLI owns flags, help and exit/report behavior. Domain calculations, runtime capture behavior, frontend, schemas and release pins remain unchanged. Use strict concrete types, no any/interface{}, no new runtime dependency and no emojis.

All persisted ledger writers must share layout, including SaveFactLedgerPreservingHistory, backfill, migration and the existing computed-metrics export writer SaveMetricsLedger where applicable. Generation and automatic publication already route through SaveFactLedgerPreservingHistory; preserve their existing validation and fact semantics.

## Delivery boundaries

Exercise only disposable fixture ledgers; do not mass-format repository history. Document opt-in use and the invoked release's policy with no hooks, CI mandate, downloads or pin updates. Add meaningful serializer, real-filesystem, CLI and publication regressions using self-contained fixtures. Complete focused review, integrated make, lifecycle artifacts, final actual-harness telemetry, commits, push and PR. Review approval does not authorize merging.
