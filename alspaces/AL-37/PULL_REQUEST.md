# docs(setup): cross-project setup manual and agentic skill distribution (AL-37)

## Overview
Delivers the cross-project setup manual (`docs/SETUP_MANUAL.md`) and establishes the native agentic distribution model for the setup skill (`.opencode/skills/agentic-lens-setup/SKILL.md`). Reorganizes the manual to lead with setup pathways, elevates agnostic adoption, removes internal maintainer methodology text, documents realistic adopter role mappings, and relocates architecture to an Appendix. Removes obsolete packaging and test scripts (`scripts/package-setup-skill.sh`, `scripts/test-setup.sh`) and reverts Makefile packaging additions.

- Contract: [alspaces/AL-37/CONTRACT.md](alspaces/AL-37/CONTRACT.md)
- Summary: [alspaces/AL-37/SUMMARY.md](alspaces/AL-37/SUMMARY.md)

## Summary of Changes

### Documentation (`docs/SETUP_MANUAL.md`)
- Quick Start: Actionable setup for Pathway A (autonomous agentic setup via `SKILL.md`) and Pathway B (manual installation via GitHub Releases).
- Agnostic Adoption: Emphasizes workflow neutrality across arbitrary harnesses, topologies, custom prefixes, workspaces, and branching; excises internal maintainer methodology text and internal skills.
- Configuration Reference: Schema reference for `.agenticlens-config.yaml` (`project_name`, `ticket`, `workspaces`, `database`, `live`, `code_metrics`, `role_mappings`, `role_order`).
- Agnostic Workflow Integration: Delayed Telemetry Ingestion Mandate, nonce generation (`od -N 4 -tx1 /dev/urandom ...`), conventional commit classification, and custom persona mappings (`security-scanner: security`, `perf-analyzer: tester`, `integration-tester: tester`, `frontend-specialist: developer`, `backend-specialist: developer`).
- CLI & Dashboard Reference: Subcommands (`init`, `generate`, `scan`, `server`, `version`, `backfill`, `migrate`) and 5 dashboard views.
- Troubleshooting: SQLite concurrency (WAL mode, busy timeout, read-only mode), cross-platform database paths, executable permissions, macOS quarantine, git branch inference, and workspace resolution.
- Appendix: Details FCIS pattern, single-binary zero-dependency model, local immutable fact ledgers (`agentic_metrics.json`), live SSE streaming, anomaly detection, and context window health states.

### Agentic Skills & Repository Layer (`.opencode/`, `Makefile`, `scripts/`)
- `.opencode/skills/agentic-lens-setup/SKILL.md`: Self-contained for an autonomous agent reading via shared URL or repo link. Guides host detection, direct binary downloads from GitHub Releases, placement in `.opencode/skills/agentic-lens-setup/bin/agentic-lens`, protective `.gitignore`, `init` execution, and confirmation gate.
- `Makefile`: Reverted packaging additions: removed `package-setup-skill` from `.PHONY` and `help`, deleted target block, and removed dependency from `release-package`.
- `scripts/`: Deleted obsolete scripts (`package-setup-skill.sh`, `test-setup.sh`).
