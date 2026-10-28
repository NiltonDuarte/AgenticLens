# FIX-9 Summary: Stable Codex Agent Role Names

## Configuration

- Added Codex custom-agent definitions for `planner`, `coder`, `tester`, and `reviewer` under `.codex/agents/`.
- Each wrapper points to the shared persona guidance under `.agents/agent/` rather than duplicating its content.
- Documented that Codex custom agents inherit the parent session's sandbox and permissions.

## Domain, Adapters, and Presentation

- No application domain, telemetry adapter, or presentation code changed. Existing Codex role extraction already uses canonical `agent_role` or `agent_path` identity.

## Verification

- Coder reported successful TOML parsing and `git diff --check`.
- Reviewer verdict: `APPROVED`, no findings.
- Automated test suites were not run.
