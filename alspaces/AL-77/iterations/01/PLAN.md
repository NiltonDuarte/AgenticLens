# AL-77 iteration 01: Persisted session overlap in Fleet Overview

Status: Approved by explicit user response "Approve implementation" on 2026-11-04.

## Baseline and grounding

- Initial revision: a84adbe5be46c2ef52ede0987a417562715d38e1. Initial working tree was clean. Parent selected feat/AL-77-warn-duplicate-sessions and owns ticket bookkeeping. The configured workspace root is alspaces.
- Contract: [CONTRACT.md](../../CONTRACT.md). No previous implementation or verification is claimed.
- web/src/domain/types.ts exposes AgenticMetricsLedger ticket_id/ticket_title and SessionFact id/source in facts.sessions. web/src/api/validation.ts decodeLedgers validates these and normalizes known nullable collections. No extra read identifiers are needed.
- web/src/api/client.ts fetchWorkspaceLedgers defaults to the unfiltered /api/workspaces listing and returns response diagnostics. fetchWorkspaces derives table rows from the same ledger response; those rows are insufficient for identity detection.
- backend/internal/adapters/http/handlers.go WorkspacesHandler calls adapters.ScanWorkspaces before applying optional query filters, returns workspace diagnostics, and serializes sessions through SessionFactResponseDTO in workspace_dto.go. backend/internal/adapters/repository.go ScanWorkspaces validates supported persisted ledgers and derives metrics without live-source hydration. Invalid siblings produce diagnostics rather than hiding valid siblings.
- web/src/views/FleetOverview.tsx currently loads filtered summary/table data together. Its loadData callback changes with date/type/prefix filters. Table search is local to TicketTable. Keep overlap inspection independent from these filtered data values.
- web/src/App.tsx handleSelectTicketFromList supplies onSelectTicket to FleetOverview and switches to TicketDeepDive. Reuse this callback for conflict tickets; no router or new route is required.
- WorkspaceLoadWarning already renders skipped-workspace diagnostics and response diagnostic-read warnings. Reuse its diagnostic semantics, while explicitly describing the effect on overlap scope.
- Existing coverage owners: web/src/views/FleetOverview.test.tsx for loading/retry/filter/selection and diagnostics; web/src/api/client.test.ts for ledger decoding and diagnostics; backend/internal/adapters/http/server_test.go TestGetWorkspaces and static/SPA tests for read projections and serving; web/src/test/sync-assets.test.ts for complete embedded-tree synchronization.
- web/package.json npm run build performs typecheck, Vite build, then web/scripts/sync-assets.mjs copying the complete tree into backend/internal/adapters/web/dist. That generated tree is currently untracked. Makefile's default gate installs dependencies, checks, tests with Go race detection, and builds all configured platforms.

## Implementation steps and ownership

1. Coder owns production paths web/src/domain/sessionOverlaps.ts (new pure projection and its concrete readonly types), web/src/components/SessionOverlapWarning.tsx (new presentation), web/src/views/FleetOverview.tsx, and necessary scoped styles in web/src/index.css. Reuse canonical API/types; avoid unrelated refactors.
2. Derive exact session-to-distinct-ticket memberships from facts.sessions only. Preserve all affected tickets, titles, and source context across repeated projections. Ignore unusable identity values without inventing replacements and return enough scope information to explain omitted records. Sort output deterministically for stable reading; do not mutate inputs or totals.
3. Add independently handled unfiltered fetchWorkspaceLedgers loading to Fleet Overview. One listing request per explicit inspection load is sufficient; no per-ticket detail/trace requests or polling. Initial mount inspects all project ledgers. Existing filter reloads may reuse the current unfiltered inspection; an explicit Refresh Data control must reload both the current metrics and the full inspection. Keep metrics available if this optional inspection fails.
4. Model overlap loading, success with complete/partial scope, and failure explicitly. A newer load supersedes older in-flight results. Do not leave old conflicts displayed as current during loading/error or restore them from a late completion. Diagnostic-read failures also mean uncertain completeness. Continue to display valid discovered conflicts from partial successful listings.
5. Render warning above the filtered metrics/table so filters and empty states do not conceal it. Use a native details/summary disclosure with a clear conflict count, full IDs, ticket IDs/titles, source context, scope explanation, and named ticket buttons using onSelectTicket. Explain metrics may overlap and the user chooses corrections. Permit text-only tickets when no callback is supplied. Scope status must avoid unsupported global absence claims.
6. Tester owns contract-focused additions to web/src/domain/sessionOverlaps.test.ts and web/src/views/FleetOverview.test.tsx, plus an existing API/HTTP test file only when it supplies missing boundary evidence. Coder may add immediate regression coverage; tester inspects it before extending it. Parent alone owns contract/plan revisions, integration, review consolidation, telemetry, commits, PR artifacts and publication. Shared-checkout agents preserve others' edits.

## Verification by regression

