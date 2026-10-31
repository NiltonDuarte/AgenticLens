# AL-51 Iteration 01 Plan

## Goal

Implement optional skill and task delegation facts across the Go ledger, OpenCode SQLite extraction, live trace parsing/streaming, and TypeScript domain contracts while preserving existing v4 and legacy-ledger behavior.

## Grounded work sequence

1. **Domain and serialization contracts**
   - Add `SkillFact` and `TaskLinkFact` and optional ledger slices in `backend/internal/domain/models.go`; add optional attributes to `LivePartFact` in `backend/internal/domain/live.go`.
   - Keep `CurrentSchemaVersion` at the repository's current value, 4, in `backend/internal/domain/constants.go`. Do not implement a schema migration.
   - Verify the adapters' DTO/conversion path preserves the new ledger facts, including `AgenticFactLedgerDTO` and domain conversion in `backend/internal/adapters`; ensure absent optional fields remain compatible.
   - Add Go serialization tests for old ledgers and populated/empty optional collections.

2. **Historical OpenCode extraction**
   - Extend the schema-family extraction paths dispatched by `backend/internal/adapters/opencode/extractor.go`, grounded in `schema_v1_session_extraction.go` and `schema_v2_session_extraction.go`, to capture skill and task tool inputs and completion outcomes with deterministic ordering.
   - Follow the extraction call sites in `backend/cmd/agentic-lens/main.go` and `backend/internal/adapters/backfill/engine.go` so extracted root-ledger facts are propagated into persisted ledgers. Preserve existing extraction APIs where practical; if a signature must change, update every production and test call site.
   - Correlate child sessions only through explicit OpenCode parent/session metadata. Missing or malformed optional input must not discard existing invocation/session facts.
   - Extend relevant schema-family tests with synthetic SQLite records for skill success/failure, task fields, optional child correlation, missing fields, and stable indexes/order.

3. **Live SQLite trace and stream propagation**
   - Extend the boundary parsing in `backend/internal/adapters/live/sqlite.go` and `backend/internal/adapters/live/service.go` (`parsePartRecord`, trace assembly, and stream publication) to populate the new `LivePartFact` values from tool input and known parent-child records.
   - Keep trace and event output backward compatible when these fields are absent. Do not change unrelated tool classification or snippet behavior.
   - Add synthetic SQLite live tests covering skill calls, task calls with and without known child sessions, and serialized trace/stream payloads.

4. **TypeScript contracts and defensive consumers**
   - Mirror the optional ledger and live-part fields in `web/src/domain/types.ts` and `web/src/domain/live.ts`, exported through the existing `web/src/types/telemetry.ts` surface as appropriate.
   - Find ledger consumers and use empty-array fallbacks for optional `skills` and `task_links` before iteration.
   - Extend `web/src/domain/events.test.ts` and `web/src/domain/live.test.ts` for legacy payloads and augmented ledger/live payloads.

## Acceptance checks

- Go JSON round trips omit absent optional fields and preserve populated skill/task facts; current schema version stays 4.
- OpenCode extraction tests prove accurate name/type/description, caller role, parent turn, message index, timestamp, success/error, and known child session association across supported schema families.
- Live adapter tests prove `skill_name`, `subagent_type`, and available `child_session_id` survive both trace retrieval and stream serialization.
- Legacy fixtures without the new fields continue through Go DTO/migration/backfill boundaries and TypeScript readers without errors.
- Run focused Go packages and frontend tests first, then `make check` (the ticket's full acceptance gate). Record commands and outcomes in the iteration review/summary.

## Out of scope

- Schema version changes or migrations.
- Skill/task UI, analytics, or presentation beyond defensive contract consumption.
- Fabricating child session links, duration, or cost when OpenCode does not record them.
