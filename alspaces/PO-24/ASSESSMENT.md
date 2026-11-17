# PO-24: Can AgenticLens work with OpenSpec?

**Partly ready: direct storage for active changes is supported by the existing layout, but a reliable full OpenSpec lifecycle is not ready through configuration alone.** The intended target is one structure: `workspaces.dir: openspec/changes`, with each change's `agentic_metrics.json` beside its OpenSpec artifacts. No separate AgenticLens ticket directory is required for this model. Archive-aware dashboard lookup and publication are the principal blockers.

The user clarified that OpenSpec must be the workspace structure, then requested reclassification of the entire assessment as PO-24 and deletion of its iteration plans. The current proposed design below supersedes both the earlier separate-workspace recommendation and the later legacy-root/OpenSpec-specific-mode framing: arbitrary user-selected workspace directories are equal first-class platform roots. This PO workspace remains under the repository's currently configured root until the implementation cutover.
This is an assessment, not an implemented integration or an end-to-end compatibility certification. No application code, tests or runtime configuration were changed.

## Evidence boundary

AgenticLens: `460c511ef96f9e42f2ecbfdba451828b52d3ce91`, matching fetched origin/main at assessment start. OpenSpec: [`d12c2817776779e06891d68230773b3bb673b106`](https://github.com/Fission-AI/OpenSpec/tree/d12c2817776779e06891d68230773b3bb673b106), package version `2.0.0` in that checkout. Assessment date: 2026-11-13. The OpenSpec SHA identifies inspected main source; it does not prove this exact build is published in a package registry. Store support is explicitly beta.

Source references below are relative to this repository for AgenticLens and pinned links for OpenSpec. Searches covered backend, web/src, relevant AgenticLens skills, docs and README for explicit OpenSpec integration; none was found. This is a focused boundary study, not a reconciled whole-repository correctness audit.

## What can work now

OpenSpec supplies planning artifacts and assistant guidance; it does not replace the assistant's session store. AgenticLens selects `codex` or `opencode` in [source resolution](../../backend/internal/adapters/config/source.go). OpenSpec supports those assistants through generated skills; its [supported tools](https://github.com/Fission-AI/OpenSpec/blob/d12c2817776779e06891d68230773b3bb673b106/docs/supported-tools.md) documents Codex skills under `.agents/skills/openspec-*/SKILL.md`. The shared use of Markdown skills creates no inherent telemetry conflict. Support for OpenSpec's other assistants does not imply AgenticLens has adapters for them.

AgenticLens can retain normal session, token, tool, cost-availability, anomaly and delegation facts while an agent works on an OpenSpec change. This is a source-based compatibility inference, contingent on the supported harness version, correct source selection and exact session binding. An OpenSpec workflow was not executed here. Cost remains unavailable where the harness lacks pricing; adopting OpenSpec does not fix that.

Use the OpenSpec change name itself as AgenticLens's supplied ticket ID, such as `add-login`. Configure the existing shared configuration's workspace fields as follows (illustrative subset, not a complete configuration and not applied here):

```yaml
workspaces:
  dir: openspec/changes
  artifacts:
    - path: proposal.md
      title: Proposal
      description: OpenSpec change proposal
    - path: design.md
      title: Design
      description: OpenSpec technical design
    - path: tasks.md
      title: Tasks
      description: OpenSpec implementation checklist
```

The resulting single structure would be:

```text
openspec/
  specs/...
  changes/
    add-login/
      .openspec.yaml
      proposal.md
      design.md
      tasks.md
      specs/.../spec.md
      agentic_metrics.json
    archive/
      2026-11-13-add-login/...
```

AgenticLens still requires its own configuration and raw telemetry ledger; it does not derive session facts from OpenSpec Markdown. For manual generation, `--ticket-id add-login`, an explicit conventional `--ticket-type`, and `--json-output openspec/changes/add-login/agentic_metrics.json` identify the destination. Exact source/session binding remains necessary. Generic numeric-prefix classification is optional for this slug model and does not establish a hard runtime rejection. The automatic writer already constructs `<configured workspaces.dir>/<bound ticket>/agentic_metrics.json` in [capture_worker.go:234](../../backend/cmd/agentic-lens/capture_worker.go#L234); the path shape fits an active change. Actual end-to-end slug capture was not exercised.

Configured artifact reading can show proposal/design/tasks directly because they are now inside the change directory. Known individual nested spec paths can also be configured. The remaining artifact gap is dynamic enumeration of capability deltas and schema-defined outputs, not a requirement to copy the standard files. Repeatable `--tag` is available for extra metadata but is not required to bridge two directories in the desired model.

The raw ledger is essential for portable session history. A separate configured *workspace directory* is avoidable; machine-local capture state and original harness session stores remain operational dependencies for fresh extraction. Eliminating those is a different requirement. OpenSpec's custom schema/rules can adapt guidance without adding a production Node dependency to AgenticLens.
## Gaps and proposed acceptance boundaries

Priorities indicate an adoption sequence, not defect severity or approved implementation tickets. P1 is required for reliable direct repo-local integration; P2 adds richer observability; P3 expands deployment coverage.

| ID / priority | Current boundary and consequence | Smallest useful direction / proposed acceptance check |
|---|---|---|
| G1 / P1: identity | Explicit slug IDs already read successfully. Numeric prefix classification/type inference does not describe arbitrary OpenSpec names. No automatic change-name-to-session binding exists. | Slug IDs are a currently supported primitive. The planned generic model preserves canonical ticket identity independently of the user-selected directory, with explicit binding/creation association and disambiguation. OpenSpec ticket-prefixed naming is optional workflow guidance rather than a mandatory core folder grammar. |
| G2 / P1: discovery | `ScanWorkspaces` recursively finds `agentic_metrics.json`, not OpenSpec Markdown. A planning-only change is invisible; absence of telemetry is not zero effort. | Decide whether an OpenSpec change inventory should be visible before capture. Check that an uninstrumented change has explicit unavailable telemetry and is not fabricated as an empty successful ledger. |
| G3 / P1: archive and detail paths | Scan is recursive, but HTTP detail uses `<workspaces>/<ticketId>/agentic_metrics.json`. OpenSpec moves changes under `archive/<date>-<name>`. Relocating a ledger can leave it listed but inaccessible through ticket detail; slash-containing IDs cannot repair this. | Add one shared identity/location resolver across arbitrary configured roots for dashboard detail, events, artifacts and writers. The current writer always targets `<root>/<id>` and could recreate an active folder after archive if publication continues. Check post-archive refresh, no duplicate resurrection, collision handling and reads with original harness sources absent. Do not solve this by introducing a second telemetry directory. |
| G4 / P2: dynamic artifacts | Proposal/design/tasks and explicit nested spec paths can already be configured inside a direct change directory. `ReadTicketArtifacts` does not expand a custom schema's dynamic spec inventory. | Keep static artifacts configuration for initial direct support. Add schema/output discovery only if full capability browsing is required. Check multiple spec deltas, custom schemas, missing optional files and containment. Global `openspec/specs` is outside an individual change's artifact sandbox. |
| G5 / P1: workflow authority | Repository instructions require ticket isolation, approval before coding, separate PO/development sessions, review, final telemetry and PR delivery. OpenSpec's fluid artifact workflow and default apply/archive guidance do not establish these gates. Its default [schema](https://github.com/Fission-AI/OpenSpec/blob/d12c2817776779e06891d68230773b3bb673b106/schemas/spec-driven/schema.yaml) models artifact dependencies rather than this repository's approval evidence. | Define a policy overlay in project context/rules or a custom schema, with one owner per document. Check that planning completion cannot authorize coding, archive cannot imply review approval, and telemetry/PR evidence survives archive. This is repository adoption work, not a universal AgenticLens runtime requirement. |
| G6 / P2: lifecycle and progress | No OpenSpec change state, artifact dependency state, task-checkbox progress or validation result is represented in the inspected ledger/API contracts. OpenSpec `status.isComplete` means planning-artifact completion, not completed implementation; task progress is distinct. | Consume versioned CLI JSON through a boundary adapter, separating planning readiness, task progress, validation and archive. Check skipped/custom artifacts, incomplete tasks, validation failure and CLI-unavailable states. Never treat file existence or a completed checkbox as verified correctness. |
| G7 / P2: phase attribution | Generic skill facts record name, caller and timestamp, not an authoritative OpenSpec phase or task-to-invocation allocation. Codex detects explicit native skill requests/attachments and skill calls; merely reading SKILL.md or mentioning a command is not activation evidence (`codex/occurrences.go`). OpenCode records actual `skill` tool occurrences. | Retain generic telemetry first. Any phase analysis needs source-backed activation/transition rules and an unknown category. Check direct skills, conversational invocation, resumed contexts and overlap without assigning all session cost to the last phase. OpenSpec task checkboxes are not AgenticLens delegation facts. |
| G8 / P3: stores and cross-repo work | OpenSpec can resolve a separate planning store, declared/global defaults and affected implementation areas. AgenticLens project configuration, capture scope and Git code metrics remain tied to the selected project; no store-aware association or cross-repo change rollup exists. | Defer stores from the first native integration. Later use portable store/change/repository identities, local resolution and explicit aggregation rules. Check two repos with the same slug, unavailable store/source, copied ledgers on another machine, and overlapping sessions without silent deduplication. |
| G9 / P3: assistant coverage | Only Codex/OpenCode are selectable telemetry sources. OpenSpec supports many more assistants. Its own anonymous command-count telemetry does not provide AgenticLens session/token/ownership data. | Publish the intersection explicitly. Each additional harness requires its own source adapter and fidelity evidence; never present unsupported assistants as measured zero. |

## Executed probes and counterevidence

Built the current native engine with `make build PLATFORMS=linux/amd64`; read its help for `generate`, source selection, tags and explicit source paths. A disposable fixture under `/tmp/gen26-layout-gnbrjhlc` used copied shared configuration and an existing ledger relabeled only for layout testing. Those copied records are not new source observations or session extraction.

1. A Markdown-only `add-login/` directory produced no valid ledgers (`--list`, exit 1).
2. Adding the copied schema-6 ledger made the arbitrary slug visible (`--list`, exit 0). Root CLI `--ticket-id add-login` also succeeded. This refutes a blanket assertion that AgenticLens requires numeric IDs everywhere.
3. Moving that directory to `archive/2026-11-13-add-login/` left list and root CLI filtering successful. This refutes a blanket assertion that recursive scanning loses archives. The narrower dashboard gap follows the HTTP call path: [handlers.go:230](../../backend/internal/adapters/http/handlers.go#L230) calls [LoadTicketWorkspace](../../backend/internal/adapters/repository.go#L269), which constructs the direct ticket path. No browser or live HTTP archive probe was run.

Existing repository coverage includes `TestScanWorkspaces`, `TestReadTicketArtifacts`, `TestLoadTicketWorkspace`, prefix-membership tests and explicit telemetry-source tests. The assessment-stage repository `make` result and later PO delivery verification are distinguished in SUMMARY.md. No OpenSpec installation, generation, apply, archive, store mutation or real assisted workflow was run. The installed manager failed before help on this shared configuration; the freshly built engine worked. That environment/version mismatch is separate from OpenSpec compatibility and was left unchanged.

## Direct configuration probe

A second disposable fixture at `/tmp/gen26-direct-openspec-5kj2m6qb` set the actual shared configuration to `workspaces.dir: openspec/changes` and configured proposal/tasks paths. Unlike the first CLI-override probe, it ran the native engine from the fixture root with no `--workspaces` override. Its schema-6 ledger was self-contained synthetic boundary data with no sessions, not extracted telemetry.

- Markdown-only change: `--list` exited 1 with no valid ledgers.
- After adding `openspec/changes/add-login/agentic_metrics.json`: `--list` exited 0 and found `add-login` through configuration alone.
- After moving the same folder under `archive/2026-11-13-add-login`: `--list` still exited 0, but the direct path used by the HTTP loader no longer existed.

This proves configuration-based discovery in the intended single structure. It does not prove live capture, browser artifact display or OpenSpec archive execution.

## Recommended decision

**Do not claim full direct OpenSpec support yet.** Existing configuration is enough to point discovery at `openspec/changes`, read ledgers in active change folders, and configure standard artifact paths. Keep the change slug as the supplied ID and store telemetry in that folder. No separate AgenticLens workspace is necessary for these supported primitives.

The minimum reliable direct integration needs the generic multi-root identity/location resolver shared by readers and writers, explicit binding/rename rules, and adoption guidance that replaces fixed ticket-path assumptions. Those are G3, the remaining binding portion of G1, and G5. G2 is required if planning-only changes must appear before telemetry exists. Dynamic artifacts, task/phase analytics, stores and other harnesses can follow independently.

The source/layout assessment is complete; actual OpenSpec/harness execution and browser/store interoperability remain unverified. No implementation is authorized by this recommendation.

## Proposed fit: generic user-selected workspace directories

The latest user clarification is that multiple workspace directories are a first-class platform capability for any workflow. Alspaces is not a compatibility feature and OpenSpec is not a special storage mode. [SPEC.md](SPEC.md) and the single AL-84 implementation ticket own the current product requirements; earlier proposals for a legacy-root path or mandatory shallow OpenSpec mode are superseded.

Users can select any combination of `alspaces`, `alspaces/archive`, `alspaces/al`, `alspaces/al2`, `alspaces/po`, `alspaces/fix`, documentation folders, arbitrary renamed/nested containers and OpenSpec directories. Directory spelling/depth and grouping are user decisions. Folder labels must not imply ticket type, completion date, lifecycle state, creation destination or ownership. Optional routing by ticket prefix/type must be explicitly configured by the user, not inferred from a folder called al, fix or docs.

One root collection and one discovery/binding/location/publication pipeline should serve all of them. Existing single-directory input is that same model with one member. Discovery scope must be explicit and bounded: immediate-child enumeration is useful for configured containers, while nested layouts need an appropriate recursive/bounded scope or roots selected at the desired depth. There should be no second compatibility scanner and no requirement that every project use one global depth. Normalize overlapping roots/physical aliases so one workspace is not loaded twice, preserve containment and report unavailable roots alongside readable siblings.

A ticket's canonical recorded identity and exact native session binding stay independent of its current directory. Users can explicitly associate a ledger-free workspace with a ticket even when the folder name has no identifier. An existing valid ledger supplies its recorded ticket identity; creation/naming association supports first publication. Where several distinct locations match, require explicit user choice rather than first-root precedence or history merging. Today's [capture schema](../../backend/internal/adapters/capture/schema.go) already scopes registrations by worktree/native identity with a ticket field; a root change must not broaden ownership across unrelated worktrees.

Every reader and writer must consume the same resolved location. Moving AL-123 from alspaces/al into alspaces/archive, or into a different user-chosen root, must preserve detail, artifacts, events, metrics, source ownership and refresh. Publication into an existing bound workspace cannot recreate its old location when a move races the write. New workspace creation is a distinct explicit operation with a selected destination or unambiguous user-defined route; a missing bound destination is not a creation instruction. Original harness sources/private capture state remain unnecessary for ordinary portable ledger reads.

OpenSpec fits this generic capability when its changes/archive directories are configured. Its setup supplies ordinary creation/naming/artifact guidance rather than a different scanner/binder/publisher. The pinned [name validator](https://github.com/Fission-AI/OpenSpec/blob/d12c2817776779e06891d68230773b3bb673b106/src/utils/change-utils.ts#L44) already accepts `openspec new change al-123-add-login`; uppercase canonical `AL-123` can remain the ticket identity as the folder changes to `2026-11-13-al-123-add-login`. The core must also handle folders whose names do not encode the ticket. Automatic allocation or mandatory naming enforcement needs a supported setup integration, not an assumed upstream hook or blindly ignored metadata field.

This repository chooses OpenSpec for future work and keeps its existing alspaces directories as ordinary configured roots. Other adopters remain free to actively use alspaces or any other layout, alone or together, without deprecation or compatibility labels. The delivery remains one end-to-end AL-84 slice including repository guidance and cutover; this study does not implement it or estimate the engineering effort. Dynamic spec browsing, phase analytics and cross-repo stores remain separate capabilities.
