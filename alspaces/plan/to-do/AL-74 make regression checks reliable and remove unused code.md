# AL-74: Make regression checks reliable and remove unused code

## Goal

Maintainers can trust isolated regression checks and immutable transformations while removing confirmed unused code without losing meaningful assertions.

## Context

Pure transformations must not mutate inputs, tests must exercise their claimed boundary in isolated environments, and unused production surfaces and duplicated setup should be removed while preserving meaningful assertions.

## Acceptance criteria

### Remove discarded core work and preserve migration input values

V1 migration mutates its input through a shared session slice. Core aggregation builds unread mapped-event structures and duplicates model grouping; some tests construct and discard their purported sorting fixtures.

* Clone mutable migration containers before transformation; prove input nonmutation and alias independence for nil/non-nil events and appendable metadata.
* Remove unused mappedEvents and anomalyEvidenceSelection.events work and reuse existing model aggregation where behavior is equivalent. Preserve role order, thread ceilings, precise and incomplete-cost behavior; claim no unmeasured speedup.
* Repair the actual sorting oracle or remove its unsupported nil-turn claim, preserve meaningful tool-name ordering, remove overwritten parity setup, and consolidate completion-date duplicates only with explicit assertion recipients.

### Isolate backend tests from checkout and unrelated listeners

Tests depend on live repository configuration/branch inventory. A fixed-port server smoke test can pass against an unrelated HTTP service, and readiness requests can exceed their stated timeout.

* Use self-contained configuration and temporary real Git repositories with explicit branch inventories; no dependency on this checkout policy or assumed absent branches.
* Use isolated listeners and verify the launched process and expected fixture identity. Detect early child exit; give readiness requests and cleanup bounded lifetimes.
* Exercise an occupied-port/foreign-service scenario and an accepting nonresponding listener. Prove parser and branch-precedence regressions remain detectable.

### Repair adapter test oracles and reassess database accessor

Required-column omission tests reject an already mixed schema and some no-write tests inspect absent or unrelated directories. A database accessor previously described as test-only also serves production hydration at the AL-67 baseline; reassess its real consumers before changing it.

* Build each required-column omission from one valid schema family, assert the missing-structure cause, and keep mixed-family rejection separate. Disabling required-column checks in a disposable mutation must fail the repaired test.
* Observe the exact configured populated workspace for no-write checks, fail traversal errors, and compare bytes. Preserve trace/hydration assertions when consolidating into existing effective tests; a write mutation must be detected.
* Revalidate the DB accessor's current consumers before removal: AL-67 found production historical hydration calls SQLiteHook.DB(), so it is no longer accurately described as test-only. Preserve that runtime path; remove or narrow the accessor only when its actual responsibility has a safe replacement in this approved scope. Coordinate broader source-boundary redesign with REV-7 rather than silently expanding this cleanup. Retain read-only enforcement proof and compile all consumers; do not replace it with another unnecessary production hook.
* AL-67 F10 independently confirms why the omission oracle is weak: each case creates both complete table inventories before removing one column and expects mixed-family rejection. Even disabled column validation could leave that earlier guard green. Keep mixed-family cases, but prove each isolated family's required-column check is the reason its omission fails.

### Remove unused frontend contracts, dependencies and props

Unused api/types contracts, lucide-react and ignored component props enlarge the production surface without consumers.

* Revalidate reachability and remove unused AsyncState/ApiConfig/ApiError/HealthBadgeVariant, dependency and no-op props where the intended contract permits.
* If message identity is required for an export, implement explicit provenance in its owning feature rather than keeping a no-op prop; do not invent consumers to justify speculative abstractions.
* Preserve callers and meaningful assertions; validate lockfile consistency, typecheck and canonical embedded asset build. No test-count target or unmeasured performance claim.

### Decouple compatibility checks from ticket artifacts and fixed ports

Maintained compatibility targets depend on AL-58 historical evidence and fixed ports; a foreign listener or exhausted readiness loop can falsely satisfy setup.

* Move maintained provider/config fixtures to a stable owned test location without altering historical AL-58 evidence. Relocating archival ticket workspaces must not break the checks.
* Allocate/configure isolated ports, verify launched child and expected protocol identity, fail readiness exhaustion before launching the client, and clean up processes.
* Exercise missing provider, occupied port and stalled startup with local fixtures; run supported real OpenCode compatibility environments when available and label unavailable environments explicitly.

## Implementation notes

Related work: Coordinate pure aggregation changes with AL-69, UI cleanup with AL-72, and dependency/lockfile edits with AL-75. These are shared-file coordination needs, not speculative blocking dependencies.

Finding references: AL-67 F10; ADP-07, ADP-08, ADP-09, ADP-13, CC-09, CC-10, CC-11, CC-12, D4, FE-D1

* [REV-5 ADAPTERS_REPORT.md](../../REV-5/iterations/03/ADAPTERS_REPORT.md)
* [REV-5 CORE_CLI_REPORT.md](../../REV-5/iterations/03/CORE_CLI_REPORT.md)
* [REV-5 DELIVERY_REPORT.md](../../REV-5/iterations/03/DELIVERY_REPORT.md)
* [REV-5 FRONTEND_REPORT.md](../../REV-5/iterations/03/FRONTEND_REPORT.md)

* [AL-67 test and consumer evidence](../../AL-67/ARCHITECTURE_ASSESSMENT.md), F10 and F07 (production accessor use). Original REV-5 evidence remains historical.

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Retain remaining migration, oracle, isolation and unused-code corrections. REV-6 (#113) already split test lanes and gates integration on unit success; do not repeat that orchestration. Current domain/events.go retains anomalyEvidenceSelection.events; AL-67 F10 reconfirms the required-column oracle issue.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
