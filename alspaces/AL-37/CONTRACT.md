# Contract: AL-37 Cross-Project Platform Setup Manual and Telemetry Integration Guide

## 1. Problem Statement

External engineering teams, platform architects, and autonomous coding agents require comprehensive, authoritative documentation to onboard and operate AgenticLens across diverse software repositories.

In Iteration 01, AL-37 attempted to package the setup skill into compressed release archives (`tar.gz` and `zip`) via custom shell packaging scripts and Makefile targets. This approach introduced unnecessary packaging overhead and contradicted the modern agentic distribution model. In agentic workflows, autonomous agents consume skills natively via Markdown instruction manifests (`SKILL.md`) read directly through shared repository URLs or cloned paths. Packaging archives of markdown files is redundant and counter-productive.

Furthermore, the documentation manual requires restructuring:
1. Adopters and autonomous agents need immediate actionable setup steps at the very top of the manual (Section 1 Quick Start covering Agent-Driven and Manual Developer pathways).
2. Adopting repositories define their own ways of working, agent personas, review procedures, and branching strategies. Adoption guidelines (Section 2 Agnostic Adoption) must be completely decoupled from internal maintainer methodology text, focusing purely on cross-project adoption.
3. Realistic adopter role mappings must be documented (such as `security-scanner: security`, `perf-analyzer: tester`, `integration-tester: tester`, `frontend-specialist: developer`, and `backend-specialist: developer`).
4. Platform architecture, value propositions, and context health details belong logically in an Appendix rather than blocking early adoption reading.
5. Obsolete packaging scripts (`scripts/package-setup-skill.sh`, `scripts/test-setup.sh`) and Makefile packaging targets must be removed and reverted.
6. The onboarding skill manifest (`.opencode/skills/agentic-lens-setup/SKILL.md`) must be fully self-contained for an autonomous agent reading via shared URL or repo link.

Ticket AL-37 Iteration 02 delivers the restructured `docs/SETUP_MANUAL.md`, enhances `.opencode/skills/agentic-lens-setup/SKILL.md`, deletes the obsolete packaging and test scripts, reverts Makefile packaging additions, and aligns the entire repository with the agentic distribution model.

---

## 2. Functional Core, Imperative Shell (FCIS) and Architectural Invariants

All deliverables must strictly satisfy `TECHNICAL_GUIDELINES.md` and repository architectural invariants:

### Functional Core, Imperative Shell (FCIS) Invariants
1. Pure Functional Core Isolation:
   - Go domain models, mathematical rollups, token aggregation, anomaly thresholds, and context window health logic in `backend/internal/domain/` remain pure, immutable value-semantic functions with zero external dependencies, zero disk I/O, zero network access, and zero clocks.
   - Documentation and skill instructions strictly honor this boundary: telemetry ledgers (`agentic_metrics.json`) are immutable factual records generated deterministically from SQLite execution traces and git diff analysis.
2. Imperative Shell Discipline:
   - Configuration file parsing and generation (`backend/internal/adapters/config/`), SQLite querying (`backend/internal/adapters/sqlite/`), git analysis (`backend/internal/adapters/git/`), HTTP serving (`backend/internal/adapters/http/`), and CLI command routing (`backend/cmd/agentic-lens/`) reside exclusively within the imperative shell.
   - Documentation, skill manifests, Makefiles, and repository operations belong to the imperative shell and must never introduce runtime dependencies to the pure domain core.

### Single Binary Distribution and Runtime Independence
1. Zero External Runtime Dependency:
   - The compiled Go binary is completely self-contained. The TypeScript single-page application dashboard is embedded into the Go binary via Go `embed.FS`.
   - Adopters require zero Node.js runtime, zero npm dependencies, and zero external database servers in production.
   - SQLite connectivity uses a pure Go driver (`modernc.org/sqlite`) compiled with `CGO_ENABLED=0`, eliminating C compiler toolchain requirements on target host machines.
2. Precompiled Release Consumption:
   - External projects must never be instructed or required to compile AgenticLens from Go or TypeScript source code. Setup pathways strictly consume official precompiled release binaries from GitHub Releases (`https://github.com/NiltonDuarte/AgenticLens/releases`).

### Agentic Distribution Model
1. Markdown Skill Manifest Distribution:
   - Skills are distributed natively as markdown files accessed directly via shared URLs (e.g. raw GitHub URLs) or repository links.
   - No archive packaging (`tar.gz`, `zip`) or packaging scripts are maintained for skills.
   - Precompiled binary executables must never be checked into git; binaries are retrieved dynamically on target host machines by the autonomous agent or developer.
2. Self-Contained Autonomous Execution:
   - `.opencode/skills/agentic-lens-setup/SKILL.md` must contain all necessary detection logic, download URLs, parameter prompts, configuration generation commands, policy recommendations, and the mandatory confirmation gate for an agent operating from a URL or repo link.

