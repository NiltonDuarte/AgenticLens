# feat(opencode): support OpenCode v1 and v2 concurrently (AL-58)

## Summary

- Detect supported OpenCode storage families from complete SQLite table and column sets, with actionable errors for unsupported, partial, and mixed schemas.
- Extract v2 sessions, messages, tools, and telemetry into the existing normalized facts while preserving v1 behavior and read-only database access.
- Separate shared dispatch and detection from v1/v2-specific extraction and resolution files; versioned filenames, symbols, diagnostics, and tests identify both family and responsibility.
- Remove redundant v1 schema metadata queries and run synthetic schema rejection tests against parallel, in-memory SQLite databases.
- Add isolated per-version and combined compatibility workflows, runtime-derived fixtures, and setup/architecture documentation.
- Resolve Codex session users consistently across generation and backfill, treating the exact `unknown` placeholder as missing and falling back to a known same-session or Git identity.

## Context

- Technical contract: `alspaces/AL-58/CONTRACT.md`
- Work summary: `alspaces/AL-58/SUMMARY.md`
- Runtime evidence: `alspaces/AL-58/evidence/RUNTIME_EVIDENCE.md`

Schema detection identifies a storage family rather than an exact OpenCode binary version. The verified evidence targets are v1.19.5 and v2.2.0; broader migration and pricing behavior remains documented as unverified.
