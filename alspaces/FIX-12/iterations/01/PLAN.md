# FIX-12 iteration 01: Keep recorded trace facts visible without hydration

Role: AgenticLens planner only. Classification: `fix`. Baseline: `alspaces/FIX-12/CONTRACT.md`.

Status: Implementation plan ready for the development orchestrator's approval gate. Only the contract and this plan are produced in this assignment. Tests and application changes are planned below and have not been performed.

## Grounding and change targets

- Configuration: `.agenticlens-config.yaml` defines `workspaces.dir: alspaces` and the FIX prefix with ticket type `fix`. No matching FIX-12 specification or plan-index entry was found; none is required for these artifacts.
- Read path: `backend/internal/adapters/repository.go`, `ScanWorkspaces` and `LoadTicketWorkspace`, currently prune events by calling `domain.PersistAnomalySummaries` before `DeriveMetricsLedger`.
- Pure models: `backend/internal/domain/models.go`, `SessionEventFact`, `AgenticFactLedger.EventFacts`, and `WorkspaceEventsResponse`; `backend/internal/domain/events.go`, `SortEvents`, `RefreshAnomalySummaries`, and `PersistAnomalySummaries`.
- HTTP: `backend/internal/adapters/http/handlers.go`, `WorkspaceDetailHandler` and `WorkspaceEventsHandler`. `backend/internal/adapters/http/dto.go` serializes the detail's domain ledger directly, so retaining its session events reaches the frontend without a new DTO envelope.
- Local hydration: `backend/internal/adapters/live/service.go`, `LiveService.HydrateWorkspaceEvents` and `hydrateCodexEvents`, already funnel source failures into a fallback response, but that fallback currently contains no events.
- Schema: `backend/internal/adapters/schemas.go`, `SessionEventFactDTO`, `ValidateRuntimeFactLedger`, and `AgenticFactLedgerDTOFromDomain`. Optional nested events decode; the write conversion deliberately omits them.
- Frontend: `web/src/views/TicketDeepDive.tsx`, `loadData` and the conditional around Row 6; `web/src/components/TraceWaterfall.tsx`, source labels, offline banner, payload expansion, and empty state.
- Types/client: `web/src/domain/types.ts`, `SessionFact` and session helpers; `web/src/types/telemetry.ts`, `EventFact` and `WorkspaceEventsResponse`; `web/src/api/client.ts`, `fetchWorkspaceEvents`.

## Implementation sequence after approval

### 1. Preserve accepted event facts during backend reads

Coder ownership: `backend/internal/domain/events.go` and `backend/internal/adapters/repository.go`; use existing domain models without widening persisted DTOs.

1. Introduce a pure read-normalization function in the domain events module that retains the current normalized session and ticket anomaly summaries while preserving a copied collection of each original session's accepted events. Reuse the existing summary normalization; do not change the write-oriented `PersistAnomalySummaries` contract.
2. Use this read normalization in both `ScanWorkspaces` and `LoadTicketWorkspace`. Continue validation before projection, retain configured role mapping and aggregation behavior, and keep filesystem error handling in the repository adapter.
3. Ensure `DeriveMetricsLedger` exposes the preserved session facts in `Facts`. Do not add flattened root collections to serialized ledger files or infer missing events from tool totals.
4. Keep `SaveFactLedger` and `AgenticFactLedgerDTOFromDomain` unchanged in purpose: serializing the read projection back through a writer still produces the established compact ledger.

### 2. Return a baseline from the optional events API

Coder ownership: `backend/internal/adapters/http/handlers.go`, `backend/internal/adapters/live/service.go`, and a pure reconciliation helper in `backend/internal/domain/events.go` if needed.

1. Initialize handler and service fallback events from `ledger.Facts.EventFacts()`, normalized to an allocated empty slice when no facts exist. Keep committed provenance, a false hydration flag, and existing anomaly summaries.
2. Preserve that fallback for an absent hydrator, a nil/unavailable service, missing local records, unavailable database or Codex rollout files, and explicit hydration errors. Existing source-selection and failure-handling policies remain intact.
3. Reconcile a nonempty successful local trace with the baseline using the contract's identity rules. Matching rows gain available local details; unmatched baseline rows survive; local-only rows are added. An empty success cannot erase a nonempty baseline or falsely advertise completed enrichment.
4. Apply reconciliation once at each public boundary as necessary and make it idempotent so a service response subsequently processed by the HTTP handler does not duplicate rows. Avoid changing local extraction or recomputing unrelated metrics.
5. Preserve response fields and current error statuses. Ensure no event fallback or successful hydration writes to workspace files.

### 3. Add typed frontend baseline projection and selection

Coder ownership: `web/src/domain/types.ts` and a small pure helper module `web/src/domain/events.ts` if extraction/reconciliation would otherwise be embedded in the view.

