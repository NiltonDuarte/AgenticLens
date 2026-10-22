# Implementation Plan: AL-37 Iteration 02 - Agentic Distribution Model and Agnostic Platform Manual

## 1. Goal and Iteration Scope

Pivot the distribution and documentation model of ticket AL-37 from archive packaging to an agentic distribution model:
1. Restructure and author the definitive cross-project platform setup manual at docs/SETUP_MANUAL.md starting with Section 1 Quick Start (Pathways A and B), Section 2 Agnostic Adoption (strictly omitting internal maintainer methodology text), Section 3 Configuration Reference, Section 4 Agnostic Agent Workflow Integration Guide (featuring realistic adopter role mappings), Section 5 CLI and Web Dashboard Reference, Section 6 Operational Troubleshooting, and Appendix Platform Architecture.
2. Update .opencode/skills/agentic-lens-setup/SKILL.md to be completely self-contained for an autonomous agent reading via shared URL or repository link.
3. Delete obsolete packaging and test scripts: scripts/package-setup-skill.sh and scripts/test-setup.sh.
4. Revert Makefile packaging additions: remove package-setup-skill from .PHONY, help, target list, and release-package dependencies.
5. Verify repository health through standard quality gates: make check, make test, and make build, ensuring zero emojis across all deliverables.

---

## 2. Grounded Reference Architecture & File Paths

The plan is grounded against the concrete files, packages, commands, and schemas of the AgenticLens repository:

### Documentation Deliverables
- docs/SETUP_MANUAL.md:
  - Primary cross-project platform onboarding and operational reference manual.
  - Written in clear technical prose with zero emojis.

### Autonomous Setup Skill Deliverables
- .opencode/skills/agentic-lens-setup/SKILL.md:
  - Self-contained OpenCode skill manifest and autonomous agent setup workflow.
  - Consumable directly via URL or repository reference.
- .opencode/skills/agentic-lens-setup/scripts/download.sh:
  - Co-located artifact downloader retrieving precompiled binaries from GitHub Releases without source compilation.
- .opencode/skills/agentic-lens-setup/bin/.gitignore:
  - Protective gitignore preventing downloaded executables from being committed.

### Script Cleanup Targets
- scripts/package-setup-skill.sh:
  - File to be deleted.
- scripts/test-setup.sh:
  - File to be deleted.

### Build Automation
- Makefile:
  - Targets to preserve: all, install, install-backend, install-frontend, check, test, test-backend, test-frontend, lint, vet, build, clean, help, server, serve, release-build, release-package, release-verify, release-clean.
  - Targets and references to remove: package-setup-skill.

### CLI Subcommands and Domain Config Grounding
- backend/cmd/agentic-lens/main.go:
  - Subcommands: init, generate, scan (root flags), server, version, backfill, migrate.
- backend/internal/domain/config.go:
  - Configuration structure: ProjectConfig, TicketConfig, WorkspacesConfig, DatabaseConfig, LiveConfig, CodeMetricsConfig, AnomalyThresholds, RoleMappings, RoleOrder.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Revert Makefile Packaging Additions and Delete Obsolete Scripts
- Task 1.1: Remove package-setup-skill from .PHONY declaration in Makefile.
- Task 1.2: Remove package-setup-skill line from the help target in Makefile.
- Task 1.3: Delete the entire package-setup-skill target block from Makefile.
- Task 1.4: Revert release-package: package-setup-skill in Makefile back to release-package:.
- Task 1.5: Delete scripts/package-setup-skill.sh from disk and git tracking.
- Task 1.6: Delete scripts/test-setup.sh from disk and git tracking.

### Phase 2: Restructure and Author Cross-Project Platform Setup Manual (docs/SETUP_MANUAL.md)
- Task 2.1: Author Section 1 - Quick Start:
  - Document Pathway A: Agent-Driven Setup via SKILL.md. Explain how an autonomous agent or developer points to SKILL.md (via shared URL or local clone), downloads the precompiled release binary from GitHub Releases, initializes configuration with agentic-lens init, verifies the binary with agentic-lens version --json, and reviews agent instructions with the mandatory confirmation gate.
  - Document Pathway B: Manual Developer Installation. Explain downloading the appropriate precompiled archive from GitHub Releases, extracting the binary into ./bin/agentic-lens, granting executable permissions with chmod +x, running ./bin/agentic-lens init to scaffold .agenticlens-config.yaml, starting the dashboard with ./bin/agentic-lens server --port 8085, and integrating the CLI into existing agent or git workflows.
