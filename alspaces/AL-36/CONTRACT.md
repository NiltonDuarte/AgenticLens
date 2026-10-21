# Contract: AL-36 Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator

## 1. Problem Statement

Adopting AgenticLens in external repositories currently presents adoption friction. Engineering teams must manually locate and retrieve compiled binaries from GitHub release assets, manually author a valid .agenticlens-config.yaml file from memory or documentation, and manually update repository agent guidelines such as AGENTS.md.

To achieve zero-friction adoption, an autonomous AI agent in any target repository must be capable of bootstrapping AgenticLens self-reliantly when instructed by an engineer. This capability requires a dedicated setup skill (agentic-lens-setup) that:
1. Detects host platform architecture and downloads the precompiled AgenticLens release binary directly into a self-contained, git-ignored skill folder without requiring local compilation or toolchains.
2. Guides an interactive user interview to determine repository parameters and writes a well-documented .agenticlens-config.yaml manifest at the project root.
3. Explains the core capabilities of AgenticLens across CLI commands, real-time live session monitoring, and embedded web dashboard visualization.
4. Inspects existing agent instructions (such as AGENTS.md) and proposes standardized operational policies, strictly requiring explicit user confirmation before modifying any project files.

Ticket AL-36 implements and packages this complete setup workflow across skill definitions, shell automation, Go configuration adapters, CLI entrypoints, and automated verification suites.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All architecture and implementation must strictly obey TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go Functional Core (backend/internal/domain/):
   - Pure domain models: domain.ProjectConfig, domain.TicketConfig, domain.WorkspacesConfig, domain.DatabaseConfig, domain.LiveConfig, domain.CodeMetricsConfig, domain.RoleMappings.
   - Domain defaults: domain.DefaultProjectConfig and ticket normalization functions domain.NormalizeTicketID, domain.NormalizeTicketType.
   - Pure string formatting and data validation without side effects.
   - Zero imports of net/http, os, database/sql, or filesystem packages.
   - Zero clocks (time.Now()) or non-deterministic generators in the domain layer.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Configuration loader (backend/internal/adapters/config/loader.go) handles upward directory traversal and unmarshaling.
   - Configuration writer and generator adapter (backend/internal/adapters/config/writer.go) provides deterministic YAML formatting, default template rendering, and atomic file creation with collision detection.
   - Filesystem repository scanner (backend/internal/adapters/repository.go) discovers workspaces relative to root configuration.
3. Go CLI Entrypoint (backend/cmd/agentic-lens/):
   - CLI command parsing in backend/cmd/agentic-lens/main.go.
   - Dedicated init subcommand providing interactive or flag-based .agenticlens-config.yaml scaffolding at the project root.
   - Existing subcommands (generate, scan, server, version) wire cleanly to domain and adapter layers.
4. Setup Skill and Downloader Shell Scripts (.opencode/skills/agentic-lens-setup/ and scripts/):
   - Skill definition (.opencode/skills/agentic-lens-setup/SKILL.md) specifies autonomous agent instructions, prompt flow, and mandatory safety gates.
   - Downloader script (scripts/download-agentic-lens.sh and .opencode/skills/agentic-lens-setup/scripts/download.sh) detects host OS/architecture, retrieves precompiled release archives from GitHub Releases, unpacks the binary, enforces executable permissions, and guarantees a protective .gitignore.
