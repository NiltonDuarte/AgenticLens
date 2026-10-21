h3. Title: AL-32: Telemetry Event Pruning, Database Message Stripping, Agent Data Investigation, and Ledger Backfill

*TARGET*: AgenticLens (domain | adapters | cli | migration | tests)

*CONTEXT*

Currently, `agentic-lens generate` dumps raw chronological event streams into `workspaces/<TICKET_ID>/agentic_metrics.json`. Persisting raw messages and error strings from the database creates critical security vulnerabilities (permanently leaking environment secrets, API keys, credentials, local filesystem paths, and customer data into git history) and repository bloat (thousands of lines of routine completed tool calls and turn messages).
Attempting regex-based secret redaction on database messages is inherently fragile and leak-prone. Therefore, the committed Git plane must strictly enforce zero database message retention: no prompts, no assistant messages, and no raw database error strings are ever written to Git.
Failed tool calls in committed ledgers must be recorded as bare failure indicators (`tool_name`, `status: "error"`, turn number, timestamp, duration) without any contextual message from the database. The full diagnostic context is reserved strictly for local hydration from the developer's SQLite database. Additionally, an investigation step is required to determine whether clean, structured metadata can be obtained directly from the agent itself (e.g. self-reported exit codes, error classifications) versus relying solely on local hydration.
Finally, historical workspace ledgers must be retroactively migrated and backfilled to purge routine events and strip all database message text across existing workspaces.

*GOAL*

Enforce zero database message persistence in committed ledgers by pruning routine events, stripping all database error and prompt messages from persisted ledgers (saving bare error indicators without DB context), investigating agent-reported metadata pathways, and implementing a schema migration backfill to sanitize historical workspace ledgers.

*VERTICAL SLICE SCOPE*

* Domain Core (`internal/domain/`, `internal/domain/migration/`):
  - Event Fact Serialization: ensure committed `EventFact` models omit database messages (`ErrorMessage: nil` or omitted when writing to Git plane; no prompts or reasoning strings).
  - Bare Error Indicator: preserve only structural metadata for failed tool calls (`ToolName`, `Status: "error"`, `Turn`, `Timestamp`, `DurationMS`).
  - In-Memory Aggregations: compute all token totals, active durations, cost, tool health rollups, and anomaly evaluations over the complete raw event stream in memory *before* pruning/stripping events for ledger persistence.
  - Migration Step: register schema migration step `StepV2ToV3Sanitized` in `internal/domain/migration/registry.go` to prune routine completed events and strip legacy database message strings from existing ledgers while asserting bit-identical aggregate metrics.
* Boundary Adapter & CLI (`internal/adapters/`, `internal/adapters/backfill/`, `cmd/agentic-lens/`):
  - Event Filtering: in `internal/adapters/opencode/extractor.go` and `cmd/agentic-lens/main.go`, filter persisted events to strictly:
    - Tool calls with `Status == "error"` (stripped of raw database error messages / text).
    - Compaction events (`EventType == "compaction"`).
    - Events directly associated with detected domain anomalies.
    - Omit all routine `completed` turn messages and successful tool calls.
  - Investigation Task: document and evaluate options for capturing structured error metadata directly from the agent runtime (e.g. exit codes, error enum types, agent self-reported diagnostics) without reading raw error text from the database.
  - Ledger Backfill Runner: wire migration step into `agentic-lens migrate` and `internal/adapters/backfill/engine.go` with `--dry-run` support, verifying historical ledgers drop to <5 KB.
* Quality & Tests:
  - Table-driven tests asserting that zero message strings, prompts, or raw database error outputs are present in serialized ledger JSON.
  - Tests verifying bare error events retain tool name, turn, and error status correctly.
  - Table-driven tests verifying all in-memory calculations (tokens, cost, anomalies, tool health) match raw values exactly.
  - End-to-end backfill test on historical workspace fixtures verifying file size reduction (<5 KB) and zero residual database text.
  - Static checks and test suite pass with zero `any` and zero emojis.

*ACCEPTANCE CRITERIA*

* `workspaces/<TICKET_ID>/agentic_metrics.json` contains strictly bare error events, compactions, and anomaly-associated events with zero raw database messages, prompts, or error strings.
* No regex sanitization is performed; database messages are omitted completely rather than masked.
* In-memory token totals, active durations, costs, tool execution metrics, and anomaly flags match exact raw values.
* Investigation report or architecture note evaluates pathways for obtaining structured data directly from the agent versus dynamic local hydration.
* `agentic-lens migrate` successfully prunes routine events and strips legacy database messages across existing workspace ledgers.
* File size of generated and migrated `agentic_metrics.json` ledgers is reduced to <5 KB for typical sessions.
* Zero `any` and zero emojis across all code and tests.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