- Task 2.2: Author Section 2 - Agnostic Adoption:
  - Strictly omit any text describing this repository's internal maintainer methodology, internal development skills (agentic-lens-dev, agentic-lens-po, agentic-lens-review), or the internal five-agent persona loop.
  - Focus purely on how external adopters adopt AgenticLens in their own repositories:
    - Adopting teams define their own ways of working, agent personas, review procedures, and branching models.
    - Zero coupling to any single-agent loop or multi-agent framework. Works with OpenCode, Claude Code, Cursor, Aider, custom shell scripts, or CI/CD pipelines.
    - Adopters configure custom ticket prefixes, workspace locations, and commit conventions.
    - AgenticLens acts strictly as an agnostic telemetry ingestion, aggregation, and visualization engine.
- Task 2.3: Author Section 3 - Configuration Reference:
  - Exhaustively document all root configuration blocks in .agenticlens-config.yaml:
    - project_name: Descriptive project identifier shown in headers and CLI summaries.
    - ticket: prefixes (list of prefix, description, ticket_type), case_insensitive boolean, custom_patterns regex list.
    - workspaces: dir (root workspace directory, default workspaces), src_folders (monitored source directories for git diff analysis), artifacts (tracked documentation filenames).
    - database: path (local OpenCode SQLite database path), query_timeout_ms, busy_timeout_ms, read_only.
    - live: active_threshold_seconds, poll_interval_ms, max_trace_buffer_messages, anomaly_thresholds (stall_threshold_seconds, error_rate_threshold, context_exhaustion_pct).
    - code_metrics: base_branch, include, exclude, test_patterns, prod_patterns.
    - role_mappings: Mapping of agent persona strings to canonical roles.
    - role_order: Ordered sequence of canonical roles for display tables and summaries.
- Task 2.4: Author Section 4 - Agnostic Agent Workflow Integration Guide:
  - Explain the Delayed Telemetry Ingestion Mandate: why running telemetry generation must be the very final turn before committing work to prevent omitting tool calls, turn counts, compactions, and subagent handoffs.
  - Document random execution nonce generation using od command to ensure deterministic turn resolution.
  - Document conventional commit classification (feat, fix, docs, chore, refactor, test, ci, perf, build) and how to map task types cleanly.
  - Provide concrete role_mappings integration examples using realistic external adopter personas:
    - security-scanner: security
    - perf-analyzer: tester
    - integration-tester: tester
    - frontend-specialist: developer
    - backend-specialist: developer
  - Show how any agent harness invokes ./bin/agentic-lens generate with ticket-id, nonce, ticket-type, json-output, and append flags as the final step before git commit.
- Task 2.5: Author Section 5 - CLI and Web Dashboard Reference:
  - Document all CLI subcommands and flags:
    - init: Scaffolding root configuration.
    - generate: Session telemetry ingestion and markdown summary table updates.
    - scan / root flags (--list, --ticket-id, --workspaces, --anomalies, --append, --json-output): Workspace discovery and summary tables.
    - server: Starting embedded web dashboard and real-time SSE live stream service.
    - version: Inspecting binary version, git commit, build date, and schema version.
    - backfill: Batch ledger backfill, git diff recomputation, and schema upgrades.
    - migrate: Stepwise schema migration across ledger versions.
  - Document embedded Web Dashboard views and analytical capabilities:
    - Fleet Overview (/): Cross-ticket rollups, context health distribution, cost velocity, prefix filtering, and search.
    - Live Observability (/live): Active agent session monitoring, SSE event streaming, trace logs, and anomaly cards.
    - Ticket Detail (/ticket/:id): In-depth ticket metrics, SDLC iteration loops, tool invocation distribution, and token usage breakdown.
    - Ticket Comparisons (/comparisons): Side-by-side comparative analysis across tickets.
    - Metric Progression (/progression): Longitudinal trends in active time, token costs, and context saturation.
