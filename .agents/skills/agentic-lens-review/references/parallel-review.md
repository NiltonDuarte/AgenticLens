# Specialized parallel review

Use this protocol for substantive application changes and broad audits. The parent orchestrator owns dispatch, evidence reconciliation, fixes and delivery; each child uses its dedicated focus persona plus shared reviewer rules. This is an agent-operated skill protocol, not an installed CI service or automatic scheduler. It supplements the canonical review procedure in SKILL.md and does not authorize implementation, publication or merges.

## Select coverage before dispatch

For substantive behavior changes, assign the four core focuses below to separate independent reviewers by default. Examples include new features, cross-module refactors, persistence/migration changes and release/runtime changes. Add a specialist when the changed boundary needs expertise beyond the core assignments. For documentation-only or mechanical low-risk edits, use a narrower review and record why omitted focuses cannot materially affect the change. Do not silently reduce coverage for convenience or cost. If a focus is truly inapplicable, record the reason in the plan rather than manufacture a review.

| Focus | Persona / Codex agent name |
| --- | --- |
| Behavior and integration | [reviewer-behavior](../../../agent/reviewer-behavior.md) |
| Data integrity and failure handling | [reviewer-data](../../../agent/reviewer-data.md) |
| Architecture and maintainability | [reviewer-architecture](../../../agent/reviewer-architecture.md) |
| Test effectiveness | [reviewer-tests](../../../agent/reviewer-tests.md) |

Each persona owns its investigation details. Shared mission/evidence/permission context lives in [review-context.md](../../../agent/references/review-context.md); the canonical investigation procedure stays in SKILL.md. Each persona loads the shared file directly, without loading another persona. The coordinator reads this matrix; each child receives only its own persona and common rules, not the other role descriptions. This limits irrelevant instruction context, not filesystem access or correlated model reasoning. Codex registrations live under .codex/agents/. Other harnesses load the matching shared Markdown persona through their supported registration mechanism; these files do not install an OpenCode configuration automatically. If a focused role is not registered in the running harness, use an available reviewer with an explicit pointer to exactly that persona and record the fallback; do not silently substitute a different focus. Generic reviewer remains the entry point for a selected specialist or narrow low-risk review.

Specialists are selected by concrete risk: frontend async/accessibility for interactive behavior; build/release integrity for artifact provenance and workflow side effects; security for an actual trust boundary; performance for a stated workload/budget. Record the specialist's question and affected boundary, not just a title. Any reviewer should report a concrete serious defect outside its focus rather than discard it; the parent assigns its verification owner.

## Model policy and attribution

Before dispatch, load [model-policy.yaml](model-policy.yaml) and follow [model execution and attribution](model-attribution.md). Every selected focus expands into the configured model profiles, currently Sol 6.1/high AND Luna 6/medium. The coordinator owns run/finding records and reports discovery credit by reviewer/model. Role selection does not remove model coverage.

## Freeze evidence and dispatch independently

