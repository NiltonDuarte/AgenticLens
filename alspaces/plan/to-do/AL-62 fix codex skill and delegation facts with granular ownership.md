h3. Title: AL-62: Fix Codex skill and delegation facts with granular source ownership

*TARGET*: AgenticLens (Domain / Codex and OpenCode Adapters / Schema Migration / Backfill / CLI / API / Dashboard / Tests)

*Status*: Ready
*Work type*: fix
*Blocked by*: None
*Blocks*: AL-63
*Origin*: PO-18 iteration 04; user explicitly requires AL-62 to fix missing facts and merges AL-67's granular data-structure scope into this ticket. AL-67 is retired as merged here.

*CONTEXT*

User-supplied observations from the reviewed build: Fleet Overview showed 0 unique skills, Overall skill success rate: Unavailable, and No recorded skills with no filters; all 102 inspected saved project ledgers lacked populated top-level skills; the selected Skills & Subagents panel was empty. AL-55 nevertheless contains a root and four child Codex sessions. These are reported review observations, not independently repeated checks in PO-18.

AgenticFactLedger currently stores skills and task_links as ticket-level collections. SessionFact already owns invocations, tools and events. Existing SkillFact and TaskLinkFact contain turn/message positions and timestamps, but those positions alone do not establish ownership across sessions. InvocationFact describes role/model execution metrics and is not necessarily one raw message/tool occurrence. Ticket-level collections obscure where a fact originated, particularly when a ticket contains a root and multiple children or multiple sources.

Investigation is a required implementation step, not the final deliverable. The original cause is not assumed to be ticket-level storage: identify source, extraction, persistence and presentation losses separately, then fix the supported paths within this same ticket.

*GOAL*

As an engineer, I want source-recorded Codex skill and delegation facts to survive extraction, generation, persistence and presentation under their narrowest proven session or invocation owner, so that ticket/fleet views are trustworthy and raw provenance is retained. Deliver the missing-fact fix together with the granular schema, compatible migration and all affected consumers. Broader analytics derive from canonical owned facts.

*VERTICAL SLICE SCOPE*

* Investigation and Remediation: Establish which Codex facts and identities are actually recorded using real source evidence and existing tests; diagnose losses independently and implement the corrections in extraction, persistence, hydration and/or presentation where required. Include the diagnosed fixes in this delivery, not merely recommendations for another ticket.
* Domain Core: Define source-owned skill/delegation facts and pure flattening/aggregation semantics. Select the narrowest granularity actually supported by source identities; distinguish invocation metrics rows from raw invocation occurrences. Preserve repeated occurrences and parent/child relationship direction.
* Boundary Adapter: Carry authoritative ownership and available source session/message/tool/invocation identifiers through OpenCode v1, OpenCode v2 and Codex extraction where those facts are recorded, live ingestion, persistence, optional hydration and backfill. Missing harness evidence is explicit, never fabricated to achieve parity.
* Schema and Compatibility: Use the existing versioned migration framework and expand-contract sequence: expand granular ownership and compatible reads, migrate extractors/writers/consumers coherently, then retire legacy canonical writes. The developer contract selects the schema version and representation. Preserve unassignable legacy facts with explicit unresolved provenance rather than forcing an owner.
* Presentation: Keep CLI, API, Skills & Subagents, timeline badges/child navigation, fleet analytics and comparison filters functional using owner-derived projections. Expose the owning session and finer source identity where available; ticket-wide responses may retain derived compatibility views without duplicate canonical storage.
* Quality: Self-contained Go domain/adapter/migration fixtures and focused TypeScript tests verify ownership through actual persistence/read paths, aggregates, hydration failures, child navigation, and safe repeated refresh.

*ACCEPTANCE CRITERIA*

