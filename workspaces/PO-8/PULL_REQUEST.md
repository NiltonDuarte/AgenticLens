# docs(plan): pre-release source code audit, cross-project setup skill, and integration manual (PO-8)

## Summary
Defines the product specification, architectural contracts, and vertical tracer-bullet Jira tickets (AL-35, AL-36, AL-37) for Phase 7 in preparation for the v1.0.0 release: executing a multi-tier codebase audit with `agentic-lens-review`, purging project-specific configuration leaks, exposing `/api/config` for dynamic observed project display alongside permanent AgenticLens platform branding, building the autonomous `agentic-lens-setup` skill with automated binary downloader and interactive interview engine, and authoring the cross-project setup manual and portable skill suite.

## Initiatives Covered

1. **Pre-Release Code Review & Configuration Leakage Elimination (AL-35)**:
   - Audits all Go backend packages and TypeScript views using `agentic-lens-review`.
   - Profiles test suite execution speed to diagnose all tests exceeding 100ms, removing artificial sleeps and enforcing microsecond domain execution.
   - Purges hardcoded repository paths (`workspaces/`), source folder prefixes (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`), and default branch (`origin/main`) from git analyzer and classifier adapters, binding dynamically to `.agenticlens/config.yaml`.
   - Exposes `GET /api/config` in Go HTTP server delivering sanitized `ProjectConfig`.
   - Displays the observed project name dynamically in `Navbar.tsx` and `LandingPage.tsx` alongside permanent AgenticLens platform branding.
   - Generates formal audit report artifact (`workspaces/AL-35/AUDIT_REPORT.md`).

2. **Cross-Project Interactive Setup Skill & Downloader (AL-36)**:
   - Packages `agentic-lens-setup` skill in `.opencode/skills/agentic-lens-setup/SKILL.md`.
   - Automated multi-platform binary downloader script supporting Linux and macOS (amd64, arm64) from GitHub Releases with local compilation fallback.
   - Interactive user interview engine prompting for project name, ticket prefixes, workspaces directory, monitored source directories, SQLite database path, and agent role mappings.
   - Generates tailored, validated `.agenticlens/config.yaml`.
   - Formulates clear platform capability explanations (CLI commands, live monitoring, web dashboard).
   - Audits repository agent instruction files (`AGENTS.md`) and injects mandatory ticket tracking rules and delayed telemetry generation commands.
   - Self-verifies installation via `./bin/agentic-lens version --json`.

3. **Cross-Project Setup Manual & Skill Suite Packaging (AL-37)**:
   - Comprehensive cross-project manual (`docs/SETUP_MANUAL.md`) covering architecture, quickstart onboarding, full YAML schema reference, agent workflow integration, CLI flags, and dashboard walkthrough.
   - Skill suite packager script (`scripts/install-skills.sh`) to distribute `agentic-lens-setup`, `agentic-lens-po`, `agentic-lens-dev`, and `agentic-lens-review` to external repositories.
   - Genericizes all skill manifests with repository-relative paths and configuration parameters.

## Ticket Breakdown
- Ready: `AL-35: Pre-Release Source Code Audit, Configuration Leakage Elimination, and Follow-Up Artifact Generation` (Phase 7, Tier 1)
- Pending: `AL-36: Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator` (Phase 7, Tier 2, Blocked by AL-35)
- Pending: `AL-37: Cross-Project Setup Manual, Skill Suite Usage Guide, and Distribution Packaging` (Phase 7, Tier 3, Blocked by AL-36)

## Pointers
- Specification: `workspaces/PO-8/SPEC.md`
- Shared Roadmap: `workspaces/plan/TICKETS.md`
- Development Tickets:
  - `workspaces/plan/to-do/AL-35 pre-release source code audit configuration leakage elimination and follow-up artifact generation.md`
  - `workspaces/plan/to-do/AL-36 cross-project interactive setup skill automated artifact downloader and agent workflow integrator.md`
  - `workspaces/plan/to-do/AL-37 cross-project setup manual skill suite usage guide and distribution packaging.md`
- Session Summary: `workspaces/PO-8/SUMMARY.md`
