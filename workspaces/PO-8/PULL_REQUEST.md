# docs(plan): pre-release source code audit, config migration, setup skill, and integration manual (PO-8)

## Summary
Defines the product specification, architectural contracts, and vertical tracer-bullet Jira tickets (AL-35, AL-38, AL-36, AL-37) for Phase 7 in preparation for the v1.0.0 release: executing a multi-tier read-only codebase audit with `agentic-lens-review` focusing on configuration leakage and test suite profiling (generating review artifacts for follow-up implementation tickets), migrating the configuration manifest to root `.agenticlens-config.yaml` with backward-compatible fallback, building the autonomous `agentic-lens-setup` skill with release binary downloader and interactive interview engine, and authoring the cross-project setup manual and telemetry integration guide.

## Initiatives Covered

1. **Pre-Release Code Review, Configuration Leakage Inspection & Review Artifact Generation (AL-35)**:
   - Strictly read-only codebase audit with zero application code or test modifications.
   - Audits all Go backend packages and TypeScript views using `agentic-lens-review`.
   - Profiles test suite execution speed to diagnose all tests exceeding 100ms, locating artificial sleeps and test execution bottlenecks.
   - Inspects for configuration leaks: hardcoded repository paths (`workspaces/`), source folder prefixes (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`), default branch (`origin/main`), and static brand assumptions.
   - Generates formal audit report artifact (`workspaces/AL-35/AUDIT_REPORT.md`) and actionable task specifications for subsequent implementation tickets.

2. **Configuration Manifest Migration to Root `.agenticlens-config.yaml` (AL-38)**:
   - Migrates default configuration path from `.agenticlens/config.yaml` to `.agenticlens-config.yaml` at project root.
   - Maintains seamless backward-compatible fallback in `loader.go` with deprecation notice logging.
   - Updates CLI flags, test fixtures, and documentation to reference `./.agenticlens-config.yaml`.

3. **Cross-Project Interactive Setup Skill & Downloader (AL-36)**:
   - Packages `agentic-lens-setup` skill in `.opencode/skills/agentic-lens-setup/SKILL.md`.
   - Automated multi-platform binary downloader script supporting Linux and macOS (amd64, arm64) from GitHub Releases.
   - Strict release-only mandate: never prompts or asks the user to compile the platform from source.
   - Binary co-location & protective `.gitignore`: downloads binary into `.opencode/skills/agentic-lens-setup/bin/agentic-lens` alongside a local `.gitignore` preventing it from being committed to git.
   - Interactive user interview engine prompting for project name, ticket prefixes, workspaces directory, monitored source directories, SQLite database path, and agent role mappings.
   - Generates tailored, validated `.agenticlens-config.yaml` at the project root.
   - Explains platform capabilities (CLI commands, live monitoring, web dashboard).
   - Audits repository agent instruction files (`AGENTS.md`) and proposes ticket tracking rules and delayed telemetry generation commands, strictly requiring explicit user confirmation before modifying any files.
   - Self-verifies installation via `./.opencode/skills/agentic-lens-setup/bin/agentic-lens version --json`.

4. **Cross-Project Platform Setup Manual & Telemetry Integration Guide (AL-37)**:
   - Comprehensive cross-project manual (`docs/SETUP_MANUAL.md`) covering architecture, quickstart onboarding, full `.agenticlens-config.yaml` schema reference, agent workflow integration, CLI flags, and dashboard walkthrough.
   - Decouples the AgenticLens product platform from internal project development methodologies: clearly documents that external projects define their own ways of working, agent personas, and delivery pipelines.
   - Distributes solely the platform onboarding skill (`agentic-lens-setup`); excludes internal repository skills (`agentic-lens-dev`, `agentic-lens-po`, `agentic-lens-review`).

## Ticket Breakdown
- Ready: `AL-35: Pre-Release Source Code Audit, Configuration Leakage Inspection, and Review Artifact Generation` (Phase 7, Tier 1 - Ready, unblocked)
- Ready: `AL-38: Migrate Configuration Manifest to Root .agenticlens-config.yaml` (Phase 7, Tier 1 - Ready, unblocked, blocks AL-36)
- Pending: `AL-36: Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator` (Phase 7, Tier 2 - Pending, blocked by AL-38, blocks AL-37)
- Pending: `AL-37: Cross-Project Platform Setup Manual and Telemetry Integration Guide` (Phase 7, Tier 3 - Pending, blocked by AL-36)

## Pointers
- Specification: `workspaces/PO-8/SPEC.md`
- Shared Roadmap: `workspaces/plan/TICKETS.md`
- Development Tickets:
  - `workspaces/plan/to-do/AL-35 pre-release source code audit configuration leakage inspection and review artifact generation.md`
  - `workspaces/plan/to-do/AL-38 migrate configuration manifest to root agenticlens-config-yaml.md`
  - `workspaces/plan/to-do/AL-36 cross-project interactive setup skill automated artifact downloader and agent workflow integrator.md`
  - `workspaces/plan/to-do/AL-37 cross-project setup manual and telemetry integration guide.md`
- Session Summary: `workspaces/PO-8/SUMMARY.md`