| Regression or acceptance case | Evidence owner and focused verification |
| --- | --- |
| Two tickets share a genuine stored session ID | Pure projection test plus FleetOverview consumer renders the session and every affected ticket |
| Repeated rows or repeated projection of one ticket create a false positive | Pure test reuses one ticket ID with repeated sessions and invocations; conflict threshold remains distinct ticket count |
| Different or missing source labels suppress a real match | Pure test matches by ID alone and keeps labels as context |
| Single-ticket, disjoint, empty or blank identities create invented conflicts | Small table of pure cases; no warning and explicit limitation for unusable IDs where needed |
| Filtered visible rows hide a project overlap | FleetOverview test supplies a conflict outside filtered rows, changes date/type/prefix and table search, verifies unfiltered request/inspection and preserved metrics/table filtering |
| Skipped ledgers or unreadable diagnostic headers masquerade as complete | Consumer test returns diagnostics with valid ledgers; display loaded scope and partial status without losing valid conflicts; reuse existing client diagnostic parsing tests |
| Failure masquerades as no overlap or breaks valid metrics | Independently fail the unfiltered listing and verify unavailable status, working metrics, and retry recovery |
| Corrected records retain stale warning | Reload with a corrected second ticket; warning clears, metrics and navigation remain valid |
| Old response overwrites corrected reload | Deferred promises complete out of order; newest inspection wins and unmount is safe |
| Details/navigation inaccessible | Actual consumer keyboard interaction using testing-library user-event for disclosure and ticket button, accessible names, callback destination, full long session ID; inspect containment styles and focus behavior |
| Receiving checkout lacks original sources/private capture state | Extend existing HTTP workspace fixture coverage only if necessary: two current-schema ledger files under a temporary project, no hydration hook; GET contains both stored IDs/source labels and leaves bytes unchanged. Consumer fixture has persisted facts only and requires no traces |
| Warning changes totals | Consumer asserts supplied fleet totals and table values stay unchanged with conflict present; pure input remains unchanged |

Do not create redundant declarations per cell; combine coherent cases using existing fixture/test owners. No live AL-77 history or source transcripts become test fixtures. No source availability is inferred merely from synthetic data. Record exactly what source-unavailable fixture evidence establishes.

After focused checks, run frontend typecheck, lint, tests and build through canonical commands and the integrated make gate after merging upstream main. Record tested revision/diff and real results. Reuse unchanged valid checks; rerun affected checks after fixes. Compare web/dist with backend/internal/adapters/web/dist after the build. Verify the newly built host binary's index and referenced assets through a temporary standalone-server run; do not use an installed pinned release to prove this source build. Record unavailable browser/assistive-technology or runtime stages explicitly.

## Independent review roster

Freeze one candidate revision/fingerprint and stop production/test edits during discovery. Each selected focus runs both gpt-6.1-sol/high and gpt-6-luna/medium under the configured model policy. Requested settings must be passed to real dispatch, with acknowledged versus source-observed identity kept distinct. Capacity permits three child runs at once; execute waves without omitting cells. Parent owns consolidated REVIEW.md and private run/finding records; each child owns only its named report. Reports are temporary active-review evidence, compacted after final approval into REVIEW_SUMMARY.json with all finding/reporting tuples.

| Focus owner | Risk question and scope | Sol report | Luna report |
| --- | --- | --- | --- |
| reviewer-behavior | Does FleetOverview expose all real conflicts while preserving filters, metrics and ticket navigation through App? | REVIEW_BEHAVIOR_SOL.md | REVIEW_BEHAVIOR_LUNA.md |
| reviewer-data | Does the exact-ID index preserve distinct membership and partial/unavailable scope without source dependence, writes, or aggregate changes? Follow client, handler and scan dependencies. | REVIEW_DATA_SOL.md | REVIEW_DATA_LUNA.md |
| reviewer-architecture | Is the change a minimal pure read projection with bounded traversal/request work and appropriate domain/API/view separation? | REVIEW_ARCHITECTURE_SOL.md | REVIEW_ARCHITECTURE_LUNA.md |
| reviewer-tests | Do self-contained tests detect actual threshold, filtering, reload and failure regressions through consumers without mirroring implementation? | REVIEW_TESTS_SOL.md | REVIEW_TESTS_LUNA.md |
| reviewer with frontend async/accessibility specialty | Can delayed reloads restore stale conflicts, or can keyboard/assistive-technology users lose disclosure, full IDs or ticket navigation? Inspect loading ownership and semantic controls. | REVIEW_FRONTEND_SOL.md | REVIEW_FRONTEND_LUNA.md |

Report paths are under iterations/01/. The specialist is justified by new asynchronous inspection state and interactive disclosure/navigation. Use generic reviewer with an explicit specialty brief and canonical review/shared context because no dedicated specialist persona is registered. Queue example: behavior Sol/Luna and data Sol; then data Luna and architecture Sol/Luna; then tests Sol/Luna and frontend Sol; finally frontend Luna. Actual scheduling can vary while preserving frozen evidence and independent first passes.

Parent verifies/deduplicates findings, routes fixes to coder/tester, then rechecks affected focuses under both models. Integrated make and current review evidence gate final consolidated approval and publication. A missing model/focus or material verification gap remains INCOMPLETE. No child may publish or merge.

## Planning decision and approval gate

The proposed surface is Fleet Overview with navigation into existing ticket details; this meets the ticket/fleet-context requirement without adding a second fetch owner in TicketDeepDive. Existing read projections suffice, so production scope is frontend only. No product ambiguity currently blocks approval. If implementation discovers a need for API/schema changes or aggregate behavior changes, return to planning rather than silently expanding scope.

User approved CONTRACT.md and this plan explicitly on 2026-11-04. Implementation may proceed within these boundaries; completed-ticket lifecycle and final telemetry remain deferred to delivery.
