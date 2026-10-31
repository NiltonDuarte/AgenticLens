# REV-3 Iteration 02: Steering plans across all tickets

## Scope and evidence limits

User steering expanded REV-3 from REV/FIX history to every ticket's steering plans. Baseline f3b7391; initial tree clean; PR #71 open. The [inventory](INVENTORY.md) accounts for 94 other ticket workspaces: 69 with 148 surviving iteration plans and 25 without plans, plus 16 specification-only tickets. REV-3 is excluded from its own evidence corpus.

Method: scan the complete 1,412,423-byte plan corpus structurally, inspect goals/steering/acceptance/boundary sections, and read selected contradictory decisions and associated context in depth. This is a documentation/process review, not a line-by-line audit of all implementation recipes. No raw conversation transcripts were re-extracted. A plan's description of user feedback is attributed to that plan; it does not independently prove approval, execution, or failure. Specifications and summaries provide context where iteration plans are absent; no missing steering is invented. Consolidated histories are not treated as lost work or policy violations.

Ordered plan-corpus SHA256 (path + NUL + content, sorted by workspace/iteration): 5aa57d7d3b66b94b71573783de671cb63aa1828fbc99fd69e20d8424c9ae3a2c.

## Findings and documentation improvements

### 1. Explicitly supersede decisions instead of accumulating incompatible plans

[AL-56/10](../../../AL-56/iterations/10/PLAN.md) requires durable source-readiness metadata; [AL-56/11](../../../AL-56/iterations/11/PLAN.md) explicitly corrects this to temporary metadata and adds partial workspace availability. [FIX-11/01](../../../FIX-11/iterations/01/PLAN.md) permits legacy source-root bridging; [FIX-11/05](../../../FIX-11/iterations/05/PLAN.md) removes that acceptance. [AL-37/02](../../../AL-37/iterations/02/PLAN.md) proposes agnostic harness claims, while [AL-37/03](../../../AL-37/iterations/03/PLAN.md) narrows support; later Codex work shows why neither historical statement is a permanent compatibility reference.

These are recorded decisions and corrections, not evidence that every earlier choice was negligent. The risk is replaying a detailed obsolete plan without reconciling the current contract. AGENTS.md and CONTRIBUTING.md now require a concise decision delta: source, baseline, acceptance change, superseded decision, preserved guarantees, verification impact, and evidenced status. New numbers and plan files do not establish approval.

### 2. Separate scope change from status, integration, and review repair

[AL-45/02](../../../AL-45/iterations/02/PLAN.md) and [03](../../../AL-45/iterations/03/PLAN.md) document base-branch merges. [AL-46/02](../../../AL-46/iterations/02/PLAN.md) repairs keyboard/visibility tests while preserving behavior. [GEN-11/02](../../../GEN-11/iterations/02/PLAN.md) records branch naming only. These histories illustrate why a new plan for every message is unnecessary under current policy. [GEN-10/04](../../../GEN-10/iterations/04/PLAN.md) consolidates accepted scope and prior work rather than preserving empty scaffolding.

Contributor guidance now keeps routine integration, review corrections, and status within the current iteration unless acceptance changes; explicit consolidation preserves decisions and verification. Directory counts are not source-backed SDLC iteration counts. This does not retroactively rewrite historical artifacts.

### 3. Check adjacent UI guarantees when fixing the reported symptom

[AL-25/03](../../../AL-25/iterations/03/PLAN.md) removes tooltip clipping; [AL-25/04](../../../AL-25/iterations/04/PLAN.md) explains that visible overflow exposed missing table containment. [AL-28/02](../../../AL-28/iterations/02/PLAN.md) records an empty-result state that hid the very filters required to recover. [AL-33/04](../../../AL-33/iterations/04/PLAN.md) distinguishes no filtering from selecting all known categories. [AL-45/01](../../../AL-45/iterations/01/PLAN.md), [AL-49/01](../../../AL-49/iterations/01/PLAN.md), and [FIX-12/02](../../../FIX-12/iterations/02/PLAN.md) add long-ID access, stable chronology, and collapse/inspection constraints.

The prevention checklist and technical guide now name these preservation cases together: empty-to-recovered transitions, controls independent of data state, tooltip plus table overflow, long-value copy access, keyboard behavior, stable logical identity, and existing collapse defaults. Tests of one symptom alone are insufficient evidence for adjacent behavior.

### 4. Preserve source facts; map presentation separately

[AL-24/02](../../../AL-24/iterations/02/PLAN.md) preserves raw personas before configured mapping. [AL-31/02](../../../AL-31/iterations/02/PLAN.md), [03](../../../AL-31/iterations/03/PLAN.md), and [04](../../../AL-31/iterations/04/PLAN.md) progressively remove role-name heuristics from structural root/child classification. [AL-39/03](../../../AL-39/iterations/03/PLAN.md) moves iteration counting away from workspace directories; [AL-56/04](../../../AL-56/iterations/04/PLAN.md) preserves exact zero rather than inventing a minimum. [AL-50/01](../../../AL-50/iterations/01/PLAN.md) identifies sanitizer transformations applied to technical source bodies.

