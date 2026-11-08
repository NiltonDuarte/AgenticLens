---
name: agentic-lens-dev
description: Use when building features, fixing bugs, or refactoring the standalone AgenticLens application in Go (backend/CLI/engine) and TypeScript (frontend dashboard) with multi-agent orchestration (planner, coder, tester, reviewer). Requires a Jira ticket or task identifier.
---

# AgenticLens development

Coordinate application changes through planner, coder, tester and independent
reviewers. Follow [repository policy](../../../AGENTS.md) for ticket isolation,
role boundaries, approvals, telemetry and delivery, and
[technical guidelines](../../../TECHNICAL_GUIDELINES.md) for engineering rules.
Read `.agenticlens-config.yaml` for ticket prefixes and `workspaces.dir`
(default `workspaces`); paths below are relative to the active ticket workspace.

## Setup & Workspace Structure

Reuse the supplied ticket and branch/workspace when resuming. Record initial Git
status, preserve unrelated changes and read any matching specification. Update an
existing index entry to In Progress; ad hoc work needs no specification/index.
Create a feature branch if needed; never develop on main.

| Artifact | Content | Usual prose length |
| --- | --- | --- |
| `CONTRACT.md` | Frozen problem, agreed outcome, exclusions and observable acceptance; link canonical technical rules. | 200–400 words |
| `iterations/<NN>/PLAN.md` | Steps, affected components/owners, verification and review coverage; later plans record acceptance deltas. | 200–400 words |
| `SUMMARY.md` | Current behavior, decisions, actual checks/revision, review outcome and limits. | 300–600 words before metrics |
| `PULL_REQUEST.md` | Conventional title first; problem, resulting behavior, validation and limits; link contract/summary. | Usually 100–250 words; at most 3000 characters |
| `agentic_metrics.json` | Final CLI-generated session ledger; rendered metrics follow the authored summary. | Structured output |

Use the [shared writing rules](../../../AGENTS.md#artifact-writing-rules).
Review records and retained findings follow the
[review lifecycle](../agentic-lens-review/references/parallel-review.md#completed-review-retention).

## Iteration lifecycle

1. **Plan:** Give the planner the request, baseline, contract and active iteration.
   Ground in existing modules/tests; write the initial contract/plan or a scope
   delta under the repository's iteration rules. Select relevant
   [regression boundaries](../../../docs/REGRESSION_PREVENTION.md), including
   cross-machine cases for persistence/refresh. Tests cannot impose a new
   authenticity policy on portable history.
2. **Approve:** Present contract/plan and obtain explicit scope approval before
   dispatching a coder. Existing explicit approval satisfies this gate; record it.
   Material scope changes need renewed approval, not routine corrections.
3. **Implement:** Assign the coder owned paths, contract, current plan and any
   findings. The orchestrator coordinates; it never implements application fixes.
4. **Verify:** Assign the tester relevant regressions and missing contract cases
   under [test proportionality](../../../TECHNICAL_GUIDELINES.md#test-purpose-and-proportionality).
   Reuse valid existing checks; record actual results and skipped stages.
5. **Review:** Follow the [parallel protocol](../agentic-lens-review/references/parallel-review.md)
   for focus/model selection, independent frozen-revision investigations,
   attribution, findings, corrections and re-review. Pass file pointers and
   task-specific boundaries rather than copying common instructions. Each child
   owns a separate report; the parent alone consolidates.
6. **Integrate and deliver:** Once focused investigations have no unresolved
   blockers/material gaps, commit owned changes, integrate main and run `make`
   under [delivery policy](../../../AGENTS.md#1-ticketed-and-ad-hoc-work).
   Assess integration changes against the review evidence. Final consolidated
   approval requires passing integrated checks; child approval alone cannot open
   publication. Complete artifact/index lifecycle, final telemetry, commits, push
   and PR handoff under that same policy.

`CHANGES_REQUESTED` returns to coder/tester. `INCOMPLETE` requires the missing
material evidence before delivery. A blocked scope decision returns to planning;
ask again only for changed approval boundaries. `APPROVED` permits publication,
not merging. Preserve applicable approvals/checks for unchanged behavior and
compact completed review records only through the review retention protocol.