1. Add an optional readonly `events` collection to `SessionFact`, representing nested events without a repeated `session_id`. Reuse the established event field types while keeping persisted details optional; do not add `facts.events` to `AgenticFactLedger`.
2. Add a pure session-event projection consistent with `sessionTools`: absent collections yield an empty array, each row receives its enclosing session ID, and original arrays and objects remain untouched.
3. Implement the same deterministic enrichment precedence as the backend: valid response events are usable regardless of hydration status; a failed or empty response retains the baseline. Reconcile partial nonempty responses without collapsing separate events from the same turn. Sort by timestamp and turn with stable exact ties.
4. Leave the client wrapper/raw-array normalization intact unless a narrow boundary adjustment is required. Do not weaken strict types or add a new legacy-file reader.

### 4. Decouple detail display from optional enrichment

Coder ownership: `web/src/views/TicketDeepDive.tsx`.

1. After `fetchWorkspaceById` succeeds, publish its ledger-derived baseline events, detail anomalies, and committed provenance, and finish required page loading. Fetch enrichment without keeping the page behind its loading screen.
2. Publish a successful response through the pure selection/reconciliation helper. Use response anomalies when supplied and keep detail anomalies when absent. Preserve valid anomaly summaries when enrichment is unavailable.
3. Handle request rejection and decoding failures as optional enrichment failures. Preserve the selected ticket's baseline, false hydration flag, committed source, and usable detail view.
4. Add a request-generation or cancellation guard covering ticket selection and refresh. Discard obsolete detail/enrichment completions; clear event selection, drawer state, and old enrichment metadata when the ticket changes.
5. Always render the trace component for a loaded detail, including its empty state. Remove the empty-events-to-legacy-schema inference. If the legacy notice is retained for explicit older-schema frontend inputs, render it as separate metadata and never as the alternate branch to the waterfall.
6. Keep aggregate anomaly inspection accessible beside a no-events trace with a neutral label. Preserve the existing keyboard and drawer interactions, thread records, metrics, and artifact display.

### 5. Make waterfall copy reflect optional details

Coder ownership: `web/src/components/TraceWaterfall.tsx`; use existing CSS unless a minimal layout adjustment is necessary.

1. Label the unhydrated source `Source: Committed Ledger` in both collapsed and expanded states. Update the offline banner to explain that recorded facts remain visible and additional payload details require the local source.
2. Continue rendering events without a hydration prerequisite. Preserve the collapsed summary, accurate total count, current default filter, All Events control, role/tool/anomaly filters, debug identifiers, and anomaly selection.
3. Render only available rich fields. Missing rich details produce an explicit unavailable/uncaptured notice for that event, including baseline rows retained in an otherwise hydrated trace. Do not imply the whole trace lacks facts because one payload is unavailable.
4. Use the existing chronological empty-state message for an actual empty collection, and retain the distinct no-matching-filters state. Do not describe a current summary-only ledger as schema v1.

## Planned test cases

Tester ownership is limited to the files listed below and directly related fixture updates. Existing assertions that encode the defect must change deliberately; unrelated coverage remains in place. Use real domain values for pure tests and real temporary ledger files for repository/HTTP tests. Adapter-boundary hydrator doubles and frontend network stubs may reuse the existing harness. No domain mocks or silent skips.

### Frontend domain and boundary tests

In `web/src/domain/events.test.ts`:

1. Project unsorted events from two canonical sessions and restore the correct session ID, including equal turn numbers in different sessions. Verify every metadata field, stable ordering, and immutability.
2. Omitted session events, explicit empty collections, empty sessions, and absent sessions in a compatibility payload return an empty collection without a schema inference. Events without optional IDs or payloads remain renderable.
3. Reconcile an enriched matching event, an unmatched baseline event, and a local-only event. Verify details are added and rows are not lost or duplicated. Cover different IDs on the same turn, ambiguous missing-ID rows, and repeated reconciliation.

In `web/src/api/client.test.ts`, preserve the `fetchWorkspaceEvents` wrapper, raw-array, empty-stream, network, HTTP, and JSON-decoding cases. Extend only when the implementation changes that boundary; keep raw-array compatibility distinct from ledger schema compatibility.

### Ticket detail regressions

In `web/src/views/TicketDeepDive.test.tsx`, create typed detail fixtures with event facts under `ledger.facts.sessions[].events`; remove reliance on the extra `facts.events` properties used in existing hydration fixtures.

1. A detail with recorded error/tool facts plus a rejected events request still shows Execution Trace Waterfall, the committed source, the correct count, and the expected expanded rows. Assert there is no legacy notice and no fatal page error.
2. Repeat with a resolved empty unhydrated response and an empty response marked hydrated. Both retain baseline facts and committed provenance.
3. Use a deferred events promise to prove the baseline renders before enrichment resolves or rejects; do not use timing sleeps.
4. A nonempty unhydrated response renders even if the detail baseline is empty. A nonempty local SQLite or Codex response adds rich details and uses the correct source. Partial enrichment retains unmatched baseline facts.
5. A schema-v4 summary-only detail, with optional event collections omitted or empty, shows the trace empty state without a legacy-schema claim. Aggregate anomaly items still open the drawer, and threads and artifacts remain accessible.
6. Explicit legacy metadata in a frontend-only compatibility fixture cannot replace supplied events; unknown schema metadata does not imply legacy. Keep backend legacy rejection tested separately.
7. A successful unflagged baseline event is counted, appears after selecting All Events, and does not open the anomaly drawer. Errors and correlated anomalies remain inspectable without rich payloads.
8. Select ticket B while ticket A enrichment is pending, then resolve A. Assert B keeps its own rows, source, count, and drawer state. Repeat for an older refresh response completing after a newer request.

