# AL-69: Preserve accurate telemetry and evidence

## Goal

Operators can trust tool-failure metrics, configured anomalies and source completeness without losing saved summary or trace evidence.

## Context

Execution status, configured anomaly policy, partial-data diagnostics, summary preservation, source fidelity and anomaly attribution must agree across extraction, persistence and presentation.

## Acceptance criteria

### Classify tool failures from execution evidence

Successful reads containing exit code or crash phrases are classified as failed tools, corrupting live and offline failure metrics.

* Use only authoritative structured status and structured process-exit evidence. No textual process fallback is permitted, including execution-tool wrappers and legacy records. Free-form output/error text is preserved content and never determines failure, severity or no-match behavior. Preserve a reported tool status without inventing an unavailable process result; missing structured evidence stays unknown. This boundary follows the user's explicit AL-69 clarification recorded in the active iteration plan.
* Exercise successful reads containing exit code 2 and segmentation fault, actual process failures, grep/glob no-match behavior, and live/offline parity through the real consumers.

### Apply anomaly thresholds consistently through persistence and views

Custom anomaly thresholds disappear during persistence and fleet recomputation; ErrorRateThreshold is serialized but ignored in favor of fixed levels.

* Require explicit configuration-file values for minimum tool-call count, warning rate and critical rate. No invented numeric defaults, hardcoded minimum or derived critical rate is permitted. Missing/invalid required values must fail affected analysis clearly. Apply the configured policy to event detection and aggregate fallback, and display truthful thresholds.
* Prove generate/save/reload/detail/fleet and live parity with distinct explicitly configured values, including exact boundaries, missing/invalid configuration and event pruning. Do not silently replace an already configured projection with default detection. The user's required-configuration clarification is recorded in AL-69 iteration 02.

### Surface partial scan and child-linkage diagnostics

CLI scan hides malformed siblings and rejected targets; live child lookup failures look like a successful absence of children.

* Present per-ticket rejection and fatal scan diagnostics while retaining valid siblings under the existing partial-scan policy. Distinguish rejected targets from nonexistent tickets.
* Keep supported live messages/parts visible when optional child lookup fails, with a typed completeness diagnostic reaching HTTP/CLI consumers.
* AL-67 F05 confirms the reachable case: remove session.parent_id while leaving parent/message/part reads usable. The trace still has useful content, but its caller currently swallows the child-query error. Report unavailable relationship lookup distinctly from a proven empty child set; do not trust a tool payload as native parent evidence. Repair the existing TestLiveTraceSurvivesChildSessionLookupFailure oracle to require that diagnostic while retaining its content-preservation assertions.
* Test corrupt/unsupported/unreadable scan siblings and failed, successful-empty and successful-linked child lookup. Do not turn optional enrichment failure into an empty or wholly failed trace.

### Preserve summary evidence when replacing generated metrics

Updating Workflow Efficiency consumes a subsequent peer-level Verification heading and deletes user evidence.

* Bound generated-block replacement with stable markers or correct heading-aware boundaries and a conservative legacy path. Preserve all unrelated prose and evidence.
* Test peer/parent headings, heading-like text inside fenced code, legacy headers, absent-block append and repeat-run idempotence through the actual writer and CLI boundary.

### Align live and offline source decoding fidelity

Live raw tool output trims meaningful whitespace that offline v1 decoding preserves; duplicated interpretation owns inconsistent field/status semantics.

* Share only the actual version-specific decoded representation needed by live/offline consumers; keep I/O/cache ownership in adapters and explicit projections in consumers.
* Preserve transient raw strings verbatim and verify parity for whitespace, structured values, malformed/unknown states and normal records.
* Preserve ledger privacy: no raw-payload expansion in serialization. Keep status behavior aligned with the execution-classification section above without inventing a universal source abstraction.

### Preserve anomaly identity and count semantics

Fallback matching ignores contradictory tool identities; generic error totals are exported as consecutive-error streaks.

* Require all supplied identity constraints to agree before weaker fallbacks; preserve authoritative evidence references and documented legacy formats.
* Test same-session/same-turn different tools, part-1 versus part-10, and read versus read_file in live and historical matching.
* Only consecutive/retry rules supply streak counts. Test separated errors under a rate anomaly versus a real streak through badges, filtering, inspector and export; preserve original evidence text.

## Implementation notes

Related work: AL-62 owns auxiliary-fact preservation and granular ownership; AL-63 owns capability parity; AL-67 owns broader architecture investigation. Reuse their contracts without duplicating their scope.

Finding references: AL-67 F05; ADP-06, ADP-10, ADP-12, CC-01, CC-03, CC-04, CC-08, FE-09, FE-12

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)

* [AL-67 child-lookup evidence](../../AL-67/ARCHITECTURE_ASSESSMENT.md), F05. This is shared discovery of the existing correction, not a second ticket.

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Retain the scoped corrections. Current domain/live.go still inspects crash phrases, and AL-67 F05 independently supports the child-lookup diagnostic gap. AL-62 lean ownership and AL-66 optional output are shipped baselines to preserve, not work to repeat.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