5. Agent Instruction Layer (AGENTS.md, .opencode/agent/*.md):
   - Standardized policy proposals.
   - Mandatory User Confirmation Gate: zero modifications to target repository files without presenting the proposed text and receiving affirmative confirmation from the user.

### Strong Typing and Zero Any Invariant
- Absolute prohibition of any or interface{} across all Go backend packages and TypeScript modules.
- Concrete domain types, typed string constants, and explicit struct definitions for all configuration parameters.

### Zero Emojis Mandate
- Absolute prohibition of emojis across code, comments, docstrings, shell scripts, CLI output, skill manifests, agent guidelines, commit messages, and workspace artifacts.

---

## 3. Scope Boundaries: Skills, CLI, Adapters, Scripts, and Tests

### In Scope for AL-36
1. Skill Manifest Authoring (.opencode/skills/agentic-lens-setup/SKILL.md):
   - Author complete specification defining binary download, interactive interview, config generation, platform capability overview, and agent file review.
   - Sized for single-session autonomous execution by OpenCode or harness agents.
2. Automated Artifact Downloader:
   - Shell script supporting host OS detection (linux, darwin, windows) and architecture detection (amd64, arm64).
   - Retrieves precompiled release archives or binaries from GitHub Releases (https://github.com/agenticlens/AgenticLens/releases/latest/download/... or gh release download).
   - Strict release-only mandate: never requests the user to compile the binary from source; requires zero Go or Node.js toolchain installations in target project.
   - Binary co-location in skill directory (.opencode/skills/agentic-lens-setup/bin/agentic-lens).
   - Protective .gitignore co-located in binary directory containing wildcard pattern to prevent accidental git tracking.
   - Executable permissions enforcement (chmod +x).
3. Root Configuration Generator:
   - Interactive questionnaire structured into the skill.
   - Go adapter (backend/internal/adapters/config/writer.go) for structured, commented YAML serialization.
   - CLI init subcommand in backend/cmd/agentic-lens/main.go supporting flags (--project-name, --workspaces, --db-path, --src-folders, --force, --dry-run).
   - Outputs valid, commented .agenticlens-config.yaml at project root.
4. Platform Capability Explainer:
   - Comprehensive documentation in the skill covering CLI subcommands (generate, scan, server, version), live session tracking, anomaly detection, and embedded web dashboard navigation.
5. Agent Guidelines Inspection with Mandatory User Approval:
   - Inspection logic scanning for AGENTS.md, .opencode/agent/*.md, or CLAUDE.md.
   - Mandatory confirmation gate: agent must present proposed additions and request explicit user confirmation before writing or modifying any agent files.
   - Standardized operational policies covering ticket tracking mandate (<PREFIX>-<ID>), dedicated workspace allocation, and delayed telemetry ingestion.
6. Verification Suites and Tests:
   - Test script (scripts/test-setup.sh) validating OS/arch detection, asset extraction, protective gitignore presence, root config generation, and binary execution (./.opencode/skills/agentic-lens-setup/bin/agentic-lens version --json).
   - Go unit tests in backend/internal/adapters/config/writer_test.go and backend/cmd/agentic-lens/main_test.go.

### Out of Scope for AL-36
1. Modifying OpenCode SQLite database schema, trace extractor queries, or event pruning logic.
2. Modifying web dashboard React frontend components or TypeScript views.
3. Altering git commit diff classification rules or git metric extraction logic (governed by AL-39).
4. Writing external user documentation manuals or standalone integration markdown outside the skill (governed by downstream ticket AL-37).

---

## 4. Architectural Contracts and Data Specifications

### Contract 1: Artifact Downloader and Binary Co-Location
- Host OS mapping: Darwin maps to darwin, Linux maps to linux, Windows (or Msys/Mingw) maps to windows.
- Host Architecture mapping: x86_64 and amd64 map to amd64, arm64 and aarch64 map to arm64.
- Target archive naming convention: agentic-lens-<VERSION>-<OS>-<ARCH>.tar.gz for Linux and Darwin; agentic-lens-<VERSION>-<OS>-<ARCH>.zip for Windows. Raw binary fallback: agentic-lens-<OS>-<ARCH> (with .exe extension on Windows).
- Destination directory: .opencode/skills/agentic-lens-setup/bin/.
- Target binary path: .opencode/skills/agentic-lens-setup/bin/agentic-lens (or agentic-lens.exe on Windows).
- Protective .gitignore path: .opencode/skills/agentic-lens-setup/bin/.gitignore. Content must include wildcard pattern ignoring all binary artifacts while retaining the gitignore itself.

### Contract 2: Root Configuration Generator
- Target file path: .agenticlens-config.yaml at repository root.
- Required sections and defaults:
  - project_name: string identifier defaulting to root directory name.
  - ticket.prefixes: structured list containing prefix, description, and optional conventional ticket type.
  - ticket.case_insensitive: boolean defaulting to true.
  - workspaces.dir: relative workspace path defaulting to workspaces.
  - workspaces.src_folders: list of source code directories to observe (e.g. src, backend, web).
  - database.path: SQLite database path defaulting to ~/.local/share/opencode/opencode.db.
  - database.query_timeout_ms: integer defaulting to 5000.
  - database.busy_timeout_ms: integer defaulting to 5000.
  - database.read_only: boolean defaulting to true.
  - live.active_threshold_seconds: integer defaulting to 300.
  - live.poll_interval_ms: integer defaulting to 2000.
  - live.max_trace_buffer_messages: integer defaulting to 500.
  - live.anomaly_thresholds.stall_threshold_seconds: integer defaulting to 120.
  - live.anomaly_thresholds.error_rate_threshold: float defaulting to 0.25.
  - live.anomaly_thresholds.context_exhaustion_pct: float defaulting to 0.85.
  - role_mappings: map of persona names to canonical AgentRole values.
  - role_order: ordered list of roles for tabular presentation.

### Contract 3: Agent Workflow Integration and User Gate
- Inspection target files: AGENTS.md, .opencode/agent/*.md, CLAUDE.md.
- Safety rule: If no agent files exist, suggest creating AGENTS.md. If agent files exist, show diff or preview of the proposed policy section.
- User Confirmation Gate: The agent MUST execute a user prompt or question tool call asking for confirmation. If the user declines or requests adjustments, the agent must honor the decision without writing.
- Proposed policy contents:
  - Section 1: Ticket Tracking Mandate with prefix requirement and single-ticket exclusivity.
  - Section 2: Dedicated Workspace Allocation under configured workspaces directory.
  - Section 3: Delayed Telemetry Ingestion Mandate executing the co-located binary as the final turn before git commit.
