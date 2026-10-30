# AL-58 Iteration 02 Delta Plan: OpenCode Schema Naming

## Goal and boundary

Apply the user's naming steering to the approved iteration 01 implementation: make ownership explicit across shared code, schema family v1 code, and schema family v2 code. Every version-specific source and test filename must include its family version and responsibility; version-specific contents and symbols must also identify their family. Names such as `v2.go` and `session_message.go` are insufficient. This is a naming-only refactor, with references/tests updated as needed; preserve all behavior and iteration 01 acceptance criteria in [`../../CONTRACT.md`](../../CONTRACT.md) and [`../01/PLAN.md`](../01/PLAN.md).

The database detector identifies a storage schema family, not the exact OpenCode executable release. Use `v1`/`v2` or “schema family v1/v2” in code names and diagnostics. Keep `1.19.5` and `2.2.0` only in runtime-evidence/version contexts such as evidence fixtures and compatibility documentation; do not encode these releases into family detection, extraction symbols, or error context.

## Ownership map and target names

### Shared files

- `client.go` remains shared: database opening and connection lifecycle are family-neutral.
- Rename `schema.go` to `schema_family_detection.go`; move/rename its tests to `schema_family_detection_test.go` (or extract the detection cases from the current combined tests into that file). Detection classifies complete v1/v2 schema families and remains independent of extraction and resolver implementation.
- `extractor.go` remains shared for exported extraction entry points, family dispatch, shared fact assembly/helpers, and family-neutral steering-turn dispatch. It must not retain a v1 implementation merely because it was historically located there.
- `resolver.go` remains shared for public routing (`ResolveActiveRootSession`, `ResolveSessionByWorkspacePath`), shared validation/error declarations, and schema-family dispatch. It must not retain either family’s query implementation or v2-specific helpers.
- Keep genuinely shared tests in `extractor_test.go` where they exercise the public shared entry points or shared helpers. Move database-open/read-only and cross-family path normalization cases out of `v2_test.go` into shared test ownership; move schema classification cases into `schema_family_detection_test.go`. Shared file and symbol names must not imply v1 or v2 ownership.

### Schema family v1-only files

- Move the v1 session extraction implementation currently embedded in `extractor.go` into `schema_v1_session_extraction.go`; name its entry point `extractSchemaV1Sessions` and make v1 payload/row-only symbols explicit (for example, `schemaV1MessageData`, `schemaV1PartData`). Move its schema-specific steering-turn count query there as `countSchemaV1SteeringTurns`. Keep public dispatch in `ExtractSessions` and `ExtractSteeringTurns` in `extractor.go`.
- Move the v1 root-session and workspace-path SQL implementations currently embedded in `resolver.go` into `schema_v1_session_resolution.go`; use family-explicit helper names such as `resolveSchemaV1RootSession` and `resolveSchemaV1WorkspacePath`. The public shared resolver functions continue dispatching to this implementation.
- Place v1 extraction and resolver cases in `schema_v1_session_extraction_test.go` and `schema_v1_session_resolution_test.go`. Retain shared-entry-point tests in `extractor_test.go` only when their assertions are family-neutral or exercise dispatch across families.

### Schema family v2-only files

- Rename `v2.go` to `schema_v2_session_extraction.go` and `v2_test.go` to `schema_v2_session_extraction_test.go`. These files own the v2 session/message extraction implementation and its focused tests.
- Rename v2 extraction functions and payload/row-only symbols to state family ownership: use `extractSchemaV2Sessions` and names such as `schemaV2SessionRow`, `schemaV2MessageRow`, and `schemaV2MessagePayload`, rather than bare `extractV2Sessions`, `v2Session`, or `v2Payload`. Move its schema-specific steering-turn count query there as `countSchemaV2SteeringTurns`.
- Move the v2 root-session and workspace-path SQL helpers currently in `resolver.go` into `schema_v2_session_resolution.go`; use family-explicit helper names such as `resolveSchemaV2RootSession` and `resolveSchemaV2WorkspacePath`. Put their focused cases in `schema_v2_session_resolution_test.go`. The v2 extraction test file remains responsible for v2 extraction behavior.

