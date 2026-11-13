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

- Put the coverage roster in the current change design/tasks for the active iteration: focus, risk question, scope/call paths, evidence owner and unique report path. The parent alone owns the consolidated REVIEW.md. Children write only their assigned reports if the harness permits it; otherwise return content for the parent to save.
- Give every reviewer the same base/head commit and requirements. For uncommitted code, record a diff/content fingerprint covering changed and relevant untracked files. Include active product/deployment constraints and useful raw test evidence; do not seed an approval or suspected answer. Fresh context is preferred where available. Do not distribute other reviewers' conclusions before their first pass.
- Freeze product/test edits during this pass. Shared-checkout reviewers may overlap in reading, but cannot change production/test code. Probes must use disposable paths and separate resource names/ports. A reviewer that observes source changes reports stale evidence instead of approving a mixed revision.
- Dispatch independent focuses concurrently up to actual harness capacity. Reserve capacity for the parent as needed. If all persona/model runs cannot run at once, queue remaining focuses in waves against the same frozen revision; do not drop coverage or pretend sequential work was parallel. Use supported native/delegated tools and report unavailable capabilities. Do not create top-level threads or change models/providers merely to obtain parallelism.
- Each reviewer follows SKILL.md: independently form failure hypotheses, trace relevant unchanged dependencies, seek refuting evidence and report findings with bounded confidence. Coordinate expensive identical tests through the parent rather than running four whole suites. Shared valid test results are reusable; independent reasoning is still required.

### Dispatch brief

Fill in this brief rather than copying four general review requests:

    Ticket/iteration: <active ticket and iteration>
    Role: <registered focus reviewer or explicit persona pointer>; question: <concrete risk>
    Run: <run ID, phase, profile and acknowledged model/effort, report path>
    Evidence: <base/head or immutable fingerprint, raw diff, requirements,
               contract/plan, canonical review skill, relevant source paths,
               test results and supported deployment assumptions>
    Inspect: <entry points, consumers and important unchanged dependencies>
    Own: <unique report path only; no product/test edits>
    Shared checkout: preserve others' changes; isolated probes only.
    First pass: investigate independently; do not read sibling conclusions.
    Return: reviewed revision, hypotheses/disposition, findings, verification,
            gaps and APPROVED / CHANGES_REQUESTED / INCOMPLETE.
    Parent owns fixes, consolidated verdict, telemetry and publication.

The focus is a responsibility, not a source-read boundary. Give enough context to follow a value across layers. Never give a reviewer only a file list and ask it to certify end-to-end behavior.

## Verify and consolidate findings

Keep one finding record per distinct root cause/observable failure, retaining all contributing reviewers' evidence and disagreement. Do not merge unrelated failures merely because they affect the same file. Use stable IDs such as BEH-01 or DATA-01, then reference those IDs from the consolidated report.

Each candidate records location, introduced/pre-existing/suggestion classification, supported expectation, reachable trigger, observed or inferred consequence, evidence/command, numeric severity score from 0 (low) to 10 (high), and verification owner. The parent assigns a disposition: confirmed, refuted with evidence, or unresolved. A disagreement or majority approval does not refute a demonstrated failure. One confirmed blocker is sufficient to prevent delivery. Deduplication must preserve distinct regression assertions.

Validate candidates against actual callers and guards. For disputed or consequential findings, the parent or a separately assigned verifier checks the counterexample using raw source and relevant requirements; a fresh verifier may see the candidate because its task is validation, not blind discovery. Do not launch a mandatory extra reviewer when the existing reproduction/source proof settles the issue. Unresolved material hypotheses are evidence gaps, not silently discarded findings. Existing out-of-scope debt remains visible with an owner/follow-up; it blocks this delivery only if it defeats the accepted behavior or a required invariant.

## Corrections and revision-aware re-review

After the first pass, route confirmed fixes to the coder and missing regression coverage to the tester. Freeze a new candidate revision after corrections. While review is open, preserve earlier reports as evidence of their own revisions; append/revise with an explicit revision marker rather than erasing prior findings. Completed-review cleanup follows the retention rule below.

Map each fix to affected focus owners and dependencies. Re-review the corrected behavior, regression tests and affected interactions; re-run affected checks. Retain approvals for unchanged scopes only with an explicit dependency/diff rationale in the consolidated report. For a new cross-cutting change or uncertain impact, re-run all affected focuses; a new head alone is neither proof of validity nor a reason to repeat every unchanged test. If the source keeps changing, stop and obtain a stable candidate rather than loop on stale approvals.

## Consolidated delivery gate

When focused investigations have no unresolved blockers or material review gaps, the parent may commit the candidate, integrate main and run the required verification phase. The final consolidated approval is issued only after those checks pass. Integration changes invalidate affected review evidence and require the same impact assessment/re-review; this ordering avoids requiring final approval before its verification can run.

