# Agent Instructions for AgenticLens

Mandatory operational policies for all agents and contributors working in this repository.

---

## 1. Ticketed and Ad Hoc Work

Every task has one canonical configured ticket ID and one isolated execution home. Roadmap tickets and execution workspaces serve different purposes.

### Roadmap, changes and configured ledger roots

- This repository's long-term roadmap remains in `alspaces/plan/to-do`, `alspaces/plan/done` and `alspaces/plan/TICKETS.md`. Tickets own evolving intent; the index owns priority, readiness, dependencies and status. Creating or reprioritizing a ticket does not create an OpenSpec change.
- Selected work uses a linked change under `openspec/changes/<ticket-prefixed-description>/`; archive moves that same home under `openspec/changes/archive/`. Define `<CHANGE_DIR>` from the selected actual change location, never reconstruct it from a ticket ID. Use the pinned [OpenSpec workflow](docs/OPENSPEC_WORKFLOW.md). A ticket links to its execution change and that change links back to the ticket.
- Read explicit `workspaces.roots` in the selected `.agenticlens-config.yaml` or `.agenticlens-config.yml`; both suffixes together are an error. Each entry has `path` and `discovery.max_depth`/`max_directories`. No root ID, minimum depth, scalar `workspaces.dir`, runtime single-directory override or invented root is supported. Writers initialize depth 0/budget 300; this repository explicitly selects depth 1/budget 300 for its containers. Missing/invalid policy requires setup.
- AgenticLens discovers only directories containing `agentic_metrics.json`; ledger ticket identity is authoritative. Folder names, OpenSpec metadata and planning files establish no runtime identity. Before first generation, supply the canonical ID and explicit ledger destination. No workspace registration, identity sidecar or bind/create/resolve command is needed. Distinct ledgers with the same ID require explicit location selection, never root-order precedence or merging.
- OpenSpec creates its folders; arbitrary other configured layouts remain first-class platform choices. Moving a ledger across configured roots does not change ticket/session identity, type, completion or ownership. All readers/writers use the same resolved location; refresh must not recreate former/deleted destinations.

### Rules

1. **Ticket identity and exclusivity**: Use the supplied ID. If absent, ask or allocate an unused ID from an explicitly configured literal prefix with description; never invent a prefix. Check the roadmap and existing work before allocation. Ad hoc tickets need no roadmap entry. An agent remains locked to one ticket for the session; subsequent prompts steer that ticket, not another task.
2. **Execution isolation and preserved work**: Read the actual branch/worktree, initial Git status, ticket, selected change and current approval before writing. Never mix ticket artifacts or discard unrelated changes. Assign one owner per writable path; stage only task-owned deliverables. Do not move another active owner's writable artifacts without coordination. Existing historical audits and supported ledgers remain intact; no wholesale roadmap migration is required.
3. **Scope and iterations**: `design.md` owns the execution contract, explicit user approval and material steering. `tasks.md` owns current steps and progress. Keep clarification/review corrections in the current iteration; advance an iteration only for a material acceptance boundary change unless the user explicitly consolidates it. Record source, changed acceptance, preserved guarantees, superseded decision and verification impact in design, without a parallel editable plan/contract. A roadmap edit, planning readiness, later iteration or checkbox is not scope approval. Preserve existing approval/evidence across compaction.
4. **Change artifacts**: `proposal.md` links the roadmap intent; `specs/<capability>/spec.md` owns selected observable acceptance; `design.md` owns technical decisions/approval/steering; `tasks.md` owns execution sequencing; `delivery.md` owns current outcome, tested revision/diff, actual checks, attributed review, limits and PR context; `agentic_metrics.json` owns final raw session telemetry. PO planning changes use the planning schema and never author developer contracts. Do not keep independently editable CONTRACT/PLAN/SUMMARY/PULL_REQUEST copies alongside these artifacts. Review evidence is change-local under `evidence/iterations/<NN>/`, following [canonical retention](.agents/skills/agentic-lens-review/references/parallel-review.md#completed-review-retention); retain compact findings and linked human evidence before deleting completed review working files. Keep unresolved findings/incomplete coverage visible. Tests use self-contained fixtures, never a live ticket workspace or change path.
5. **Roadmap lifecycle**: Update roadmap status from verified delivered outcomes. Only when a matching ticket is actually complete, move its existing specification from `alspaces/plan/to-do` to `alspaces/plan/done` and remove its existing index entry; never create a ticket/spec/Done section to satisfy this rule. Promote downstream tickets to Ready only when all dependencies are delivered. OpenSpec artifact readiness, task completion, archive and roadmap-ticket completion are separate decisions. Keep navigable links to archived execution evidence.
6. **Operational publication state**: Local setup explicitly writes `agenticlens.state_dir`, initially `~/.agenticlens/state`, separately from versioned binaries and capture storage. Writers may initialize it without capture enablement/registry access. Journals/backups and locks stay there; only a transient pending-publication indicator accompanies an interrupted ledger. It is transaction status, never workspace identity. Reads require no private state and must report a pending set as incomplete when recovery data is unavailable. Retired journals/configuration require explicit update/recovery, not compatibility fallbacks or deletion of accepted facts.
7. **Final telemetry and commit type**:
   - Delay telemetry to the last extraction step immediately before the final artifact commit, so later calls and child sessions are included. Generate a unique literal nonce directly and put that literal in the command; no shell variable or command substitution for the nonce.
   - Confirm supported options with the built CLI and select the actual source explicitly. A Codex session uses `--source codex`; use `--codex-home` only for an actual nonstandard session-store location.
   - Publish to the selected change, with an explicit conventional type:
     ```bash
     agentic-lens --al-offline --al-project . generate --source codex --nonce <LITERAL_NONCE> --ticket-id <TICKET_ID> --ticket-type <TYPE> --json-output <CHANGE_DIR>/agentic_metrics.json --append <CHANGE_DIR>/delivery.md
     ```
   - Verify root/source/expected children and preservation of unavailable owners and other harnesses. Exit zero alone does not establish complete extraction. Record missing sources, unknown costs and incomplete facts; never manufacture an empty/successful ledger. Commit raw ledger and rendered delivery metrics; keep transcripts, private journals and source stores outside Git. Backfill needs the machine with the original sources.
   - Types are `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf` or `build`. PO uses `docs`; GEN uses `chore` unless explicitly a feature. Development uses the actual conventional branch/change classification, never blanket `feature`.
8. **PR delivery and handoff**:
   - Inspect/stage owned implementation paths and commit the candidate before integration; preserve unrelated work. Fetch/merge main into the feature branch, resolve both sides' accepted behavior, then run required integrated `make` and affected verification/review.
   - `delivery.md` starts with the conventional PR title and contains the authored PR context. Derive a temporary body from its Pull request section, excluding generated metrics; do not maintain another editable PR document. Create/update the PR using that derived body file.
   - Check existing PR state before follow-up delivery: a merged PR cannot deliver later changes. Commit final evidence and telemetry, push, create/update and link the correct PR. Approval to publish is not approval to merge. Never directly check out or merge main within an agent worktree.
   - Display the PR URL prominently at the top of the final handoff, with delivered behavior, verification and limits.

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
I/O in adapters/CLI, typed states and zero any/interface{}. Distribute a native installation/router manager and a separate engine; embed the compiled
frontend in the engine. The manager must not import engine CLI or web assets; the
engine must not dispatch to another build. No production Node runtime. Zero emojis in
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
