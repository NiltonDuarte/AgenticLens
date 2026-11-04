h3. Title: AL-77: Warn when a session appears in multiple tickets

*TARGET*: AgenticLens (Frontend / read-only session-to-ticket projection)

*Status*: Done
*Work type*: feat
*Blocked by*: None
*Blocks*: None
*Origin*: Explicit user request during MGT-2 platform-definition clarification.

*CONTEXT*

Ledgers are trusted project records shared through Git. A single session ID can be recorded in multiple tickets, making ticket comparisons and combined metrics overlap. AgenticLens cannot determine which attribution the user intended. The platform should make the overlap visible and provide enough context for the user to resolve it at their end.

*GOAL*

Show a clear frontend warning identifying each shared session and its affected tickets, while preserving recorded data and existing aggregation behavior. Session ID is the stable identity; no machine-local receipts, source access or new identity scheme is required to detect the overlap.

*VERTICAL SLICE SCOPE*

* Derive an index from stored session IDs to distinct ticket IDs in the loaded project ledger dataset. Keep source labels as context, not an additional machine-specific matching requirement.
* Use existing read projections; if necessary, expose the minimal read-only identifiers needed by the frontend. No extraction, ledger-schema or storage redesign is included.
* Surface the warning in the frontend's ticket/fleet context with a discoverable detail view. Reuse existing navigation to affected tickets where available.
* This ticket is queued only. No development thread or implementation is started by its creation.

*ACCEPTANCE CRITERIA*

* A session ID appearing in at least two distinct ticket IDs produces a warning. Repeated rows or repeated projections of the same ticket alone do not create a cross-ticket warning.
* Warning details identify the session ID and all affected tickets within the inspected dataset, with readable labels and available ticket links. Explain that metrics may overlap and the user must decide how to correct attribution; do not declare one ticket correct.
* Detection uses the available project dataset rather than silently interpreting a filtered subset as complete. State the inspected scope when data is partial; do not claim that no overlap exists outside it.
* Valid persisted ledgers suffice even when all original sources and private capture databases are unavailable. Missing session IDs do not get guessed from names, timestamps, paths or developer identities.
* Do not delete, move, reassign or silently deduplicate sessions across tickets. Do not change ledger contents, per-ticket totals or fleet allocation as an automatic response to the warning. Refreshing the view is read-only.
* A single-ticket session produces no warning. When the user corrects their records and the data is reloaded, the warning reflects the new dataset without retaining stale conflicts.
* Warning details and ticket navigation work with keyboard and assistive technology; preserve existing filtering and selection behavior.
* Verification exercises a genuine two-ticket overlap, same-ticket repetition, no-overlap, source-unavailable records, partial/filtered scope and resolution after reload. Use self-contained fixtures, not live ticket history. Follow normal development/review/build/embed gates when implementing.

*DEFINITION OF DONE*

* The warning and its read-only projection are implemented and verified through actual frontend consumers.
* Applicable checks and required integrated make pass; shipped embedded assets match the frontend build.
* No automatic data repair or aggregate-policy change is introduced. The user retains the decision about correcting attribution.

*CANONICAL REFERENCES*

* [Portable ledger contract](../../../TECHNICAL_GUIDELINES.md#portable-ledger-contract)
* [Portable ledger acceptance scenarios](../../../docs/REGRESSION_PREVENTION.md#portable-ledgers-across-developers-and-machines)
