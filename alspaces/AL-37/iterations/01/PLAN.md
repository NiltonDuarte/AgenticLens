# Implementation Plan: AL-37 Cross-Project Platform Setup Manual and Telemetry Integration Guide

## 1. Goal and Iteration Scope

Deliver the complete implementation of ticket AL-37 across documentation, packaging scripts, Makefile integration, and automated verification:
1. Author the definitive, comprehensive cross-project platform setup manual and telemetry integration guide at docs/SETUP_MANUAL.md.
2. Implement the standalone setup skill packaging script at scripts/package-setup-skill.sh.
3. Wire the packaging script into Makefile via the package-setup-skill target, update release-package to include skill distribution, and update release-clean.
4. Extend scripts/test-setup.sh to verify docs/SETUP_MANUAL.md compliance, validate skill archive packaging, inspect archive contents, verify internal skill exclusion, and ensure zero emojis across all deliverables.
5. Execute the full verification suite (make check, make test, make build, ./scripts/test-setup.sh) to confirm all quality gates pass.

---

## 2. Grounded Reference Architecture & File Paths

The implementation plan is grounded against the concrete files, packages, commands, and schemas of the AgenticLens repository:

### Documentation Layer
- docs/SETUP_MANUAL.md (new file):
  - Primary cross-project platform onboarding and operation manual.
  - Written in clear prose with zero code blocks and zero emojis.

### Configuration Reference & Schema
- .agenticlens-config.yaml (existing root configuration):
  - project_name: string identifier.
  - ticket: prefixes (structured list of prefix, description, ticket_type), case_insensitive (bool), custom_patterns (regex list).
  - workspaces: dir (ledger directory), src_folders (monitored source directories for git diff analysis), artifacts (tracked workspace documentation).
  - database: path (OpenCode SQLite database), query_timeout_ms (int), busy_timeout_ms (int), read_only (bool).
  - live: active_threshold_seconds (int), poll_interval_ms (int), max_trace_buffer_messages (int), anomaly_thresholds (stall_threshold_seconds, error_rate_threshold, context_exhaustion_pct).
  - code_metrics: base_branch (upstream comparison branch), include, exclude, test_patterns, prod_patterns.
  - role_mappings: map of agent persona to canonical role (orchestrator, planner, coder, tester, reviewer).
  - role_order: ordered slice of canonical roles for display sorting.
- backend/internal/domain/config.go:
  - ProjectConfig, TicketConfig, TicketPrefix, WorkspacesConfig, DatabaseConfig, LiveConfig, CodeMetricsConfig, AnomalyThresholds.
  - DefaultProjectConfig(): Canonical baseline defaults.
- backend/internal/adapters/config/writer.go:
  - GenerateConfigYAML(): Scaffolds fully commented .agenticlens-config.yaml with mandatory fields and default blocks.

### CLI Application Entrypoints
- backend/cmd/agentic-lens/main.go:
  - init: Scaffolds root .agenticlens-config.yaml. Required flags: --project-name, --src-folders, --ticket-prefixes. Optional flags: --workspaces, --db-path, --output, --force, --dry-run.
  - generate: Ingests session telemetry and git diffs into agentic_metrics.json. Flags: --ticket-id, --nonce, --session-id, --ticket-title, --ticket-type, --strict-type, --started-at, --completed-at, --json-output, --append, --db-path, --anomalies, --allow-empty.
  - scan / root flags: Discovers workspaces and outputs metrics tables. Flags: --list, --ticket-id, --workspaces, --anomalies, --append, --json-output.
  - server: Embedded web dashboard and live SSE streaming API. Flags: --port, --workspaces, --db-path.
  - version: Binary build metadata and schema version. Flags: --json, --short.
  - backfill: Batch telemetry backfill and schema migrations. Flags: --workspaces, --ticket-id, --recompute-git, --recompute-db, --migrate-schema, --dry-run, --update-summaries, --db-path, --force.
  - migrate: Universal stepwise schema migration engine. Flags: --workspaces, --ticket-id, --target-version, --dry-run, --backup, --update-summaries.