The parent records a compact roster (focus, model/effort, run ID/phase, reviewed revision, report, outcome, any valid carry-forward rationale), deduplicated findings/dispositions, remaining debt, verification and limits in evidence/iterations/<NN>/REVIEW.md. A child approval never independently opens the gate.

- CHANGES_REQUESTED when a substantiated blocker remains. Also list any incomplete required reviews so fixing the blocker cannot hide the gaps.
- INCOMPLETE when no confirmed blocker remains but a required focus, material hypothesis, revision match or necessary verification is unresolved. A cancelled/timed-out/unavailable reviewer is incomplete, not an approval. Recover the missing coverage or report the limitation; do not waive it silently.
- APPROVED only when all applicable focus/model cells have current or justified carried-forward evidence, no blocking finding/material gap remains, and required integrated checks pass. Local approval is scoped and does not authorize a remote merge. Hosting/CI or runtime stages not exercised must be stated explicitly.

Stop when this gate is satisfied for the requested scope. Do not add endless speculative passes, demand unanimous style preferences, or claim bug-free software. Measure usefulness through confirmed defects, escaped regressions and false positives with honest sampling limits, not number of agents or approval rate.

## Completed review retention

Apply the [shared writing rules](../../../../AGENTS.md#artifact-writing-rules) to the human-facing summary. This protocol owns review artifact content and retention; group repeated outcomes rather than duplicating run narratives.

Review reports and structured run/finding files support the active investigation, correction loop and approval gate. Once findings are addressed and consolidated approval is issued, the parent compacts them into `evidence/iterations/<NN>/REVIEW_SUMMARY.json` before deleting any source reports. Keep this small structured file permanently and link it from the ticket's `delivery.md`. Then delete individual review reports, consolidated `REVIEW.md`, `REVIEW_RUNS.json`, and `FINDINGS.json` for that completed review; remove stale links. Keep plans, regression tests, telemetry and other required deliverables. Historical audits in unrelated tickets remain outside this cleanup scope.

### Compact findings format

Use `schema_version: 1` and a `findings` array. Omit `reviewed_revision` from this compact file; revision evidence remains in the ticket's `delivery.md` and active-review records. Each entry contains:

- `finding_id`: the stable canonical finding ID.
- `severity`: a JSON number from 0 (low) to 10 (high), inclusive, reconciled against the supported impact. Do not use severity labels such as `"low"` or `"high"`.
- `issue`: one very brief, self-contained sentence describing the defect.
- `reporters`: every distinct reporting persona/model/effort tuple, including shared discovery across models and later reports of the same issue. Use the actual persona name and exact model identifier, not internal run IDs or agent namespace paths. Deduplicate repeated reports from the same tuple; never keep only the first reporter.
- `disposition`: a brief outcome/fix, including refutation or accepted debt and its owner/follow-up where applicable. Preserve every canonical candidate, including refuted claims; do not make them look like confirmed defects.

For example (illustrative values, not session evidence):

```json
{
  "schema_version": 1,
  "findings": [
    {
      "finding_id": "DATA-01",
      "severity": 8,
      "issue": "Failed hydration hides persisted ledger rows.",
      "reporters": [
        {"persona": "reviewer-data", "model": "gpt-6.1-sol", "effort": "high", "identity_basis": "acknowledged"},
        {"persona": "reviewer-behavior", "model": "gpt-6-luna", "effort": "medium", "identity_basis": "observed"}
      ],
      "disposition": "Fixed: retain persisted rows when enrichment fails."
    }
  ]
}
```

Resolve reporters from the canonical findings' observations and run records before removing those records. Prefer source-observed model/effort, otherwise dispatch-acknowledged settings; label `identity_basis` as `observed`, `acknowledged`, or `requested`. When execution identity is unavailable, use `null` for unknown model/effort and a short `identity_note`; requested settings may be recorded only as requested, never as proof of execution. A verifier confirming a supplied issue belongs in the verification summary, not in `reporters` as a discoverer. If a verifier reports a new distinct issue, retain its persona/model/effort with that finding and state its verification-phase origin in the disposition. With no candidates, write `"findings": []`; do not invent a placeholder issue.

Before cleanup, check that every canonical finding and all its reporting tuples survived compaction, including shared credit and refuted candidates; parse the JSON and confirm `delivery.md` links to it. Keep issue/disposition text short; do not copy raw evidence, transcripts or run rosters into the compact file.

The ticket's `delivery.md` retains reviewed scope/revision and final approval, a link to the compact file, verification actors separately from reporters, initial/final review outcomes including no-findings/incomplete runs where relevant, actual verification and remaining limits. Use readable reviewer responsibilities with exact model/effort for any findings or fixes mentioned there. Do not delete unresolved findings, accepted-debt ownership or material evidence gaps; cleanup cannot convert an incomplete review into approval. Do not retain verbose resolved-finding reports as duplicate evidence.
