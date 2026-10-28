# fix(codex): normalize subagent role names (FIX-9)

## Summary

- Register canonical `planner`, `coder`, `tester`, and `reviewer` Codex agents under `.codex/agents/`.
- Reuse each corresponding `.agents/agent/*.md` persona as the instruction source.
- Document Codex's inherited parent permission and sandbox model.

## Context

- Contract: `alspaces/FIX-9/CONTRACT.md`
- Summary: `alspaces/FIX-9/SUMMARY.md`
- Plan: `alspaces/FIX-9/iterations/02/PLAN.md`

## Verification

TOML parsing and `git diff --check` passed. `make build` completed successfully. Reviewer approved with no findings. Automated tests were not run.
