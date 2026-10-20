# Feature Specification: Pre-Release Source Code Audit, Configuration Leakage Elimination, Cross-Project Setup Skill, and Integration Manual

## Problem Statement

As AgenticLens approaches its v1.0.0 first public release, two critical capability and architectural gaps must be resolved:

1. **Configuration Leakage and Codebase Hygiene Deficits**:
   - AgenticLens was developed inside this specific repository, causing repo-specific assumptions to leak into source code and domain adapters.
   - In `backend/internal/adapters/git/analyzer.go` and `classifier.go`, source folder prefixes (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`), workspace paths (`workspaces/`), and the default branch (`origin/main`) are hardcoded instead of dynamically evaluating `ProjectConfig.Workspaces.SrcFolders`, `ProjectConfig.Workspaces.Dir`, and git configuration.
   - In `backend/internal/domain/config.go`, `DefaultProjectConfig()` hardcodes repository-specific values (`ProjectName: "AgenticLens"`, prefixes `AL`, `GEN`, `PO`, folders `web` and `backend`).
   - In `backend/internal/adapters/http/`, there is no `/api/config` endpoint. The frontend dashboard (`web/src/components/Navbar.tsx`, `LandingPage.tsx`) statically hardcodes the brand string "AgenticLens" without dynamic project name binding or configured prefix awareness.
   - The test suite and codebase require a deep, structured audit using the `agentic-lens-review` skill to profile test execution times, eliminate dead code, enforce strict FCIS separation, and generate actionable follow-up tickets.

2. **Absence of Cross-Project Onboarding Skill and Setup Manual**:
   - External engineering teams and AI agents cannot easily adopt AgenticLens in their own repositories without manual reverse engineering of this project's setup.
   - There is no interactive setup skill (`agentic-lens-setup`) allowing an AI agent to autonomously download the compiled binary artifact from GitHub Releases, interview the user about their repository topology, generate a valid `.agenticlens/config.yaml`, and explain platform capabilities.
   - There is no automated inspection or instruction injection for repository agent files (`AGENTS.md`, `.opencode/agent/*.md`) to enforce ticket tracking and delayed telemetry generation.
   - There is no comprehensive cross-project setup manual (`docs/SETUP_MANUAL.md`) or skill packaging script detailing how human developers and AI agents collaborate using AgenticLens across different languages, frameworks, and agent harnesses.

---

## Target Personas & Use Cases

- **AI Platform Engineers & Tech Leads**: Need to onboard AgenticLens into new and existing codebases within minutes, configuring ticket conventions, tracked source directories, and agent personas without hardcoded assumptions.
- **Autonomous AI Agents (OpenCode / Harness Agents)**: Need a standardized setup skill (`agentic-lens-setup`) that enables them to self-install the platform binary, interview their human operator, generate project configuration, and update their own instructions (`AGENTS.md`) to produce telemetry ledgers.
- **Software Engineers & Reviewers**: Need clean, leak-free source code where repository-specific paths do not compromise portability, alongside a profiled test suite executing with microsecond domain speed and low adapter latency.

---

## User Stories

1. As an AI platform engineer, I want a complete code audit executed with `agentic-lens-review`, so that all architectural drift, test bottlenecks, and dead code are documented with concrete follow-up tasks.
2. As an adopter running AgenticLens on a non-Go/TS codebase, I want all source paths, workspace directories, and git base branches to respect `.agenticlens/config.yaml`, so that git diff and code churn metrics accurately reflect my project's source tree.
3. As a dashboard viewer, I want the web interface to query `/api/config` and display my configured project name and ticket prefixes, so that the dashboard reflects my actual project identity.
4. As an AI agent invoked in a new repository, I want an `agentic-lens-setup` skill that downloads the platform binary for my operating system and architecture, so that I can install the executable without human manual compilation.
5. As an agent operator, I want an interactive interview during setup, so that I can configure my project name, ticket prefixes, source folders, database path, and agent role mappings effortlessly.
6. As an agent operator, I want the setup skill to explain CLI and dashboard capabilities, so that I understand how to run the server, inspect live sessions, and analyze ticket ledgers.
7. As an agent operator, I want the setup skill to check and update my repository's `AGENTS.md` file, so that all active agents automatically receive instructions to enforce ticket tracking and extract telemetry upon task completion.
8. As an engineering lead, I want a comprehensive `docs/SETUP_MANUAL.md` and portable skill packaging script, so that my organization can distribute the AgenticLens skill suite across any team repository.

---

## Functional Requirements

### 1. Pre-Release Code Review & Configuration Leakage Elimination (AL-35)
- **Multi-Tier Audit with `agentic-lens-review`**:
  - Run test suite profiling to identify all test bottlenecks exceeding 100ms:
    ```bash
    go test -json ./... | jq -r 'select(.Action=="pass" or .Action=="fail") | select(.Test != null) | "\(.Elapsed)s \(.Package) \(.Test)"' | sort -nr | head -20
    ```
  - Audit FCIS boundaries in `internal/domain` (zero I/O, zero clocks, microsecond execution SLA).
  - Audit module depth and eliminate dead wrappers, unused structs, and speculative code.
  - Produce formal audit report artifact (`workspaces/AL-35/AUDIT_REPORT.md`) containing test diagnostics, architectural findings by severity, subtraction targets, and concrete follow-up tickets.
- **Configuration Leakage Purge (Backend & Adapters)**:
  - `backend/internal/adapters/git/analyzer.go`:
    - Replace hardcoded folder prefixes (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`) with dynamic evaluation of `cfg.Workspaces.SrcFolders`.
    - Replace hardcoded `origin/main` with dynamic base branch detection (evaluate `git symbolic-ref refs/remotes/origin/HEAD`, check `origin/main`, `origin/master`, `main`, `master`, or fallback gracefully).
  - `backend/internal/adapters/git/classifier.go`:
    - Replace hardcoded `"workspaces/"` string in universal ignore rules with dynamic `cfg.Workspaces.Dir`.
  - `backend/internal/adapters/http/handlers.go` & `server.go`:
    - Add `GET /api/config` endpoint returning sanitized `ProjectConfig` (project name, ticket prefixes, workspaces dir, role mappings, role order, live thresholds).
- **Dynamic Frontend Branding (TypeScript Web)**:
  - Update `web/src/api/client.ts` with `fetchConfig()` calling `GET /api/config`.
  - Update `web/src/components/Navbar.tsx` and `web/src/views/LandingPage.tsx` to display dynamic `config.project_name` (falling back to "AgenticLens" if unconfigured or offline).

### 2. Cross-Project Interactive Setup Skill & Artifact Downloader (AL-36)
- **Skill Manifest & Workflow (`.opencode/skills/agentic-lens-setup/SKILL.md`)**:
  - Registered as `agentic-lens-setup` with description and step-by-step agent instructions.
  - Sized for single-turn execution or interactive user dialogue.
- **Automated Artifact Downloader**:
  - Platform detection: `uname -s` (Linux, Darwin) and `uname -m` (x86_64/amd64, arm64/aarch64).
  - Release artifact retrieval:
    - Attempt download of binary archive from GitHub Releases (`https://github.com/<owner>/AgenticLens/releases/latest/download/agentic-lens-<VERSION>-<OS>-<ARCH>.tar.gz` or via `gh release download`).
    - Unpack binary to `./bin/agentic-lens` and set executable permissions (`chmod +x`).
    - Fallback: If network release is unavailable, check for local binary or compile via `make build` if Go toolchain exists.
  - Verification: Run `./bin/agentic-lens version --json` and assert valid JSON output with version and commit metadata.
- **Interactive User Interview**:
  - Ask user structured questions (via `question` tool or prompt interaction):
    1. **Project Name**: Human-readable name of the target repository.
    2. **Ticket Identification**: Desired ticket prefix(es) (e.g. `CORE`, `PROJ`, `GEN`), casing rules, and conventional commit mappings.
    3. **Workspaces Directory**: Path for session artifacts and telemetry (default: `workspaces`).
    4. **Source Folders**: Top-level source code folders to monitor for code churn (e.g. `src`, `lib`, `app`).
    5. **Database Path**: Path to OpenCode SQLite database (default: `~/.local/share/opencode/opencode.db`).
    6. **Agent Roles**: Persona names and role mappings used in the project.
  - Generate clean, commented `.agenticlens/config.yaml` reflecting the user's answers.
- **Platform Capability Explanation**:
  - Clear, structured overview explaining:
    - `agentic-lens generate`: Ingests session facts, calculates aggregations, writes `agentic_metrics.json`, and updates `SUMMARY.md`.
    - `agentic-lens scan`: Scans workspace ledgers and validates schema compliance.
    - `agentic-lens server`: Launches embedded HTTP server and web dashboard on port 8085.
    - Live Observability: Non-locking SQLite monitoring, session trees, and anomaly detection.
- **Agent Guidelines File Audit & Telemetry Injection**:
  - Search repository for agent instructions (`AGENTS.md`, `.opencode/agent/*.md`, `CLAUDE.md`).
  - If missing, generate standard `AGENTS.md`.
  - If present, inspect for telemetry generation instructions:
    - Single-ticket exclusivity rule.
    - Workspace deliverables checklist (`SUMMARY.md`, `PULL_REQUEST.md`, `agentic_metrics.json`).
    - Delayed telemetry ingestion mandate with exact CLI command:
      ```bash
      NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
      ./bin/agentic-lens generate --nonce $NONCE --ticket-id <TICKET_ID> --ticket-type <TYPE> --json-output <WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json --append <WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md
      ```

### 3. Cross-Project Setup Manual & Skill Suite Packaging (AL-37)
- **Comprehensive Setup Manual (`docs/SETUP_MANUAL.md`)**:
  - Quickstart guide: 3-step onboarding (download binary, configure YAML, update agent prompt).
  - Detailed `.agenticlens/config.yaml` schema reference covering every field and default.
  - CLI reference: Complete flag documentation for `generate`, `scan`, `server`, `version`, `migrate`, and `backfill`.
  - Architecture and operational patterns: FCIS, delayed telemetry ingestion, database pragmas, and dual-plane waterfall.
  - Cross-project troubleshooting guide: handling missing session trees, database locks, and unmapped git branches.
- **Multi-Project Skill Suite Packager**:
  - Installer script (`scripts/install-skills.sh`) to copy or package the four core skills into any target repository:
    - `agentic-lens-setup`: Onboarding and configuration.
    - `agentic-lens-po`: Requirements elicitation and Jira ticket decomposition.
    - `agentic-lens-dev`: Multi-agent orchestration and lifecycle execution.
    - `agentic-lens-review`: Architectural review and performance profiling.
  - Ensure skill manifests use repository-relative paths and parameter variables rather than hardcoded AgenticLens paths.

---

## Non-Functional & Architectural Invariants

1. **FCIS (Functional Core, Imperative Shell)**:
   - `internal/domain`: Pure aggregation and evaluation functions. Zero I/O, zero network, zero clocks.
   - `internal/adapters` and `cmd/`: Imperative shell for filesystem operations, network downloads, and HTTP handlers.
2. **Strict Typing & Zero Any**:
   - Zero `any` or `interface{}` across Go backend and TypeScript frontend.
   - Strictly typed configuration DTOs and API response models.
3. **Zero Emojis**:
   - Zero emojis permitted in code, comments, docstrings, commit messages, PR drafts, UI strings, logs, or workspace artifacts.
4. **Binary Self-Sufficiency**:
   - Distributed AgenticLens executable remains 100% self-contained with embedded frontend assets (`embed.FS`). Zero Node.js runtime dependency in production.
5. **PO Non-Execution Invariant**:
   - The Project Owner only authors specifications, decomposes tickets, and updates roadmap registries. All code and test changes are performed via `agentic-lens-dev`.

---

## Decomposition & Delivery Roadmap

- [ ] **AL-35**: Pre-Release Source Code Audit, Configuration Leakage Elimination, and Follow-Up Artifact Generation *(Phase 7, Tier 1 - Ready, Unblocked)*
- [ ] **AL-36**: Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator *(Phase 7, Tier 2 - Pending, Blocked by AL-35)*
- [ ] **AL-37**: Cross-Project Setup Manual, Skill Suite Usage Guide, and Distribution Packaging *(Phase 7, Tier 3 - Pending, Blocked by AL-36)*
