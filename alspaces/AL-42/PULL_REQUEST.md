# fix(domain): unify tool execution error classification (AL-42)

## Summary

Unify offline and live tool execution failure classification through a pure domain classifier. Keep semantic output content and benign grep/glob no-match results from inflating execution failure metrics, while preserving hard failure detection, severity escalation, and sanitized error messages.

## Scope

- Domain classifier and failure regression tests
- OpenCode offline extraction and live service parity
- SQLite-backed offline/live parity test

See [CONTRACT.md](CONTRACT.md) and [SUMMARY.md](SUMMARY.md).
