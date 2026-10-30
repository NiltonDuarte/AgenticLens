# feat(config): add local multi-harness source selection (AL-59)

## Summary

- Add optional local telemetry source configuration for OpenCode SQLite and Codex rollout home, with explicit source and path precedence and clear failures without cross-harness fallback.
- Extend `agentic-lens init` to generate shared and local config files with sanitized defaults and safe dry-run/overwrite behavior.
- Update setup guidance, documentation, ignored local state, and focused tests.

See [CONTRACT.md](CONTRACT.md) and [SUMMARY.md](SUMMARY.md) for acceptance boundaries and implementation details.
