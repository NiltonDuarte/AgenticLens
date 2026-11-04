# AL-62 iteration 02 frontend

Baseline: 19d70574d0e22b00f185c32d048baaf62792743e. Current acceptance is CONTRACT.md's superseding section and iteration 02 PLAN.md. The user explicitly authorized ownerless deletion and removal of legacy ledger compatibility. No new scope gate was required.

## Delivered behavior

- AgenticFactLedger is current schema 6 and contains sessions only for skill/delegation facts. Removed root skills/task_links and unresolved_facts types, selectors, normalization and UI copy. Shared selectors flatten canonical sessions once, inherit owner/source, preserve repeated occurrences and scope by exact session.
- API validation requires schema 6, rejects every retired root collection (including empty arrays), and reports migration required for missing/older schema without converting or mutating payloads. Current nested invocation/delegation costs require explicit valid evidence.
- Skill outcomes and fleet known_outcomes/unknown_outcomes are explicit current fields. Removed boolean outcome and total-invocation denominator fallbacks; unknown activation requests do not count as failures or measured success.
- Mirrored native item_id, message/call/turn/event identifiers, optional diagnostic source_ref, skill evidence_references and session agent_path. Panel distinguishes diagnostic source references from native identities and displays supplemental evidence references without counting another occurrence.
- Preserved exact session filters, recorded owner/source display, native trace badge joins, unavailable owned historical facts and UUID child navigation. Raw child_reference remains distinct from child_session_id and reference-only records explicitly show Thread unavailable.
- Removed obsolete schema-v1 ledger banner and its unused styles. The events endpoint's unrelated raw-array transport fallback remains unchanged by explicit parent direction; ledger compatibility removal does not redesign that transport.

## Test ownership and assertion preservation

Tester first updated schema/owned fixtures in domain events, aggregations, formatters, progression, comparison and shared mockData. The parent then transferred exclusive web/src production/test/fixture ownership to the frontend coder. Existing tester edits were preserved and completed with required outcomes, schema literals, cost evidence and native owned skill/task fixtures.

Replaced retired compatibility expectations with current-boundary assertions. Successful v5 optional-cost reads became current v6 explicit-cost assertions; old schemas now have explicit migration-required rejection coverage. Root/unresolved API acceptance became rejection of all six retired root collections. Preserved typed cost states/zero/partial/unavailable, opaque source payloads, nil current collections, nonmutating decoding, owner multiplicity, filters and keyboard child navigation. Unavailable historical skill coverage now places those facts under a real remote owner rather than unresolved storage. No live workspace fixtures or raw transcripts were added.

## Verification

- npm run typecheck: passed after fixture adaptation; the canonical build also ran TypeScript validation successfully.
- npm run lint: passed with zero warnings.
- npm test: all 810 tests passed in all 36 suites on the final frontend diff (2026-11-04).
- npm run build: passed; Vite emitted index-D-FEfTs4.js and index-Bn43kjvY.css and synchronized them into backend/internal/adapters/web/dist.
- diff -qr web/dist backend/internal/adapters/web/dist: passed.
- git diff --check -- web/src: passed.

The existing greater-than-500 kB bundle advisory remains informational. Parent owns integrated make, standalone HTTP embedded-asset verification, review/integration, telemetry and publication; none are claimed complete here. Backend was concurrently changing and the frontend checks exercised self-contained declared current API fixtures.

Frontend implementation is paused for integration/review. Parent may assign focused corrections; no concurrent edits will be made while review is frozen.

## Child relationship diagnostic correction before review freeze

Backend added optional child_link_diagnostic, separate from execution outcome/error. Frontend mirrors and validates the field, renders a full relationship note in the delegation panel and a diagnostic tooltip on native trace badges. When a diagnostic is present, both entry points preserve the raw child UUID/source reference but disable navigation and show Child relationship unresolved; copy-ID remains available. This does not relabel execution success as failure.

Extended the existing delegation test with a contradictory locally available UUID while retaining the prior duplicate unavailable rows and reliable keyboard navigation case. Extended native-call badge coverage to verify diagnostic tooltip/no navigation and the nested API boundary case to reject nonstring diagnostics.

