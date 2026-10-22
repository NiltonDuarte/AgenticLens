# Contract: AL-37 Cross-Project Platform Setup Manual and Telemetry Integration Guide

## 1. Problem Statement

External engineering teams, platform architects, and autonomous coding agents require comprehensive, authoritative documentation to onboard and operate AgenticLens across software repositories.

In Iteration 01, AL-37 attempted to package the setup skill into compressed release archives (`tar.gz` and `zip`) via custom shell packaging scripts and Makefile targets. This approach introduced unnecessary packaging overhead and contradicted the modern agentic distribution model. In agentic workflows, autonomous agents consume skills natively via Markdown instruction manifests (`SKILL.md`) read directly through shared repository URLs or cloned paths. Packaging archives of markdown files is redundant and counter-productive.

In Iteration 02, the manual was restructured to lead with practical setup pathways (Pathway A via `SKILL.md` and Pathway B via precompiled GitHub Releases), elevated agnostic external adoption, eliminated internal maintainer methodology text, documented realistic adopter persona mappings (such as `security-scanner`, `perf-analyzer`, `frontend-specialist`), and relocated platform architecture theory to an Appendix. Obsolete packaging scripts (`scripts/package-setup-skill.sh`, `scripts/test-setup.sh`) and Makefile packaging targets were deleted.

In Iteration 03, user steering feedback identified a critical honesty defect across the documentation: `docs/SETUP_MANUAL.md` and `.opencode/skills/agentic-lens-setup/SKILL.md` contained false claims of universal, harness-agnostic compatibility (claiming support for Claude Code, Cursor, Aider, GitHub Copilot, and generic agent harnesses). In reality, AgenticLens is deeply coupled to OpenCode internals. Its telemetry extraction adapter (`backend/internal/adapters/opencode`), live SSE streaming service (`/api/live/stream`), and CLI `generate` command query OpenCode's local SQLite database (`~/.local/share/opencode/opencode.db`) and its specific schema tables (`session`, `message`, `part`). AgenticLens is completely incompatible with anything other than OpenCode and cannot function without an active OpenCode installation and its SQLite database.

Furthermore, upstream `main` merged FIX-5, which updated the telemetry generation contract: OpenCode records tool inputs into SQLite before subshell expansion occurs. Using shell variable substitutions (`$NONCE`, `$(...)`) results in the unexpanded token being stored in the database, causing exact nonce matching to fail. Agents must generate a literal nonce token directly and inject the concrete token into `--nonce <GENERATED_NONCE>`.

Documentation must be completely honest and transparent:
1. State clearly that AgenticLens is an observability and telemetry platform built exclusively for OpenCode, requiring OpenCode and its SQLite database.
2. Remove all false claims of universal harness compatibility and all references to Claude Code, Cursor, Aider, GitHub Copilot, or generic external agent harnesses across `docs/SETUP_MANUAL.md` and `.opencode/skills/agentic-lens-setup/SKILL.md`.
3. Clarify what IS agnostic: Within OpenCode, AgenticLens is agnostic to development methodology. Teams can use single-agent loops, custom subagents, custom ticket prefixes, custom workspace paths, and custom branching models without being forced into this repository's maintainer personas.
4. Enforce the literal nonce mandate from FIX-5: The CLI `--nonce <GENERATED_NONCE>` argument requires the agent to generate a literal nonce string directly without shell variable expansion (`$NONCE`), because OpenCode records tool inputs into SQLite before subshell expansion occurs, causing database lookups for the expanded token to fail.

---

## 2. Functional Core, Imperative Shell (FCIS) and Architectural Invariants

All deliverables must strictly satisfy `TECHNICAL_GUIDELINES.md` and repository architectural invariants:

### Functional Core, Imperative Shell (FCIS) Invariants
1. Pure Functional Core Isolation:
   - Go domain models, mathematical rollups, token aggregation, anomaly thresholds, and context window health logic in `backend/internal/domain/` remain pure, immutable value-semantic functions with zero external dependencies, zero disk I/O, zero network access, and zero clocks.
   - Documentation and skill instructions strictly honor this boundary: telemetry ledgers (`agentic_metrics.json`) are immutable factual records generated deterministically from SQLite execution traces and git diff analysis.
