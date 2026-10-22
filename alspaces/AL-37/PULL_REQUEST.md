# docs(setup): cross-project setup manual and agentic skill distribution (AL-37)

## Overview
Delivers the platform setup manual (`docs/SETUP_MANUAL.md`) and refines the autonomous setup skill (`.opencode/skills/agentic-lens-setup/SKILL.md`) under the native agentic distribution model. Documents honest, exclusive coupling to OpenCode and its local SQLite database schema (`~/.local/share/opencode/opencode.db`), removing false claims of universal harness compatibility (Claude Code, Cursor, Aider, Copilot). Clarifies that engineering methodology within OpenCode is agnostic (custom topologies, arbitrary ticket taxonomies, flexible workspaces, and custom persona role mappings). Enforces the literal nonce generation mandate from FIX-5 across all documentation to prevent SQLite argument capture failures. Reverts obsolete Makefile packaging targets and deletes temporary packaging scripts.

- Contract: [alspaces/AL-37/CONTRACT.md](alspaces/AL-37/CONTRACT.md)
- Summary: [alspaces/AL-37/SUMMARY.md](alspaces/AL-37/SUMMARY.md)

## Summary of Changes

### Documentation (`docs/SETUP_MANUAL.md`)
- Quick Start: Actionable onboarding pathways for Pathway A (autonomous setup via `SKILL.md`) and Pathway B (manual installation via GitHub Releases). Removed `CLAUDE.md`.
- Agnostic Adoption for OpenCode: Documents honest coupling to OpenCode SQLite internals (`session`, `message`, `part` tables). Excised claims of universal harness compatibility and third-party harness mentions. Clarifies intra-OpenCode methodology neutrality across topologies, ticket prefixes, workspaces, and persona role mappings.
- Configuration Reference: Schema reference for `.agenticlens-config.yaml` (`project_name`, `ticket`, `workspaces`, `database`, `live`, `code_metrics`, `role_mappings`, `role_order`).
- OpenCode Agent Workflow Integration: Delayed Telemetry Ingestion Mandate, conventional commit classification, and literal nonce generation (`--nonce <GENERATED_NONCE>`). Explains FIX-5 SQLite argument capture timing.
- CLI & Dashboard Reference: Subcommand reference (`init`, `generate`, `scan`, `server`, `version`, `backfill`, `migrate`) and 5 dashboard views.
- Troubleshooting & Appendix: SQLite concurrency, cross-platform paths, permissions, FCIS pattern, single binary architecture, immutable fact ledgers, SSE streaming, anomaly detection, and context window health states.

### Skills & Repository Layer (`.opencode/`, `Makefile`, `scripts/`)
- `.opencode/skills/agentic-lens-setup/SKILL.md`: Clarified exclusive OpenCode support and SQLite requirements. Removed `CLAUDE.md`. Mandated literal nonce generation without shell variables. Guides host detection, binary downloads from GitHub Releases, placement in `.opencode/skills/agentic-lens-setup/bin/agentic-lens`, protective `.gitignore`, and confirmation gate.
- `Makefile`: Removed `package-setup-skill` target and help entries; decoupled `release-package`.
- `scripts/`: Deleted `package-setup-skill.sh` and `test-setup.sh`.
