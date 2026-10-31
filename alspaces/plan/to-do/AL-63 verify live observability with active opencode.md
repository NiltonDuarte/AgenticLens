h3. Title: AL-63: Verify live observability with an active OpenCode session

*TARGET*: AgenticLens (OpenCode live tracing / dashboard)

*Phase*: Phase 1 (post-v2.0 platform review)
*Status*: Ready
*Work type*: test
*Blocked by*: None
*Blocks*: AL-64
*Origin*: PO-18; user review dated 2026-10-31, build v2.0.0-rc.106+9707d41-1-g45cb679.

*CONTEXT*

The Codex walkthrough had no Running or Active session and verified none of these live checks. Active OpenCode evidence is required; prior implementation tickets and static UI inspection do not prove runtime success.

*GOAL*

As an engineer observing an active session, I want a recorded live walkthrough so that refresh, navigation, and diagnostic evidence can be trusted.

*VERTICAL SLICE SCOPE*

* Domain Core: Validate live event identity, error classification, ordering, invocation facts, and child relationships through existing paths.
* Boundary Adapter: Exercise an actual supported OpenCode installation, recording exact version/schema family, session IDs, connection/reconnection times, and source location availability.
* Presentation: Exercise live inspector refresh, ordering/navigation, state preservation, failure expansion, invocation badges, child links, and Copy Prompt.
* Quality: Prefer existing automated checks plus a reproducible controlled live walkthrough. This is verification only; report defects for separate scope rather than implementing fixes.

*ACCEPTANCE CRITERIA*

* Verify automatic refresh of new turns and edits/tool results within existing turns; record delivery evidence and observed latency.
* Verify initial connection delivery and the startup subscription race with repeatable connection timing; reconnect does not lose or duplicate accepted facts.
* Verify latest-first/oldest-first ordering, navigation controls, reading-position preservation, and expanded-state preservation during updates.
* Verify collapsed failed-tool details, visible counts and expansion; invocation badges, available-child navigation and explicit unavailable-child behavior.
* Verify Copy Prompt retains full identifiers and captured source input/output/error evidence under the approved redaction policy.
* Record each check as passed, failed, or not executed with commands/build/session evidence. An unavailable active source is a blocker to live acceptance, never a pass. Preserve user data and do not reuse the changed AL-55 ledger as a disposable fixture.

*DELIVERY CONSTRAINTS*

* Phase: Phase 1 (post-v2.0 platform review); use AL identifiers, not AL2.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. FCIS; Go 1.23+ and TypeScript 5.x strict; zero any/interface{} and zero emojis.
* Tests and experiments use self-contained fixtures/disposable copies, never live ticket workspaces. Preserve alspaces/AL-55/agentic_metrics.json.
* No new capability is inferred from missing data. Separate source evidence, recorded observations, and hypotheses.

*DEFINITION OF DONE*

* Deliver the evidence report, limitations and follow-up recommendation in the active ticket workspace; no application fixes are part of this ticket.
* Record exact checks executed, failures, unavailable evidence and unexecuted checks. Existing test results may be reused for unchanged code.
* Maintain ticket summary/review/PR artifacts and source-grounded telemetry; publish a reviewable PR. Completion requires evidence, not just a plan.
