h3. Title: AL-76: Compact ledger JSON formatting and add a selective format command

*TARGET*: AgenticLens (Ledger Serialization / Filesystem Adapters / CLI / Tests)

*Status*: Done
*Work type*: feat
*Blocked by*: None
*Blocks*: None
*Origin*: AL-68 follow-up: reduce unnecessary lines in persisted metrics and provide an explicit way to reformat selected existing ledgers. Implementation belongs to a separate AL-76 development session.

*CONTEXT*

The standard JSON pretty-printer expands small tool, token, cost-evidence and anomaly records across many lines. This makes agentic_metrics.json harder to scan and produces unnecessarily long diffs. Generation, backfill and migration currently have separate serialization call sites, so a format change limited to one writer would alternate layouts between commands.

Existing ledgers also need a source-independent formatting operation. Regenerating or migrating them to change whitespace can alter telemetry, require unavailable source stores, normalize unknown fields or change completion timestamps.

*GOAL*

As an operator, I want consistently readable metrics with selected records on one line, and a format command that applies the current layout to all ledgers or an explicit ticket selection, without changing their facts, schema or observation boundary.

*VERTICAL SLICE SCOPE*

* Serialization: Define one deterministic JSON-aware formatting policy shared by ledger writers and the format command. The current formatting policy means the policy shipped by the invoked installed/pinned AgenticLens release; formatting must not fetch a newer release or change the project version pin. Keep the outer ledger/session/invocation structure indented. Compact each tool record, invocation tokens record, invocation cost_evidence record and each root/session anomaly record, including its nested context, onto one line. Preserve an invocation evidence object when present in a supported input family and apply the corresponding compact layout. Do not add absent fields or replace cost_evidence with a newly invented evidence field.
* Writer integration: Ground the work in adapters/repository.go SaveFactLedger, backfill/engine.go and backfill/migration_runner.go serialization. Generation and automatic publication must reach the shared formatter through their existing writers. Identify any additional persisted-ledger writer before implementation; presentation/API JSON unrelated to persisted ledgers is outside scope.
* Existing-file adapter: Reformat ledger bytes without rebuilding facts through a domain DTO, deriving metrics, querying a source, hydrating events, migrating a schema or recalculating anomalies. Preserve unrecognized members and exact numeric/string values. Reuse established output locking, physical-path validation and safe publication behavior.
* CLI: Add format with workspace-wide, single-ticket and ticket-set selection, dry-run and check modes. Document behavior in command help and the operator CLI reference.
* Quality: Extend coherent serializer/repository and CLI tests with self-contained ledgers and real filesystem boundaries. No domain calculation or frontend feature is required; existing readers/dashboard behavior must remain compatible.

*CLI CONTRACT*

* agentic-lens format --all: format every discovered agentic_metrics.json under the configured workspaces.dir.
* --workspaces DIR: explicitly select the workspace root for any mode; otherwise use the normal configuration resolver.
* --ticket-id ID: format exactly one ticket ledger.
* --ticket-ids ID1,ID2: format an explicit set of ticket ledgers. Trim list separators, reject empty items and deduplicate IDs before resolving files.
* Exactly one selection mode is required: --all, --ticket-id or --ticket-ids. Missing/conflicting selectors fail before any write; do not implicitly select the latest ticket or all workspaces.
* --dry-run: report which selected files would change without modifying files or timestamps.
* --check: report files that differ from the current formatting policy and exit nonzero if any require formatting, without modifying files. Distinguish formatting differences from input/I/O failures in diagnostics. Dry-run and check are mutually exclusive.
* Default operation writes only selected ledgers that differ. Report changed, unchanged and failed targets deterministically. Success requires every selected target to complete; empty --all selection is a successful no-op with an explicit zero-ledger result. An explicitly requested missing ticket is an error.

*ACCEPTANCE CRITERIA*