### Zero Emojis Mandate
Zero emojis are permitted across all documentation files, skill manifests, Makefiles, Go code, TypeScript code, commit messages, pull requests, CLI outputs, and workspace artifacts.

---

## 3. Scope Boundaries: Documentation, Skill Enhancement, and Cleanup

### In Scope
1. Documentation Authoring and Restructuring (`docs/SETUP_MANUAL.md`):
   - Section 1: Quick Start (Pathway A: Agent-Driven Setup via `SKILL.md`, Pathway B: Manual Developer Installation via GitHub Releases).
   - Section 2: Agnostic Adoption (purely methodology-agnostic cross-project adoption, zero internal maintainer methodology text).
   - Section 3: Configuration Reference (exhaustive schema reference for `.agenticlens-config.yaml` covering `project_name`, `ticket`, `workspaces`, `database`, `live`, `code_metrics`, `role_mappings`, and `role_order`).
   - Section 4: Agnostic Agent Workflow Integration Guide (delayed telemetry ingestion mandate, random execution nonce generation, conventional commit classification, and realistic adopter role mappings including `security-scanner: security`, `perf-analyzer: tester`, `integration-tester: tester`, `frontend-specialist: developer`, and `backend-specialist: developer`).
   - Section 5: CLI and Web Dashboard Reference (full command-line reference for `init`, `generate`, `scan` / root flags, `server`, `version`, `backfill`, and `migrate`, alongside Fleet Overview, Live Observability, Ticket Detail, Comparisons, and Metric Progression).
   - Section 6: Operational Troubleshooting (SQLite concurrency, platform database paths, binary permissions, macOS quarantine, git branch inference, and workspace path resolution).
   - Appendix: Platform Architecture (FCIS architecture, single binary zero-dependency model, local `agentic_metrics.json` ledgers, live SSE streaming, anomaly detection, and context window health states).
2. Autonomous Skill Manifest Enhancement (`.opencode/skills/agentic-lens-setup/SKILL.md`):
   - Update `SKILL.md` to be completely self-contained for an autonomous agent reading via shared URL or repository link.
   - Include direct architecture detection and release download commands from GitHub Releases (`https://github.com/NiltonDuarte/AgenticLens/releases`).
   - Include configuration interviewing, scaffolding, feature explanation, and the mandatory user confirmation gate before touching repository files.
3. Script Cleanup:
   - Delete `scripts/package-setup-skill.sh`.
   - Delete `scripts/test-setup.sh`.
4. Makefile Reversion:
   - Remove `package-setup-skill` from `.PHONY` and `help`.
   - Remove `package-setup-skill` target definition.
   - Revert `release-package` target so it does not depend on `package-setup-skill`.

### Out of Scope
1. Altering Go domain logic, mathematical aggregation formulas, context window health rules, or anomaly detection thresholds in `backend/internal/domain/`.
2. Modifying TypeScript frontend UI components, views, or routing in `web/src/`.
3. Altering OpenCode SQLite schema parsing or SSE live event streaming formats.
4. Changing internal repository skills (`agentic-lens-dev`, `agentic-lens-po`, `agentic-lens-review`).

---

## 4. Architectural Contracts and Deliverables Specification

### Documentation Contract: docs/SETUP_MANUAL.md
The setup manual must be authored as a clean, highly structured Markdown file at `docs/SETUP_MANUAL.md` with the following exact top-level sections:
1. Section 1: Quick Start
2. Section 2: Agnostic Adoption
3. Section 3: Configuration Reference (.agenticlens-config.yaml)
4. Section 4: Agnostic Agent Workflow Integration Guide
5. Section 5: CLI and Web Dashboard Reference
6. Section 6: Operational Troubleshooting
7. Appendix: Platform Architecture

### Skill Manifest Contract: .opencode/skills/agentic-lens-setup/SKILL.md
1. Self-contained instructions enabling an agent accessing only the markdown file (via URL or path) to bootstrap AgenticLens.
2. Direct binary download commands using `curl`/`wget` targeting official precompiled releases from `https://github.com/NiltonDuarte/AgenticLens/releases`.
3. Co-located binary placement at `.opencode/skills/agentic-lens-setup/bin/agentic-lens` or `./bin/agentic-lens` with protective `.gitignore`.
4. Non-interactive and interactive configuration generation via `init`.
5. Strict enforcement of the Mandatory User Confirmation Gate before modifying any target repository instructions (`AGENTS.md`, `CLAUDE.md`, etc.).

### Repository Cleanliness Contract
1. `scripts/package-setup-skill.sh` must be completely removed from the repository.
2. `scripts/test-setup.sh` must be completely removed from the repository.
3. `Makefile` must contain zero references to `package-setup-skill`.
4. Standard quality gates (`make check`, `make test`, `make build`) must succeed cleanly.
