# FIX-9 Iteration 01 Plan: Stable Codex Agent Role Names

## Goal

Provide Codex-discoverable custom agents for the canonical `planner`, `coder`, `tester`, and `reviewer` roles while keeping the `.agents` persona material as the cross-harness source of role instructions.

## Grounded Findings

- OpenCode persona definitions and constraints are in `.agents/agent/planner.md`, `.agents/agent/coder.md`, `.agents/agent/tester.md`, and `.agents/agent/reviewer.md`.
- Codex's documented project custom-agent discovery path is `.codex/agents/`; custom-agent files require `name`, `description`, and `developer_instructions`. The `name` field is the identity Codex uses when spawning or referring to an agent.
- Codex reads `AGENTS.md` for project guidance and `.agents/skills` for skills, but does not treat `.agents/agent/*.md` as a named-agent registry.
- `backend/internal/adapters/codex/rollout.go`: `decodeSessionMetadata` extracts `agent_role` and `agent_path`; `readSessionMetadata` gives explicit `agent_role` precedence, then derives the role from the basename of `agent_path`.
- `.agenticlens-config.yaml` already maps `planner`, `coder`, `tester`, and `reviewer` to their canonical roles.
- The ticket board `alspaces/plan/TICKETS.md` has no FIX-9 entry, and no `alspaces/plan/to-do/FIX-9*.md` specification exists. The feature branch is already checked out and the working tree was clean at planning start.

## Proposed Tasks

1. Resolve the Codex file-location design before implementation. Recommended: create four TOML wrappers under `.codex/agents/`, each named for its canonical role, and have its `developer_instructions` direct the agent to read and follow the corresponding `.agents/agent/<role>.md` file. This avoids duplicating the persona content while using Codex's documented discovery path.
2. If keeping TOMLs physically under `.agents/agents/` is a hard requirement, identify and implement an explicit synchronization or exposure mechanism to `.codex/agents/`. Do not assume Codex scans `.agents/agents/` automatically. Avoid undocumented symlink behavior unless confirmed for the target Codex runtime.
3. For each Codex definition, translate only the persona's role and task guidance into the TOML contract. Do not copy OpenCode YAML frontmatter fields such as `mode` or `permission` into Codex settings; Codex custom agents inherit the parent permission mode and do not define separate sandbox policies.
4. Verify by inspecting Codex rollout metadata for one invocation of each custom agent. Confirm `agent_role` or the final `agent_path` component is canonical and that the current adapter emits the existing canonical role mappings.
5. Update ticket tracking/workspace deliverables for FIX-9 in line with the repository mandate. Add an In Progress entry and a concise ticket specification if required by the maintainer's workflow, then complete `SUMMARY.md`, `PULL_REQUEST.md`, and final telemetry at delivery time.

## Verification Approach

- Validate all four TOML definitions parse and include the required Codex fields.
- Confirm Codex discovers each definition by canonical name in the selected file layout.
- Inspect generated Codex rollout metadata for all four agents and confirm existing role extraction classifies them as `planner`, `coder`, `tester`, and `reviewer`.
- Run the repository checks required for the config/adapter scope after implementation; avoid changing adapter code unless observed rollout metadata proves the stable name is not preserved.

## Open Question

The user requested TOMLs under `.agents/agents/`, but Codex auto-discovery expects `.codex/agents/`. Confirm whether Codex-native wrappers that reference the shared Markdown are acceptable, or whether FIX-9 must keep TOMLs under `.agents/agents/` and add a synchronization mechanism.
