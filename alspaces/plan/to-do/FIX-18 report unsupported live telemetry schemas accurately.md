h3. Title: FIX-18: Report unsupported live telemetry schemas accurately

*TARGET*: AgenticLens (OpenCode / Live Service / HTTP SSE / Dashboard)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: AL-67 architecture assessment; user-requested follow-up ticketing.
*Findings*: F03

*CONTEXT*

The server can open an OpenCode v2 database and announce that live telemetry is active, even though its live queries still read v1 tables. The live session list then fails, while background polling can drop the error. Historical trace hydration is a separate path and can still work with v2.

Example: select a valid database containing session_v2/session_message but no session/message/part. Startup succeeds and advertises an active hook; opening the live view cannot load a tree. This is a source-level counterexample, not a recorded production outage.

*GOAL*

Operators can tell whether live observation is supported, unavailable or failing, while retaining usable historical traces and ledger-backed views.

*VERTICAL SLICE SCOPE*

* Adapter/composition: inspect the selected database for the live operations actually supported before reporting activation; keep OpenCode schema detection inside its owning boundary.
* Presentation: carry truthful live availability through startup messages, REST/SSE and the existing dashboard error/availability flow.
* Quality: exercise production composition with isolated v1/v2/unknown/partial/mixed layouts and historical hydration.

*ACCEPTANCE CRITERIA*

* A valid v1 source retains its active tree, trace, polling and cleanup behavior.
* A v2-only source is explicitly reported as unsupported for current live operations; no active-hook notice, successful empty tree or silent stream is used to imply support. Historical v2 hydration and persisted facts remain available independently.
* Unknown, malformed, missing and unavailable sources have distinguishable diagnostics. Startup and subsequent REST/SSE behavior agree; polling/read failures become visible through the intended consumer flow.
* Verify ordinary v1/v2 extraction and exact-owner mixed-layout hydration retain their documented, different policies. Do not reject useful historical access just because live observation is unsupported.
* Tests reach the actual server-to-service selection path, including a source that connects successfully but lacks live tables; injected extractor tests alone are insufficient.

*SCOPE LIMITS*

No requirement to implement v2 or Codex live observation. AL-63 retains capability-matrix ownership; AL-74 owns the independent required-column test-oracle repair. Coordinate shared files without blocking this correction on those tickets.

*DEFINITION OF DONE*

* Revalidate the AL-67 source evidence at the implementation/review revision. Record actual checks, remaining limits and disposition; do not claim a production incident from a static counterexample.
* Work in a separate dedicated ticket session under the appropriate development/review workflow. Preserve portable ledger history, exact source/session/parent identity, legitimate repeated occurrences and explicit unavailable values. No application implementation is authorized by this planning document alone.
* Use self-contained fixtures; keep I/O outside the pure domain and follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md. Implementation requires affected checks, independent review and the required make gate; frontend changes also require canonical embedded-asset verification. Review-only work records evidence and recommendations without application/test edits.
* Publish the scoped deliverables and source-grounded telemetry. Do not mark completion from passing tests or a proposed design alone.

*SOURCE EVIDENCE*

* [AL-67 assessment](../../AL-67/ARCHITECTURE_ASSESSMENT.md), F03; application baseline a84adbe5be46c2ef52ede0987a417562715d38e1. Findings are unimplemented.
* [Follow-up ownership](../../AL-67/iterations/02/FOLLOW_UP_TICKETS.md).

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Still open: AL-67 F03 distinguishes v2 historical support from v1 live queries. Current live/sqlite.go still queries session/message/part; no merged corrective delivery was found.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.
