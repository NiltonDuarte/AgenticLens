# AL-62 frontend implementation and verification

Implemented session-owned skill/task projections with inherited owner/source, lossless unresolved root history, and transitional echo suppression only for proven owner plus source occurrence identity (and skill name). Equal-looking occurrences remain distinct. Added schema 6 boundary validation, nested task cost validation, source identities/evidence/outcomes, ownership diagnostics and child namespace references. Existing schema 4/5 responses remain supported.

Skills & Subagents now exposes owner/source event or invocation and message identity, unavailable execution outcomes, unresolved ownership and explicit child references without invented UUIDs. Skill file reads remain recorded loading evidence with execution unavailable. Fleet success rates use known outcomes only; comparison skill filters consume canonical projections. Timeline facts require proven owner and source event/call identity, or an unambiguous native-tool position within that owner; unresolved legacy facts remain visible in the panel but cannot bind arbitrary timeline events. Session-local overlap and duplicate positional events do not misattribute facts.

Verification on the working AL-62 diff (baseline 122e983f8c365a2aa248b1b6d492252555e922ae):

- npm run typecheck: passed.
- npm run lint: passed, zero warnings.
- npm test -- src/domain/events.test.ts src/domain/aggregations.test.ts src/api/client.test.ts src/components/TraceWaterfall.test.tsx src/views/TicketDeepDive.test.tsx src/views/TicketComparisons.test.tsx: 307 tests passed across six files.
- A final duplicate positional-event guard was then added to the existing badge safety table. npm test -- src/components/TraceWaterfall.test.tsx -t 'rich invocation badges|retrospective invocation metadata safety': seven tests passed; 79 unrelated tests explicitly excluded by the name filter. Parent integrated make remains required.
- npm run build: passed canonical tsc/Vite/sync-assets pipeline. Vite reports the existing advisory for a JS bundle larger than 500 kB.
- SHA-256 comparison: web/dist and backend/internal/adapters/web/dist contain identical index.html and JS/CSS asset trees. Generated ignored assets are not staged. Standalone binary verification belongs to parent integration.
- git diff --check -- web: passed.

One first-pass test failed because an old child-navigation fixture relied on assigning an ownerless legacy fact to a timeline event. Converted that fixture to a canonical session-owned fact with explicit event identity while preserving all child navigation and ticket-reset assertions. Separate existing tests retain repeated unresolved root history, panel rows and child references; an explicit ownerless-task case rejects arbitrary timeline binding while preserving input-only badges.

No browser geometry/interactive browser run claimed. Existing local historical timestamp formatters and their tests were preserved. All tests use self-contained fixtures; no live AL-55 ledger is read by tests. Exact backend wire agreement: sessions[].skills/task_links; optional session_id/source/source_event_id/source_message_id/source_invocation_id; skill evidence skill_tool|skill_file_read and outcome succeeded|failed|unavailable; task child_reference; ledger ownership_diagnostics; fleet known_outcomes and nullable failure_rate.

## Candidate02 review corrections

Behavior review B2 is corrected by returning all distinct source-proven skill badges for a shared invocation. Conflicting/ambiguous task facts still cannot select a child link; ownerless legacy rows cannot attach to arbitrary timeline events. Original session/call/event IDs remain the join authority, including hydrated events whose event ID differs but original source invocation ID agrees.

Contract gap G1 is corrected through pure ledgerTimelineEvents, a transient read projection of compact owned facts into the ticket timeline. It requires recorded owner, timestamp and source event/call identity; it never invents event/call IDs or stores raw payloads. Multiple skills for one call use one timeline row and separate badges. Real recorded/hydrated events take precedence on exact source identity; missing children/other-owner evidence remain visible after partial enrichment and successful includes-baseline responses. No ledger tools, invocation metrics, session roots/parent links, failure rates or canonical fact storage are modified. Mutation checks preserve input facts. Projected rows explicitly display unavailable context and unknown delegation outcomes; placeholder values required by EventFact are not displayed as measured context.

Introduced frontend fields: optional source_invocation_id on EventFact/SessionEventFact (matching backend wire); UI-only transient fact_projection marker on read-derived timeline rows; optional auxiliary_extraction_version on SessionFact accepted as the backend extractor representation marker. No new persisted frontend store or server write side effect exists. Child namespace references without a source UUID display an unavailable identity state rather than a fabricated navigation target.

Verification after corrections:

- npm run typecheck and npm run lint passed.
- npm test -- src/domain/events.test.ts src/components/TraceWaterfall.test.tsx src/views/TicketDeepDive.test.tsx src/api/client.test.ts: 291 tests passed.
- Final changed UI/event scope rerun after source-call reconciliation indexing and unavailable-reference rendering: npm run lint and npm test -- src/domain/events.test.ts src/components/TraceWaterfall.test.tsx src/views/TicketDeepDive.test.tsx: 180 tests passed.
- Final one-line preservation of the new source invocation identity during ordinary event enrichment was then covered by the extended existing domain enrichment case: npm test -- src/domain/events.test.ts: 26 tests passed.
- Final npm run build passed TypeScript checking, Vite and asset synchronization. SHA-256 verification confirms identical frontend/embed index.html, JS and CSS trees; generated ignored assets remain untracked.
- git diff --check -- web passed. Parent full make, standalone binary checks and required affected-focus re-review remain delivery gates.

The compact persisted-shape deep-dive fixture has no session events, two skill-file facts from one wrapper call, explicit linked child, unavailable child UUID and unresolved namespace. It proves offline badges, one row per source invocation, unavailable outcomes/context, and child navigation with optional enrichment rejected. Domain cases also exercise pending baseline, empty failed enrichment, partial independent responses and successful includes-baseline responses without duplicated source-call rows or input mutation. Existing raw-event child-navigation, ticket-reset, ambiguity and empty-state guards remain intact.

The first correction verification exposed two null-ledger empty-state regressions from removing optional access at the timeline initialization boundary. Restored optional ledger handling, retained both existing tests, and reran the entire affected UI scope successfully. No application/test changes after final verification and asset comparison.
