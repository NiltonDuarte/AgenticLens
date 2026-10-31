h3. Title: AL-62: Investigate missing Codex skill and delegation facts

*TARGET*: AgenticLens (Codex ingestion / persistence / analytics)

*Phase*: Phase 1 (post-v2.0 platform review)
*Status*: Ready
*Work type*: docs
*Blocked by*: None
*Blocks*: AL-64
*Origin*: PO-18; user review dated 2026-10-31, build v2.0.0-rc.106+9707d41-1-g45cb679.

*CONTEXT*

User-supplied observations from the reviewed build: Fleet Overview showed 0 unique skills, Overall skill success rate: Unavailable, and No recorded skills with no filters; all 102 inspected saved project ledgers lacked populated top-level skills; the selected Skills & Subagents panel was empty. AL-55 nevertheless contains a root and four child Codex sessions. These are reported review observations, not independently repeated checks in PO-18.

*GOAL*

As an operator, I want an evidence-backed explanation of missing skill and delegation facts so that unavailable recording is not confused with no usage or no children.

*VERTICAL SLICE SCOPE*

* Domain Core: Audit distinctions between skill invocation, delegation invocation, child session ownership, and unavailable/empty facts.
* Boundary Adapter: Trace real Codex source records through discovery, extraction, serialization, persistence, optional hydration, and API projection separately.
* Presentation: Inspect ticket and fleet projections and filters against persisted data with hydration available and unavailable.
* Quality: Produce a bounded source-to-UI evidence report using existing tests and real records where available. No application, schema, or new test changes in this investigation.

*ACCEPTANCE CRITERIA*

* Identify which skill and delegation facts the inspected Codex versions actually record, with source/session IDs and privacy-safe event references; explicitly mark inaccessible evidence.
* For each missing fact, classify source absence/unavailability, extraction loss, persistence loss, or presentation loss; do not claim a cause without evidence.
* Account for AL-55 root 01a19267-636b-7d61-85c6-db8d3d3eba77 and its four recorded children separately from delegation facts. Missing invocation facts do not imply missing child sessions.
* Distinguish known zero, no recorded facts, unsupported recording, and unavailable source evidence; do not infer success rates from an empty collection.
* Deliver findings and narrowly scoped follow-up recommendations with verification criteria; any implementation requires a separate approved development scope.

*DELIVERY CONSTRAINTS*

* Phase: Phase 1 (post-v2.0 platform review); use AL identifiers, not AL2.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. FCIS; Go 1.23+ and TypeScript 5.x strict; zero any/interface{} and zero emojis.
* Tests and experiments use self-contained fixtures/disposable copies, never live ticket workspaces. Preserve alspaces/AL-55/agentic_metrics.json.
* No new capability is inferred from missing data. Separate source evidence, recorded observations, and hypotheses.

*DEFINITION OF DONE*

* Deliver the evidence report, limitations and follow-up recommendation in the active ticket workspace; no application fixes are part of this ticket.
* Record exact checks executed, failures, unavailable evidence and unexecuted checks. Existing test results may be reused for unchanged code.
* Maintain ticket summary/review/PR artifacts and source-grounded telemetry; publish a reviewable PR. Completion requires evidence, not just a plan.
