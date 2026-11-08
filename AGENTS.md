# Agent Instructions for AgenticLens

Mandatory operational policies for all agents and contributors working in this repository.

---

## 1. Ticketed and Ad Hoc Work

Every task, feature, bugfix, or refactoring in this repository must have a ticket identifier and an isolated ticket workspace. A planning specification or index entry under `plan/` is not required for an ad hoc ticket.

### Configurable Workspace Root
- Define `<WORKSPACES_DIR>` as the `workspaces.dir` value in `.agenticlens-config.yaml`.
- If `.agenticlens-config.yaml` or the `workspaces.dir` key is absent, use `workspaces` as the fallback.
- Use `<WORKSPACES_DIR>` as the root for all ticket workspaces and plan paths below.

### Rules
1. **Mandatory Ticket Identifier**:
   - A ticket ID is required before starting any work. If the request does not provide one, ask the user or assign an ad hoc `GEN-#` identifier using the configured prefixes in `.agenticlens-config.yaml` (fallback: `GEN`).
   - Use an existing ticket ID when one is provided; do not create a duplicate.
   - An ad hoc ticket ID does not require a corresponding specification or entry under `<WORKSPACES_DIR>/plan/`.
2. **Single-Ticket Exclusivity**:
   - Agents are strictly prohibited from working on more than one ticket at a time.
   - Once work begins on a ticket, the agent is locked to that ticket. Agents must refuse any request to switch to another ticket or work on multiple tickets simultaneously within the same session.
3. **Prompt Steering & Iterations**:
   - Once a ticket is active, every subsequent user prompt must be interpreted as steering feedback for that specific ticket.
   - Advance iterations (`iterations/<NN>/`) when scope or acceptance boundaries change. Keep clarification, status questions, and review corrections within the current iteration; do not create a new plan for each message. Preserve approval and completed verification across compaction. Honor explicit requests to consolidate plans and remove empty iteration folders. Record each material steering delta with its source, changed acceptance criteria, preserved guarantees, superseded decision, and verification impact. A later plan number alone does not establish approval or supersede accepted behavior; current explicit user instructions govern. Distinguish new requirements from review corrections, routine integration, and previously verified work.
4. **Dedicated Workspace Allocation**:
   - Every task must have its isolated workspace under `<WORKSPACES_DIR>/<TICKET_ID>/`, including ad hoc tickets.
   - Never mix artifacts across tickets or leave task-owned changes untracked at handoff. Record the initial Git status and preserve unrelated user or agent changes; stage only files owned by this task.
   - Check the existing workspace and ticket index before allocating an ad hoc identifier. A generated branch name is not proof that its ticket number is unused.
   - A rename or relocation of ticket artifacts must not require application or test changes. Tests must use self-contained fixtures, never the live ticket workspace.
