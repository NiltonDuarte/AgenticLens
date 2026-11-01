h3. Title: AL-74: Strengthen test reliability and remove unused code

*TARGET*: AgenticLens (Domain Migration and Aggregation / Adapter and CLI Tests / Frontend Cleanup / Compatibility Fixtures)

*Status*: Ready
*Work type*: refactor
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit; consolidated draft batch at the user's request in iteration 09.
*Findings*: ADP-07, ADP-08, ADP-09, ADP-13, CC-09, CC-10, CC-11, CC-12, D4, FE-D1
*Related*: Coordinate pure aggregation changes with AL-69, UI cleanup with AL-72, and dependency/lockfile edits with AL-75. These are shared-file coordination needs, not speculative blocking dependencies.

*CONTEXT*

Pure transformations must not mutate inputs, tests must exercise their claimed boundary in isolated environments, and unused production surfaces and duplicated setup should be removed while preserving meaningful assertions.

The source reports retain affected paths, triggers, evidence and limits at their frozen baseline. Revalidate before implementation; this specification does not claim remediation. Overlapping findings and test/architecture analyses are counted once in the current finding map.

*GOAL*

Correct the scoped behavior through its production consumers and verification boundaries, satisfying every acceptance section below.

*VERTICAL SLICE SCOPE*

* Own the changes through Domain Migration and Aggregation / Adapter and CLI Tests / Frontend Cleanup / Compatibility Fixtures, including real consumer behavior and regression proof. Keep pure calculations in the core and I/O in adapters; preserve supported stored facts when enrichment is unavailable.
* Use bounded implementation and review iterations within this ticket. Grouping reduces tracking overhead; it does not authorize dropping acceptance cases, skipping focused reviews or calling a partial fix complete. No broad framework rewrite or new speculative abstraction is required.

*ACCEPTANCE CRITERIA*

h4. Remove discarded core work and preserve migration input values

V1 migration mutates its input through a shared session slice. Core aggregation builds unread mapped-event structures and duplicates model grouping; some tests construct and discard their purported sorting fixtures.

* Clone mutable migration containers before transformation; prove input nonmutation and alias independence for nil/non-nil events and appendable metadata.
* Remove unused mappedEvents and anomalyEvidenceSelection.events work and reuse existing model aggregation where behavior is equivalent. Preserve role order, thread ceilings, precise and incomplete-cost behavior; claim no unmeasured speedup.
* Repair the actual sorting oracle or remove its unsupported nil-turn claim, preserve meaningful tool-name ordering, remove overwritten parity setup, and consolidate completion-date duplicates only with explicit assertion recipients.

h4. Isolate backend tests from checkout and unrelated listeners

Tests depend on live repository configuration/branch inventory. A fixed-port server smoke test can pass against an unrelated HTTP service, and readiness requests can exceed their stated timeout.

* Use self-contained configuration and temporary real Git repositories with explicit branch inventories; no dependency on this checkout policy or assumed absent branches.
* Use isolated listeners and verify the launched process and expected fixture identity. Detect early child exit; give readiness requests and cleanup bounded lifetimes.
* Exercise an occupied-port/foreign-service scenario and an accepting nonresponding listener. Prove parser and branch-precedence regressions remain detectable.

h4. Repair adapter test oracles and remove test-only database accessor

Required-column omission tests reject an already mixed schema, no-write tests inspect absent or unrelated directories, and an exported production DB accessor exists only for test introspection.

* Build each required-column omission from one valid schema family, assert the missing-structure cause, and keep mixed-family rejection separate. Disabling required-column checks in a disposable mutation must fail the repaired test.
* Observe the exact configured populated workspace for no-write checks, fail traversal errors, and compare bytes. Preserve trace/hydration assertions when consolidating into existing effective tests; a write mutation must be detected.
* Remove the exported test-only DB accessor while retaining read-only enforcement proof through a package-internal or real boundary test. Compile all consumers and avoid replacing it with another unnecessary production hook.

h4. Remove unused frontend contracts, dependencies and props

Unused api/types contracts, lucide-react and ignored component props enlarge the production surface without consumers.

* Revalidate reachability and remove unused AsyncState/ApiConfig/ApiError/HealthBadgeVariant, dependency and no-op props where the intended contract permits.
* If message identity is required for an export, implement explicit provenance in its owning feature rather than keeping a no-op prop; do not invent consumers to justify speculative abstractions.
* Preserve callers and meaningful assertions; validate lockfile consistency, typecheck and canonical embedded asset build. No test-count target or unmeasured performance claim.

h4. Decouple compatibility checks from ticket artifacts and fixed ports

Maintained compatibility targets depend on AL-58 historical evidence and fixed ports; a foreign listener or exhausted readiness loop can falsely satisfy setup.

* Move maintained provider/config fixtures to a stable owned test location without altering historical AL-58 evidence. Relocating archival ticket workspaces must not break the checks.
* Allocate/configure isolated ports, verify launched child and expected protocol identity, fail readiness exhaustion before launching the client, and clean up processes.
* Exercise missing provider, occupied port and stalled startup with local fixtures; run supported real OpenCode compatibility environments when available and label unavailable environments explicitly.

* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with an approved concrete contract. Complete every scoped correction and evidence case; this planning session changes no application, test or runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: strict typing, zero any/interface{}, zero emojis, no domain mocks and self-contained fixtures. Pass affected checks, independent focused review and the required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Preserve source identity, unavailable values, meaningful existing assertions and applicable local-only boundaries. Record tested revision, reproduction/probe evidence, reused checks and environment limits; update finding dispositions only after verified fixes. Publish the implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