Affected verification passed: typecheck, lint, all 279 tests in API client/TicketDeepDive/TraceWaterfall suites, canonical npm build, embedded-tree comparison and git diff check. Current bundle is index-BzPA6ihA.js (CSS index-Bn43kjvY.css). The earlier 810-test full run predates this bounded correction; unaffected suites retain their evidence and parent integrated make will validate the complete candidate. Frontend is paused again for review.

## Discovery review corrections

Frontend implementation actor: coder, gpt-6.1-sol/high (inherited runtime configuration). Backend production/test owners independently corrected source extraction, refresh and read projection findings; this section records only frontend changes and checks.

- BEH-06: Added current skill_activation event type to telemetry validation/rendering with an explicit unavailable outcome. Replaced the invented turn_message activation fixture with test/fixtures/codexActivation.json, copied from the backend tester's sanitized ParseRollout -> SessionFacts -> HTTP SessionFactResponseDTOsFromDomain output. The component test passes this exact wire payload through decodeDetail, sessionEvents and ledgerSkills, then checks the single token+attachment activation occurrence, native selection message, supplemental attachment reference and unknown-outcome badge. It does not use live source files or fabricate native provenance fields.
- BEH-07: Mirrored optional opaque child_agent_id, displayed it as an agent reference, and never used it as the session navigation target. Tests cover simultaneous known child UUID plus opaque agent reference and an agent-only record with unavailable session identity; existing exact keyboard navigation remains covered.
- ARCH-01: Current TaskLinkFact outcome is required in TypeScript and current API validation. Native fixtures explicitly state unknown; missing/empty outcomes reject rather than default. Added malformed opaque-agent boundary coverage.
- DATA-04: Kept explicit unknown/succeeded/failed outcomes and known-outcome skill-rate denominator; no success boolean or opaque tool-completion fallback was introduced. Native activation timeline explicitly renders Outcome unavailable. Backend outcome evidence corrections feed these existing explicit fields.
- BEH-03/BEH-04: Frontend retains read-only child relationship projections and diagnostic-aware navigation. It does not persist derived UUID bindings or derive source-native IDs from namespaces.

Final correction verification: npm run typecheck and lint passed; full npm test passed all 815 tests in 36 suites; canonical npm run build passed and produced index-Bk2Om3Tv.js / index-Bn43kjvY.css. Embedded-tree comparison and frontend diff check passed. This supersedes prior frontend suite counts/artifact IDs for the corrected candidate. The informational bundle-size advisory remains unchanged. Parent owns final frozen revision, paired re-review, integrated make/standalone verification and publication.

Frontend paused again after correction checks; no concurrent edits during the next review freeze.

## Native call links and partial saved timeline integration

Mirrored native optional EventFact/SessionEventFact call_id, including validation and optional enrichment retention. Fact-to-event badge joins compare native call IDs within the proven session owner before recorded explicit event-ID fallback. Removed codex-call prefix inference; a mismatched native call cannot be matched by timestamps. Added normalized load_skill alongside skill to the supported source tool badges.

The source-shaped test/fixtures/codexCalls.json was produced by the backend tester from sanitized native load_skill and spawn_agent records through ParseRollout, SessionFacts, SaveFactLedger, reload and HTTP session DTO serialization. It preserves distinct event_id and call_id fields, explicit unavailable child session and separate opaque agent reference. The DOM test decodes/projects the exact fixture with payload-free saved events, checks both badges and unavailable-child controls, and exercises opaque event-ID representation while retaining native call IDs. The representation variant is a synthetic event-contract check, not a claim that the original source emitted alternative IDs; mismatched native calls reject.

Mirrored optional session events_are_partial boolean and reject malformed markers. Existing source badge and offline banner label the partial saved occurrence subset; no additional warning flow was added. The detail view conservatively retains that indication whenever any saved owner remains partial, rather than assuming a global hydration flag proves every owner's source trace is complete. Existing full-source/default presentation remains unchanged.

Affected final verification passed: typecheck, lint, all 310 tests across API client, event enrichment/projection, TraceWaterfall and TicketDeepDive; canonical build; embedded-tree comparison; frontend diff check. Current artifact is index-mKcHXOp5.js / index-Bn43kjvY.css. The preceding 815-case full run remains evidence for unchanged scopes, and parent owns final integrated make and standalone binary verification. Frontend paused again for the final freeze.
