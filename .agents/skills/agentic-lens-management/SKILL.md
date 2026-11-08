---
name: agentic-lens-management
description: Supervise AgenticLens ticket threads and coordinate non-release pull requests through integration and merge under an MGT ticket. Use for repository-wide PR management and cross-ticket delivery; application implementation stays with each ticket owner.
---

# AgenticLens management

Apply the [shared artifact writing rules](../../../AGENTS.md#artifact-writing-rules); the artifact definitions below belong to this workflow.

Use one configured MGT ticket for coordination. Follow [AGENTS.md](../../../AGENTS.md) for artifacts, ownership, telemetry and delivery. Read the active management contract and iteration before acting. A management thread controls independent ticket threads as a user proxy; it does not become their application developer or telemetry parent.

## Management artifact definitions

Keep the contract focused on authorized supervision/merge boundaries; iteration
plans record priority or acceptance deltas. Use a single control ledger for the
ownership/status fields below, updating rows instead of appending polling history.
`SUMMARY.md` explains delivered coordination outcomes, remaining blockers and
external review ownership (usually 150–300 words before generated metrics).
`PULL_REQUEST.md` describes management guidance/decisions and validation, with a
conventional title first (usually 100–200 words; at most 3000 characters). Link
the ledger and owning ticket evidence rather than copying their implementation
reports. The final `agentic_metrics.json` records this management session only.

## Authority and release exclusion

Record the user's actual authorization to supervise, send instructions, fix reviewed defects through ticket owners, and merge. A review-only request does not authorize messages or merges. Explicit authorization to supervise until merge permits routine owner instructions and eligible non-release merges after the gates below; do not ask again for those steps. The owning management thread's own management PR is excluded from that merge authority.

**Release pull requests are human controlled. Never merge, enable auto-merge, approve on behalf of a human, close, or instruct another agent to merge a release PR.** Identify release candidates by branch (including release-please), title, labels, manifest/version/changelog changes and automation provenance. Keep them in the inventory as `human_only`; report their state without treating an open release as unfinished agent work. When classification is uncertain, hold the merge for a human decision. This exclusion also applies to a replacement release PR and to otherwise broad instructions to merge every PR. Only the human performs a release merge.

**Do not self-merge management work.** The management thread may prepare, push and report its own MGT PR, but must leave review/merge to a human or a separately authorized independent controller. Do not enable its auto-merge or ask a child to perform the same merge as a workaround. Determine ownership from the management ticket/control record and creating/publishing thread, not a shared GitHub author account. Keep the management PR open and draft while its workflow is evolving; an accepted checkpoint still needs external merge ownership. Green CI alone does not authorize self-merge. Report it as `external_review`, not as an automatically mergeable item in the product queue.

## Discover ownership

1. Capture all open PRs from the actual Git remote: URL, head SHA, base, title, author, labels, draft state, mergeability, checks, reviews and unresolved discussions. Rediscover after merges; cached heads and earlier audit findings may be stale.
2. Match each non-release PR to an existing independent ticket thread using linked PR URL first, then corroborated repository, ticket and branch. Read the thread before describing or sending to it. A thread label may lag a renamed Git branch; verify the actual PR head. Search/list with pagination when needed. Never treat a native Codex session ID as a T3 thread ID.
3. Record a small control ledger in the management workspace: PR/ticket, thread link, current head/base, dependency/order, owner, last instruction/run, verified checks, blockers, and next action. Distinguish `preparing`, `ready`, `waiting_checks`, `blocked`, `merged`, `closed_unmerged`, `unmapped`, and `human_only`. Capture actual merge SHAs, not just owner reports.
4. If the original thread is unavailable, record the exact lookup failure. Do not impersonate it or invent a successful mapping. Continue independent work. Manage a verified documentation-only PR directly if authorized; application repairs require an available ticket owner. A replacement root for that existing ticket requires a user request; do not bypass missing ownership by allocating a duplicate ticket. Newly discovered defects follow the routing rule below.

## Route newly discovered defects

During authorized management supervision, every newly discovered issue or bug must get a separate `FIX-#` ticket and a new independent thread unless it is **very closely related** to an existing ticket. Keep a defect with an existing owner only when it directly regresses that ticket's approved behavior or must be corrected to meet the same accepted contract. Sharing files, failing the same test suite or blocking the same PR is not sufficient; do not expand or repeatedly reopen a feature ticket to absorb unrelated bugs.

Proactively create the required FIX ticket and launch its isolated owner using the launch protocol below. The user's authorization to supervise and fix discovered bugs includes these routine FIX launches; do not ask again merely to allocate the ticket or thread. Review-only authority does not authorize execution, and material product decisions still require approval. Reuse an already owning FIX thread instead of creating a duplicate.

Record the defect, concrete reproduction/expected behavior, relevance decision, FIX identifier, owner/thread and dependencies in the management ledger. Check existing ticket usage and ownership before allocating an unused `FIX-#`; the new owner maintains its own configured ticket workspace, implementation, verification and telemetry. Keep the original tickets' acceptance boundaries and completed work intact. Independent fixes may run in parallel; serialize dependent integration and merges after actual predecessor delivery.

## Instruct the owning thread

Use the available T3 thread read/send tools for existing independent roots. Do not replace them with native subagents or delegate_task merely to control their tickets. User-authorized sends may start idle owners or steer running ones; use a stable client request ID and record the returned run ID to avoid duplicate instructions. Preserve provider/model/modes. Do not restart, cancel or replace a running owner unless necessary and authorized.

### Message contract

Start with `$agentic-lens-dev`, `$agentic-lens-po` or `$agentic-lens-review` as
appropriate. State the desired outcome, actual approval/acceptance delta and
concrete inputs, observed behavior and expected result. Include only accessible
repository paths or durable links; never depend on another thread's temporary
files or local worktree. Name finding IDs only for accessible shared records.

Keep follow-ups to the decision/correction needed, preserving completed work.
Establish writable ownership/isolation at assignment or transfer; keep routing
checks and coordination bookkeeping in the manager ledger. Include operational
metadata only to resolve an actual ambiguity. State specific handoff/merge
ownership; let the invoked skill govern standard execution and delivery.
Before sending, remove unnecessary, nonportable or duplicated instructions.

Allow independent defect fixes in parallel. Serialize upstream integration and merges where files or dependencies overlap. Send an upstream-update instruction only once its predecessor is actually merged; do not repeatedly restart work to chase every unrelated push. Preserve passed checks for unchanged code and require `make` after actual integration, with embedded assets checked for frontend changes.

Use the invoked skill’s handoff evidence to assess readiness. Ask only for a missing fact needed by the merge gate; do not resend the entire workflow. Do not equate a completed thread turn with a delivered ticket.

## Launch ticket owners

Before each launch for requested backlog work or a newly discovered FIX under the routing rule above, inspect the current roadmap/dependency graph, ticket usage and existing PR, thread, branch and worktree ownership. Match complete ticket IDs, not substring references in another ticket's conversation. Reuse a live or paused owning thread instead of duplicating it. A Ready label alone does not settle superseded scope or contradict an explicit planning hold. This FIX authorization does not release unrelated roadmap work.

Create independent top-level threads with the host launch tool and an explicit isolated worktree strategy bound before the first message; do not create checkouts in the shell or prompt the new agent to make its own worktree. Inherit provider/model/settings unless the user requested otherwise. Use a stable client request ID, retry only with that ID after an uncertain response, then read the created thread and record its exact link/run/ticket ownership. Before the owner writes files, require it to verify its actual cwd and Git toplevel against the assigned host worktree. Corroborate T3 workspace metadata with a distinct Git worktree path/gitdir, not just a different branch label. If a mismatch is observed, stop that root, preserve files and repair the host binding; do not move files or create shell worktrees to conceal the mismatch. Explicit user stop instructions take precedence over pending dispatch.

Invoke `$agentic-lens-dev` for development owners in their own ticket workspace, with explicit planning/code acceptance boundaries and existing approval evidence. Invoke `$agentic-lens-po` for product owners in a separate PO ticket/root; they own only planning/specification artifacts and never implement application code or author developer contracts. Assign one owner to shared planning files; serialize updates with developer status transitions. When establishing a new owner or transferring responsibility, distinguish its separate worktree and shared Git history/remotes from native roles sharing one ticket checkout. Keep routine follow-ups within the message contract above. A PO session ends at planning delivery; subsequent development belongs in an independent developer session.

Ask the user through the question tool when priorities, overlapping acceptance criteria, ticket-series scope, or schema/cost semantics need a product decision. Continue independent grounding while waiting, but do not treat elapsed time or absent answers as approval for dependent implementation. Publish the decision and its source before releasing affected developers. Recheck dependencies and ownership after each relevant merge and dispatch only newly unblocked, accepted work; do not launch every backlog item merely because its label says Ready. Preserve release exclusion throughout expanded supervision.

## Integration and merge gate

Preserve the [portable ledger contract](../../../TECHNICAL_GUIDELINES.md#portable-ledger-contract) when disposing findings and coordinating fixes. Establish a rejection rule's product authority before expanding a ticket to preserve it. Do not turn local capture evidence, fingerprints or existing test assertions into a new authenticity policy for Git-shared history. Select the [cross-machine scenarios](../../../docs/REGRESSION_PREVENTION.md#portable-ledgers-across-developers-and-machines) for affected integration; route a genuine contract change to an explicit product decision.

Reconcile cross-PR behavior, not just textual conflicts. Use fetched immutable heads for merge-tree simulations; keep speculative conflict trees outside product branches. Assign each shared production-file resolution to the later owner, with adjacent guarantees and tests from both features. Do not resolve by discarding one side. Reconcile planning dependencies against what is actually on main.

Before each non-release merge, freshly verify:

- It is the intended repository/base/PR and still classified non-release, open, and non-draft; it is not this thread's own management PR.
- The head is exactly the reviewed/tested head and includes the required current base. Any new head invalidates earlier readiness until its changes and checks are assessed.
- Required checks passed for that head; there are no pending/failed required checks, unresolved blocking review findings/discussions, or changes-requested reviews. An empty required-check list is not proof that verification ran.
- The owner completed applicable review and full integration checks, and cross-ticket dependencies are met. Explain any unavailable boundary rather than treating it as passed.

Use the repository-supported merge method with an expected-head guard (for example `gh pr merge --squash --match-head-commit <SHA>` when squash is enabled). Do not use admin bypass, force-push, or enable auto-merge to evade readiness. A race, changed head, protection failure or new objection returns the PR to preparation. Verify GitHub reports MERGED and record the merge commit; refresh main before releasing the next owner. Link any PR taken over or pushed using the host's linking tool. Do not check out main in agent worktrees.

## Asynchronous supervision

Use native thread completion notifications and host PR watches instead of sleep loops, repeated status polling, or `gh --watch`. Follow a tool's returned `next` instruction; register the relevant completion/watch before yielding. When a watch requires ending the turn, finish any needed edits, push and management records first. A wake is a reason to recheck readiness, never permission to merge blindly. Preserve the control ledger across continuations and resume its next action. Remove obsolete watches when their target is merged, replaced, or no longer managed; follow the host's handoff/unwatch rules.

Do not say supervision is complete while an eligible PR remains open, blocked, or merely has auto-merge requested. A true blocker needs evidence and a specific required user action; continue other unblocked tickets. Finish when every in-scope non-release PR is verified merged, or explicitly report the remaining blocked/closed-unmerged work. Report human-only releases and externally owned management merges separately. Do not hold the user in an automatic merge loop for your own management PR; hand it off explicitly for external review. Recheck late-arriving PRs for scope and ownership before declaring completion.

## Telemetry and finalization

Keep the manager's coordination records and native root in its MGT workspace. Each independent ticket thread maintains its own contract, summary and source ledger; never copy its token totals into the manager ledger or invent parent links. Reading a result adds manager context but does not transfer telemetry ownership. T3 MCP tool-count omissions in the current decoder are a known completeness limit, not zero control activity.

Append scope deltas and verification to the active iteration. Generate manager telemetry only at the final step of each delivered iteration using its own literal nonce, actual source and native root. Preserve existing source sessions and record unavailable facts. Update or replace the manager PR according to its actual merged state; a merged PR cannot deliver later control-ledger changes. Final handoff identifies merged PRs and merge evidence, remaining blockers, and excluded release PRs without claiming future supervision has already happened.
