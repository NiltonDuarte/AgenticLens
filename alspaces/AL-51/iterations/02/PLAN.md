# AL-51 Iteration 02 Plan

## Goal

Address the iteration 01 review findings without expanding the approved AL-51 scope: align OpenCode v1 skill failure classification with v2, and prove malformed or incomplete optional tool inputs do not disrupt extraction.

## Review-driven work

1. **Normalize v1 tool failure status**
   - In `backend/internal/adapters/opencode/schema_v1_session_extraction.go`, normalize status casing and classify both `error` and `failed` as unsuccessful tool outcomes, consistent with the v2 extractor.
   - Preserve the existing output-signature classification and unrelated invocation/event behavior.
   - Add a focused synthetic SQLite regression case in `backend/internal/adapters/opencode/schema_v1_session_extraction_test.go` proving that `failed` and uppercase `ERROR` skill invocations produce unsuccessful `SkillFact` records, including any available error detail.

2. **Cover malformed and incomplete optional inputs**
   - Extend OpenCode schema-family tests in `backend/internal/adapters/opencode/schema_v1_session_extraction_test.go` and `schema_v2_session_extraction_test.go` for malformed JSON, a missing skill name, and missing task fields.
   - Assert optional skill/task facts are skipped or retain only valid source fields according to the current extractor contract, while ordinary session, invocation, and event facts remain available.
   - Add a live boundary malformed-input case in `backend/internal/adapters/live/sqlite_test.go` if the existing live fixtures do not already prove malformed optional input is ignored safely.

## Acceptance checks

- v1 skill status matching is case-insensitive and treats both `failed` and `error` as failure outcomes.
- Synthetic v1 and v2 cases cover malformed JSON and absent required skill/task input fields without losing unrelated extracted facts.
- Live malformed optional input does not fail trace retrieval or stream serialization.
- Current schema version remains 4; no migration or unrelated tool classification changes are introduced.
- Focused OpenCode and live adapter tests, followed by `make check`, pass; record exact outcomes in `iterations/02/REVIEW.md` and the ticket summary.

## Boundaries

- Keep this iteration within the existing AL-51 contract and data shape.
- Do not add UI presentation or alter migration behavior.
- Preserve additive compatibility for legacy ledgers and live trace payloads.
