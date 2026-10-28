# FIX-9 Contract: Stable Codex Agent Role Names

## Problem

Codex subagent telemetry can carry a generated or task-specific identity instead of one of AgenticLens' canonical personas. That creates noisy role names in platform metrics. OpenCode personas are defined in `.agents/agent/`, while Codex discovers its own named custom-agent definitions from `.codex/agents/`.

The user wants the project to remain cross-harness and specifically requested Codex TOML agent definitions inside `.agents/agents/`. Official Codex documentation describes `.codex/agents/` as the discovery location for project-scoped custom agents; TOMLs stored only in `.agents/agents/` are not automatically discovered by Codex.

## Intended Outcome

Make Codex delegate using stable canonical names `planner`, `coder`, `tester`, and `reviewer`, so Codex session metadata can resolve to those personas and AgenticLens metrics remain consistent across harnesses. Keep the existing `.agents/agent/*.md` persona guidance as the cross-harness instruction source where feasible.

## Existing System Context

- `.agents/agent/planner.md`, `.agents/agent/coder.md`, `.agents/agent/tester.md`, and `.agents/agent/reviewer.md` define the current OpenCode personas and contain OpenCode-specific frontmatter.
- `.agenticlens-config.yaml` already maps the four canonical persona names to themselves and defines their metric ordering.
- `backend/internal/adapters/codex/rollout.go`'s `decodeSessionMetadata` reads Codex `agent_role` and `agent_path`; `readSessionMetadata` prefers `agent_role`, then uses the final `agent_path` component as the role.
- Codex custom agent TOMLs require `name`, `description`, and `developer_instructions`; each custom agent is a separate file under Codex's documented project directory.

## Scope

- Define the Codex-side named agents for the four canonical AgenticLens personas.
- Preserve or reference the shared persona instructions rather than silently letting Codex-specific text drift from `.agents/agent/*.md`.
- Place Codex-discoverable definitions in the location Codex supports, or provide an explicit supported synchronization mechanism if canonical TOMLs must be kept under `.agents/agents/`.
- Document the constraint that Codex custom agents inherit the parent session's permissions; the OpenCode permission frontmatter cannot be copied as an enforceable Codex per-agent sandbox policy.

## Out of Scope

- Changes to AgenticLens telemetry schemas, role mappings, or Codex rollout parsing unless planning evidence shows stable role names alone are insufficient.
- Changes to OpenCode agent behavior or the persona responsibilities themselves.
- Changes to project-wide sandbox or approval settings.

## Acceptance Criteria

1. Codex can discover and spawn the four agents using the stable names `planner`, `coder`, `tester`, and `reviewer`.
2. Codex agent instructions stay aligned with the corresponding shared persona markdown, without copying OpenCode-only permission metadata as if Codex enforced it.
3. The chosen file layout is compatible with Codex's documented agent discovery behavior and honors the user's cross-harness organization intent as far as the tools allow.
4. Existing AgenticLens role mapping and Codex role extraction require no change if the canonical names are present in `agent_role` or `agent_path`.

## Design Decision Needed

Codex does not automatically discover `.agents/agents/*.toml`. Choose between native `.codex/agents/*.toml` wrappers that direct each agent to its matching `.agents/agent/*.md` persona (recommended for avoiding persona duplication), or canonical TOMLs under `.agents/agents/` plus an explicit mechanism that exposes those definitions under `.codex/agents/`. Do not deliver TOMLs only in `.agents/agents/` while claiming Codex will auto-discover them.
