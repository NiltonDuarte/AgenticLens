# FIX-9 Iteration 02 Plan: Codex-Native Persona Wrappers

## Steering Decision

The user selected Codex's native project agent directory. Create the four Codex TOML definitions only under `.codex/agents/`. Do not add canonical TOML copies under `.agents/agents/` and do not rely on symlinks or undocumented discovery behavior.

## Goal

Register `planner`, `coder`, `tester`, and `reviewer` as stable Codex custom-agent names while keeping `.agents/agent/<role>.md` as the shared persona instruction source.

## Implementation Tasks

1. Add `.codex/agents/planner.toml`, `.codex/agents/coder.toml`, `.codex/agents/tester.toml`, and `.codex/agents/reviewer.toml`.
2. Set each TOML agent's `name` to its exact canonical persona name and provide the required Codex `description` and `developer_instructions` fields.
3. In each `developer_instructions`, direct the agent to read and follow the matching shared persona file under `.agents/agent/`. The wrapper should make this source-of-truth relationship explicit and preserve the role's existing purpose.
4. Do not copy OpenCode frontmatter such as `mode` or `permission` into TOML as though Codex enforces it. Codex custom agents inherit the parent session's permission mode and sandbox; instructions may request read-only behavior but cannot establish per-agent permission boundaries.
5. Leave `.agenticlens-config.yaml` role mappings and `backend/internal/adapters/codex/rollout.go` unchanged unless verification shows Codex does not record the canonical name in `agent_role` or `agent_path`.

## Verification

- Parse or otherwise validate each TOML definition and confirm the four required fields are present.
- Confirm Codex discovers all four names from `.codex/agents/` and that each wrapper points to the matching `.agents/agent/<role>.md` persona.
- Inspect a rollout for each role, if the active Codex environment permits invocation, and verify the recorded `agent_role` or `agent_path` basename is canonical.
- Confirm the existing `readSessionMetadata` path in `backend/internal/adapters/codex/rollout.go` yields the expected canonical persona role without adapter changes.
- Run relevant repository checks after implementation, following the repository's verification instructions.

## Delivery Notes

Codex's custom agent definitions are harness-specific registration wrappers; shared persona behavior remains in `.agents/agent/`. This avoids a second persona-content copy while respecting Codex's documented discovery path. Complete FIX-9 tracking and workspace deliverables at delivery time as required by `AGENTS.md`.
