# docs(setup): cross-project platform setup manual and standalone skill packaging (AL-37)

## Overview
Delivers the cross-project setup manual (`docs/SETUP_MANUAL.md`) and packaging automation for the `agentic-lens-setup` onboarding skill (`scripts/package-setup-skill.sh`, `make package-setup-skill`). Delineates the AgenticLens product platform from internal development methodology, enabling external teams to onboard AgenticLens across diverse repositories and custom agent workflows without inheriting internal maintainer skills.

- Contract: [alspaces/AL-37/CONTRACT.md](alspaces/AL-37/CONTRACT.md)
- Summary: [alspaces/AL-37/SUMMARY.md](alspaces/AL-37/SUMMARY.md)

## Summary of Changes

### Documentation (`docs/SETUP_MANUAL.md`)
- Platform Overview: Single-binary zero-dependency architecture (`CGO_ENABLED=0`, embedded TypeScript SPA, pure Go SQLite driver), local fact ledgers (`agentic_metrics.json`), live observability via SQLite polling and SSE streaming, anomaly detection, and context window health states.
- Architectural Delineation: Separates the AgenticLens platform from internal development skills (`agentic-lens-dev`, `agentic-lens-po`, `agentic-lens-review`, and orchestrator-planner-coder-tester-reviewer role structure); external projects define their own workflows, personas, and branch strategies.
- Setup Pathways: Pathway A (agent-driven onboarding via `agentic-lens-setup` with user confirmation gate) and Pathway B (manual developer installation via precompiled GitHub Releases).
- Configuration Reference: Schema reference for `.agenticlens-config.yaml` (`project_name`, `ticket`, `workspaces`, `database`, `live`, `code_metrics`, `role_mappings`, `role_order`).
- Agnostic Workflow Integration: Delayed Telemetry Ingestion Mandate, invoking `./bin/agentic-lens generate` as the final turn before committing work, conventional commit classifications, and custom persona mapping.
- CLI & Web Dashboard Reference: Operational guide for CLI subcommands (`init`, `generate`, `scan`, `server`, `version`, `backfill`, `migrate`) and 5 dashboard views.
- Troubleshooting: SQLite concurrency (WAL mode, busy timeout, read-only mode), cross-platform DB paths, binary permissions, git branch inference, and workspace paths.

### Packaging & Automation (`scripts/`, `Makefile`, `.opencode/`)
- `scripts/package-setup-skill.sh`: Packages `.opencode/skills/agentic-lens-setup` into `dist/skills/agentic-lens-setup.tar.gz` and `.zip` with `SHA256SUMS`. Enforces inclusion of `SKILL.md`, `download.sh`, `bin/.gitignore`, and zero inclusion of precompiled binaries or internal repository skills.
- `Makefile`: Added `package-setup-skill` target, wired `release-package` dependency, and updated `release-clean`.
- `.opencode/skills/agentic-lens-setup/bin/.gitignore`: Created protective `.gitignore` and updated root `.gitignore` to track it.
- `scripts/test-setup.sh`: Added tests validating setup manual headers, packaging archives/contents/exclusions, and zero emojis.