* Use two-space indentation for the expanded structure, deterministic compact separators inside selected records and one trailing newline. Each tools array element occupies one line; each invocation tokens, cost_evidence or supported evidence object occupies one line; each anomaly array element occupies one line. Keep their enclosing arrays/objects expanded and readable. Do not compact arbitrary unrelated objects merely because a key name matches at another location. Cover canonical session-owned data and supported legacy root-owned shapes deliberately.
* Preserve all ledger values, schema_version, field presence, nulls, unknown cost evidence, array order, member order and unknown fields. Preserve exact number literals and string contents, including integers above floating-point exact precision, decimal/exponent forms and escaped braces/quotes/newlines. Whitespace inside string literals is data and must not be changed. Formatting is JSON-aware, never regex replacement of arbitrary serialized text or a float-based generic object round trip.
* The operation does not require Codex/OpenCode records, a capture registry or network access. It does not change started_at, completed_at, updated_at, code facts, invocation totals, anomalies, skills or task_links. It does not regenerate SUMMARY.md or introduce a formatting metadata field or schema bump.
* All relevant persisted-ledger writers use the same policy. Repeated formatting is byte-identical, and already formatted ledgers are not rewritten. Subsequent generation/backfill/migration preserves the layout while independently retaining each command's existing data semantics.
* Selection resolves only expected ledger files within the explicitly selected workspace root. Never rewrite arbitrary JSON, backups, review files or configuration. Prevent traversal and unsafe physical/symlink escapes; coordinate aliases of the same selected ledger rather than formatting it twice. A ticket name, branch name or cwd must not silently broaden the selection.
* Validate selection and parse selected input files before applying changes. Invalid JSON, duplicate/ambiguous member definitions that prevent safe preservation, unsupported ledger shapes and missing explicit targets produce actionable errors without replacing the affected original or being counted as successful. Do not create missing ledgers or silently skip requested failures.
* Acquire existing publication ownership before reading/replacing a target so concurrent automatic/manual publishers cannot be overwritten from a stale pre-lock read. Preserve existing permissions and supported alias semantics. Each replacement is atomic/recoverable under the existing writer contract; no claim of one atomic transaction across the entire workspace is required. If a later I/O failure leaves earlier targets formatted, report that partial result and return nonzero without damaging originals.
* If the implementation baseline still supports a digest-paired capture report, formatting must not silently invalidate its ledger association. Maintain an already valid supported pairing transactionally by updating only its byte-digest reference, or reject that target unchanged with a precise unsupported/stale-pair diagnostic. Do not create a report, refresh coverage or claim a new observation boundary. If AL-68 has retired that sidecar by implementation time, verify the retired-artifact compatibility policy instead of reintroducing it. Shared writer changes must preserve existing publication journal/recovery guarantees.
* Verify all selection modes, explicit/configured roots, duplicate and missing IDs, invalid/empty selections, dry-run/check no-write behavior, malformed/unsupported input, unchanged-file idempotence, safe failure and concurrent publication. Use a nonempty fixture containing nested token/evidence/tool/anomaly records, unknown fields, null costs, large/exponent numbers and escaped strings. Assert both the requested line layout and value/lexeme preservation; meaningful formatting tests must fail against the old fully expanded printer.
* Document that this is an opt-in invocation, not an automatically installed Git hook or mandatory project CI check. No mass reformat of unrelated committed ledgers is part of implementation delivery; exercise the command on disposable fixtures and only reformat real ticket ledgers when explicitly requested.

*COORDINATION*

* No blocking dependency is required. Coordinate persistence/automatic-publication seams with AL-68 and optional field-output behavior with AL-66; formatting must preserve whatever facts/categories those commands intentionally emit. Do not implement either ticket's unrelated requirements.
* Keep pure formatting decisions free of filesystem/source work and retain I/O in adapters/CLI. Do not add unnecessary domain structs, frontend controls or a new runtime dependency.

*DEFINITION OF DONE*

* Execute in a separate AL-76 agentic-lens-dev session with the normal scope approval, implementation, tests and focused review. This follow-up definition does not start AL-76 implementation or create its developer workspace/CONTRACT.md.
* Deliver the shared serialization policy, all three selection modes, dry-run/check behavior, safe source-independent rewriting and updated help/operator guidance end to end.
* Pass appropriate formatter/CLI/concurrency regressions and the required integrated make gate. Preserve schema/read compatibility and existing publication/source-preservation invariants. Follow strict typing, FCIS, zero emojis and self-contained fixture requirements.
* Retain concise finding/reviewer/model and verification summaries in the ticket artifacts; do not commit raw reports, duplicated manifests or build-log JSON archives.
