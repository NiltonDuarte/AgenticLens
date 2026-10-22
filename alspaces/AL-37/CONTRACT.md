# Contract: AL-37 Cross-Project Platform Setup Manual and Telemetry Integration Guide

## 1. Problem Statement

External engineering teams, platform architects, and developers require comprehensive, authoritative documentation to onboard and operate AgenticLens across diverse software repositories. Currently, external adopters lack a unified setup manual and an automated packaging mechanism for platform onboarding skills.

Crucially, an architectural confusion exists between the AgenticLens product platform and this repository's internal development methodology. The internal skills resident in this repository (.opencode/skills/agentic-lens-dev, .opencode/skills/agentic-lens-po, and .opencode/skills/agentic-lens-review), alongside the multi-agent orchestrator-planner-coder-tester-reviewer delivery lifecycle, are specific to developing AgenticLens itself. They are not part of the AgenticLens product platform.

External repositories adopting AgenticLens must never be forced or expected to adopt the internal agent personas, ticket prefixes, or delivery workflows of this repository. Instead, adopting teams define their own ways of working, agent personas, review procedures, and branching strategies. They require a neutral, standalone platform documentation manual explaining:
1. Platform architecture, value proposition, and the single-binary zero-dependency deployment model.
2. Setup pathways for both agent-driven bootstrapping via the standalone agentic-lens-setup skill and manual developer installation via GitHub Releases.
3. Exhaustive reference documentation for all root configuration parameters in .agenticlens-config.yaml.
4. Agnostic agent workflow integration demonstrating how any agent harness, single-agent loop, or continuous integration script invokes the AgenticLens CLI as the final turn before committing work.
5. Complete CLI command reference and embedded web dashboard navigation.
6. Operational troubleshooting and best practices covering SQLite locks, database path resolution across operating systems, binary permissions, and git branch inference.

To enable clean distribution, the standalone onboarding skill agentic-lens-setup must be cleanly packaged into standalone release archives (dist/skills/agentic-lens-setup.tar.gz and dist/skills/agentic-lens-setup.zip) that strictly exclude internal development skills, temporary binaries, and build artifacts.

Ticket AL-37 delivers this complete documentation manual at docs/SETUP_MANUAL.md, implements the standalone skill packaging script at scripts/package-setup-skill.sh, exposes the corresponding Makefile packaging targets, and verifies the end-to-end setup and packaging pipeline via scripts/test-setup.sh.

---

## 2. Functional Core, Imperative Shell (FCIS) and Architectural Invariants

All implementation deliverables must strictly adhere to TECHNICAL_GUIDELINES.md and core system invariants:

### Functional Core, Imperative Shell (FCIS) Invariants
1. Pure Functional Core Isolation:
   - Go domain models and mathematical rollups in backend/internal/domain/ (such as token aggregation, cost derivation, context window health evaluation, anomaly detection, and version formatting) remain pure value-semantic functions with zero external dependencies, zero disk I/O, zero network calls, and zero clocks.
   - Documentation and integration instructions must preserve this boundary: telemetry ledgers (agentic_metrics.json) are immutable factual records generated deterministically from SQLite execution traces and git diff analysis.
2. Imperative Shell Discipline:
   - All filesystem operations, configuration file parsing and generation (backend/internal/adapters/config/), SQLite database queries (backend/internal/adapters/sqlite/), git diff extraction (backend/internal/adapters/git/), HTTP serving and asset embedding (backend/internal/adapters/http/ and backend/internal/adapters/web/), and CLI flag handling (backend/cmd/agentic-lens/) reside exclusively within the imperative shell.
   - Packaging scripts and verification harnesses operate cleanly within the imperative shell without introducing runtime dependencies to the pure domain core.

### Single Binary Distribution and Runtime Independence
1. Zero External Runtime Dependency:
   - The compiled Go binary is completely self-contained. The TypeScript single-page application dashboard is embedded into the Go binary via Go embed.FS.
   - Adopters require zero Node.js runtime, zero npm dependencies, and zero external database servers in production.
   - SQLite connectivity uses a pure Go driver (modernc.org/sqlite) compiled with CGO_ENABLED=0, eliminating C compiler toolchain requirements on target host machines.
