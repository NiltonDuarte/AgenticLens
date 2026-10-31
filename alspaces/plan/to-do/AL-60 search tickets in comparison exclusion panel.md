h3. Title: AL-60: Add Search tickets to the comparison exclusion panel

*TARGET*: AgenticLens (Dashboard comparisons)

*Status*: Ready
*Work type*: feat
*Blocked by*: None
*Blocks*: None
*Origin*: PO-18; user review dated 2026-10-31, build v2.0.0-rc.106+9707d41-1-g45cb679.

*CONTEXT*

The review requests a text input labeled Search tickets within Comparisons > Exclude tickets. Long ticket lists make selection difficult; searching must never change chart membership.

*GOAL*

As an analyst, I want to find tickets by ID or title inside the selection panel so that I can manage exclusions without losing selections.

*VERTICAL SLICE SCOPE*

* Domain Core: Preserve existing inclusion/exclusion semantics and chart datasets; no new metric or aggregation.
* Boundary Adapter: Reuse existing ticket identity/title data; verify titles remain available through the current API without adding extraction requirements.
* Presentation: Add an accessible Search tickets input and filter only the panel choices by ticket ID or title.
* Quality: Focused TypeScript interaction tests cover query changes, hidden selections, restoration, and empty results; existing Go/API contracts remain compatible.

*ACCEPTANCE CRITERIA*

* The input has the visible and accessible label Search tickets and matches ticket IDs or titles case-insensitively.
* Typing or clearing a query changes only visible selection choices, never ticket inclusion, exclusion counts, or plotted points.
* Previously selected/excluded tickets remain selected/excluded while hidden and when the query is cleared; toggling a visible choice still works.
* An empty result explains that no tickets match; the search input and query clearing remain usable.
* Keyboard access, existing ticket exclusion/restoration, point navigation, chart defaults, axis behavior, tooltips, and logarithmic rendering remain functional.

*DELIVERY CONSTRAINTS*

* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. FCIS; Go 1.23+ and TypeScript 5.x strict; zero any/interface{} and zero emojis.
* Tests and experiments use self-contained fixtures/disposable copies, never live ticket workspaces. Preserve alspaces/AL-55/agentic_metrics.json.
* No new capability is inferred from missing data. Separate source evidence, recorded observations, and hypotheses.

*DEFINITION OF DONE*

* Deliver reviewed implementation and meaningful regression tests through agentic-lens-dev in a separate dedicated session with approved scope.
* Pass required make verification; frontend changes include the canonical build and verification of embedded standalone assets.
* Maintain ticket summary/review/PR artifacts and source-grounded telemetry; publish a reviewable PR. Completion requires evidence, not just a plan.
