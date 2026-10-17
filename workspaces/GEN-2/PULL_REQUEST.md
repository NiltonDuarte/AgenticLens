# GEN-2: Agent Single-Ticket Exclusivity and Prompt Steering Mandate

## Summary
Updates repository operational policy in `AGENTS.md` and `SKILL.md` to mandate ticket tracking, enforce single-ticket exclusivity, and define prompt steering conventions.

## Changes
- Mandate explicit ticket identifier prior to any work execution (with automatic fallback to `GEN-#` or user prompt).
- Prohibit agents from working on multiple tickets simultaneously or switching tickets mid-session.
- Define all subsequent user prompts as steering directions for the active ticket, creating new iterations as needed.
- Align multi-agent development invariants in `.opencode/skills/agentic-lens-dev/SKILL.md`.
