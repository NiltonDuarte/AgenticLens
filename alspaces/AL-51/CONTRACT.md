# AL-51 Contract: Skills and Task Delegation Facts

## Problem and outcome

AgenticLens does not expose OpenCode `skill` invocations or `task` delegations as first-class facts. This ticket adds optional historical ledger facts and live trace attributes so existing consumers continue to read ledgers that omit them.

## Compatibility and boundaries

- Extend the Go domain ledger and live part model additively, with optional JSON fields omitted when empty. Mirror those contracts in the TypeScript domain types.
- Preserve the repository's active `domain.CurrentSchemaVersion` value of 4 (`backend/internal/domain/constants.go`). The supplied ticket prose names version 3, but the repository has already advanced to v4; this work must not downgrade it, add a migration, or alter migration behavior.
- Keep domain structs and any domain transformations pure. SQLite reads, JSON decoding, session correlation, SSE/API propagation, and malformed-input handling stay in adapters.
- Preserve existing OpenCode schema-family handling, legacy ledger decoding, live trace shape, and stream behavior. Missing, malformed, or unavailable optional skill/task data must not break ordinary extraction or traces.
- Emit child session identifiers only when a concrete parent-child relationship is present in OpenCode records. Do not infer a child from a task description or ordering.
- Do not introduce UI presentation for these facts in this ticket.

## Data contract

`AgenticFactLedger` gains optional `skills` and `task_links` collections. A skill fact records its name, turn and message indices, caller role, timestamp, success state, and optional error. A task link records subagent type, description, parent turn, timestamp, and optional child session identifier, duration, and cost. `LivePartFact` gains optional skill name, child session ID, and subagent type attributes.

The JSON field names and scalar types must match the approved AL-51 specification and corresponding TypeScript interfaces. Empty collections remain omitted so old serialized ledgers retain their established shape. Older ledgers without these fields decode as empty/absent optional data.

## Acceptance boundaries

1. Go and TypeScript contracts represent the optional historical and live fields consistently.
2. OpenCode extraction recognizes `skill` input name and `task` input subagent type/description, with turn/message/role/timestamp and completion status derived from the owning invocation records.
3. Task child IDs and optional completion metadata are populated only when supported by source records; otherwise they remain absent.
4. Live trace parts carry skill and task attributes through the existing trace and event serialization paths.
5. Schema version remains 4, no migration runs, legacy ledgers still decode, and existing unrelated facts/stream fields are unchanged.
6. Deterministic Go tests cover legacy and augmented serialization, both OpenCode schema families with synthetic SQLite data, live extraction, task-child correlation, and malformed/missing input. Frontend tests cover legacy and augmented domain/live payload typing or parsing.
7. All applicable Go and frontend checks pass; no `any`/`interface{}` additions or emojis are introduced.

## Grounded implementation surface

- Go domain contracts: `backend/internal/domain/models.go`, `backend/internal/domain/live.go`; current version: `backend/internal/domain/constants.go`.
- OpenCode SQLite extraction: `backend/internal/adapters/opencode/extractor.go` dispatches to schema-family extractors, including `schema_v1_session_extraction.go` and `schema_v2_session_extraction.go`.
- Live SQLite boundary and trace assembly: `backend/internal/adapters/live/sqlite.go` (`RawPartRecord`, `FetchSessionTrace`) and `backend/internal/adapters/live/service.go` (`parsePartRecord`, `GetSessionTrace`, stream publication).
- Frontend contracts: `web/src/domain/types.ts` and `web/src/domain/live.ts`; telemetry re-exports: `web/src/types/telemetry.ts`.
- Existing test surfaces: corresponding domain model tests, OpenCode schema extraction tests, live adapter SQLite/service tests, `web/src/domain/events.test.ts`, and `web/src/domain/live.test.ts`.