2. Imperative Shell Discipline:
   - Configuration file parsing and generation (`backend/internal/adapters/config/`), SQLite querying (`backend/internal/adapters/opencode/` and `backend/internal/adapters/live/`), git analysis (`backend/internal/adapters/git/`), HTTP serving (`backend/internal/adapters/http/`), and CLI command routing (`backend/cmd/agentic-lens/`) reside exclusively within the imperative shell.
   - Documentation, skill manifests, Makefiles, and repository operations belong to the imperative shell and must never introduce runtime dependencies to the pure domain core.

### Single Binary Distribution and Runtime Independence
1. Zero External Runtime Dependency:
   - The compiled Go binary is completely self-contained. The TypeScript single-page application dashboard is embedded into the Go binary via Go `embed.FS`.
   - Adopters require zero Node.js runtime, zero npm dependencies, and zero external database servers in production.
   - SQLite connectivity uses a pure Go driver (`modernc.org/sqlite`) compiled with `CGO_ENABLED=0`, eliminating C compiler toolchain requirements on target host machines.
2. Precompiled Release Consumption:
   - External projects must never be instructed or required to compile AgenticLens from Go or TypeScript source code. Setup pathways strictly consume official precompiled release binaries from GitHub Releases (`https://github.com/NiltonDuarte/AgenticLens/releases`).

### OpenCode Coupling and SQLite Invariants
1. OpenCode Exclusivity:
   - Telemetry ingestion (`backend/internal/adapters/opencode/extractor.go`, `resolver.go`, `client.go`), live SSE streaming (`backend/internal/adapters/live/sqlite.go`), and the CLI `generate` command are architected exclusively to parse OpenCode's internal SQLite database schema.
   - AgenticLens does not support, integrate with, or claim compatibility with any alternative agent harnesses or tools.
