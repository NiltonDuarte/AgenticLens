# AL-62 frontend implementation

Historical iteration 01 working record. The accepted iteration 02 contract, implementation and final SUMMARY.md supersede unfinished progress and compatibility assumptions below.

Baseline: 19d70574d0e22b00f185c32d048baaf62792743e. Scope approved in CONTRACT.md and PLAN.md. Frontend coder exclusively owns web/src and this report; backend implementation is concurrent and preserved.

## Progress

- Mirrored v6 session-owned skill/task arrays, native provenance, explicit outcomes/evidence, and unresolved history.
- Owner-derived selectors ignore root compatibility projections under canonical representation and preserve legacy fallback. Session selectors never assign ownerless history.
- API validation includes owned/unresolved occurrences and v6 cost evidence; nil optional collections normalize without payload mutation.
- Panel exposes owner/native references, activation requests with unavailable outcomes, session filter and explicit unresolved ownership. Child links retain unavailable behavior.
- Trace badges require a proven session owner for fact joins; native occurrence IDs take precedence, and Codex spawn tools and request message badges are supported. Ownerless indexes/timestamps no longer attach historical child links.
- Fleet success denominator uses known_outcomes when provided; unknown requests remain visible invocation totals.

## Verification

- npm run typecheck: passed on changed frontend sources.
- npm run lint: passed with zero warnings.
- Focused API/domain/fleet/trace/deep-dive run: 286 tests passed in five suites before the final native-ID trace regression was added.
- Full npm test: 800 passed of 801 in 36 suites. The one failing new trace test used a broad text locator that matched the parent row's child-link text instead of the child session row. Corrected the locator to the exact Session: child debug label; no production change was required.
- Final affected TraceWaterfall and TicketDeepDive suites: 153 tests passed, including native-call owner separation, unknown activation outcome, session filtering, unresolved retention, unavailable raw child reference and keyboard child navigation. Other 648 full-run passing cases remain unchanged.
- npm run build: passed; current bundle index-BKftCX0i.js was synchronized to the Go embedded asset directory. The existing greater-than-500 kB bundle advisory remains; no bundle-size acceptance change was introduced.
- git diff --check -- web/src: passed.

Parent retains the integrated make gate and standalone embedded HTTP verification. These have not been claimed here. Backend source implementation was concurrently evolving; frontend fixtures exercise the declared stable API, not a completed live server.

## Compatibility and ownership notes

Backend coder confirmed that read-only API serialization canonicalizes compatible legacy/transitional facts first: independent ownerless rows become unresolved_facts, exact known owners relocate, and identified root projections suppress against canonical owners. Root API arrays are then derived views. Frontend selectors ignore those views when schema v6, owned collections or unresolved_facts establish canonical representation; old root-only records retain fallback behavior. No content deduplication erases repeated occurrences.

Task child_reference preserves a native task namespace independently from a resolved child_session_id; reference-only records display Thread unavailable and are never passed to UUID navigation. SkillAdoptionFact known_outcomes and unknown_outcomes are optional only for compatibility with older APIs; v6 producers supply them.

Changed frontend paths: api/client.test.ts, api/costValidation.ts, api/validation.ts, components/SkillsSubagentsPanel.tsx, components/TraceInvocationBadge.tsx, components/TraceWaterfall.tsx and .test.tsx, domain/aggregations.ts and .test.ts, domain/traceBadges.ts, domain/types.ts, views/FleetOverview.test.tsx and views/TicketDeepDive.test.tsx. Existing shared helpers route fleet/comparison consumers through owned skills without view-specific persistence copies. No backend or runtime-script edits. No raw transcripts or live ticket-workspace fixtures.

Frontend coder is paused and relinquishes web/src test ownership to the parent-assigned tester. Production corrections remain available upon explicit parent reassignment; no concurrent edits will be made while tests/reviews run.