- Task 2.6: Author Section 6 - Operational Troubleshooting:
  - SQLite concurrency, WAL mode, busy timeout handling, and read-only connections.
  - Cross-platform database path resolution for Linux, macOS, and Windows.
  - Executable permissions (chmod +x) and macOS Gatekeeper quarantine removal.
  - Git upstream base branch inference and fallback probing (origin/main, origin/master, main, master).
  - Workspace directory path resolution and relative path handling.
- Task 2.7: Author Appendix - Platform Architecture:
  - Functional Core, Imperative Shell (FCIS) design pattern in Go.
  - Single binary zero-dependency architecture with embedded TypeScript web dashboard via Go embed.FS.
  - Local immutable fact ledgers (agentic_metrics.json) ensuring data privacy and zero cloud telemetry egress.
  - Real-time live session observability and SSE streaming engine.
  - Automated anomaly detection (stalls, error rate spikes, context exhaustion).
  - Context window health classifications (HEALTHY, ELEVATED, BLOATED, CRITICAL).

### Phase 3: Enhance Standalone Setup Skill Manifest (.opencode/skills/agentic-lens-setup/SKILL.md)
- Task 3.1: Update SKILL.md frontmatter and description to highlight the autonomous agentic distribution model.
- Task 3.2: Structure the workflow so an autonomous agent reading SKILL.md via shared URL or repository link can execute setup end-to-end without assuming local repository files exist in advance.
- Task 3.3: Document direct binary retrieval instructions via curl or wget targeting official precompiled releases from GitHub Releases.
- Task 3.4: Include platform detection commands (uname -s and uname -m) to download the appropriate asset (linux-amd64, darwin-arm64, windows-amd64, etc.).
- Task 3.5: Detail placing the binary in ./bin/agentic-lens or .opencode/skills/agentic-lens-setup/bin/agentic-lens, applying chmod +x, and creating a protective .gitignore.
- Task 3.6: Provide clear interactive interview prompts for configuration (--project-name, --ticket-prefixes, --src-folders, --workspaces, --db-path).
- Task 3.7: Retain the strict Mandatory User Confirmation Gate before modifying any target repository files (AGENTS.md, CLAUDE.md, etc.).
- Task 3.8: Verify zero emojis across SKILL.md.

### Phase 4: Verification and Quality Gates
- Task 4.1: Run make check to verify Go vet, frontend linting, frontend typechecking, and git diff cleanliness.
- Task 4.2: Run make test to verify Go unit tests with race detection and frontend test suite.
- Task 4.3: Run make build to verify single-binary compilation with embedded web dashboard.
- Task 4.4: Verify that scripts/package-setup-skill.sh and scripts/test-setup.sh are deleted.
- Task 4.5: Verify that Makefile contains zero occurrences of package-setup-skill.
- Task 4.6: Run a repository-wide emoji audit on all modified and newly created files to assert zero emojis.

---

## 4. Test and Verification Commands

The coder and tester subagents will execute the following concrete verification commands:

1. Static Code Analysis and Code Invariant Checks:
   - Run make check
   - Expected Output: Go vet exits cleanly, npm typecheck exits cleanly, npm lint exits cleanly, and git diff check passes.

2. Automated Unit Test Suite:
   - Run make test
   - Expected Output: All Go tests in backend/ pass with race detection enabled; all frontend tests in web/ pass.

3. Single Binary Build:
   - Run make build
   - Expected Output: Multi-platform binaries compile without errors and bin/agentic-lens is generated successfully.

4. Packaging Cleanliness Verification:
   - Run git status --porcelain scripts/package-setup-skill.sh scripts/test-setup.sh
   - Expected Output: Files do not exist on disk or in git index.
   - Run grep -n "package-setup-skill" Makefile || true
   - Expected Output: Zero matches returned.

5. Zero Emojis Verification:
   - Run grep on modified files to assert zero unicode emoji characters.