2. Literal Nonce Invariant (FIX-5):
   - The execution nonce passed to `--nonce` must be a literal token generated directly by the agent (e.g. an 8 to 16 character alphanumeric or hex string).
   - Shell variable expansions (`$NONCE`) or command substitutions (`$(...)`) are strictly prohibited in documentation and agent instructions because OpenCode records tool inputs into SQLite prior to subshell execution.

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
   - Introduction: Explicitly state that AgenticLens is an observability and telemetry platform built exclusively for OpenCode, requiring OpenCode's local SQLite database.
   - Section 1: Quick Start (Pathway A: Agent-Driven Setup via `SKILL.md` within OpenCode, Pathway B: Manual Developer Installation via GitHub Releases). Remove references to non-OpenCode instruction files such as `CLAUDE.md`.
   - Section 2: Agnostic Adoption within OpenCode. Completely eliminate false claims of universal harness compatibility and all mentions of Claude Code, Cursor, Aider, GitHub Copilot, and generic agent harnesses. Honestly state that AgenticLens is coupled to OpenCode internals. Clarify what IS agnostic: engineering methodology within OpenCode (single-agent vs multi-agent, custom personas, custom ticket prefixes, custom workspace paths, custom git branching models, decoupled from this repository's maintainer personas).
   - Section 3: Configuration Reference (schema reference for `.agenticlens-config.yaml` covering `project_name`, `ticket`, `workspaces`, `database`, `live`, `code_metrics`, `role_mappings`, and `role_order`, reinforcing `database.path` as the OpenCode SQLite database).
   - Section 4: OpenCode Agent Workflow Integration Guide. Enforce the literal nonce mandate from FIX-5: replace shell variable commands (`NONCE=$(...)` and `--nonce "$NONCE"`) with literal token generation (`--nonce <GENERATED_NONCE>`) and explain why shell expansion fails in OpenCode. Remove `CLAUDE.md` from the agent guidelines policy block.
   - Section 5: CLI and Web Dashboard Reference (command-line reference for `init`, `generate`, `scan` / root flags, `server`, `version`, `backfill`, and `migrate`, alongside Fleet Overview, Live Observability, Ticket Detail, Comparisons, and Metric Progression).
   - Section 6: Operational Troubleshooting (OpenCode SQLite concurrency, platform database paths, binary permissions, macOS quarantine, git branch inference, and workspace path resolution).
   - Appendix: Platform Architecture (FCIS architecture, single binary zero-dependency model, local `agentic_metrics.json` ledgers, live SSE streaming, anomaly detection, and context window health states).
2. Autonomous Skill Manifest Enhancement (`.opencode/skills/agentic-lens-setup/SKILL.md`):
   - Clarify exclusive OpenCode support in frontmatter and description.
   - Remove references to `CLAUDE.md` from operational mandates and Step 4 instructions inspection.
   - Ensure literal nonce generation is consistently enforced without shell variables across all examples and policy templates.
3. Quality Assurance and Auditing:
   - Perform repository-wide search to verify zero occurrences of Claude Code, Cursor, Aider, GitHub Copilot, or `CLAUDE.md` in `docs/SETUP_MANUAL.md` and `.opencode/skills/agentic-lens-setup/SKILL.md`.
   - Verify zero occurrences of shell variable substitution (`$NONCE`, `$(...)`) in documented `generate` commands.
   - Verify zero emojis across all deliverables.
   - Run `make check`, `make test`, and `make build`.

### Out of Scope
1. Modifying Go domain logic, mathematical aggregation formulas, context window health rules, or anomaly detection thresholds in `backend/internal/domain/`.
2. Modifying Go adapters or CLI implementation in `backend/internal/adapters/` and `backend/cmd/agentic-lens/`.
3. Modifying TypeScript frontend UI components, views, or routing in `web/src/`.
4. Adding adapters, drivers, or compatibility layers for non-OpenCode harnesses.
5. Altering internal repository skills (`agentic-lens-dev`, `agentic-lens-po`, `agentic-lens-review`).

---

## 4. Architectural Contracts and Deliverables Specification

### Documentation Contract: docs/SETUP_MANUAL.md
The setup manual must be maintained as a clean, highly structured Markdown file at `docs/SETUP_MANUAL.md` with the following exact top-level sections:
1. Section 1: Quick Start and Setup Pathways
2. Section 2: Agnostic Adoption for OpenCode Projects
3. Section 3: Configuration Reference (.agenticlens-config.yaml)
4. Section 4: OpenCode Agent Workflow Integration Guide
5. Section 5: CLI Commands and Embedded Web Dashboard Reference
6. Section 6: Operational Troubleshooting and Best Practices
7. Appendix: Platform Architecture and Core Concepts

Section 2 must explicitly disclaim compatibility with other harnesses and document that OpenCode SQLite is strictly required, while highlighting intra-OpenCode methodology flexibility.
Section 4 must document literal nonce generation and forbid shell variable expansion.

### Skill Manifest Contract: .opencode/skills/agentic-lens-setup/SKILL.md
1. Self-contained instructions enabling an OpenCode agent accessing only the markdown file (via URL or path) to bootstrap AgenticLens.
2. Direct binary download commands using `curl`/`wget` targeting official precompiled releases from `https://github.com/NiltonDuarte/AgenticLens/releases`.
3. Co-located binary placement at `.opencode/skills/agentic-lens-setup/bin/agentic-lens` or `./bin/agentic-lens` with protective `.gitignore`.
4. Non-interactive and interactive configuration generation via `init`.
5. Strict enforcement of the Mandatory User Confirmation Gate before modifying target repository instructions (`AGENTS.md`, `.opencode/agent/*.md`), with zero references to `CLAUDE.md`.
6. Documentation of literal nonce generation for the `generate` command.

### Quality and Invariant Contract
1. `docs/SETUP_MANUAL.md` and `.opencode/skills/agentic-lens-setup/SKILL.md` must contain zero references to Claude Code, Cursor, Aider, GitHub Copilot, or `CLAUDE.md`.
2. Documented `generate` CLI examples must use literal nonce generation and forbid `$NONCE` or `$(...)`.
3. Zero emojis across all modified files and deliverables.
4. Standard quality gates (`make check`, `make test`, `make build`) must pass cleanly.
