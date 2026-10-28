# FIX-9: Normalize Codex Subagent Names to Canonical AgenticLens Roles

## Problem

Codex subagents can appear under task-specific names in AgenticLens metrics, while the project defines canonical personas as planner, coder, tester, and reviewer. The generated names create noisy role metrics.

## Goal

Configure Codex to spawn named agents using the canonical roles and keep their instructions aligned with the shared persona guidance in `.agents/agent/`.

## Scope

- Add project-scoped Codex custom-agent TOML definitions under `.codex/agents/` named `planner`, `coder`, `tester`, and `reviewer`.
- Each definition must reference and follow its corresponding `.agents/agent/<role>.md` persona file.
- Preserve existing AgenticLens role mappings and Codex telemetry extraction unless real Codex rollout metadata proves they do not preserve the canonical names.
- Respect Codex's permission model: custom agents inherit the parent session's permissions; OpenCode permission frontmatter is not enforceable Codex configuration.

## Acceptance Criteria

- Codex discovers all four custom agents by their canonical names.
- The agents follow the shared persona markdown without maintaining a second copy of the persona instructions.
- A Codex telemetry sample for each role resolves to `planner`, `coder`, `tester`, or `reviewer` in the generated metrics.
- Existing role mapping and extraction behavior remain unchanged unless an observed telemetry gap requires a focused fix.