The technical guide now explicitly separates structural hierarchy, persona facts, display mappings, and source-based steering counts. Diagnostic payloads stay separate from generated narrative, with only approved redaction applied to source evidence. This refines iteration 01's generic sanitization advice; it does not authorize rewriting raw payloads or imply unavailable cost is already nullable in the product.

### 5. Distinguish initialization defaults, runtime policy, and verified support

[AL-36/02](../../../AL-36/iterations/02/PLAN.md) and [03](../../../AL-36/iterations/03/PLAN.md) remove platform assumptions from artifact/prefix configuration. [AL-40/02](../../../AL-40/iterations/02/PLAN.md) requires API values to reflect actual configuration. [AL-56/02](../../../AL-56/iterations/02/PLAN.md) separates useful generated template defaults from silent runtime guessing. [AL-58/01](../../../AL-58/iterations/01/PLAN.md) names fixture and runtime compatibility evidence, while [AL-58/02](../../../AL-58/iterations/02/PLAN.md) separates shared code from version-family responsibilities. [AL-59/02](../../../AL-59/iterations/02/PLAN.md) repairs tests that failed on a runner's absent source database before reaching their intended assertions.

The checklist now requires evidence-backed support statements, isolated source/configuration fixtures, and a distinction between defaults explicitly written by init and guesses made at runtime. Historical plans are not command help for a currently installed release.

### 6. Availability needs an explicit failure granularity

[AL-51/03](../../../AL-51/iterations/03/PLAN.md) deliberately preserves usable traces on optional child lookup failure. This contextualizes iteration 01's REV-2 finding: making that optional read fatal would reverse the accepted resilience decision. A future improvement should preserve trace availability while making missing correlation observable; this audit does not implement a new response shape.

[AL-56/11](../../../AL-56/iterations/11/PLAN.md) similarly preserves valid sibling workspaces while diagnosing individual rejected ledgers. The technical guide and checklist now distinguish project-fatal errors from per-ledger rejection and optional-enrichment failure. Review recommendations must preserve accepted availability while improving diagnostic truthfulness.

### 7. Recovery and handoff need preserved evidence, not repeated ceremony

[AL-33/02](../../../AL-33/iterations/02/PLAN.md) records verification after cleanup/disruption; [AL-19/02](../../../AL-19/iterations/02/PLAN.md) calls for recovery and contract reconciliation. [GEN-9/01](../../../GEN-9/iterations/01/PLAN.md) keeps cross-machine source-backed migration incomplete when data is unavailable. [FIX-14/02](../../../FIX-14/iterations/02/PLAN.md) preserves accepted implementation/verification during ticket reassignment; [FIX-8/04](../../../FIX-8/iterations/04/PLAN.md) records a narrowly authorized specification deletion.

The checklist adds generated-path cleanup boundaries and integrity checks after disruption. The contributor delta format carries verification and delivery status forward. These examples do not authorize new ticket switches, bulk rewrites, deleting history, or source-backed backfill on machines without the data.

## Coverage conclusions and remaining work

The inventory covers every existing ticket workspace and plan-only ticket ID. Single-plan tickets mainly contribute initial acceptance context; no new defect is alleged without specific evidence. PO specifications describe planned capability rather than executed steering. Twenty-five workspaces have no surviving iteration plan; no claim is made about their unrecorded conversations.

Documentation changes are limited to AGENTS.md, CONTRIBUTING.md, TECHNICAL_GUIDELINES.md, the existing prevention guide, and REV-3 artifacts. No historical plan, skill, code, test, runtime script, configuration, or other ledger is changed. Open runtime findings from iteration 01 remain open; the optional-lookup recommendation is refined by its steering context. Automation to enforce these policies remains separate implementation work.

## Verification

Self-review accepted for documentation scope; no independent reviewer was launched. No runtime remediation is claimed.

- Local Markdown targets and zero-emoji checks passed across all 12 REV-3/current guidance documents; the inventory accounts for 110 other ticket IDs. git diff --check passed.
- Tested documentation implementation: d794b85. Fetched and merged origin/main; already up to date. Subsequent changes record evidence only.
- Full make passed: Go vet, TypeScript typecheck/lint, 12 backend race-test packages (valid cached Go results reused), all 793 frontend tests across 36 files, frontend production build, and four platform builds. Log: /tmp/rev3-iteration02-make.log (local, outside Git).
- Generated web/dist and backend/internal/adapters/web/dist match byte-for-byte. Historical source/tests/skills and other-ticket ledgers remain unchanged.
- Existing dependency audit findings and React act warnings remain environment observations; no dependencies changed. No browser run, source-backed migration, or new performance profile was needed for the documentation-only delta.
- Final telemetry targets the same verified Codex root as iteration 01, with zero child sessions expected. Existing session IDs/source must survive the refresh; extraction outcome is recorded in SUMMARY.md. Finalization remains a snapshot before subsequent delivery calls.
