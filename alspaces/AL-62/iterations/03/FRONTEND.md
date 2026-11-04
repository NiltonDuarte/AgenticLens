# AL-62 iteration 03 frontend implementation

Approved scope: CONTRACT.md current iteration and iterations/03/PLAN.md; baseline 5e4eb34019370150786085d69c47987b328e4e79. Frontend coder owns web/, this report, and (by explicit parent transfer) backend/internal/adapters/http/*_test.go. Shared backend, schema, ticket, and tester changes are preserved. No commits, publication, or telemetry performed by this subagent.

## Delivered behavior

Current schema 6 types and HTTP validators accept exactly the lean skill and delegation members, preserve required null unknowns, reject retired root/session fields (even empty/false), reject rich occurrence fields and older schemas, and retain invocation financial validation. Owner selectors pair lean rows with the enclosing session ID only and preserve repeated equal rows. No source/provenance or canonical occurrence identity is invented.

The skills/delegations panel retains session filters, keyboard disclosure/row access, timestamps, and exact child navigation. Unknown role/type/time values are visibly unavailable; known child IDs remain readable when the thread is absent. Removed columns are execution status, error details, turn indexes, delegation duration/cost, and rich source metadata. Fleet adoption retains unique/ranked activation counts; comparison skill filtering and unrelated tool outcomes/financial metrics remain.

Source traces are separate transient responses with unavailable, local SQLite, local Codex or mixed native source attribution. Successful empty hydration remains complete; pending, failed, partial and unavailable states carry distinct copy and preserve stored skills, delegations, metrics and anomaly summaries. No saved-event fallback, union/reconciliation, or lean-to-event synthesis remains. A separate recorded-summary disclosure keeps every saved anomaly summary inspectable even after complete-empty source hydration. Recorded inspection stays bound to its stored summary; source trace associations can refresh complete owners while retaining unavailable and unknown owners. Badges use actual source tool arguments or transient native skill_name / proven child_session_id metadata. The live inspector continues using its existing native part metadata.

## Assertion transfers and retired oracles

| Previous coverage | Current recipient / rationale |
| --- | --- |
| Skill success weighting and unknown-outcome denominator in aggregations.test.ts / FleetOverview.test.tsx | Metric retired by approved contract. Removed only outcome/rate assertions. aggregations.test.ts retains Unicode ordering, uniqueness, empty collections, owner scope and multiplicity; FleetOverview.test.tsx verifies ranked activation counts, unique skills, empty refetch and no skill success-rate text. |
| Rich owned provenance, status, diagnostic and delegation-cost acceptance in client.test.ts | v6 lean owned occurrence boundary tests replace rich acceptance with exact keys, required nulls, multiplicity, unknown/retired-key rejection, missing/empty/malformed required members, explicit migration requirement, root/session retired-field rejection and response immutability. Existing invocation cost/evidence and aggregate finance tests remain. |
| sessionEvents and baseline reconciliation / includes_baseline tests in events.test.ts | Persistence and baseline merge are retired. transient source traces tests exercise production source selection, immutable chronological ordering with multiplicity, complete/partial/unavailable zero results, session helper recovery, and native source selection identity ambiguity across owner/IDs. API client rejects legacy event arrays and includes_baseline and validates source availability/coverage. |
| Saved-baseline and saved-event lifecycle tests in TicketDeepDive.test.tsx | Source trace and stored facts lifecycle cases cover facts/metrics/known unavailable children before and after deferred failed/unavailable/partial/complete-empty hydration, no artificial event rows, retained recorded summaries for incomplete coverage, ticket-response races (including stale resolve and reject after returning to a ticket), summary-to-source disclosure/inspector rebinding, and independent recorded-summary inspection surviving complete-empty hydration. Baseline-echo selection and unmatched persisted row union are retired oracles. |
| Repeated rich skills / delegation panel details | Skills and subagents autopsy preserves repeated activations, nullable role/time, complete timestamp tooltip, row keyboard access, owner/session filtering and mixed-source owners. Exact child keyboard reactivation still expands/focuses the precise thread and preserves parent/sibling collapse state; unavailable children remain visible. Retired auxiliary delegation financial/status/source fields have no replacement store. |
| Saved native badge joins, provenance/call-ID matching and partial saved trace labels in TraceWaterfall.test.tsx | Native source invocation badges and production wire fixture coverage use actual transient event payloads, recognized activation metadata and proven child IDs. No lean matching by name/time/turn/index occurs. Payload-less/malformed source events cannot borrow a lean skill name. Old codexActivation.json/codexCalls.json saved-event fixtures are superseded by isolated production schema6 detail and transient source-event fixtures. |
| Full source payload/inspect/copy/filter/accessibility behavior | Existing TraceWaterfall filtering, recovery, chronology, selection, rich payload, anomaly, clipboard and keyboard suites remain. TicketDeepDive source waterfall/AnomalyDrawer integration retains tool-error and anomaly drawers, clear selection, summary keyboard disclosure, real payload display, dates and KPI behavior. Live trace source/tests remain. |

## Fixture provenance

codexOwnedDetail.json and codexSourceEvents.json are sanitized synthetic native-format fixtures, initially exported by the backend tester and regenerated by the frontend coder after source payload fixes from production ParseRollout -> SessionFacts -> SaveFactLedger -> WorkspaceDetailHandler and NewLiveServiceWithCodexHome.HydrateWorkspaceEvents -> WorkspaceEventsHandler. The source reproduces TestCodexOwnedOccurrencesUseNativeCallsAndActivationEvidence: explicit selection with attachment, skill call, out-of-order spawns with explicit real child IDs, pending spawn and ordinary read. No live alspaces ledger or developer source store is read. They are deterministic verification inputs, not evidence of a real user session.

## Verification

Verified against the shared working diff based on 5e4eb34019370150786085d69c47987b328e4e79, after final recorded-summary disclosure and partial-owner anomaly association changes:

- npm run typecheck: passed, including canonical lean type key/null/retired-session compiler guards in src/test/leanFactTypes.ts.
- npm run lint: passed with zero warnings.
- npm test: 36 files, 825 tests passed (28.24s final run). Earlier focused/current8-file and 824-test results are superseded by this final result.
- npm run build: passed; 339 transformed modules. Vite's existing advisory chunk-size warning remains (508.31 kB JavaScript before gzip); this is not a failing gate.
- Canonical asset synchronization: diff -qr web/dist backend/internal/adapters/web/dist returned no differences. Both generated trees remain ignored. Final bundle assets: index-C3WE_otj.css and index-0OOsU4ic.js.
- go test ./internal/adapters/http: passed (0.552s) after lean current-schema strict rejection/sibling cases and source payload restoration.
- Temporary production native-wire exporter: passed; removed after regenerating both frontend fixtures.
- git diff --check for frontend and assigned HTTP test path: passed.

Parent owns integrated make, standalone binary asset smoke verification, review, commits, publication and telemetry. This subagent has not claimed those gates or final ticket completion.

## Delegated HTTP test adaptation

The parent explicitly transferred HTTP test ownership after the backend tester paused those paths. Only server_test.go changed. All source fixtures are isolated temporary directories or maintained backend/testdata inputs; no live ticket workspace is read.

- Current fixtures now use schema 6 and omit persisted events. Legacy rejection covers versions 1 through 5; current root/session retired keys and rich occurrence fields receive explicit 422 invalid-current diagnostics while valid siblings remain listed with warnings. GET preserves both valid and invalid ledger bytes.
- TestLeanFactsSurviveSourceTraceAvailabilityWithoutGETWrites replaces the saved-event projection fallback oracle with lean ownership/multiplicity/nulls/known unavailable child IDs, stored metrics and summaries, full/empty/partial/unavailable coverage, actual transient payloads, and the original recursive directory/byte snapshot guard.
- TestWorkspaceSourceEventsPreserveIndependentEqualOccurrencesAcrossReads replaces includes_baseline/echo union expectations with independent equal source occurrences over repeat requests and unchanged ledger bytes. The deleted baseline union is an approved retired oracle; no event is reconstructed from lean rows.
- Summary-only response tests retain allocated empty events, unavailable source state and recorded anomalies. Real SQLite HTTP hydration retains user prompt, reasoning, context tokens, tool input/output and source-only counts, plus missing-source diagnostics, traversal rejection and unchanged ledger files/directories.
- Discovered missing transient native Codex tool_input and SQLite message prompt/reasoning payloads during producer/HTTP verification. Backend coder restored those source payloads. Existing meaningful payload assertions and new frontend production-wire assertions pass after the fixes.
- The backend tester converted shared runtime fixtures with the actual migration engine, preserving event-derived AL-2 anomaly summaries. Existing testdata anomaly assertions remain.

## Schema 6 correction after review

User steering: schema 6 is introduced by this still-unmerged AL-62 ticket while main remains schema 5; final lean ownership, retired-field rejection and source-only traces therefore belong to schema 6 with a direct 5-to-6 migration. This supersedes the prior proposed 6-to-7 transition. Correction baseline: 7cec253d6e62f0440df1f1c906263f645d075c9d; initial task-owned paths were clean. Parent-owned contract, summary, telemetry and lifecycle edits already present are preserved.

Only explicit current-version discriminators and corresponding fixture/test labels changed in web/ and HTTP tests. Frontend types/validators accept schema 6; obsolete-version tables no longer reject current 6 and continue rejecting older versions without implicit migration. Native wire, financial, formatter, progression and HTTP fixtures now identify the same lean shape as version 6. Opaque source/map values, including schema_version: 75 in role metadata, remain unchanged. No compatibility reader or 6-to-7 acceptance path was added. All occurrence members, null semantics, trace availability/coverage, recorded summaries, child proof, filters and keyboard behavior remain intact.

Verification for this narrow correction is recorded below. Earlier 825-test/version-7 evidence above remains historical and is superseded for the changed version contract. Parent owns the integrated make gate and migration implementation/tests outside HTTP.

Narrow correction verification:

- npm run typecheck: passed.
- npm test: all 36 files and 824 tests passed (28.66s). The count decreased by one because the old-schema table no longer treats current schema 6 as obsolete; all behavior and strict lean-shape guards remain.
- npm run build: passed; 339 transformed modules. The pre-existing advisory large-chunk warning remains. Final JavaScript asset is index-CqJY9wej.js; CSS remains index-C3WE_otj.css.
- diff -qr web/dist backend/internal/adapters/web/dist: no differences; generated assets remain ignored.
- git diff --check on owned paths: passed.
- go test ./internal/adapters/http: passed (0.591s) after the production constant changed to 6; the parent retains full make/integration ownership.

The native detail fixture's version discriminator was corrected to 6; its previously production-exported lean fields and separate transient source-event fixture were preserved. No new source observation or fixture re-export is claimed for this version-only edit.