2. Precompiled Release Consumption:
   - External projects must never be instructed or required to compile AgenticLens from Go or TypeScript source code. Setup pathways strictly consume official precompiled release binaries from GitHub Releases (https://github.com/NiltonDuarte/AgenticLens/releases).

### Clean Skill Distribution and Boundary Isolation
1. Exclusive Distribution of agentic-lens-setup:
   - Only .opencode/skills/agentic-lens-setup/ is packaged for external distribution.
   - Internal development skills (.opencode/skills/agentic-lens-dev, .opencode/skills/agentic-lens-po, and .opencode/skills/agentic-lens-review) must never be packaged, exported, or distributed in release archives.
2. Binary Hygiene in Skill Distribution:
   - Distribution archives package SKILL.md, scripts/download.sh, and the protective bin/.gitignore.
   - Precompiled binary executables (.opencode/skills/agentic-lens-setup/bin/agentic-lens) must never be bundled into the git repository or into the skill distribution archive; binaries are downloaded dynamically on target host machines during onboarding.

### Zero Emojis Mandate
Zero emojis are permitted across all documentation files, skill manifests, bash scripts, Makefiles, Go code, TypeScript code, commit messages, pull requests, CLI outputs, and workspace artifacts.

---

## 3. Scope Boundaries: Documentation, Packaging, and Automated Verification

### In Scope
1. Documentation Authoring (docs/SETUP_MANUAL.md):
   - Section 1: Platform Overview and Architecture (value proposition, single binary with embedded web dashboard, local fact ledgers, live observability, context window health analytics).
   - Section 2: Delineation of Product Platform vs. Internal Development Methodology (explicitly establishing that internal skills and multi-agent roles belong to this repository's own git project structure, whereas adopters configure their own workflows).
   - Section 3: Setup Pathways (Agent-Driven Onboarding via agentic-lens-setup, and Manual Developer Installation via precompiled GitHub Releases).
   - Section 4: Configuration Reference (exhaustive schema reference for .agenticlens-config.yaml covering project_name, ticket, workspaces, database, live, code_metrics, role_mappings, and role_order).
   - Section 5: Agnostic Agent Workflow Integration (the delayed telemetry ingestion mandate, invoking generate as the final turn before committing work, integrating custom personas, and conventional commit classification).
   - Section 6: CLI and Embedded Web Dashboard Reference (full command-line reference for init, generate, scan / root flags, server, version, backfill, and migrate, along with navigation for Fleet Overview, Live Observability, Ticket Detail, Ticket Comparisons, and Metric Progression).
   - Section 7: Operational Troubleshooting and Best Practices (SQLite concurrency and locks, platform-specific database paths, binary permissions, git branch inference, and workspace resolution).
2. Standalone Skill Packaging (scripts/package-setup-skill.sh and Makefile):
   - Implement scripts/package-setup-skill.sh to package .opencode/skills/agentic-lens-setup into dist/skills/agentic-lens-setup.tar.gz and dist/skills/agentic-lens-setup.zip.
   - Enforce packaging hygiene: include SKILL.md, scripts/download.sh, and bin/.gitignore; explicitly exclude any precompiled binaries, temporary files, or internal skills.
   - Generate SHA256 checksums for packaged skill archives.
   - Add package-setup-skill target to Makefile, integrate it with release-package, and ensure release-clean removes dist/skills/.
3. Automated Verification (scripts/test-setup.sh):
   - Extend scripts/test-setup.sh with automated assertions verifying:
     - docs/SETUP_MANUAL.md exists, is non-empty, and satisfies the zero-emoji mandate.
     - scripts/package-setup-skill.sh executes successfully and outputs valid tar.gz and zip archives in dist/skills/.
     - Packaged skill archives contain SKILL.md, scripts/download.sh, and bin/.gitignore.
     - Packaged skill archives do not contain precompiled binaries (bin/agentic-lens) or internal skills (agentic-lens-dev, agentic-lens-po, agentic-lens-review).
     - SHA256 checksums are generated and verify cleanly.

### Out of Scope
1. Altering Go domain logic, aggregation formulas, context window health rules, or anomaly detection thresholds in backend/internal/domain/.
2. Modifying TypeScript frontend UI components, views, or routing in web/src/.
3. Altering OpenCode SQLite schema parsing or SSE live event streaming formats.
4. Changing the behavior, manifest, or instructions of internal repository skills (agentic-lens-dev, agentic-lens-po, agentic-lens-review).

---

## 4. Architectural Contracts and Data Specifications

### Documentation Structure Contract: docs/SETUP_MANUAL.md
The setup manual must be authored as a clean, highly structured Markdown file at docs/SETUP_MANUAL.md containing the following exact top-level sections:
1. Title and Document Overview
2. Platform Architecture and Core Concepts
3. Architectural Delineation: Product Platform vs. Internal Methodology
4. Installation and Setup Pathways
5. Configuration Reference (.agenticlens-config.yaml)
6. Agnostic Agent Workflow Integration Guide
7. CLI and Web Dashboard Reference
8. Troubleshooting and Operational Best Practices

### Standalone Skill Packaging Contract: scripts/package-setup-skill.sh
1. Input Source Directory: .opencode/skills/agentic-lens-setup
2. Output Directory: dist/skills/
3. Output Artifacts:
   - dist/skills/agentic-lens-setup.tar.gz
   - dist/skills/agentic-lens-setup.zip
   - dist/skills/SHA256SUMS
4. Validation and Packaging Rules:
   - Fail immediately if .opencode/skills/agentic-lens-setup/SKILL.md is missing.
   - Fail immediately if .opencode/skills/agentic-lens-setup/scripts/download.sh is missing.
   - Ensure bin/.gitignore is present in the archive with wildcard pattern '*' and '!/.gitignore'.
   - Strictly exclude any existing binary files in bin/ (such as bin/agentic-lens or bin/agentic-lens.exe).
   - Strictly exclude any parent or sibling directories containing internal skills (agentic-lens-dev, agentic-lens-po, agentic-lens-review).
   - Support optional command-line flags: --dest-dir <dir>, --version <ver>, --dry-run, --help.

### Makefile Integration Contract
1. package-setup-skill: Standalone target invoking scripts/package-setup-skill.sh.
2. release-package: Dependencies updated to build cross-compiled binaries and invoke package-setup-skill.
3. release-clean: Updated to remove dist/ including dist/skills/.

### Verification Contract: scripts/test-setup.sh
1. Documentation Check: Verifies existence and zero-emoji compliance of docs/SETUP_MANUAL.md.
2. Packaging Check: Executes packaging script in a temporary test directory, verifies tar.gz and zip archive creation, inspects archive contents for expected files, and asserts zero inclusion of internal skills or binary executables.
3. Checksum Verification: Validates SHA256SUMS file generation and integrity.