### Standalone Setup Skill & Scripts
- .opencode/skills/agentic-lens-setup/SKILL.md:
  - OpenCode skill manifest and interactive onboarding workflow.
- .opencode/skills/agentic-lens-setup/scripts/download.sh:
  - Co-located artifact downloader retrieving precompiled binaries from GitHub Releases without source compilation.
- scripts/download-agentic-lens.sh:
  - Root repository binary downloader with OS and architecture detection.
- scripts/package-setup-skill.sh (new file):
  - Packaging script bundling .opencode/skills/agentic-lens-setup into dist/skills/agentic-lens-setup.tar.gz and dist/skills/agentic-lens-setup.zip.
- Makefile:
  - Targets: install, check, test, build, release-build, release-package, release-verify, release-clean, clean.
  - Adding target: package-setup-skill.

### Verification Harness
- scripts/test-setup.sh:
  - Automated verification suite testing binary compilation, downloader platform resolution, protective .gitignore generation, version JSON output, init command scaffolding, edge cases, skill manifest presence, and zero emojis.
  - Adding test stages for docs/SETUP_MANUAL.md validation, skill packaging execution, archive content verification, internal skill exclusion, and checksum verification.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Author Cross-Project Platform Setup Manual (docs/SETUP_MANUAL.md)
- Task 1.1: Initialize docs/ directory if not present and create docs/SETUP_MANUAL.md.
- Task 1.2: Author Section 1 - Executive Summary and Platform Architecture:
  - Detail the zero-dependency single-binary architecture with bundled TypeScript web dashboard via Go embed.FS.
  - Detail local telemetry ledgers (agentic_metrics.json), live session observability via SQLite polling and SSE streaming, automated anomaly detection (stalls, tool errors, context saturation), and context window health states (HEALTHY, ELEVATED, BLOATED, CRITICAL).
- Task 1.3: Author Section 2 - Core Architectural Invariant: Platform vs. Development Methodology:
  - Explicitly delineate the AgenticLens product platform from this repository's internal development methodology.
  - Clarify that skills like agentic-lens-dev, agentic-lens-po, and agentic-lens-review, as well as the orchestrator-planner-coder-tester-reviewer role structure, belong to this repository's internal development process.
  - State clearly that external repositories define their own software delivery workflows, agent personas, and branch strategies, using AgenticLens purely as an observability and analytics layer.
