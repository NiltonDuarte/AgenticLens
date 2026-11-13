---
name: agentic-lens-dev
description: Use when building features, fixing bugs, or refactoring the standalone AgenticLens application in Go (backend/CLI/engine) and TypeScript (frontend dashboard) with multi-agent orchestration (planner, coder, tester, reviewer). Requires a Jira ticket or task identifier.
---

# AgenticLens development

Coordinate application changes through planner, coder, tester and independent
reviewers. Follow [repository policy](../../../AGENTS.md) for ticket isolation,
role boundaries, approvals, telemetry and delivery, and
[technical guidelines](../../../TECHNICAL_GUIDELINES.md) for engineering rules.
Read the selected canonical shared YAML for explicit ticket prefixes and `workspaces.roots`; missing/invalid policy fails with setup guidance. Both canonical suffixes together are an error. Roadmap tickets remain separate from execution changes. Select/create the ticket-linked OpenSpec home through the [pinned workflow](../../../docs/OPENSPEC_WORKFLOW.md); paths below are relative to that actual change directory, including after archive.

## Setup & Workspace Structure

Read configuration and Git status, identify the supplied ticket and actual
branch/workspace, and record the baseline. Preserve unrelated changes. For a
resumed ticket, read design/tasks, approval, review status and valid
verification before selecting the next action. Reuse its workspace/branch;
create a feature branch only when needed, never develop on main. Read the matching roadmap specification and update its existing index entry to In Progress when work starts; ad hoc work requires neither a roadmap specification nor an index entry. Link ticket and change in both directions without copying roadmap status into execution artifacts.

| Artifact | Content | Usual prose length |
| --- | --- | --- |
| `proposal.md` and `specs/<capability>/spec.md` | Selected execution intent and observable acceptance linked to the roadmap ticket. | Concise, sufficient for independent review |
| `design.md` | Frozen technical contract, explicit approval, owners, decisions and material iteration steering. | 200–400 words before necessary decisions/evidence |
| `tasks.md` | Current implementation/verification steps and progress; design owns scope. | Proportionate task groups |
| `delivery.md` | Current outcome, actual verification/review, remaining limits and authored PR context; metrics appended last. | 300–600 words before metrics |
| `evidence/iterations/<NN>/` | Test/review evidence and permanent compact review findings under canonical retention. | Skill-owned evidence definitions |
| `agentic_metrics.json` | Final CLI-generated session ledger, beside the execution artifacts. | Structured output |