5. **Workspace Deliverables Checklist**:
   - Apply the [shared writing rules](#artifact-writing-rules); each skill owns its artifact purposes, content and length guidance. Required evidence and unresolved limits must survive shortening.
   - `<WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md`: Current outcome, material decisions, verification and remaining limits.
   - `<WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md`: Pull request description and context.
   - `<WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md`: Frozen technical contract and acceptance boundaries, when required by the configured artifact list.
   - **Plan Lifecycle Synchronization**: When a matching specification exists in `<WORKSPACES_DIR>/plan/to-do/<TICKET_ID>*.md`, move it to `<WORKSPACES_DIR>/plan/done/` when the ticket is completed. Remove the completed ticket entry from `<WORKSPACES_DIR>/plan/TICKETS.md` only if one exists; do not create a specification, ticket entry, or Done section just to satisfy this lifecycle rule. For ticketed work, unblock downstream tickets whose dependencies are now fully satisfied by promoting them to `## Ready` with `Status: Ready`.
   - `<WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json`: Raw session telemetry ledger (generated at the very final step of the iteration).
   - Store plans under `iterations/<NN>/PLAN.md`. Review artifacts, all finding/actor attribution and cleanup follow the [canonical retention protocol](.agents/skills/agentic-lens-review/references/parallel-review.md#completed-review-retention). Retain REVIEW_SUMMARY.json and its linked human summary before deleting approved-review working files. Never hide unresolved findings, accepted debt or incomplete evidence; unrelated historical audits are not cleanup targets.
6. **Telemetry Ingestion as the Final Step & Conventional Commit Type Classification**:
   - **Delayed Ingestion Mandate**: Generating telemetry must always be delayed to the very last step of the workflow. Running telemetry early causes subsequent turns, tool calls, compactions, and subagent sessions to be omitted from the ledger.
   - **Literal Nonce Generation**: The agent must generate a unique random nonce directly (by any means, e.g. an 8 to 16 character alphanumeric or hex string) and supply the literal token directly in the command. Shell variables (`$NONCE`) or command substitutions (`$(...)`) must not be used because OpenCode records tool inputs into SQLite before subshell expansion occurs, which causes database searches for the expanded token to fail.
   - Execute AgenticLens CLI with an explicit conventional commit `type` immediately before final commit:
     ```bash
     agentic-lens --al-offline --al-project . generate --nonce <GENERATED_NONCE> --ticket-id <TICKET_ID> --ticket-type <CONVENTIONAL_TYPE> --json-output <WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json --append <WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md
     ```
   - Select the actual harness explicitly when defaults are ambiguous. For a Codex session, add `--source codex`; use `--codex-home` only when its session store is outside the normal configured location. Confirm supported options with the built CLI.
   - Before committing telemetry, verify the root session, source, expected child sessions, and preservation of existing sessions from other harnesses. A successful exit alone does not establish complete extraction. Record unavailable facts or source errors; never manufacture a successful or empty ledger.
   - Commit the generated ledger and rendered summary. Keep raw transcripts outside Git. Backfill requires the machine holding the original source sessions; unavailable source data must remain explicitly incomplete.
   - `<CONVENTIONAL_TYPE>` must match standard conventional commit classifications (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).
   - PO tickets (`PO-#`) must strictly use `--ticket-type docs`.
   - Generic tickets (`GEN-#`) must use `--ticket-type chore` unless explicitly a feature (`feat`).
   - Developer tickets must use the corresponding conventional commit type matching their branch prefix (`feat`, `fix`, `refactor`, etc.). Never default all tickets to `feature`.
   - Ad hoc tickets use their assigned ticket ID and isolated workspace for telemetry. They do not require a specification or index entry under `plan/`.
7. **Pull Request Delivery & Handoff**:
   - Commit all implementation deliverables on the feature branch first (stage the reviewed task-owned paths, then commit). Git requires a clean working tree before merging.
   - Merge `main` into the feature branch (`git fetch origin main && git merge origin/main --no-edit`) to resolve any upstream conflicts, and verify test suites pass (`make`).
   - Extract the PR title directly from the first line of `PULL_REQUEST.md` (e.g. `PR_TITLE=$(head -n 1 <WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md | sed 's/^# *//')`).
   - Check the existing PR state before follow-up delivery. If it has merged, compare the remaining branch changes with current `origin/main` and open a new PR for any undelivered changes; do not report a merged PR as delivering later work.
   - Commit final deliverables and telemetry facts on the feature branch, push upstream, and open a GitHub Pull Request via `gh pr create --title "$PR_TITLE" --body-file <WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md`.
   - Avoid directly merging or checking out `main` within agent worktrees.
   - The Pull Request URL must be prominently displayed at the top of the final handoff report.

### Artifact writing rules

- Write for a human without the conversation: lead with the goal or outcome, use
  plain language and include technical detail only when it explains a decision.
- Give each document one job. Keep it understandable on its own, then link shared
  context and evidence instead of copying sections or repository rules.
- Maintain the current account, not a session diary. Omit routine tool attempts,
  dispatch logs, repeated status and file inventories. Record material steering
  once in its iteration plan; link it from the summary.
- Follow the producing skill's artifact definitions. Lengths are prose targets,
  not quotas; omit empty sections and never pad or remove necessary evidence to
  meet a target. Add an artifact only when it needs a distinct purpose.
- Preserve verification, discovery credit, unresolved findings and limits under
  their owning protocols. Keep generated metrics intact without repeating them
  in prose; keep raw transcripts and private diagnostics outside Git.

---

## 2. Multi-Agent SDLC & Role Boundaries

Application development in this repository is orchestrated through specialized agents and skills. Documentation-only maintenance and session audits may directly update instructions and ticket artifacts without launching a full development cycle. If the work reaches application code, tests, schemas, or runtime scripts, use the development workflow below.

For delegated work, name one owner for each writable path, state the role and acceptance boundaries, and pass the contract, current plan, review findings, and baseline commit. Tell agents that they share the checkout and must preserve others' changes. The parent owns integration, telemetry, commits, and PR delivery unless explicitly delegated. A tester or reviewer must not publish or continue product implementation on its own.

Role workflows have one owner:

- [Management](.agents/skills/agentic-lens-management/SKILL.md) supervises independent ticket owners under MGT, without taking their implementation/telemetry. Never merge releases or the owning thread's management PR, enable their auto-merge or delegate that merge. Keep an evolving MGT PR open/draft for external review; later work after a merged checkpoint needs a new PR.
- [PO](.agents/skills/agentic-lens-po/SKILL.md) authors specs/tickets only: no application, tests, runtime scripts, adapters or developer contracts. End the PO session at planning delivery; development requires a separate session.
- [Development](.agents/skills/agentic-lens-dev/SKILL.md) coordinates planner -> coder -> tester -> focused reviewers -> consolidated review. The orchestrator never implements application code. Obtain explicit scope approval before coding; record existing pre-approval instead of asking again. Material boundary changes require renewed approval.
- Planner owns zero-code contracts/plans; coder owns implementation; tester owns adversarial tests with zero domain mocks and no silent skips. [Reviewers](.agents/skills/agentic-lens-review/SKILL.md) independently investigate relevant changed/unchanged dependencies. Product code stays read-only; report writes need explicit permission. INCOMPLETE evidence cannot pass delivery.

---

## 3. Architecture & Code Invariants

[TECHNICAL_GUIDELINES.md](TECHNICAL_GUIDELINES.md) owns engineering rules:
Go 1.23+ in backend, TypeScript 5.x+ strict in web, pure immutable domain logic,
I/O in adapters/CLI, typed states and zero any/interface{}. Embed the compiled
frontend in the single Go binary; no production Node runtime. Zero emojis in
code, comments, UI/logs, commits, PRs or artifacts.


## 4. Session-Grounded Execution Practices

Select relevant [regression boundaries](docs/REGRESSION_PREVENTION.md) and record
their evidence in the active iteration. Canonical technical rules stay in
[TECHNICAL_GUIDELINES.md](TECHNICAL_GUIDELINES.md).

- Ground changes in configured paths/classification, owning modules, existing tests
  and installed command help, not ticket names or one developer's checkout.
- Distinguish real source/session facts from inference and synthetic fixtures.
  Surface unknown pricing, missing children, source errors and partial migrations.
- Preserve [portable history](TECHNICAL_GUIDELINES.md#portable-ledger-contract)
  across machines/harnesses without private capture receipts. Select cross-machine
  cases for persistence/refresh; tests/reviews cannot invent authenticity policy.
  Optional hydration must not hide supported persisted facts when unavailable.
- Name version-specific adapters/tests by schema family and responsibility; shared
  logic/comments must reflect actual dispatch. Bound searches/output and change
  the hypothesis after repeated failure; retain progress across compaction.
- Follow [test proportionality](TECHNICAL_GUIDELINES.md#test-purpose-and-proportionality):
  identify regressions, inspect coverage, justify deletions and preserve distinct
  assertions. Scores/counts/ratios guide investigation, never automatic thresholds.
- Record tested revision/diff, commands, results and skipped/environment limits.
  Reuse valid unchanged-code results; rerun affected checks after fixes and `make`
  after integration. Frontend work requires canonical build, current embedded
  assets under backend/internal/adapters/web/dist and standalone-binary verification;
  respect generated-file tracking policy, never restore stale assets for a small diff.
- Close with delivered behavior, verification, limits and the correct PR URL.
  Plans, child reports and earlier PR results alone do not establish completion.
- Correct recurring guidance at its source; search active docs/personas/skills/setup
  templates for conflicts. Link canonical rules; preserve historical audit records
  and distinguish proposed recommendations from implemented fixes.

Rationale: [GEN-12 session audit](alspaces/GEN-12/iterations/01/REVIEW.md).
