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

Read configuration and Git status, identify the supplied ticket and actual
branch/workspace, and record the baseline. Preserve unrelated changes. For a
resumed ticket, read the contract, active plan, approval, review status and valid
verification before selecting the next action. Reuse its workspace/branch;
create a feature branch only when needed, never develop on main. Read a matching
specification and update an existing index entry to In Progress; ad hoc work
requires neither a specification nor an index entry.

| Artifact | Content | Usual prose length |
| --- | --- | --- |
| `CONTRACT.md` | Frozen problem, agreed outcome, exclusions and observable acceptance; link canonical technical rules. | 200–400 words |
| `iterations/<NN>/PLAN.md` | Steps, affected components/owners, verification and review coverage; later plans record acceptance deltas. | 200–400 words |
| `SUMMARY.md` | Current behavior, decisions, actual checks/revision, review outcome and limits. | 300–600 words before metrics |
| `PULL_REQUEST.md` | Conventional title first; problem, resulting behavior, validation and limits; link contract/summary. | Usually 100–250 words; at most 3000 characters |
| `agentic_metrics.json` | Final CLI-generated session ledger; rendered metrics follow the authored summary. | Structured output |

The contract records agreed outcomes and boundaries; the plan describes how the
current iteration will achieve them. A correction within accepted scope updates
the plan/implementation, not the frozen contract. Record acceptance changes as a
steering delta under repository iteration policy and obtain any renewed approval.
Use the [shared writing rules](../../../AGENTS.md#artifact-writing-rules).
Review records and retained findings follow the
[review lifecycle](../agentic-lens-review/references/parallel-review.md#completed-review-retention).

## Assignment and return contract

Give each agent its role, specific outcome, owned writable paths, preserved
behavior, baseline/candidate revision, contract/active-plan pointers and relevant
entry points or findings. State that the checkout is shared. Use the supported
harness tools; do not copy technical policy into delegation prompts.

Expect changed artifacts/paths, results with tested revision and commands,
unresolved decisions or limits, and the next required action. The parent inspects
actual artifacts/diffs and reconciles deviations; a successful child turn alone
does not establish completion. The orchestrator never implements application fixes.

## Iteration lifecycle

### 1. Plan

Give the planner the request, existing agreement/approval, baseline and relevant
code/test context. Expect the initial contract/plan, or an acceptance delta for a
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
   integrated checks pass. Update human artifacts and specification/index status;
   compact approved review records through the retention protocol.
5. Generate telemetry at the final extraction point before the final artifact
   commit: literal unique nonce, actual source and conventional ticket type.
   Verify root/child membership, prior-session preservation and source limitations;
   a successful exit alone is insufficient. Exact commands belong to delivery policy.
6. Commit final artifacts, push and create/update the correct PR. Check its actual
   state: a merged PR cannot deliver later work. Extract the title from
   PULL_REQUEST.md, link the PR through available host tools and hand off behavior,
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