`design.md` records agreed outcomes, approval and boundaries; `tasks.md` describes how the current iteration achieves them. Do not retain parallel editable CONTRACT/PLAN/SUMMARY/PULL_REQUEST documents after transfer into the change. A correction within accepted scope updates
the plan/implementation, not the approved contract in design. Record acceptance changes as a
steering delta under repository iteration policy and obtain any renewed approval.
Use the [shared writing rules](../../../AGENTS.md#artifact-writing-rules).
Review records and retained findings follow the
[review lifecycle](../agentic-lens-review/references/parallel-review.md#completed-review-retention).

## Assignment and return contract

Give each agent its role, specific outcome, owned writable paths, preserved
behavior, baseline/candidate revision, design/tasks pointers and relevant
entry points or findings. State that the checkout is shared. Use the supported
harness tools; do not copy technical policy into delegation prompts.

Expect changed artifacts/paths, results with tested revision and commands,
unresolved decisions or limits, and the next required action. The parent inspects
actual artifacts/diffs and reconciles deviations; a successful child turn alone
does not establish completion. The orchestrator never implements application fixes.

## Iteration lifecycle

### 1. Plan

Give the planner the request, existing agreement/approval, baseline and relevant
code/test context. Expect the initial design/tasks, or an acceptance delta for a
changed scope: proposed behavior, affected boundaries/owners, verification and
review coverage, plus unresolved product decisions. Ground in existing modules
and tests; select applicable [regression boundaries](../../../docs/REGRESSION_PREVENTION.md),
including cross-machine cases for persistence/refresh. Tests cannot impose a new
authenticity policy on portable history.

Before advancing, check that acceptance is observable, adjacent guarantees remain
explicit and unresolved decisions do not make implementation speculative. A
resumed session or routine correction alone does not require another iteration.

### 2. Approve

Present the contract/plan and obtain explicit approval for the proposed scope
before dispatching a coder. Record the instruction establishing approval. Reuse
existing explicit approval when acceptance is unchanged; seek renewed approval
for material boundary changes. A plan file, later iteration number or silence is
not approval. Dependent coding waits for unresolved scope decisions.

### 3. Implement

Assign the coder the approved outcome, owned application paths, relevant callers,
preserved behavior and any verified review findings. Expect implementation,
focused check evidence, deviations and blockers. Inspect the diff against the
accepted plan, including relevant untracked files; preserve unrelated work and
route incomplete implementation back to its owner. Product changes outside the
approved boundary return to planning/approval before dependent work proceeds.

### 4. Verify

Give the tester the contract, candidate diff, plausible regressions, existing
coverage and available raw results. Expect evidence of which contract cases were
inspected/executed, missing regression tests and remaining gaps under
[test proportionality](../../../TECHNICAL_GUIDELINES.md#test-purpose-and-proportionality).

Check that tests reach the claimed behavior and use a meaningful oracle. Test
existence/discovery is not execution; a planned command is not a result. Reuse
valid checks for unchanged code with a revision/scope rationale. Record command,
revision, actual result and skipped/environment limits; missing required evidence
must be recovered before delivery, without silently treating it as passed.

### 5. Review and correct

Before dispatch, load the [parallel protocol](../agentic-lens-review/references/parallel-review.md)
and its model policy. Select focus coverage, record any permitted low-risk
exception, and assign every selected focus to its configured model profiles.
Freeze a candidate revision/diff and give independent reviewers the same
requirements/raw evidence, separate report ownership and no sibling conclusions.
Use capacity-limited waves when needed.

Expect scoped verdicts, candidate findings with triggers/evidence, investigated
hypotheses and gaps. The parent verifies/deduplicates findings, preserves all
discovery credit and dissent, and alone consolidates the result. Route product
fixes to the coder and missing tests to the tester; the parent does not implement
them. After corrections, freeze the new candidate and reassess affected review
coverage. Carry forward unchanged approvals only with a dependency/diff rationale.
A child approval alone never permits publication.

### 6. Integrate and deliver

When focused investigations have no unresolved blockers/material gaps, use this
order under [delivery policy](../../../AGENTS.md#1-ticketed-and-ad-hoc-work):

1. Inspect/stage owned paths and commit the candidate; integration needs a clean
   working tree. Preserve unrelated changes rather than staging them indiscriminately.
2. Fetch/merge main into the feature branch. Assign conflicts to the coder with
   both sides' accepted behavior; inspect the resolution and reopen affected
   verification/review. Do not discard one side to obtain a textual merge.
3. Run the required integrated `make` gate and any affected checks. Record the
   actual integrated revision/results; resolve failures and review gaps before
   advancing. Reuse unaffected evidence rather than repeating every investigation.
4. Issue final consolidated approval only after applicable review coverage and
   integrated checks pass. Update delivery evidence and the separate roadmap status from verified delivered outcomes; archive alone does not complete a ticket;
   compact approved review records through the retention protocol.
5. Generate telemetry at the final extraction point before the final artifact
   commit: literal unique nonce, actual source and conventional ticket type.
   Verify root/child membership, prior-session preservation and source limitations;
   a successful exit alone is insufficient. Exact commands belong to delivery policy.
6. Commit final artifacts, push and create/update the correct PR. Check its actual
   state: a merged PR cannot deliver later work. Extract the title and derive a temporary PR body from delivery.md, link the PR through available host tools and hand off behavior,
   verification and limits with the URL first. Approval to publish is not approval
   to merge.

## Corrections and stopping conditions

| State | Next owner/action | Gate to advance |
| --- | --- | --- |
| `CHANGES_REQUESTED` | Coder fixes confirmed product defects; tester fills regression gaps; parent verifies disposition. | Affected checks/review cover the corrected candidate. |
| `INCOMPLETE` | Relevant verifier/reviewer obtains missing material evidence or capability. | Evidence gap is resolved, not waived by a lack of findings. |
| `BLOCKED` | Parent routes an implementation obstacle to its owner or a product/contract conflict to planner/user. | Obstacle is resolved; renewed approval only if scope changes. |
| `APPROVED` | Parent completes remaining publication steps. | Current review and integrated checks support delivery; merging remains separately authorized. |

Stop at the requested delivery boundary. Keep unresolved debt/limits visible;
do not restart valid approvals or add speculative review passes after the gate
is satisfied. Protocol details and engineering rules remain in their canonical
sources rather than additional per-ticket checklists.
