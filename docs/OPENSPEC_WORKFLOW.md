# Roadmap tickets and OpenSpec execution

The roadmap remains in `alspaces/plan/`. Ticket documents describe evolving product intent; `TICKETS.md` owns priority, readiness, dependencies and status. A ticket selected for work gets a linked OpenSpec change. Creating or reprioritizing a roadmap entry does not create a change. Archiving a change does not automatically complete the ticket.

## Pinned development tool

The repository pins `@fission-ai/openspec` 2.0.0 and its integrity through `tools/openspec/package-lock.json`. Node >=22.12.0 is required for this development/setup tool; AgenticLens's manager and engine remain native and need no production Node runtime.

```bash
npm ci --prefix tools/openspec
scripts/openspec.sh --version
```

The wrapper uses only that installed package and disables OpenSpec's separate anonymous command telemetry. It never installs a newer release implicitly. AgenticLens session telemetry still comes only from its selected Codex/OpenCode source adapter.

## Setup and creation

The supported project customization points are `openspec/config.yaml`, the project schemas under `openspec/schemas/`, and their templates. OpenSpec 2.0.0 labels custom schema management experimental. Project context/rules and apply/archive operation guidance are instructions for agents, not enforced authorization or completion gates.

The setup bundle installer’s explicit `--with-openspec` option installs this pinned tool/schema bundle from the selected exact AgenticLens release (or approved local source), preflighting every target and refusing differing custom files before any installation writes. It does not run npm, enable capture or generate assistant files.

After explicitly approving setup in an adopting repository, use the pinned CLI's `init --tools codex,opencode --no-animation` and install the reviewed schema bundle/configuration. Preserve custom/unrelated instructions and review generated-file conflicts before accepting updates; do not run `--force` to remove them. This repository's generated shared skills support both Codex and OpenCode. No additional AgenticLens source adapters are implied.

For this repository, create selected execution work from the project root:

```bash
node scripts/openspec-change.mjs AL-123 "add login"
scripts/openspec.sh status --change al-123-add-login --json
```

The creation helper validates the canonical ID against explicitly configured literal prefixes, derives the lowercase name, checks collision and invokes supported `openspec new change --description --schema --json`. OpenSpec owns folder creation. The helper writes no unknown metadata fields, workspace identity sidecar, binding registry or telemetry ledger. It preserves existing content on rejection. The canonical ID is subsequently supplied explicitly to first AgenticLens generation; the runtime never infers it from the folder name.

Use `--planning` only for a selected PO planning ticket's own delivery:

```bash
node scripts/openspec-change.mjs PO-123 "refine roadmap" --planning
```

`agenticlens-planning` has proposal, planning tasks and delivery artifacts, with no developer contract or application implementation authority. PO can edit roadmap tickets without opening a change; creating a PO task's home never creates downstream development changes. A separate development session uses the development schema and must obtain explicit scope approval before coding.

## One execution artifact set

| Artifact | Owner and purpose |
| --- | --- |
| `proposal.md` | Selected outcome and link to the separate roadmap ticket. |
| `specs/<capability>/spec.md` | Selected observable acceptance for development work. |
| `design.md` | Technical contract, writable-path ownership, approval and material steering for development. |
| `tasks.md` | Current implementation/planning steps and progress, appropriate to the schema. |
| `delivery.md` | Current outcome, tested revision/diff, actual checks, attributed review, limits and PR context. |
| `evidence/iterations/<NN>/` | Review/test evidence under canonical attribution and retention rules. |
| `agentic_metrics.json` | Final source-backed session ledger. |

Do not keep parallel editable CONTRACT/PLAN/SUMMARY/PULL_REQUEST copies after transferring execution work into the change. Preserve prior approval, steering, verified evidence and navigation during a transfer; the roadmap ticket remains in place. Existing historical audits and unrelated active owners are not migration targets.

OpenSpec artifact readiness and checked tasks are not user approval or verified delivery. Follow [AGENTS.md](../AGENTS.md) and the owning AgenticLens skill. Generated upstream skills use the `openspec` command name; in this checkout execute their CLI steps through `scripts/openspec.sh` so they use the pinned package. Do not fall back to an older global CLI or silently ignore unsupported fields/commands. A later roadmap edit does not silently change approved execution scope. Record its source, acceptance delta, preserved guarantees and verification impact, obtaining renewed approval only for a material scope change.

## Ledger publication and archive

Configure every chosen search directory explicitly. This repository selects `alspaces`, `openspec/changes` and `openspec/changes/archive` with depth 1 and budget 300; the generic configuration writer initializes depth 0/budget 300. Root entries have only a path and discovery policy, with optional configured artifact overrides. No scalar `workspaces.dir`, root ID, minimum depth or runtime `--workspaces` override is retained.

AgenticLens ignores a planning-only folder. First generation needs the canonical ticket ID, actual source/session selection, explicit conventional type and ledger destination. The following is a command shape, not permission to generate telemetry early:

```bash
agentic-lens generate --source codex --session-id <ACTUAL_ROOT_ID> --ticket-id AL-123 --ticket-type feat --json-output openspec/changes/al-123-add-login/agentic_metrics.json --append openspec/changes/al-123-add-login/delivery.md
```

Final session telemetry uses a literal unique nonce at the final extraction step, with root/child/source completeness checked by the orchestrator. Unknown costs and unavailable/partial owners remain explicit. Writers use the selected local `agenticlens.state_dir`, initially `~/.agenticlens/state`, for private journals/backups and locks without requiring capture installation or enablement. Completed ledger reads require no private state. A transient adjacent pending-publication indicator survives a move/copy; absent private recovery data leaves it visibly incomplete, never accepted as a mixed set.

Once verification and consolidated review pass, use the pinned schema's status/instructions and normal archive command:

```bash
scripts/openspec.sh validate al-123-add-login --strict
scripts/openspec.sh instructions archive --change al-123-add-login --json
scripts/openspec.sh archive al-123-add-login
```

Inspect task/approval/delivery evidence before archive. The static archive command itself does not enforce repository review or human authorization. After the move, refresh the actual resolved archived ledger rather than rebuilding an active folder path. Folder names never supply runtime ticket identity, type, completion or ownership.

## Adoption and rollback

Review Git changes for configuration, guidance and selected execution artifacts. Validate schema and links, verify readers/writers against the configured locations, then transfer only the active ticket's owned execution evidence. Coordinate any other active owner before moving their writable files. Roll back matching configuration/guidance and executable together if adoption fails; do not add a repository-migration framework, rewrite historical ledgers or move the whole roadmap.

Run `python3 scripts/tests/openspec_workflow_test.py` after installing the pinned toolchain for durable isolated creation/schema/archive checks. Missing dependencies fail explicitly; there are no silent skips.

The pinned package has been exercised in isolated native CLI probes: project init for Codex/OpenCode, schema validation, creation, instructions/status, strict change validation and archive with delivery evidence retained and delta specs synchronized. Those probes establish OpenSpec's CLI/schema surface, not AgenticLens capture compatibility. Native Codex/OpenCode source extraction, archive-safe publication and standalone dashboard evidence must be recorded against the actual delivered AgenticLens candidate.