### Shared symbol and diagnostic ownership

- Rename `familyV1`/`familyV2` to `schemaFamilyV1`/`schemaFamilyV2`, and `detectSchema` to `detectSchemaFamily`.
- Shared dispatch calls the family-specific extraction/resolution helpers using the detected schema family. Family-specific row, payload, and resolver symbols consistently carry `schemaV1` or `schemaV2` ownership.
- Make family-specific error context explicit, such as `OpenCode schema family v2: decoding assistant message …`; shared diagnostics should say “schema family” when referring to classification. Do not imply an exact binary release.

## Extraction, resolver, and test boundaries

- Extraction boundary: `extractor.go` validates the database, detects the schema family, handles shared empty-selection/dispatch behavior, and retains genuinely family-neutral helpers and normalized fact assembly. `schema_v1_session_extraction.go` and `schema_v2_session_extraction.go` own their respective database layouts, family-specific extraction representations, and steering-turn count queries. Shared accumulators/helpers may remain in `extractor.go` only if both implementations use them without embedding family-specific data semantics.
- Resolver boundary: `resolver.go` owns public API behavior, common input validation, family detection/dispatch, and shared errors. Each `schema_v1_session_resolution.go` / `schema_v2_session_resolution.go` owns only its family's SQL shape and query-specific helpers. Preserve routing priority and result/error behavior.
- Test boundary: preserve existing cases and assertions while moving family-specific cases alongside the implementation they exercise. Keep tests of public cross-family dispatch and common API behavior in `extractor_test.go`; detection-specific cases (including partial/mixed schemas and required-column omissions) belong in `schema_family_detection_test.go`. Ensure every family-specific test filename includes the corresponding version and responsibility.
- This plan specifies file/symbol ownership and names only. It does not call for new cases or changed assertions; the same existing coverage must remain after moves and reference updates.

## Acceptance criteria

1. Every v1-only and v2-only implementation or test filename includes its schema family version and responsibility. No bare `v1.go`, `v2.go`, or table-only filename remains for family-specific code.
2. Shared files (`client.go`, `schema_family_detection.go`, `extractor.go`, and `resolver.go`) contain shared lifecycle, family classification, public dispatch, or genuinely family-neutral helpers only; v1/v2 query and extraction implementations are owned by their corresponding versioned files.
3. Detection constants/functions, extraction/resolution functions, family-specific row/payload types, diagnostics, and focused tests communicate schema-family ownership consistently.
4. All moved/renamed references, test names, comments, and package-internal callers are updated consistently. Existing tests cover the same cases; no new behavior or release inference is introduced.
5. The diff changes names, organization, and references only: extraction, detection, resolver behavior, error meaning, CLI behavior, contracts, fixtures, and compatibility claims remain unchanged.
6. The change remains within `backend/internal/adapters/opencode` plus any references/tests required by those moves, follows Go/FCIS and error-context conventions in [`../../../../TECHNICAL_GUIDELINES.md`](../../../../TECHNICAL_GUIDELINES.md), and adds no domain/frontend changes.

## Grounding

Grounded in the frozen AL-58 contract and iteration 01 plan, all current adapter files (`client.go`, `extractor.go`, `extractor_test.go`, `resolver.go`, `schema.go`, `v2.go`, `v2_test.go`), technical guidelines, and merged `main` (`86fbcf1`, incorporated into the active ticket branch at `11d8d36`). Current `extractor.go` combines shared dispatch/helpers with the v1 extraction implementation and v1/v2 steering-turn queries; `v2.go` contains the v2 extraction implementation. Current `resolver.go` combines public schema-family routing with v2-specific resolution helpers and the v1 SQL implementation. Current schema symbols include `familyV1`/`familyV2` and `detectSchema`; v2-specific names/context include `extractV2Sessions` and “v2 message” errors. Preserve the contract's distinction between family detection and the pinned v1.19.5/v2.2.0 runtime evidence.

## Handoff

This iteration plans the naming refactor only. No application code, tests, evidence, or behavior are changed by this plan. Implementation requires the existing user confirmation gate.