Replace the existing tests named `renders legacy v1 fallback banner when viewing a ticket whose events array is empty` and `displays aggregate anomalies in legacy fallback banner for v1 tickets` with schema-aware empty-state and independent anomaly-access assertions.

### Waterfall component regressions

In `web/src/components/TraceWaterfall.test.tsx`:

1. With `isHydrated` false and populated facts, verify collapsed count, committed-ledger source, expanded chronological rows, available identifiers, filters, and payload-unavailable messaging. Update both old committed-anomaly-summary copy assertions.
2. Hydrated SQLite and Codex rows retain available payload blocks. A hydrated response containing an unmatched bare baseline row shows an honest missing-details state for that row.
3. Preserve empty-input, zero-filter-match, stable chronology, error/failed-status, multiple-tools-per-turn, anomaly matching, debug-copy, keyboard, and collapse/expand coverage. Limit new assertions to behaviors affected by FIX-12.

### Backend pure and repository regressions

In `backend/internal/domain/events_test.go`, cover the new read projection's event preservation, normalized anomaly scopes, stable ordering, empty inputs, and input immutability. Cover reconciliation identity, partial enrichment, ambiguous rows, and idempotence if a domain merge helper is introduced. Keep `TestPersistAnomalySummariesDropsEventsAndEvidenceButKeepsFacts` passing under its existing write contract.

In `backend/internal/adapters/repository_test.go`, change `TestEventFactDTOAndRepositoryEventHandling`: accepted nested v4 facts must survive `LoadTicketWorkspace` and `ScanWorkspaces`, while tool totals and summary normalization remain intact. Assert detail facts can be flattened with their owners and no source file bytes change. Serializing those facts through `AgenticFactLedgerDTOFromDomain` must still omit events and evidence.

In `backend/internal/adapters/schemas_test.go`, retain `TestAgenticFactLedgerDTO_OmitsEventsAndAnomalyEvidence` and the runtime rejection tests. Add a canonical v4 read fixture containing optional nested event metadata and a row without an event ID. Preserve omission/empty compatibility without admitting root collections, old schemas, or raw payload fields.

### Backend HTTP and hydration regressions

In `backend/internal/adapters/http/server_test.go`:

1. Use a canonical v4 event-bearing file to exercise detail and events routes. No hydrator, an erroring hydrator, an unhydrated result, and an empty result return the recorded facts with false hydration, committed provenance, allocated event arrays, and preserved anomalies.
2. Verify chronological ordering across sessions and turns, successful enrichment metadata/details, and survival of baseline rows during partial local success.
3. A summary-only ledger returns HTTP 200 with `events: []` and its summaries. Preserve missing-workspace, hostile-ticket, corrupted-ledger, and `TestHTTPRejectsLegacySchemaVersionsWithoutPartialPayload` behavior.
4. Snapshot workspace files before and after requests and retain the zero-writes assertion. Existing tests that expect zero fallback events must distinguish event-bearing files from genuinely event-free files.

In `backend/internal/adapters/live/sqlite_test.go`, extend existing hydration fallback cases using a ledger with recorded facts: nil service/hook, missing sessions, closed DB, missing tables, absent Codex roots/rollouts, and empty local results must return baseline events. Keep successful OpenCode/Codex hydration, anomaly hierarchy, and zero-disk-write coverage. Update old blanket zero-event assertions only for fixtures that contain baseline events.

## Verification and handoff for subsequent roles

No tests, builds, or application commands are executed by this planner.

After implementation approval, the tester should first run the focused frontend files using `npm test -- src/domain/events.test.ts src/api/client.test.ts src/views/TicketDeepDive.test.tsx src/components/TraceWaterfall.test.tsx` from `web/`, plus typecheck and lint. Run focused Go domain/adapter/HTTP/live packages from `backend/` with the race detector. Run the repository-required `make` before delivery and verify the production build still embeds frontend assets. These are planned commands, not recorded passing results.

The reviewer should audit both the visible rendering behavior and the entire source-to-detail-to-events path; a view-only change is insufficient while reads prune facts. Check the frozen schema boundary and compact writer tests explicitly. Expected review evidence includes ledger-byte immutability, baseline visibility during deferred enrichment, no false legacy notice, partial-enrichment fidelity, and preserved empty/anomaly behavior.

The orchestrator owns later `SUMMARY.md`, `PULL_REQUEST.md`, `REVIEW.md`, final telemetry with ticket type `fix`, commits, synchronization, and PR delivery. This planning assignment does not mark FIX-12 implemented or complete and does not move unrelated plan entries.