* Identify the Codex versions/source records inspected with privacy-safe session/event references, distinguish available evidence from unsupported/unavailable recording, and document the diagnosed loss points. Account for AL-55 root 01a19267-636b-7d61-85c6-db8d3d3eba77 and its four children independently from delegation invocation facts.
* Fix missing skill/delegation facts that the source records: prove representative populated source facts reach extraction, generated ledger, persisted reload, CLI/API and Skills & Subagents/fleet projections, including filters and child navigation where applicable. Reproduce the diagnosed failure with meaningful regression tests and show the fixed path. An investigation report or schema relocation alone does not satisfy this ticket.
* Where a source genuinely lacks the required evidence, implement accurate unsupported/unavailable/no-recorded-facts behavior without fabricated skill/delegation activity or success rates. Unavailable local source records must be reported as unverified and cannot substitute for regression evidence on a supported path. Child sessions alone do not prove delegation invocations.
* New facts with known source ownership are canonically stored with their owning session; use invocation scope when an unambiguous source invocation identity exists. No fact is stored more coarsely merely for aggregation convenience, or more finely than the source proves. Document the ownership rule per supported source/schema family in the developer contract.
* Preserve source identifiers, original available values, timestamps, outcomes/errors, turn/message positions, and repeated legitimate occurrences. Session-local indexes, role, model, name or timestamp proximity alone never establish cross-session ownership. Do not merge equal-looking facts or invent invocation IDs to force placement in a metrics row.
* A task_link belongs to the delegating parent occurrence/session and retains the explicit child-session relationship. A referenced child does not become its owner. An unavailable child remains navigable as an explicit unavailable state; recorded child sessions alone do not establish a delegation invocation.
* Root and child sessions with overlapping turn/message positions, two independent roots, repeated identical-looking calls, and mixed harnesses remain distinct. Session-scoped views show only their facts; ticket/fleet aggregations and filters count each occurrence exactly once.
* Explicit migration safely preserves existing top-level skills/task_links. Relocate a historical fact only with reliable ownership evidence. Ambiguous or source-unavailable facts remain losslessly retained with an explicit unresolved ownership diagnostic/compatibility representation; never attach them arbitrarily to the root, drop them, or report them as a new known occurrence. Known recorded totals remain available without claiming unavailable session attribution.
* Existing ledgers with omitted/empty collections, legacy populated collections, granular collections, and transitional representations have documented compatible behavior. Transitional duplicate projections do not double count facts. Ordinary read/GET does not silently rewrite ledgers; derived views are not independent canonical fact stores.
* Exact-identity backfill and optional hydration preserve stored granular facts, unresolved legacy history, previous trees and other-source sessions when source records are missing, malformed or partial. Growing snapshots replace their own facts without accumulation; unchanged repeats are stable. Preserve incomplete/mixed-age diagnostics and explicit opt-in heuristic recovery.
* Persisted facts remain visible with enrichment unavailable, pending or failed. CLI/API/dashboard projections continue to provide existing skill analytics, filters, timeline evidence and child links without requiring a live source; missing evidence is distinguished from known zero/no recorded facts.
* Schema ownership changes follow expand-contract with compatibility and rollback/preservation considerations recorded before removal of old writes. Coordinate with concurrent schema owners such as AL-57 rather than preselecting a version. Use this ticket's source investigation to implement evidence-backed fixes; AL-63 consumes the delivered verification and is not an upstream dependency. Do not invent missing Codex facts or equal harness capabilities.
* Pass the required make gate and meaningful regression checks for ownership, multiplicity, ambiguous migration and preservation. Frontend changes require canonical asset build and embedded standalone verification. Tests never use the live AL-55 ledger/workspace as a fixture.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session under the normal reviewed contract and approval gate. The user explicitly requested a fix plus granular structure; preserve that scope rather than reverting to an investigation-only plan. This PO session modifies planning artifacts only.
* Deliver and review source-grounded diagnosis, actual missing-fact corrections, granular storage, safe migration, coherent consumers and regression tests as one integrated slice. Internal expand-contract stages remain under AL-62; do not publish a schema-only release that leaves consumers broken or loses accepted facts.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: pure domain, strict Go/TypeScript, zero any/interface{}, zero emojis, no domain mocks, self-contained fixtures and canonical embedded-asset verification for frontend changes. Preserve the live AL-55 evidence.
* Record tested diff/commands/results, source limits and any unverified paths; publish the reviewed implementation PR and source-grounded telemetry. Completion requires the implemented fix and integrated verification, not a recommendation to implement later.