- Task 1.4: Author Section 3 - Platform Setup Pathways:
  - Detail Pathway A: Agent-Driven Onboarding via agentic-lens-setup. Explain running the skill in OpenCode, downloading the precompiled release binary into .opencode/skills/agentic-lens-setup/bin/agentic-lens, running init to scaffold .agenticlens-config.yaml, and presenting agent instructions with the mandatory user confirmation gate.
  - Detail Pathway B: Manual Developer Installation. Explain downloading the appropriate OS/architecture archive from GitHub Releases (https://github.com/NiltonDuarte/AgenticLens/releases), extracting to ./bin/agentic-lens, marking executable with chmod +x, and running ./bin/agentic-lens init.
- Task 1.5: Author Section 4 - Exhaustive Configuration Reference (.agenticlens-config.yaml):
  - Document all root configuration blocks with field names, data types, default values, and operational effects:
    - project_name: Human-readable project name displayed in web dashboard badges and CLI outputs.
    - ticket: prefixes (prefix, description, ticket_type), case_insensitive, custom_patterns.
    - workspaces: dir (root directory for ticket workspaces), src_folders (monitored source directories for git metrics), artifacts (tracked documentation files).
    - database: path (OpenCode SQLite database location), query_timeout_ms, busy_timeout_ms, read_only.
    - live: active_threshold_seconds, poll_interval_ms, max_trace_buffer_messages, anomaly_thresholds (stall_threshold_seconds, error_rate_threshold, context_exhaustion_pct).
    - code_metrics: base_branch, include, exclude, test_patterns, prod_patterns.
    - role_mappings: mapping of agent persona strings to canonical roles.
    - role_order: ordering sequence for canonical roles in summaries and tables.
- Task 1.6: Author Section 5 - Agnostic Agent Workflow Integration Guide:
  - Detail the delayed telemetry ingestion mandate: explain why running telemetry generation must be the very final turn before committing work to prevent omitting tool calls, turn counts, and subagent sessions.
  - Provide instructions for integrating ./bin/agentic-lens generate into any agent harness (single-agent loops, multi-agent frameworks, shell scripts, or CI pipelines).
  - Explain command-line parameters: --ticket-id, --nonce (random execution nonce for session resolution), --ticket-type (conventional commit classification: feat, fix, docs, chore, refactor, test, ci, perf, build), --json-output, and --append.
  - Explain how external projects define their own ticket naming conventions, workspace directories, and commit messages without inheriting AgenticLens-specific roles.
- Task 1.7: Author Section 6 - CLI Commands and Web Dashboard Navigation:
  - Document each CLI subcommand with its purpose and flags:
    - init: Scaffolding configuration.
    - generate: Session telemetry ingestion and markdown summary updates.
    - scan / root flags (--list, --ticket-id, --workspaces, --anomalies): Workspace discovery and summary tables.
    - server: Starting embedded web dashboard and SSE live streaming server.
    - version: Inspecting binary version, git commit, build date, and schema version.
    - backfill: Batch ledger backfilling, git recomputation, and schema upgrades.
    - migrate: Universal stepwise schema migration across ledger versions.
  - Document Web Dashboard views and analytical capabilities:
    - Fleet Overview (/): Cross-ticket rollups, context window health distribution, cost velocity, prefix filtering, and search.
    - Live Observability (/live): Real-time active agent session monitoring, SSE event streaming, trace logs, and anomaly cards.
    - Ticket Detail (/ticket/:id): In-depth ticket metrics, SDLC iteration loops, tool invocation distribution, and token usage breakdown.
    - Ticket Comparisons (/comparisons): Multi-ticket side-by-side comparative analysis.
    - Metric Progression (/progression): Longitudinal trends in active time, token consumption, and context health.
- Task 1.8: Author Section 7 - Operational Troubleshooting and Best Practices:
  - SQLite database concurrency, busy timeout handling, WAL mode, and read-only access.
  - Cross-platform OpenCode database path resolution (Linux, macOS, Windows).
  - Executable permissions and binary path resolution.
  - Git upstream base branch inference and fallback probing.
  - Workspace directory path resolution (relative vs. absolute paths).

### Phase 2: Implement Standalone Setup Skill Packaging (scripts/package-setup-skill.sh & Makefile)
- Task 2.1: Create scripts/package-setup-skill.sh:
  - Ensure script is executable (chmod +x) and uses set -euo pipefail.
  - Accept optional CLI flags: --dest-dir <dir> (default: dist/skills), --version <version> (default: resolved from git describe or v0.0.0-dev), --dry-run, --help.
  - Validate source directory .opencode/skills/agentic-lens-setup: verify SKILL.md and scripts/download.sh exist.
  - Ensure bin/.gitignore exists with wildcard pattern '*' and '!/.gitignore'.
  - Create secure temporary packaging directory.
  - Copy only the agentic-lens-setup directory into the temporary directory, explicitly omitting any binary executables (.opencode/skills/agentic-lens-setup/bin/agentic-lens*) and temporary editor artifacts.
  - Assert that internal skills (.opencode/skills/agentic-lens-dev, .opencode/skills/agentic-lens-po, .opencode/skills/agentic-lens-review) are completely excluded from the package.
  - Create dist/skills/agentic-lens-setup.tar.gz and dist/skills/agentic-lens-setup.zip.
  - Generate dist/skills/SHA256SUMS containing checksums of the packaged archives.
  - Print clean summary output with zero emojis.
- Task 2.2: Update Makefile:
  - Add package-setup-skill target to execute scripts/package-setup-skill.sh.
  - Update release-package target to call package-setup-skill.
  - Update release-clean target to remove dist/ including dist/skills/.
  - Update help target to list package-setup-skill.

### Phase 3: Extend Automated Verification Suite (scripts/test-setup.sh)
- Task 3.1: Ensure .opencode/skills/agentic-lens-setup/bin/.gitignore is created if absent so that initial setup checks succeed deterministically.
- Task 3.2: Add Test 6 to verify docs/SETUP_MANUAL.md:
  - Assert docs/SETUP_MANUAL.md exists and is non-empty.
  - Assert docs/SETUP_MANUAL.md contains key section headers: Platform Architecture, Architectural Delineation, Installation and Setup Pathways, Configuration Reference, Agnostic Agent Workflow, CLI and Web Dashboard Reference, Troubleshooting.
- Task 3.3: Add Test 7 to verify standalone skill packaging:
  - Execute scripts/package-setup-skill.sh with --dest-dir pointing to a temporary test directory.
  - Assert that agentic-lens-setup.tar.gz, agentic-lens-setup.zip, and SHA256SUMS are created.
  - Verify SHA256SUMS checksum validation succeeds.
  - Extract the tar.gz archive into a sandbox directory and inspect contents:
    - Assert SKILL.md is present.
    - Assert scripts/download.sh is present and executable.
    - Assert bin/.gitignore is present.
    - Assert no precompiled binary (agentic-lens or agentic-lens.exe) exists inside the archive.
    - Assert no internal skill files (agentic-lens-dev, agentic-lens-po, agentic-lens-review) exist inside the archive.
- Task 3.4: Update zero emojis check:
  - Add docs/SETUP_MANUAL.md and scripts/package-setup-skill.sh to the list of checked files in scripts/test-setup.sh.
  - Assert zero unicode emojis across all new and modified files.

### Phase 4: Quality Gate Execution and Verification
- Task 4.1: Run make check (go vet, frontend lint/typecheck, git diff check).
- Task 4.2: Run make test (Go unit tests with race detector and frontend unit tests).
- Task 4.3: Run ./scripts/test-setup.sh and assert all test stages pass.
- Task 4.4: Run make package-setup-skill and verify generated archives in dist/skills/.
- Task 4.5: Assert zero emojis and zero code blocks in CONTRACT.md and PLAN.md.

---

## 4. Test and Verification Commands

The coder and tester subagents will execute the following concrete verification commands:

1. Static Analysis and Code Invariant Check:
   - Command: make check
   - Expected Output: go vet clean exit, npm typecheck clean exit, npm lint clean exit, git diff check clean exit.

2. Automated Unit and Boundary Test Suite:
   - Command: make test
   - Expected Output: All Go tests in backend/ pass with race detector enabled; all frontend tests in web/ pass.

3. Automated Cross-Project Setup and Packaging Verification Suite:
   - Command: ./scripts/test-setup.sh
   - Expected Output: All setup tests pass, including:
     - Platform and architecture detection in downloader script.
     - Local artifact installation and protective bin/.gitignore creation.
     - Binary self-test and version JSON metadata verification.
     - Root configuration scaffolding via init command with edge cases.
     - Documentation presence and structure check for docs/SETUP_MANUAL.md.
     - Standalone skill packaging verification (tar.gz, zip, SHA256SUMS, file inclusion, internal skill exclusion).
     - Zero emojis check across all new and modified files.

4. Standalone Skill Packaging Target Verification:
   - Command: make package-setup-skill
   - Expected Output: Clean generation of dist/skills/agentic-lens-setup.tar.gz, dist/skills/agentic-lens-setup.zip, and dist/skills/SHA256SUMS.

5. Single Binary Build:
   - Command: make build
   - Expected Output: Executable compiled at bin/agentic-lens embedding the web dashboard without errors.