- Put the coverage roster in the current iteration PLAN.md: focus, risk question, scope/call paths, evidence owner and unique temporary report path. Use private scratch outputs, then the parent archives full reports by run ID under the [compact committed archive](model-attribution.md#compact-committed-archive); separate owners must not become separate committed files. The parent alone owns the consolidated REVIEW.md. Children write only their assigned reports if the harness permits it; otherwise return content for the parent to save.
- Give every reviewer the same base/head commit and requirements. For uncommitted code, archive the exact diff or source manifest once, including relevant untracked files, and reference it by a readable candidate ID. Include active product/deployment constraints and useful raw test evidence; do not seed an approval or suspected answer. Fresh context is preferred where available. Do not distribute other reviewers' conclusions before their first pass.
- Freeze product/test edits during this pass. Shared-checkout reviewers may overlap in reading, but cannot change production/test code. Probes must use disposable paths and separate resource names/ports. A reviewer that observes source changes reports stale evidence instead of approving a mixed revision.
- Dispatch independent focuses concurrently up to actual harness capacity. Reserve capacity for the parent as needed. If all persona/model runs cannot run at once, queue remaining focuses in waves against the same frozen revision; do not drop coverage or pretend sequential work was parallel. Use supported native/delegated tools and report unavailable capabilities. Do not create top-level threads or change models/providers merely to obtain parallelism.
- Each reviewer follows SKILL.md: independently form failure hypotheses, trace relevant unchanged dependencies, seek refuting evidence and report findings with bounded confidence. Coordinate expensive identical tests through the parent rather than running four whole suites. Shared valid test results are reusable; independent reasoning is still required.

### Dispatch brief

Fill in this brief rather than copying four general review requests:

    Ticket/iteration: <active ticket and iteration>
    Role: <registered focus reviewer or explicit persona pointer>; question: <concrete risk>
    Run: <run ID, phase, profile and acknowledged model/effort, report path>
    Evidence: <base/head or readable candidate ID with its exact archived diff/manifest, raw diff, requirements,
               contract/plan, canonical review skill, relevant source paths,
               test results and supported deployment assumptions>
    Inspect: <entry points, consumers and important unchanged dependencies>
    Own: <unique private scratch report path or full response; no product/test/archive edits>
    Shared checkout: preserve others' changes; isolated probes only.
    First pass: investigate independently; do not read sibling conclusions.
    Return: reviewed revision, hypotheses/disposition, findings, verification,
            gaps and APPROVED / CHANGES_REQUESTED / INCOMPLETE.
    Parent owns fixes, consolidated verdict, telemetry and publication.

The focus is a responsibility, not a source-read boundary. Give enough context to follow a value across layers. Never give a reviewer only a file list and ask it to certify end-to-end behavior.

## Verify and consolidate findings

Keep one finding record per distinct root cause/observable failure, retaining all contributing reviewers' evidence and disagreement. Do not merge unrelated failures merely because they affect the same file. Use stable IDs such as BEH-01 or DATA-01, then reference those IDs from the consolidated report.

Each candidate records location, introduced/pre-existing/suggestion classification, supported expectation, reachable trigger, observed or inferred consequence, evidence/command, severity and verification owner. The parent assigns a disposition: confirmed, refuted with evidence, or unresolved. A disagreement or majority approval does not refute a demonstrated failure. One confirmed blocker is sufficient to prevent delivery. Deduplication must preserve distinct regression assertions.

Validate candidates against actual callers and guards. For disputed or consequential findings, the parent or a separately assigned verifier checks the counterexample using raw source and relevant requirements; a fresh verifier may see the candidate because its task is validation, not blind discovery. Do not launch a mandatory extra reviewer when the existing reproduction/source proof settles the issue. Unresolved material hypotheses are evidence gaps, not silently discarded findings. Existing out-of-scope debt remains visible with an owner/follow-up; it blocks this delivery only if it defeats the accepted behavior or a required invariant.

## Corrections and revision-aware re-review

After the first pass, route confirmed fixes to the coder and missing regression coverage to the tester. Freeze a new candidate revision after corrections. Preserve earlier reports as evidence of their own revisions; append/revise with an explicit revision marker rather than erasing prior findings.

Map each fix to affected focus owners and dependencies. Re-review the corrected behavior, regression tests and affected interactions; re-run affected checks. Retain approvals for unchanged scopes only with an explicit dependency/diff rationale in the consolidated report. For a new cross-cutting change or uncertain impact, re-run all affected focuses; a new head alone is neither proof of validity nor a reason to repeat every unchanged test. If the source keeps changing, stop and obtain a stable candidate rather than loop on stale approvals.

## Consolidated delivery gate

When focused investigations have no unresolved blockers or material review gaps, the parent may commit the candidate, integrate main and run the required verification phase. The final consolidated approval is issued only after those checks pass. Integration changes invalidate affected review evidence and require the same impact assessment/re-review; this ordering avoids requiring final approval before its verification can run.

Before committing, the parent losslessly archives completed report bodies and candidate/supporting evidence in REVIEW_RUNS.json, rewrites references, compares original content bytes and verifies provenance before deleting scratch files.

The parent records a compact roster (focus, model/effort, run ID/phase, reviewed revision, report, outcome, any valid carry-forward rationale), deduplicated findings/dispositions, remaining debt, verification and limits in iterations/<NN>/REVIEW.md. A child approval never independently opens the gate.

- CHANGES_REQUESTED when a substantiated blocker remains. Also list any incomplete required reviews so fixing the blocker cannot hide the gaps.
- INCOMPLETE when no confirmed blocker remains but a required focus, material hypothesis, revision match or necessary verification is unresolved. A cancelled/timed-out/unavailable reviewer is incomplete, not an approval. Recover the missing coverage or report the limitation; do not waive it silently.
- APPROVED only when all applicable focus/model cells have current or justified carried-forward evidence, no blocking finding/material gap remains, and required integrated checks pass. Local approval is scoped and does not authorize a remote merge. Hosting/CI or runtime stages not exercised must be stated explicitly.

Stop when this gate is satisfied for the requested scope. Do not add endless speculative passes, demand unanimous style preferences, or claim bug-free software. Measure usefulness through confirmed defects, escaped regressions and false positives with honest sampling limits, not number of agents or approval rate.
