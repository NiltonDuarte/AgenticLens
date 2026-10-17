h3. Title: AL-6: Versioned Schema Migration and Chronological Event Extraction

*TARGET*: AgenticLens (adapters | domain | scripts | tests)

*CONTEXT*

The initial `agentic_metrics.json` ledger (Version 1) records only coarse aggregated counts per invocation and aggregate tool call/error numbers. Detecting granular anomalies (such as consecutive tool failure streaks or sudden context window spikes) requires ordered event telemetry. We need schema versioning where omitted version numbers default to 1, and version 2 introduces an ordered chronological events array.

*GOAL*

Implement schema versioning across Go adapters, domain models, and the Python extraction script, supporting seamless deserialization of unversioned v1 ledgers as version 1 while emitting and parsing version 2 ledgers with chronological tool and turn events.

*VERTICAL SLICE SCOPE*

* Domain Core: Add `SchemaVersion` integer, `EventFact` struct (`Timestamp`, `SessionID`, `Turn`, `Role`, `EventType`, `ToolName`, `Status`, `ErrorMessage`, `DurationMs`, `ContextTokens`), and slice of `Events` in `AgenticFactLedger` (`internal/domain/models.go`).
* Boundary Adapter: Update `internal/adapters/schemas.go` with `SchemaVersion` handling. In `UnmarshalJSON`, if `schema_version` is missing, null, or 0, set `SchemaVersion = 1`. Add `EventFactDTO` and wire to domain converter.
* Python Extractor: Update `scripts/extract_agentic_metrics.py` to extract chronological tool calls and compaction events ordered by timestamp from the OpenCode SQLite DB, outputting `schema_version: 2` and the `events` array.
* Quality: Go unit tests in `internal/adapters` asserting that unversioned JSON payloads correctly decode as Version 1, and v2 payloads decode chronological event streams. Tests in `scripts/` verifying extraction output.

*ACCEPTANCE CRITERIA*

* Legacy `agentic_metrics.json` without `schema_version` deserializes successfully with `SchemaVersion == 1`.
* Version 2 ledgers with `schema_version: 2` deserialize with populated, chronologically sorted `Events`.
* `scripts/extract_agentic_metrics.py` generates valid `schema_version: 2` JSON containing ordered tool calls with execution status and error text when available.
* All existing aggregations in `internal/domain` continue to pass without regression.
* `go test -race ./...` and `go vet ./...` pass with zero errors.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Go type checking clean with zero any or empty `interface{}`.
* Telemetry tested with `agentic-lens` CLI.
