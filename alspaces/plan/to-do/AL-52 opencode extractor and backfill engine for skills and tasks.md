h3. Title: AL-52: OpenCode Extractor and Backfill Engine for Skills and Tasks

*TARGET*: AgenticLens (adapters / opencode / backfill / cli / tests)

*CONTEXT*
 
OpenCode persists skill invocations and subagent delegations inside SQLite `part` records as tool calls where `tool == "skill"` and `tool == "task"`. Currently, `ExtractFacts` in `backend/internal/adapters/opencode/extractor.go` simply increments generic counters in `toolMap["skill"]` and `toolMap["task"]`, discarding the input parameters (`{"name": "..."}` for skills, `{"subagent_type": "...", "description": "..."}` for tasks) and failing to resolve the spawned child `session_id`.

*GOAL*
 
Enhance `ExtractFacts` and the backfill engine to parse skill names, errors, caller roles, and correlate task tool calls with spawned child sessions, populating `Skills`, `TaskLinks`, and enriched `EventFact` instances.

*VERTICAL SLICE SCOPE*
 
* Adapters / OpenCode Extractor:
  - In `extractor.go`, parse `p.data` when `tool == "skill"`. Unmarshal JSON arguments to extract the skill `name`. Populate `SkillFact` instances aggregating total invocations, execution errors, first turn loaded, and caller role.
  - In `extractor.go`, parse `p.data` when `tool == "task"`. Unmarshal `subagent_type` and `description`. Correlate the parent tool call to the spawned child session via parent session hierarchy and timestamp bounds.
  - Populate `EventFact.SkillName` for skill events and `EventFact.ChildSessionID` for task events.
* Adapters / Backfill:
  - Update `internal/adapters/backfill/engine.go` to re-extract skills and task links when running database-backed backfill migrations.
* CLI:
  - Ensure `./bin/agentic-lens generate` populates Schema v4 `skills` and `task_links` arrays in `agentic_metrics.json`.
* Quality:
  - Integration tests in `extractor_test.go` and `backfill_test.go` with mock SQLite databases containing skill and task parts.
  - Assert robust fallback handling when tool input contains malformed or non-standard JSON payloads.

*ACCEPTANCE CRITERIA*
 
* ExtractFacts extracts unique skill names into `[]domain.SkillFact` with accurate call counts, error counts, caller role, and first turn.
* Task tool calls link to child session IDs in `[]domain.TaskLinkFact`.
* Chronological `EventFact` stream contains non-nil `SkillName` on skill tool calls.
* Backfill engine updates v1/v2/v3 ledgers to v4 with newly extracted skill facts when SQLite database is accessible.
* `go test ./...` and `make` pass cleanly.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
