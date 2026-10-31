h3. Title: AL-64: Investigate and fix active-turn duration across Codex and OpenCode v1/v2

*TARGET*: AgenticLens (Domain / Codex / OpenCode v1 and v2 / Generation / Backfill / Ledger / Presentation)

*Status*: Ready
*Work type*: fix
*Blocked by*: None
*Blocks*: None
*Origin*: PO-18 consolidated specification; original review dated 2026-10-31, build v2.0.0-rc.106+9707d41-1-g45cb679. User explicitly expands investigation to OpenCode v1 and v2; scope is not limited to Codex.

*CONTEXT*

Diagnosed bug in reviewed build v2.0.0-rc.106+9707d41-1-g45cb679: the Codex extractor accrues paired start/completion events only. Generation during an active turn omits its elapsed time. The user explicitly requires a fix. AL-55 root 01a19267-636b-7d61-85c6-db8d3d3eba77 started its second turn at 2026-10-31T15:11:36.949Z; its own final generation command ran at 15:29:26.541Z, ledger update 15:29:28.051249Z, completion 15:30:26.217Z. Source refresh increased root active seconds 287 -> 1417 and total 2058 -> 3188; the same four children retained unchanged durations. The 1130-second delta matches the newly completed turn, not duplicate accumulation or review work. Evidence is user-supplied and must be preserved.

OpenCode v1 and v2 must be investigated independently for the same active-duration boundary. The Codex diagnosis is established in the supplied notes; an equivalent OpenCode defect is not yet established. Inspect each schema family's real start/completion/in-progress records and current duration calculation, then fix any confirmed accounting defects within this ticket.

*GOAL*

As an operator generating telemetry from Codex, OpenCode v1 or OpenCode v2, I want accurate completed and in-progress active duration through the snapshot cutoff, with provisional/unsupported evidence made explicit, so that the report includes source-proven current activity and later refresh converges without double counting. Deliver the diagnosed Codex fix and investigate and remediate any equivalent OpenCode defects in the same ticket.

*VERTICAL SLICE SCOPE*

* Domain Core: Define deterministic completed/provisional active-duration semantics and additive rollups with explicit provisional coverage; supply cutoff as data so the core has no clock I/O.
* Boundary Adapter: Trace Codex and OpenCode schema v1/v2 duration paths separately through generation, persistence and source-backed refresh. Identify authoritative in-progress boundaries for each source, extract supported unfinished activity through a recorded snapshot cutoff and persist enough provenance/state to replace provisional facts once completion is available. Implement confirmed fixes in the affected adapter(s); select compatible schema evolution only if needed.
* Presentation: Surface provisional duration and cutoff consistently in ledger/API, CLI summaries and dashboard duration consumers without implying a completed measurement.
* Quality: Source-family-specific generation and refresh regression coverage for Codex, OpenCode v1 and OpenCode v2; pure duration boundary tests and meaningful projection checks using self-contained fixtures. Reuse existing coverage for already-correct paths when it proves the required boundaries.

*ACCEPTANCE CRITERIA*

* Record a per-source investigation result for Codex, OpenCode v1 and OpenCode v2: source version/schema, authoritative activity boundaries, current calculation and generated/persisted values. Distinguish confirmed defect, verified correct behavior, unsupported evidence and unexecuted checks; do not assume the Codex event-pairing cause applies to OpenCode.
* Fix the diagnosed Codex bug and any confirmed OpenCode v1/v2 active-duration defects as part of this ticket. An OpenCode investigation report alone is insufficient when it establishes a supported-path accounting bug; already-correct behavior must be preserved.
* For each source with authoritative unfinished-activity evidence, generation during an unfinished turn includes elapsed active time from its start through the snapshot cutoff and explicitly identifies that duration as provisional.
* Completed turns retain measured duration. Time between turns is excluded; events later than the snapshot cutoff never contribute future time. Invalid/orphan event sequences are handled explicitly without negative or invented duration.
* Refresh after completion replaces the provisional contribution with completed duration, including in root/total rollups, without adding both snapshots or altering unchanged children.
* Repeated backfill against unchanged source is stable and does not duplicate duration or sessions. Refresh while still active advances only to the new cutoff.
* Verify each of Codex, OpenCode v1 and OpenCode v2 independently: generation while activity is unfinished, refresh while still active, completion/refresh and unchanged repeated backfill. Cover an active root with children, other roots/harnesses and unavailable source retention. Use source-backed runtime evidence when available plus deterministic adapter-to-output regression fixtures; inaccessible runtime environments must be explicitly unverified, never reported as passing. Never use the live AL-55 workspace as a test fixture.
* Where a source lacks reliable start/in-progress boundaries, preserve known duration and report the limitation rather than inventing elapsed activity. Missing completion alone does not prove an interval is still active; do not replace active duration with whole-session wall time. Map source-specific assistant-message/turn semantics explicitly and preserve established rollup behavior without accidental duplication of overlapping or repeated records.
* Use --recompute-db when demonstrating source refresh; --force alone is not source recomputation. Moving generation later within the same turn does not satisfy this fix.
* Document compatible completed/provisional/unsupported read and display semantics across all three source families. Use expand-contract if a shared schema change is necessary; retain all existing valid facts.

*DELIVERY CONSTRAINTS*

* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. FCIS; Go 1.23+ and TypeScript 5.x strict; zero any/interface{} and zero emojis.
* Tests and experiments use self-contained fixtures/disposable copies, never live ticket workspaces. Preserve alspaces/AL-55/agentic_metrics.json.
* No new capability is inferred from missing data. Separate source evidence, recorded observations, and hypotheses.

*DEFINITION OF DONE*

* Deliver per-source investigation evidence, the diagnosed Codex fix, all confirmed OpenCode accounting fixes and meaningful regression coverage through agentic-lens-dev in a separate dedicated session with approved scope. This is active-duration accounting work, not restoration of the removed standalone live-inspector walkthrough.
* Pass required make verification; frontend changes include the canonical build and verification of embedded standalone assets.
* Maintain ticket summary/review/PR artifacts and source-grounded telemetry; publish a reviewable PR. Completion requires evidence, not just a plan.
