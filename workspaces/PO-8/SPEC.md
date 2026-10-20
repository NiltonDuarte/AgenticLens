# Feature Specification: Pre-Release Source Code Audit, Configuration Manifest Migration, Cross-Project Setup Skill, and Integration Manual

## Problem Statement

As AgenticLens approaches its v1.0.0 first public release, four critical capability and architectural requirements must be satisfied:

1. **Pre-Release Source Code Audit and Test Bottleneck Diagnostics (AL-35)**:
   - AgenticLens was developed iteratively across six phases, accumulating technical debt, test latency, and potential architectural drift.
   - A structured, read-only audit using the `agentic-lens-review` skill is required to profile test execution times (identifying slow tests exceeding 100ms), inspect FCIS boundaries, evaluate code smells, identify configuration leakage points, and produce `workspaces/AL-35/AUDIT_REPORT.md` with actionable follow-up implementation tickets.

2. **Configuration Manifest Migration to Root `.agenticlens-config.yaml` (AL-38)**:
   - Configuration currently defaults to `.agenticlens/config.yaml`. For standard repository ergonomics and zero-friction discovery, the configuration file must reside directly at the project root as `.agenticlens-config.yaml`.
   - The configuration loader must prioritize `./.agenticlens-config.yaml` while retaining a backward-compatible fallback to `.agenticlens/config.yaml`, logging a deprecation notice when the legacy location is used.

3. **Cross-Project Interactive Setup Skill & Automated Artifact Downloader (AL-36)**:
   - Adopting AgenticLens in external repositories currently requires manual binary retrieval, manual YAML authoring, and manual editing of agent guidelines.
   - An AI agent in an external repository needs an interactive onboarding skill (`agentic-lens-setup`) that:
     - Downloads precompiled release binaries for the detected OS/architecture from GitHub Releases. Crucially, it must never ask the user to compile the platform from source.
     - Co-locates the downloaded binary inside the skill directory (e.g. `.opencode/skills/agentic-lens-setup/bin/agentic-lens`) along with a co-located `.gitignore` to prevent committing the binary to Git.
     - Conducts an interactive interview to configure project parameters and generate `.agenticlens-config.yaml` at the project root.
     - Explains platform capabilities (CLI commands, live monitoring, web dashboard).
     - Inspects existing agent instructions (`AGENTS.md`), proposing delayed telemetry ingestion guidelines, but strictly requiring explicit user confirmation before writing or modifying any files.

4. **Cross-Project Platform Setup Manual & Telemetry Integration Guide (AL-37)**:
   - External engineering teams need clear documentation on installing the platform, configuring `.agenticlens-config.yaml`, and integrating telemetry into their own agent workflows.
   - Crucially, the internal skills used to build this repository (`agentic-lens-dev`, `agentic-lens-po`, and `agentic-lens-review`) are part of this repository's git project, NOT part of the product platform. Projects using the AgenticLens platform must define their own way of working, rather than copying this repository's multi-agent SDLC.
   - A comprehensive manual (`docs/SETUP_MANUAL.md`) is needed, accompanied by clean distribution packaging of the standalone `agentic-lens-setup` skill.

---

## Target Personas & Use Cases

- **AI Platform Engineers & Tech Leads**: Onboard AgenticLens into repositories within minutes, establishing local telemetry ledgers and live observability while retaining full autonomy over team conventions and agent workflows.
- **Autonomous AI Agents (OpenCode / Harness Agents)**: Execute `agentic-lens-setup` to self-retrieve the release binary into a self-contained, git-ignored skill folder, interview the human operator, write root `.agenticlens-config.yaml`, and request permission to add telemetry instructions to `AGENTS.md`.
- **Software Engineers & Reviewers**: Benefit from a clean, audited codebase with documented test profiling metrics, zero compilation hurdles on onboarding, and permanent platform branding on the dashboard.

---

## User Stories

