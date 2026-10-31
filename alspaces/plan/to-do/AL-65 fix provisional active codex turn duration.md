h3. Title: AL-65: Include provisional active Codex turn duration at the snapshot cutoff

*TARGET*: AgenticLens (Codex duration extraction / ledger / presentation)

*Phase*: Phase 1 (post-v2.0 platform review)
*Status*: Ready
*Work type*: fix
*Blocked by*: None
*Blocks*: None
*Origin*: PO-18; user review dated 2026-10-31, build v2.0.0-rc.106+9707d41-1-g45cb679.

*CONTEXT*

Diagnosed bug in reviewed build v2.0.0-rc.106+9707d41-1-g45cb679: the Codex extractor accrues paired start/completion events only. Generation during an active turn omits its elapsed time. The user explicitly requires a fix. AL-55 root 01a19267-636b-7d61-85c6-db8d3d3eba77 started its second turn at 2026-10-31T15:11:36.949Z; its own final generation command ran at 15:29:26.541Z, ledger update 15:29:28.051249Z, completion 15:30:26.217Z. Source refresh increased root active seconds 287 -> 1417 and total 2058 -> 3188; the same four children retained unchanged durations. The 1130-second delta matches the newly completed turn, not duplicate accumulation or review work. Evidence is user-supplied and must be preserved.

*GOAL*

As an operator generating telemetry inside the current Codex turn, I want elapsed active duration through the snapshot cutoff marked provisional so that the report includes current work and later refresh converges without double counting.

*VERTICAL SLICE SCOPE*

* Domain Core: Define deterministic completed/provisional active-duration semantics and additive rollups with explicit provisional coverage; supply cutoff as data so the core has no clock I/O.
* Boundary Adapter: Extract unfinished Codex turns through a recorded telemetry snapshot cutoff; persist enough provenance/state to replace provisional facts once completion is available. Select compatible schema evolution only if needed.
* Presentation: Surface provisional duration and cutoff consistently in ledger/API, CLI summaries and dashboard duration consumers without implying a completed measurement.
* Quality: End-to-end generation and refresh regression coverage, pure duration boundary tests, and meaningful projection checks using self-contained fixtures.

*ACCEPTANCE CRITERIA*

* Generation during an unfinished turn includes elapsed active time from its start through the snapshot cutoff and explicitly identifies that duration as provisional.
* Completed turns retain measured duration. Time between turns is excluded; events later than the snapshot cutoff never contribute future time. Invalid/orphan event sequences are handled explicitly without negative or invented duration.
* Refresh after completion replaces the provisional contribution with completed duration, including in root/total rollups, without adding both snapshots or altering unchanged children.
* Repeated backfill against unchanged source is stable and does not duplicate duration or sessions. Refresh while still active advances only to the new cutoff.
* Verify direct generation from an active root, subsequent completion/refresh, other sessions and harnesses, and retained stored facts when source evidence is unavailable. Never use the live AL-55 workspace as a test fixture.
* Use --recompute-db when demonstrating source refresh; --force alone is not source recomputation. Moving generation later within the same turn does not satisfy this fix.
* Document compatibility and provisional read/display semantics. Use expand-contract if a shared schema change is necessary; retain all existing valid facts.

*DELIVERY CONSTRAINTS*

* Phase: Phase 1 (post-v2.0 platform review); use AL identifiers, not AL2.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. FCIS; Go 1.23+ and TypeScript 5.x strict; zero any/interface{} and zero emojis.
* Tests and experiments use self-contained fixtures/disposable copies, never live ticket workspaces. Preserve alspaces/AL-55/agentic_metrics.json.
* No new capability is inferred from missing data. Separate source evidence, recorded observations, and hypotheses.

*DEFINITION OF DONE*

* Deliver reviewed implementation and meaningful regression tests through agentic-lens-dev in a separate dedicated session with approved scope.
* Pass required make verification; frontend changes include the canonical build and verification of embedded standalone assets.
* Maintain ticket summary/review/PR artifacts and source-grounded telemetry; publish a reviewable PR. Completion requires evidence, not just a plan.
