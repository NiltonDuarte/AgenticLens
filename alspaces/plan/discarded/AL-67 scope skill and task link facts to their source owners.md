h3. Title: AL-67: Scope skills and task_links to their source session or invocation

*TARGET*: AgenticLens (Domain / Source Adapters / Schema Migration / Backfill / CLI / API / Dashboard / Tests)

*Phase*: Phase 1 (post-v2.0 platform review)
*Status*: Discarded (Merged into AL-62)
*Work type*: refactor
*Blocked by*: None
*Blocks*: None
*Origin*: PO-18 iteration 02; user requests granular ownership for skills and task_links, keeping facts as close to raw data as possible.

*MERGE NOTICE*

Retired by explicit user instruction in PO-18 iteration 04. All granular ownership, migration, consumer and regression acceptance is incorporated into AL-62: Fix Codex skill and delegation facts with granular source ownership. This file is a historical specification, not a separately executable ticket.

*CONTEXT*

AgenticFactLedger currently stores skills and task_links as ticket-level collections. SessionFact already owns invocations, tools and events. Existing SkillFact and TaskLinkFact contain turn/message positions and timestamps, but those positions alone do not establish ownership across sessions. InvocationFact describes role/model execution metrics and is not necessarily one raw message/tool occurrence. Ticket-level collections obscure where a fact originated, particularly when a ticket contains a root and multiple children or multiple sources.

*GOAL*

As an engineer inspecting telemetry, I want each skill invocation and delegation fact stored at its narrowest source-proven owner, session or invocation, so that I can inspect its provenance and derive trustworthy ticket/fleet analytics without losing raw occurrence detail. Move canonical storage away from unowned ticket-level collections; broader views become derived projections.

*VERTICAL SLICE SCOPE*

* Domain Core: Define source-owned skill/delegation facts and pure flattening/aggregation semantics. Select the narrowest granularity actually supported by source identities; distinguish invocation metrics rows from raw invocation occurrences. Preserve repeated occurrences and parent/child relationship direction.
* Boundary Adapter: Carry authoritative ownership and available source session/message/tool/invocation identifiers through OpenCode v1, OpenCode v2 and Codex extraction where those facts are recorded, live ingestion, persistence, optional hydration and backfill. Missing harness evidence is explicit, never fabricated to achieve parity.
* Schema and Compatibility: Use the existing versioned migration framework and expand-contract sequence: expand granular ownership and compatible reads, migrate extractors/writers/consumers coherently, then retire legacy canonical writes. The developer contract selects the schema version and representation. Preserve unassignable legacy facts with explicit unresolved provenance rather than forcing an owner.
* Presentation: Keep CLI, API, Skills & Subagents, timeline badges/child navigation, fleet analytics and comparison filters functional using owner-derived projections. Expose the owning session and finer source identity where available; ticket-wide responses may retain derived compatibility views without duplicate canonical storage.
* Quality: Self-contained Go domain/adapter/migration fixtures and focused TypeScript tests verify ownership through actual persistence/read paths, aggregates, hydration failures, child navigation, and safe repeated refresh.

*ACCEPTANCE CRITERIA*

* New facts with known source ownership are canonically stored with their owning session; use invocation scope when an unambiguous source invocation identity exists. No fact is stored more coarsely merely for aggregation convenience, or more finely than the source proves. Document the ownership rule per supported source/schema family in the developer contract.
* Preserve source identifiers, original available values, timestamps, outcomes/errors, turn/message positions, and repeated legitimate occurrences. Session-local indexes, role, model, name or timestamp proximity alone never establish cross-session ownership. Do not merge equal-looking facts or invent invocation IDs to force placement in a metrics row.
* A task_link belongs to the delegating parent occurrence/session and retains the explicit child-session relationship. A referenced child does not become its owner. An unavailable child remains navigable as an explicit unavailable state; recorded child sessions alone do not establish a delegation invocation.
* Root and child sessions with overlapping turn/message positions, two independent roots, repeated identical-looking calls, and mixed harnesses remain distinct. Session-scoped views show only their facts; ticket/fleet aggregations and filters count each occurrence exactly once.
* Explicit migration safely preserves existing top-level skills/task_links. Relocate a historical fact only with reliable ownership evidence. Ambiguous or source-unavailable facts remain losslessly retained with an explicit unresolved ownership diagnostic/compatibility representation; never attach them arbitrarily to the root, drop them, or report them as a new known occurrence. Known recorded totals remain available without claiming unavailable session attribution.
* Existing ledgers with omitted/empty collections, legacy populated collections, granular collections, and transitional representations have documented compatible behavior. Transitional duplicate projections do not double count facts. Ordinary read/GET does not silently rewrite ledgers; derived views are not independent canonical fact stores.
* Exact-identity backfill and optional hydration preserve stored granular facts, unresolved legacy history, previous trees and other-source sessions when source records are missing, malformed or partial. Growing snapshots replace their own facts without accumulation; unchanged repeats are stable. Preserve incomplete/mixed-age diagnostics and explicit opt-in heuristic recovery.
* Persisted facts remain visible with enrichment unavailable, pending or failed. CLI/API/dashboard projections continue to provide existing skill analytics, filters, timeline evidence and child links without requiring a live source; missing evidence is distinguished from known zero/no recorded facts.
* Schema ownership changes follow expand-contract with compatibility and rollback/preservation considerations recorded before removal of old writes. Coordinate with concurrent schema owners such as AL-57 rather than preselecting a version. Reuse AL-62/AL-64 findings when available; this ticket does not depend on inventing missing Codex facts or equal harness capabilities.
* Pass the required make gate and meaningful regression checks for ownership, multiplicity, ambiguous migration and preservation. Frontend changes require canonical asset build and embedded standalone verification. Tests never use the live AL-55 ledger/workspace as a fixture.

*DEFINITION OF DONE*

* A separate agentic-lens-dev session freezes the ownership/compatibility contract and obtains implementation scope approval under the normal gate; this PO ticket creates no implementation.
* Deliver the granular persistence-to-presentation slice with compatible migration and reviewed tests. No schema-only intermediate release leaves accepted consumers broken or facts missing.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: pure Go domain, I/O in adapters, strict TypeScript, zero any/interface{}, zero emojis and no domain mocks.
* Document the canonical owner, derived aggregate contract, unavailable/ambiguous behavior, compatibility strategy and executed verification in ticket artifacts. Preserve original evidence and record source limitations.
* Publish the reviewed implementation PR with source-grounded telemetry and required checks; completion requires integrated delivery evidence.