1. As an AI platform engineer, I want a complete, read-only code audit executed with `agentic-lens-review`, so that all architectural drift, test bottlenecks, and configuration leakage points are documented with concrete follow-up tasks without mutating code in the audit ticket.
2. As a repository maintainer, I want AgenticLens configuration to live at `.agenticlens-config.yaml` at the project root, so that project configuration is immediately visible and follows standard tool convention.
3. As a dashboard viewer, I want the web interface to display `AgenticLens` permanently as the platform brand alongside the configured project name, so that the dashboard always communicates the platform identity while presenting project context.
4. As an AI agent in a new project, I want to download the precompiled binary without ever being asked to compile the platform from source, co-locating it inside the skill directory with a `.gitignore` so that it is never tracked in git.
5. As an agent operator, I want an interactive interview during setup, so that my choices generate a valid `.agenticlens-config.yaml` at the project root.
6. As a repository owner, I want the agent to strictly ask for my approval before making any edits to `AGENTS.md`, so that no instructions are injected into my repository without my consent.
7. As an engineering lead adopting AgenticLens, I want a setup manual focused on the platform itself, allowing my team to define our own way of working rather than inheriting this repository's internal development skills.

---

## Functional Requirements

### 1. Pre-Release Code Audit & Test Profiling (AL-35)
- **Review-Only Mandate**:
  - Strictly read-only audit: zero application code, test files, runtime scripts, or configuration files are modified in AL-35.
  - All findings, leakage points, and test bottlenecks are documented in `workspaces/AL-35/AUDIT_REPORT.md` with concrete follow-up implementation tasks.
- **Multi-Tier Audit with `agentic-lens-review`**:
  - Profile test suite execution to identify all test bottlenecks exceeding 100ms:
    ```bash
    go test -json ./... | jq -r 'select(.Action=="pass" or .Action=="fail") | select(.Test != null) | "\(.Elapsed)s \(.Package) \(.Test)"' | sort -nr | head -20
    ```
  - Audit FCIS boundaries in `internal/domain` (zero I/O imports, zero clocks, microsecond execution SLA).
  - Audit module depth across adapters and CLI, identifying dead code, speculative abstractions, and subtraction targets.
  - Audit Fowler code smells, primitive obsession, and domain modeling invariants.
