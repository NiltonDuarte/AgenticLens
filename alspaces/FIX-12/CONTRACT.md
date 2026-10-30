# FIX-12: Render ledger event facts independently of hydration

Ticket: FIX-12. Classification: bugfix; conventional commit and eventual telemetry type: `fix`.

Status: Frozen technical baseline for iteration 01, subject to the implementation approval gate. This planning assignment creates only this contract and `iterations/01/PLAN.md`. Application changes, test execution, telemetry generation, and delivery belong to the subsequent development workflow.

## Problem and grounded findings

The execution waterfall must display chronological facts already available in a supported ledger when local session hydration fails, is unavailable, or has not completed. Hydration supplies additional local events and payload details; it is optional for displaying committed facts.

- `web/src/views/TicketDeepDive.tsx`, `loadData`, initializes `loadedEvents` to an empty array and only populates it from `fetchWorkspaceEvents`. It does not fall back to the loaded detail ledger and keeps the whole view loading while awaiting the optional request.
- The same view selects the `legacy-fallback-banner` whenever the resulting events array is empty. Empty events do not establish that a ledger uses schema v1.
- `web/src/components/TraceWaterfall.tsx` already renders supplied events with `isHydrated` false. Its offline badge and banner incorrectly describe every such collection as anomaly summaries without a timeline.
- `backend/internal/adapters/repository.go`, `ScanWorkspaces` and `LoadTicketWorkspace`, call `domain.PersistAnomalySummaries` during reads. That function explicitly removes session events. Consequently a frontend fallback alone cannot recover events removed before the detail response is serialized.
- `backend/internal/adapters/http/handlers.go`, `WorkspaceEventsHandler`, and `backend/internal/adapters/live/service.go`, `HydrateWorkspaceEvents`, both initialize fallback responses with empty events, irrespective of ledger event facts.
- `backend/internal/adapters/schemas.go` accepts optional `SessionEventFactDTO` records under sessions. `web/src/domain/types.ts`, `SessionFact`, currently lacks the corresponding optional events field.

## Required behavior

1. A successfully loaded, supported ledger is sufficient to display its recorded event facts. Neither the hydration flag nor local source availability is a condition for mounting the waterfall, showing its event count, expanding it, filtering it, or selecting a fact.
2. Detail loading and optional enrichment have independent lifecycles. Once detail facts are available, a pending events request must not leave those facts hidden behind the page loading screen. Network errors, missing endpoints, HTTP failures, malformed response decoding, and empty enrichment responses retain the ledger baseline and do not turn a successful detail load into a fatal error.
3. The baseline is the selected detail ledger's session-owned events with the enclosing session identity restored. Optional events response facts are usable even when `is_hydrated` is false. A nonempty hydrated response adds local information; it must not discard recorded baseline facts if only part of the local trace is available.
4. Hydration may enrich message identifiers, prompts, reasoning, tool inputs, tool outputs, durations, and error details that actually exist. Missing details remain explicitly unavailable. No payload content is invented from summaries or aggregate tool totals.
5. A legacy-schema notice may appear only when supported by explicit schema metadata. If retained, it is informational and appears alongside the trace or its empty state. It must never replace an available waterfall. Empty events, an absent schema version in a frontend compatibility payload, and failed hydration do not independently establish legacy schema.
6. A loaded ledger with no events from either source shows a truthful trace empty state and retains aggregate anomaly inspection, session threads, metrics, and artifacts. A recorded trace filtered down to zero rows continues to show the separate existing filter-empty message. No rows are synthesized from anomalies or invocations.
7. Existing collapse defaults and the initial errors-and-anomalies filter remain intact. A recorded successful event may be hidden by that filter but remains included in the total and is visible through All Events.
8. Ticket changes and refreshes must not display events, hydration metadata, payloads, or an open event selection left over from another ticket. Late enrichment responses are ignored after the relevant request becomes obsolete.

## Event facts and compatibility boundary

### Ledger schema

- `backend/internal/domain/constants.go` currently sets `CurrentSchemaVersion` to 4. Ordinary ingestion continues to use `ValidateRuntimeFactLedger` followed by DTO decoding and validation.
- The accepted event location is `sessions[].events`; session identity comes from `sessions[].id`. `AgenticFactLedger.EventFacts()` is the existing flattened Go projection and restores `session_id` while sorting chronologically.
- Schema versions 1, 2, and 3, a missing or unsupported version, incomplete migration state, missing required session metadata, and ambiguous root-level fact collections retain their current runtime rejection behavior. Migration and source-backed backfill remain explicit operations. FIX-12 does not add an implicit legacy reader or broaden supported ledger versions.
- Root-level ledger `events` and computed-detail `facts.events` are not canonical v4 collections. Existing frontend test objects containing the latter are not evidence of runtime schema support and must not become the fix's data source.
- Missing or empty optional session event collections are legitimate. A valid current ledger may contain only anomaly summaries. The frontend projection tolerates absent optional collections, including absent sessions in a compatibility payload, without inferring a schema version.

