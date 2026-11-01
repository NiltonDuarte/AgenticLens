h3. Title: AL-69: Correct telemetry accuracy and preserve evidence

*TARGET*: AgenticLens (Domain / Source Adapters / CLI / Persistence / Anomaly Presentation)

*Status*: Ready
*Work type*: fix
*Priority*: P2
*Blocked by*: None
*Blocks*: None
*Origin*: REV-5 iteration 03 audit; consolidated draft batch at the user's request in iteration 09.
*Findings*: ADP-06, ADP-10, ADP-12, CC-01, CC-03, CC-04, CC-08, FE-09, FE-12
*Related*: AL-62 owns auxiliary-fact preservation and granular ownership; AL-63 owns capability parity; AL-67 owns broader architecture investigation. Reuse their contracts without duplicating their scope.

*CONTEXT*

Execution status, configured anomaly policy, partial-data diagnostics, summary preservation, source fidelity and anomaly attribution must agree across extraction, persistence and presentation.

The source reports retain affected paths, triggers, evidence and limits at their frozen baseline. Revalidate before implementation; this specification does not claim remediation. Overlapping findings and test/architecture analyses are counted once in the current finding map.

*GOAL*

Correct the scoped behavior through its production consumers and verification boundaries, satisfying every acceptance section below.

*VERTICAL SLICE SCOPE*

* Own the changes through Domain / Source Adapters / CLI / Persistence / Anomaly Presentation, including real consumer behavior and regression proof. Keep pure calculations in the core and I/O in adapters; preserve supported stored facts when enrichment is unavailable.
* Use bounded implementation and review iterations within this ticket. Grouping reduces tracking overhead; it does not authorize dropping acceptance cases, skipping focused reviews or calling a partial fix complete. No broad framework rewrite or new speculative abstraction is required.

*ACCEPTANCE CRITERIA*

h4. Classify tool failures from execution evidence

Successful reads containing exit code or crash phrases are classified as failed tools, corrupting live and offline failure metrics.

* Use authoritative status and process-exit evidence. Limit any legacy text fallback to execution tools and a recognized execution envelope; arbitrary read/search output is content.
* Exercise successful reads containing exit code 2 and segmentation fault, actual process failures, grep/glob no-match behavior, and live/offline parity through the real consumers.

h4. Apply anomaly thresholds consistently through persistence and views

Custom anomaly thresholds disappear during persistence and fleet recomputation; ErrorRateThreshold is serialized but ignored in favor of fixed levels.

* Define the configured error-rate threshold and warning/critical semantics in the contract; apply the effective policy to event detection and aggregate fallback, and display truthful thresholds.
* Prove generate/save/reload/detail/fleet and live parity for settings above and below defaults, including exact boundaries and event pruning. Do not silently replace an already configured projection with default detection.

h4. Surface partial scan and child-linkage diagnostics

CLI scan hides malformed siblings and rejected targets; live child lookup failures look like a successful absence of children.

* Present per-ticket rejection and fatal scan diagnostics while retaining valid siblings under the existing partial-scan policy. Distinguish rejected targets from nonexistent tickets.
* Keep supported live messages/parts visible when optional child lookup fails, with a typed completeness diagnostic reaching HTTP/CLI consumers.
* Test corrupt/unsupported/unreadable scan siblings and failed, successful-empty and successful-linked child lookup. Do not turn optional enrichment failure into an empty or wholly failed trace.

h4. Preserve summary evidence when replacing generated metrics

Updating Workflow Efficiency consumes a subsequent peer-level Verification heading and deletes user evidence.

* Bound generated-block replacement with stable markers or correct heading-aware boundaries and a conservative legacy path. Preserve all unrelated prose and evidence.
* Test peer/parent headings, heading-like text inside fenced code, legacy headers, absent-block append and repeat-run idempotence through the actual writer and CLI boundary.

h4. Align live and offline source decoding fidelity

Live raw tool output trims meaningful whitespace that offline v1 decoding preserves; duplicated interpretation owns inconsistent field/status semantics.

* Share only the actual version-specific decoded representation needed by live/offline consumers; keep I/O/cache ownership in adapters and explicit projections in consumers.
* Preserve transient raw strings verbatim and verify parity for whitespace, structured values, malformed/unknown states and normal records.
* Preserve ledger privacy: no raw-payload expansion in serialization. Keep status behavior aligned with the execution-classification section above without inventing a universal source abstraction.

h4. Preserve anomaly identity and count semantics

Fallback matching ignores contradictory tool identities; generic error totals are exported as consecutive-error streaks.

* Require all supplied identity constraints to agree before weaker fallbacks; preserve authoritative evidence references and documented legacy formats.
* Test same-session/same-turn different tools, part-1 versus part-10, and read versus read_file in live and historical matching.
* Only consecutive/retry rules supply streak counts. Test separated errors under a rate anomaly versus a real streak through badges, filtering, inspector and export; preserve original evidence text.

* Add or repair the specific failing behavior oracle; reuse meaningful existing tests and map preserved assertions before any consolidation. Fixtures must be self-contained, never dependent on live ticket workspaces.

*DEFINITION OF DONE*

* Implement in a separate agentic-lens-dev session with an approved concrete contract. Complete every scoped correction and evidence case; this planning session changes no application, test or runtime code.
* Follow TECHNICAL_GUIDELINES.md and docs/REGRESSION_PREVENTION.md: strict typing, zero any/interface{}, zero emojis, no domain mocks and self-contained fixtures. Pass affected checks, independent focused review and the required make gate. Frontend changes require canonical build and standalone embedded-asset verification.
* Preserve source identity, unavailable values, meaningful existing assertions and applicable local-only boundaries. Record tested revision, reproduction/probe evidence, reused checks and environment limits; update finding dispositions only after verified fixes. Publish the implementation PR and source-grounded telemetry.

*SOURCE EVIDENCE*

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)