- **Configuration Leakage Inspection (Backend, Adapters & Web)**:
  - Systematically audit backend adapters for hardcoded repository assumptions:
    - `backend/internal/adapters/git/analyzer.go`: Inspect hardcoded folder prefixes (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`) and diff base branch (`origin/main`).
    - `backend/internal/adapters/git/classifier.go`: Inspect hardcoded `"workspaces/"` string in universal ignore rules.
    - `backend/internal/adapters/http/`: Inspect missing `/api/config` endpoint for dynamic project configuration exposure.
    - `web/src/components/Navbar.tsx` & `LandingPage.tsx`: Verify `AgenticLens` is preserved permanently as the platform name while supporting observed project context alongside it.
- **Review Deliverables & Follow-Up Artifacts**:
  - Author comprehensive report at `workspaces/AL-35/AUDIT_REPORT.md`:
    - Executive summary of audited codebase health.
    - Test execution profiling diagnostics table.
    - Architecture, code smell, and configuration leakage findings categorized by severity (High / Medium / Low).
    - Prioritized remediation roadmap and follow-up ticket definitions.

### 2. Configuration Manifest Migration to Root `.agenticlens-config.yaml` (AL-38)
- **Root Configuration Manifest**:
  - Primary configuration path: `./.agenticlens-config.yaml` at project root.
  - Deprecate `.agenticlens/config.yaml` with seamless fallback in `loader.go`.
  - Log a deprecation notice when the legacy path is loaded.
- **Migration & Schema**:
  - Move the repository's configuration file to `./.agenticlens-config.yaml`.
  - Update all CLI commands, tests, and documentation referencing the manifest path.
  - Verify all CLI subcommands (`generate`, `scan`, `server`, `version`) seamlessly pick up `./.agenticlens-config.yaml`.

### 3. Cross-Project Interactive Setup Skill & Downloader (AL-36)
- **Skill Manifest & Workflow (`.opencode/skills/agentic-lens-setup/SKILL.md`)**:
  - Registered as `agentic-lens-setup` with description and step-by-step agent instructions.
  - Sized for single-session autonomous execution.
- **Automated Artifact Downloader (Release-Only, Zero Compilation)**:
  - Platform detection: `uname -s` (Linux, Darwin) and `uname -m` (x86_64/amd64, arm64/aarch64).
  - Download precompiled release archive from GitHub Releases (`gh release download` or direct URL).
  - Strict release-only mandate: never ask the user to compile the platform from source; zero build toolchain required.
  - Binary Co-Location & Protective `.gitignore`:
    - Co-locate binary inside skill directory: `.opencode/skills/agentic-lens-setup/bin/agentic-lens`.
    - Co-locate a `.gitignore` inside that directory to prevent the binary from being committed to Git.
    - Set executable permissions (`chmod +x`).
  - Verification: Run `./.opencode/skills/agentic-lens-setup/bin/agentic-lens version --json` and assert valid JSON.
- **Interactive Interview & Root Config Generator**:
  - Conduct interactive interview covering project name, ticket prefixes, workspaces directory, source folders, SQLite database path, and agent role mappings.
  - Generate a formatted, commented `.agenticlens-config.yaml` at the project root.
- **Platform Capability Explainer**:
  - Clear overview of CLI subcommands (`generate`, `scan`, `server`, `version`), live session tracking, anomaly alerting, and dashboard navigation.
- **Agent Guidelines File Audit with Mandatory User Confirmation**:
  - Inspect repository for agent instruction files (`AGENTS.md`, `.opencode/agent/*.md`, `CLAUDE.md`).
  - Strict User Confirmation Rule: never modify, inject, or write instructions to `AGENTS.md` without presenting the proposed text and asking the user for explicit permission first.

### 4. Cross-Project Setup Manual & Platform Telemetry Guide (AL-37)
- **Comprehensive Setup Manual (`docs/SETUP_MANUAL.md`)**:
  - Platform overview, value proposition, and architecture.
  - Setup pathways: agent-driven onboarding (`agentic-lens-setup`) and manual developer installation.
  - Exhaustive `.agenticlens-config.yaml` schema reference for all sections.
  - Agnostic agent workflow integration: guide for teams on how to invoke `./bin/agentic-lens generate ...` as their final turn before committing work, without inheriting AgenticLens-specific roles or workflows.
  - Operational CLI reference and troubleshooting guide.
- **Platform-Only Skill Distribution**:
  - Package only the standalone `agentic-lens-setup` skill for external distribution.
  - Internal skills (`agentic-lens-dev`, `agentic-lens-po`, `agentic-lens-review`) remain strictly internal to this repository's git project and are excluded from platform packaging.

---

## Non-Functional & Architectural Invariants

1. **FCIS (Functional Core, Imperative Shell)**:
   - `internal/domain`: Pure aggregation and evaluation functions. Zero I/O, zero network, zero clocks.
   - `internal/adapters` and `cmd/`: Imperative shell for filesystem operations, network downloads, and HTTP handlers.
2. **Permanent Platform Branding**:
   - `AgenticLens` is immutable and permanent as the platform brand across Navbar and Landing Page; configured `project_name` displays alongside it as context.
3. **Strict Typing & Zero Any**:
   - Zero `any` or `interface{}` across Go backend and TypeScript frontend.
4. **Zero Emojis**:
   - Zero emojis permitted in code, comments, docstrings, commit messages, PR drafts, UI strings, logs, or workspace artifacts.
5. **PO Non-Execution Invariant**:
   - The Project Owner only authors specifications, decomposes tickets, and updates roadmap registries. All code and test changes are performed via `agentic-lens-dev`.

---

## Decomposition & Delivery Roadmap

- [ ] **AL-35**: Pre-Release Source Code Audit, Configuration Leakage Inspection, and Review Artifact Generation *(Phase 7, Tier 1 - Ready, Unblocked)*
- [ ] **AL-38**: Migrate Configuration Manifest to Root `.agenticlens-config.yaml` *(Phase 7, Tier 1 - Ready, Unblocked, Blocks AL-36)*
- [ ] **AL-36**: Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator *(Phase 7, Tier 2 - Pending, Blocked by AL-38, Blocks AL-37)*
- [ ] **AL-37**: Cross-Project Platform Setup Manual and Telemetry Integration Guide *(Phase 7, Tier 3 - Pending, Blocked by AL-36)*