### Fact fidelity and identity

- Preserve the facts supported by the existing nested DTO: event ID when present, timestamp, message ID when present, turn, role, event type, tool name, status, duration, and context tokens. Restore session identity from the owner without mutating the source object.
- Rich response-only fields remain optional under `web/src/types/telemetry.ts`, `EventFact`. The nested persisted DTO does not currently admit raw error messages or rich payload fields; this ticket does not add them to that persistence schema.
- Do not require an event ID or payload fields to display an otherwise accepted older v4 event row. Do not merge separate calls merely because they share a session, turn, or tool name.
- For enrichment reconciliation, use matching nonempty session and event IDs first. When an ID is absent, match only an unambiguous occurrence with equal session ID, timestamp, turn, event type, tool name, and available message ID. Different nonempty IDs remain distinct; ambiguous rows remain distinct. Matching hydrated rows enrich their baseline counterpart, unmatched baseline rows survive, and genuinely additional local rows are included.
- Preserve stable chronological ordering by timestamp and then turn, retaining input order for exact ties. Sorting and reconciliation must not mutate either input collection.

### Events response

- Keep the existing `WorkspaceEventsResponse` envelope: `ticket_id`, `is_hydrated`, `events`, `data_source`, and anomalies. No schema version bump or additional required response field is needed.
- Without successful local hydration, return the sorted committed ledger facts with `is_hydrated: false` and `data_source: committed_ledger`. An existing empty ledger returns `events: []`, never a fabricated row or a null collection.
- Preserve `local_sqlite` and `local_codex` for successful enrichment. These describe the enrichment source; the workspace-wide hydration flag does not guarantee payload content on every row.
- Keep `fetchWorkspaceEvents` compatibility with an older raw events-array response. That HTTP response compatibility does not authorize reading legacy ledger files.
- Keep existing invalid-ticket, missing-workspace, invalid-ledger, and HTTP error behavior. An optional enrichment failure after successful detail loading is handled locally by the frontend.

## Scope and architectural invariants

### Frontend

Work is confined to the detail view's data selection and request lifecycle, the waterfall's presentation of source and optional details, typed session event projection/reconciliation, and directly related tests. Use `web/src/domain/types.ts` for session-owned facts and pure helpers and `web/src/types/telemetry.ts` for flattened response facts. Preserve strict readonly types and runtime narrowing where needed; do not cast a legacy root events collection into the v4 model.

Offline copy must describe available ledger facts accurately. Use a committed-ledger source label in both collapsed and expanded views and explain that additional payload details require a local source. Keep aggregate anomaly access when no event can be selected. Keep the network client in the boundary layer and `TraceWaterfall` free of network fetching.

### Backend

The backend change is necessary to preserve accepted facts on read and to return those facts from fallback events responses. Use a pure read projection distinct from the compact write projection, retain existing anomaly-summary normalization, and apply the same read behavior to targeted loading and workspace scanning. Any shared reconciliation or normalization logic belongs in `backend/internal/domain`; repository, HTTP, and local-source I/O remain in adapters.

`PersistAnomalySummaries`, `SaveFactLedger`, and `AgenticFactLedgerDTOFromDomain` retain their compact persistence behavior, including omission of events, local payloads, and transient evidence references. GET requests never rewrite a ledger or persist hydrated data. This ticket does not restore event persistence, backfill absent facts, change migration rules, add an event schema, alter extraction, or revise aggregate metric formulas.

All subsequent implementation must preserve FCIS, Go and TypeScript static typing, zero `any` or `interface{}`, zero emojis, and frontend embedding in the single Go binary. No new runtime service or production Node dependency is introduced.

## Acceptance and verification boundaries

The fix is acceptable when an event-bearing canonical v4 file produces the same baseline trace through workspace detail, the no-hydrator events API, the failing-hydrator API, and the frontend's failed-request fallback. A pending request also leaves the already loaded baseline visible. Hydrated success adds usable details without losing baseline rows; summary-only current ledgers remain truthful empty traces with accessible anomalies.

Schema rejection, compact writes, anomaly scope, ordering, filtering, collapse controls, and existing successful OpenCode/Codex hydration remain covered by focused regressions. The exact implementation tasks and test cases are in `iterations/01/PLAN.md`. No tests were executed during this planning assignment; reading existing test assertions is grounding, not a passing verification result.
