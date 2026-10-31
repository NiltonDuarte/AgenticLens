h3. Title: AL-61: Show full local historical timestamps in Ticket Deep Dive

*TARGET*: AgenticLens (Dashboard historical timestamps)

*Status*: Ready
*Work type*: feat
*Blocked by*: None
*Blocks*: None
*Origin*: PO-18; user review dated 2026-10-31, build v2.0.0-rc.106+9707d41-1-g45cb679.

*CONTEXT*

Ticket Deep Dive currently shows time alone and requires hover for the full date. The user confirmed the tooltip already includes the date, but historical metrics need that context by default.

*GOAL*

As an engineer reviewing historical metrics, I want full local dates and times visible so that I can interpret sessions across different days without hovering.

*VERTICAL SLICE SCOPE*

* Domain Core: Preserve source instants and duration calculations; no change to timestamp storage.
* Boundary Adapter: Reuse existing timestamp values and verify transport preserves the instant.
* Presentation: Render full local date and time without milliseconds by default on historical timestamp labels in Ticket Deep Dive; inventory affected labels in the developer contract.
* Quality: Focused formatter/view tests cover multiple dates, local timezone conversion, missing values, and millisecond omission.

*ACCEPTANCE CRITERIA*

* Default visible historical timestamp labels include both local calendar date and time, without milliseconds or a hover requirement.
* Formatting uses the viewer local timezone consistently and retains existing timezone context; date boundaries and daylight-saving offsets do not change the represented instant.
* Unavailable or invalid timestamps have an explicit existing-compatible unavailable state, never a fabricated date.
* Full dates remain readable at narrow widths and through keyboard/accessibility surfaces; tooltips may supplement the visible value.
* Stored timestamps, ordering, durations, live inspector behavior, and unrelated chart formatting remain unchanged.

*DELIVERY CONSTRAINTS*

* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. FCIS; Go 1.23+ and TypeScript 5.x strict; zero any/interface{} and zero emojis.
* Tests and experiments use self-contained fixtures/disposable copies, never live ticket workspaces. Preserve alspaces/AL-55/agentic_metrics.json.
* No new capability is inferred from missing data. Separate source evidence, recorded observations, and hypotheses.

*DEFINITION OF DONE*

* Deliver reviewed implementation and meaningful regression tests through agentic-lens-dev in a separate dedicated session with approved scope.
* Pass required make verification; frontend changes include the canonical build and verification of embedded standalone assets.
* Maintain ticket summary/review/PR artifacts and source-grounded telemetry; publish a reviewable PR. Completion requires evidence, not just a plan.
