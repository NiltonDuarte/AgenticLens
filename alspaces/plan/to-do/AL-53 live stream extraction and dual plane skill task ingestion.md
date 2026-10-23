h3. Title: AL-53: Live Stream Extraction and Dual-Plane Skill and Task Ingestion

*TARGET*: AgenticLens (adapters / live / http / service / tests)

*CONTEXT*
 
AgenticLens employs a dual-plane architecture: historical offline workspaces ingest facts from `agentic_metrics.json`, while the Live Observability engine tails active OpenCode SQLite databases via `LiveService`. Currently, `internal/adapters/live/sqlite.go` extracts basic tool parts without parsing skill names or child session links. As a result, live sessions cannot display rich skill and task metadata in real time.

*GOAL*
 
Extend `LiveService` and `sqlite.go` to parse skill names and subagent task targets from streaming tool parts, populating `LivePartFact` and streaming enriched events to the frontend over SSE and REST endpoints.

*VERTICAL SLICE SCOPE*
 
* Domain / Live:
  - Add `SkillName *string` and `ChildSessionID *string` to `LivePartFact` in `backend/internal/domain/live.go`.
* Adapters / Live SQLite:
  - Update `QueryLiveTrace` and event polling in `sqlite.go` to unmarshal tool inputs for `skill` and `task` parts, populating `SkillName` and `ChildSessionID`.
* Live Service & HTTP:
  - Update `LiveEventTraceUpdate` broadcasting in `service.go` to convey enriched tool parts to connected browser clients.
  - Update `/api/live/traces/:sessionId` handler to serialize new fields.
* Quality:
  - Unit tests in `sqlite_test.go` and `live_handlers_test.go` verifying live extraction of skill names and child session targets.

*ACCEPTANCE CRITERIA*
 
* `LivePartFact` contains non-nil `SkillName` whenever a streaming tool part represents a `skill` call.
* Streaming `task` tool calls contain resolved `ChildSessionID`.
* Live SSE broadcast pushes enriched part facts to browser clients in real time.
* REST API `/api/live/traces/:sessionId` returns serialized skill and child session fields with backward-compatible JSON tags.
* `go test ./...` and `make` pass cleanly.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
