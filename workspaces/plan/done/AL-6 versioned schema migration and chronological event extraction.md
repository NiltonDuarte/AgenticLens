h3. Title: AL-6: Chronological Event Extraction and Schema Evolution in Go

*TARGET*: AgenticLens (adapters | domain | cli | web | tests)

*CONTEXT*

Python has been completely removed from the project, and active session telemetry is handled natively by the Go application (`cmd/agentic-lens generate`, `cmd/agentic-lens backfill`, and `internal/adapters/opencode`). While workspace ledgers (`agentic_metrics.json`) are already at `schema_version: 2` (with migration metadata introduced in AL-12), they currently record only coarse aggregated counts per invocation and aggregate tool call/error numbers. Detecting granular anomalies (such as consecutive tool failure streaks, turn thrashing, or context window spikes planned in AL-7) requires ordered event telemetry. We need native Go event extraction that builds an ordered chronological `events` array while maintaining seamless backward compatibility with existing ledgers that lack `events`.

*GOAL*

Implement chronological event extraction and schema models in the native Go engine (`internal/adapters/opencode`, `internal/domain`, `cmd/agentic-lens`), supporting seamless deserialization of existing ledgers while extracting, emitting, and parsing chronologically ordered event streams (`EventFact`).

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Add `EventFact` struct (`Timestamp`, `SessionID`, `Turn`, `Role`, `EventType`, `ToolName`, `Status`, `ErrorMessage`, `DurationMs`, `ContextTokens`) in `backend/internal/domain/models.go`.
  - Add `Events []EventFact` to `AgenticFactLedger` in `backend/internal/domain/models.go`.
* Boundary Adapter:
  - Add `EventFactDTO` in `backend/internal/adapters/schemas.go` and implement bidirectional mapping with `domain.EventFact`.
  - Update `AgenticFactLedgerDTO` with `Events []EventFactDTO` (`json:"events,omitempty"`).
  - Ensure backward compatibility in `internal/adapters/repository.go` and `schemas.go`: ledgers omitting `events` deserialize cleanly with an empty or nil slice.
* Go SQLite Extractor & CLI:
  - Update `backend/internal/adapters/opencode/extractor.go` to extract chronological tool calls, turn messages, and compaction events ordered by timestamp from OpenCode SQLite DB (`session`, `message`, `part` tables).
  - Wire extracted events to `cmd/agentic-lens/main.go` (`generate` command) and `backend/internal/adapters/backfill/engine.go` (when `--recompute-db` is executed).
* TypeScript Domain Parity:
  - Update `web/src/domain/types.ts` to include `EventFact` and optional `events?: readonly EventFact[]` on `AgenticFactLedger`.
* Quality:
  - Go unit tests in `internal/adapters/opencode` asserting accurate extraction and chronological ordering of tool executions and error captures from SQLite fixtures.
  - Go unit tests in `internal/adapters` validating backward-compatible decoding of ledgers without `events` and round-trip encoding of v2 ledgers with `events`.
  - `go test -race ./...` and `go vet ./...` pass with zero errors.

*ACCEPTANCE CRITERIA*

* Legacy `agentic_metrics.json` without `events` deserializes successfully without regression.
* Ledgers with `events` deserialize with populated, chronologically sorted `Events`.
* Native Go extraction (`agentic-lens generate`) produces valid `agentic_metrics.json` containing ordered `events` with tool names, execution statuses, duration, and error messages when present.
* `agentic-lens backfill --recompute-db` populates `events` for tickets with sessions found in the OpenCode SQLite database.
* All existing domain aggregations continue to pass without regression.
* `go test -race ./...` and `go vet ./...` pass with zero errors.
* Zero Python dependencies or references.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Go type checking clean with zero any or empty `interface{}`.
* Telemetry tested with `agentic-lens` CLI.
