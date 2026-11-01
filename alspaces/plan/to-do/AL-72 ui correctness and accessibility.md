h3. Title: AL-72: Correct UI interactions, calculations and accessibility

*TARGET*: AgenticLens (Frontend Trace Controls / Clipboard / Charts / Tables / Dialog and Tooltip)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit; consolidated draft batch at the user's request in iteration 09.
*Findings*: FE-06, FE-07, FE-08, FE-10, FE-11, FE-13, FE-14, FE-15, FE-16, FE-D3
*Related*: AL-69 owns anomaly attribution and count meaning; AL-71 owns request and transport state. Preserve their behavior during UI changes.

*CONTEXT*

Visible content, copy outcomes, chart geometry, aggregate denominators and keyboard/focus behavior must match the controls and claims presented to users.

The source reports retain affected paths, triggers, evidence and limits at their frozen baseline. Revalidate before implementation; this specification does not claim remediation. Overlapping findings and test/architecture analyses are counted once in the current finding map.

*GOAL*

Correct the scoped behavior through its production consumers and verification boundaries, satisfying every acceptance section below.

*VERTICAL SLICE SCOPE*

* Own the changes through Frontend Trace Controls / Clipboard / Charts / Tables / Dialog and Tooltip, including real consumer behavior and regression proof. Keep pure calculations in the core and I/O in adapters; preserve supported stored facts when enrichment is unavailable.
* Use bounded implementation and review iterations within this ticket. Grouping reduces tracking overhead; it does not authorize dropping acceptance cases, skipping focused reviews or calling a partial fix complete. No broad framework rewrite or new speculative abstraction is required.

*ACCEPTANCE CRITERIA*

h4. Fix trace keyboard activation and disclosure state

Parent key handlers intercept nested buttons; selected payloads cannot be hidden; Hide Reasoning changes only the label for short content.

* Give nested copy/payload/collapse controls native keyboard behavior without triggering parent selection; test Enter and Space and preserve child-session badge isolation.
* Separate selection-driven initial expansion from explicit disclosure state; selected payloads must hide and retain intended state through enrichment and controlled selection.
* Make reasoning disclosure match its label for short and long content; assert visible content, not merely label changes.
* Address FE-D3 composition-test overlap only with explicit assertion recipients preserving filters, denominator invariance, empty-to-recovered transitions and callback isolation. The clipboard and tooltip sections below retain the other FE-D3 oracles.

h4. Report clipboard outcomes accurately across consumers

The fallback ignores execCommand false and several consumers display Copied despite helper failure; ArtifactViewer maintains another fallback.

* Honor fallback booleans and exceptions, always remove temporary DOM nodes, and share result-aware behavior where it removes real duplication.
* All copy consumers show success only after a successful write; provide accessible failure/retry feedback.
* Test denied writes, fallback false, fallback throws and retry success at utility and representative consumers. Preserve verbatim payload and defensive prompt-fence assertions; this supplies the clipboard portion of FE-D3.

h4. Correct multi-series chart geometry and fleet success aggregation

Positive bar series share identical geometry and obscure each other. Overall fleet success averages ticket percentages, reporting 50 percent for 99 successes in 100 calls across unequal tickets.

* Define grouped or stacked positive series with a matching scale; verify geometry for multiple positives, mixed signs, unknown gaps and the shipped token-distribution configuration.
* Compute overall success from selected total calls/errors in a pure function, preserving search/type/prefix/skill scope and explicit zero-call behavior. Verify unequal denominators.
* Address FE-D2 benchmark aggregation ownership as part of the fix, preserving configured scope and unknown-cost semantics. Do not introduce a charting framework rewrite.

h4. Make table, drawer and tooltip interactions keyboard accessible

Table sorting/navigation is mouse-only, the modal drawer does not own focus, and hover CSS overrides Escape tooltip dismissal.

* Use keyboard-operable sorting/navigation with appropriate aria-sort and preserve tooltip propagation; cover TicketTable, comparison and landing headers.
* Implement drawer initial focus, containment/background behavior, Escape and focus return using a bounded native/existing mechanism.
* Give tooltip visibility one owner, preserve keyboard focus on Escape and keep it dismissed while the pointer remains hovered. Verify actual computed visibility with stylesheet-aware or browser evidence; class-only assertions are insufficient.
* Test keyboard flows and real visibility rather than role/label-only evidence, satisfying the tooltip portion of FE-D3.

* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with an approved concrete contract. Complete every scoped correction and evidence case; this planning session changes no application, test or runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: strict typing, zero any/interface{}, zero emojis, no domain mocks and self-contained fixtures. Pass affected checks, independent focused review and the required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Preserve source identity, unavailable values, meaningful existing assertions and applicable local-only boundaries. Record tested revision, reproduction/probe evidence, reused checks and environment limits; update finding dispositions only after verified fixes. Publish the implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
